/*

 Copyright 2019 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/
package iam

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/go-ldap/ldap"
	"github.com/go-redis/redis"
	"io/ioutil"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/models/kubectl"
	"kubesphere.io/kubesphere/pkg/server/params"
	clientset "kubesphere.io/kubesphere/pkg/simple/client"
	ldappool "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/utils/jwtutil"
)

type IdentityManagementInterface interface {
	GetUserInfo(username string) (*User, error)
}

type imOperator struct {
	config    Config
	ldap      ldappool.Client
	redis     redis.Client
	initUsers []initUser
}

type initUser struct {
	User
	Hidden bool `json:"hidden"`
}

const (
	authRateLimitRegex         = `(\d+)/(\d+[s|m|h])`
	defaultMaxAuthFailed       = 5
	defaultAuthTimeInterval    = 30 * time.Minute
	mailAttribute              = "mail"
	descriptionAttribute       = "description"
	preferredLanguageAttribute = "preferredLanguage"
	createTimestampAttribute   = "createTimestampAttribute"
	dateTimeLayout             = "20060102150405Z"
)

func IdentityManagementInit(ldap ldappool.Client, config Config) (IdentityManagementInterface, error) {

	//maxAuthFailed, authTimeInterval := parseAuthRateLimit(authRateLimit)

	imOperator := &imOperator{ldap: ldap, config: config}

	err := imOperator.checkAndCreateDefaultUser()

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	err = imOperator.checkAndCreateDefaultGroup()

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	return imOperator, nil
}

func parseAuthRateLimit(authRateLimit string) (int, time.Duration) {
	regex := regexp.MustCompile(authRateLimitRegex)
	groups := regex.FindStringSubmatch(authRateLimit)

	maxCount := defaultMaxAuthFailed
	timeInterval := defaultAuthTimeInterval

	if len(groups) == 3 {
		maxCount, _ = strconv.Atoi(groups[1])
		timeInterval, _ = time.ParseDuration(groups[2])
	} else {
		klog.Warning("invalid auth rate limit", authRateLimit)
	}

	return maxCount, timeInterval
}

func (im *imOperator) checkAndCreateDefaultGroup() error {

	conn, err := im.ldap.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	groupSearchRequest := ldap.NewSearchRequest(
		im.ldap.GroupSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=posixGroup))",
		nil,
		nil,
	)

	_, err = conn.Search(groupSearchRequest)

	if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
		err = im.createGroupsBaseDN()
		if err != nil {
			return fmt.Errorf("GroupBaseDN %s create failed: %s\n", im.ldap.GroupSearchBase(), err)
		}
	}

	if err != nil {
		return fmt.Errorf("iam database init failed: %s\n", err)
	}

	return nil
}

func (im *imOperator) checkAndCreateDefaultUser() error {
	conn, err := im.ldap.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(
		im.ldap.UserSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=inetOrgPerson))",
		[]string{"uid"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
		err = im.createUserBaseDN()
		if err != nil {
			return fmt.Errorf("UserBaseDN %s create failed: %s\n", im.ldap.UserSearchBase(), err)
		}
	}

	if err != nil {
		return fmt.Errorf("iam database init failed: %s\n", err)
	}

	data, err := ioutil.ReadFile(im.config.userInitFile)

	if err == nil {
		json.Unmarshal(data, &im.initUsers)
	}

	im.initUsers = append(im.initUsers, initUser{User: User{Username: constants.AdminUserName, Email: im.config.adminEmail, Password: im.config.adminPassword, Description: "Administrator account that was always created by default.", ClusterRole: constants.ClusterAdmin}})

	for _, user := range im.initUsers {
		if result == nil || !containsUser(result.Entries, user) {
			_, err = im.CreateUser(&user.User)
			if err != nil && !ldap.IsErrorWithCode(err, ldap.LDAPResultEntryAlreadyExists) {
				klog.Errorln(err)
				return err
			}
		}
	}

	return nil
}

func containsUser(entries []*ldap.Entry, user initUser) bool {
	for _, entry := range entries {
		uid := entry.GetAttributeValue("uid")
		if uid == user.Username {
			return true
		}
	}
	return false
}

func (im *imOperator) createUserBaseDN() error {

	conn, err := im.ldap.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	groupsCreateRequest := ldap.NewAddRequest(im.ldap.UserSearchBase(), nil)
	groupsCreateRequest.Attribute("objectClass", []string{"organizationalUnit", "top"})
	groupsCreateRequest.Attribute("ou", []string{"Users"})
	return conn.Add(groupsCreateRequest)
}

func (im *imOperator) createGroupsBaseDN() error {

	conn, err := im.ldap.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	groupsCreateRequest := ldap.NewAddRequest(im.ldap.GroupSearchBase(), nil)
	groupsCreateRequest.Attribute("objectClass", []string{"organizationalUnit", "top"})
	groupsCreateRequest.Attribute("ou", []string{"Groups"})
	return conn.Add(groupsCreateRequest)
}

func (im *imOperator) RefreshToken(refreshToken string) (*models.AuthGrantResponse, error) {
	validRefreshToken, err := jwtutil.ValidateToken(refreshToken)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	payload, ok := validRefreshToken.Claims.(jwt.MapClaims)

	if !ok {
		err = errors.New("invalid payload")
		klog.Error(err)
		return nil, err
	}

	claims := jwt.MapClaims{}

	// token with expiration time will not auto sliding
	claims["username"] = payload["username"]
	claims["email"] = payload["email"]
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(im.config.tokenIdleTimeout * 4).Unix()

	token := jwtutil.MustSigned(claims)

	claims = jwt.MapClaims{}
	claims["username"] = payload["username"]
	claims["email"] = payload["email"]
	claims["iat"] = time.Now().Unix()
	claims["type"] = "refresh_token"
	claims["exp"] = time.Now().Add(im.config.tokenIdleTimeout * 5).Unix()

	refreshToken = jwtutil.MustSigned(claims)

	return &models.AuthGrantResponse{TokenType: "jwt", Token: token, RefreshToken: refreshToken, ExpiresIn: (im.config.tokenIdleTimeout * 4).Seconds()}, nil
}

func (im *imOperator) PasswordCredentialGrant(username, password, ip string) (*models.AuthGrantResponse, error) {

	records, err := im.redis.Keys(fmt.Sprintf("kubesphere:authfailed:%s:*", username)).Result()

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(records) >= im.config.maxAuthFailed {
		return nil, restful.NewError(http.StatusTooManyRequests, "auth rate limit exceeded")
	}

	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(
		im.ldap.UserSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", username, username),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		return nil, err
	}

	if len(result.Entries) != 1 {
		return nil, ldap.NewError(ldap.LDAPResultInvalidCredentials, errors.New("incorrect password"))
	}

	uid := result.Entries[0].GetAttributeValue("uid")
	email := result.Entries[0].GetAttributeValue("mail")
	dn := result.Entries[0].DN

	// bind as the user to verify their password
	err = conn.Bind(dn, password)

	if err != nil {
		klog.Infoln("auth failed", username, err)

		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			loginFailedRecord := fmt.Sprintf("kubesphere:authfailed:%s:%d", uid, time.Now().UnixNano())
			im.redis.Set(loginFailedRecord, "", im.config.authTimeInterval)
		}

		return nil, err
	}

	claims := jwt.MapClaims{}

	// token with expiration time will not auto sliding
	claims["username"] = uid
	claims["email"] = email
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(im.config.tokenIdleTimeout * 4).Unix()

	token := jwtutil.MustSigned(claims)

	if !im.config.enableMultiLogin {
		// multi login not allowed, remove the previous token
		sessions, err := im.redis.Keys(fmt.Sprintf("kubesphere:users:%s:token:*", uid)).Result()

		if err != nil {
			klog.Errorln(err)
			return nil, err
		}

		if len(sessions) > 0 {
			klog.V(4).Infoln("revoke token", sessions)
			err = im.redis.Del(sessions...).Err()
			if err != nil {
				klog.Errorln(err)
				return nil, err
			}
		}
	}

	claims = jwt.MapClaims{}
	claims["username"] = uid
	claims["email"] = email
	claims["iat"] = time.Now().Unix()
	claims["type"] = "refresh_token"
	claims["exp"] = time.Now().Add(im.config.tokenIdleTimeout * 5).Unix()

	refreshToken := jwtutil.MustSigned(claims)

	im.loginLog(uid, ip)

	return &models.AuthGrantResponse{TokenType: "jwt", Token: token, RefreshToken: refreshToken, ExpiresIn: (im.config.tokenIdleTimeout * 4).Seconds()}, nil
}

// User login
func (im *imOperator) Login(username, password, ip string) (*models.AuthGrantResponse, error) {

	records, err := im.redis.Keys(fmt.Sprintf("kubesphere:authfailed:%s:*", username)).Result()

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(records) >= im.config.maxAuthFailed {
		return nil, restful.NewError(http.StatusTooManyRequests, "auth rate limit exceeded")
	}

	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(
		im.ldap.UserSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", username, username),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		return nil, err
	}

	if len(result.Entries) != 1 {
		return nil, ldap.NewError(ldap.LDAPResultInvalidCredentials, errors.New("incorrect password"))
	}

	uid := result.Entries[0].GetAttributeValue("uid")
	email := result.Entries[0].GetAttributeValue("mail")
	dn := result.Entries[0].DN

	// bind as the user to verify their password
	err = conn.Bind(dn, password)

	if err != nil {
		klog.Infoln("auth failed", username, err)

		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			loginFailedRecord := fmt.Sprintf("kubesphere:authfailed:%s:%d", uid, time.Now().UnixNano())
			im.redis.Set(loginFailedRecord, "", im.config.authTimeInterval)
		}

		return nil, err
	}

	claims := jwt.MapClaims{}

	// token without expiration time will auto sliding
	claims["username"] = uid
	claims["email"] = email
	claims["iat"] = time.Now().Unix()

	token := jwtutil.MustSigned(claims)

	if !im.config.enableMultiLogin {
		// multi login not allowed, remove the previous token
		sessions, err := im.redis.Keys(fmt.Sprintf("kubesphere:users:%s:token:*", uid)).Result()

		if err != nil {
			klog.Errorln(err)
			return nil, err
		}

		if len(sessions) > 0 {
			klog.V(4).Infoln("revoke token", sessions)
			err = im.redis.Del(sessions...).Err()
			if err != nil {
				klog.Errorln(err)
				return nil, err
			}
		}
	}

	// cache token with expiration time
	if err = im.redis.Set(fmt.Sprintf("kubesphere:users:%s:token:%s", uid, token), token, im.config.tokenIdleTimeout).Err(); err != nil {
		klog.Errorln(err)
		return nil, err
	}

	im.loginLog(uid, ip)

	return &models.AuthGrantResponse{Token: token}, nil
}

func (im *imOperator) loginLog(uid, ip string) {
	if ip != "" {

		im.redis.RPush(fmt.Sprintf("kubesphere:users:%s:login-log", uid), fmt.Sprintf("%s,%s", time.Now().UTC().Format("2006-01-02T15:04:05Z"), ip))
		im.redis.LTrim(fmt.Sprintf("kubesphere:users:%s:login-log", uid), -10, -1)
	}
}

func (im *imOperator) LoginLog(username string) ([]string, error) {
	data, err := im.redis.LRange(fmt.Sprintf("kubesphere:users:%s:login-log", username), -10, -1).Result()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (im *imOperator) ListUsers(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	pageControl := ldap.NewControlPaging(1000)

	users := make([]User, 0)

	filter := "(&(objectClass=inetOrgPerson))"

	if keyword := conditions.Match["keyword"]; keyword != "" {
		filter = fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=*%s*)(mail=*%s*)(description=*%s*)))", keyword, keyword, keyword)
	}

	if username := conditions.Match["username"]; username != "" {
		uidFilter := ""
		for _, username := range strings.Split(username, "|") {
			uidFilter += fmt.Sprintf("(uid=%s)", username)
		}
		filter = fmt.Sprintf("(&(objectClass=inetOrgPerson)(|%s))", uidFilter)
	}

	if email := conditions.Match["email"]; email != "" {
		emailFilter := ""
		for _, username := range strings.Split(email, "|") {
			emailFilter += fmt.Sprintf("(mail=%s)", username)
		}
		filter = fmt.Sprintf("(&(objectClass=inetOrgPerson)(|%s))", emailFilter)
	}

	for {
		userSearchRequest := ldap.NewSearchRequest(
			im.ldap.UserSearchBase(),
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			filter,
			[]string{"uid", "mail", "description", "preferredLanguage", "createTimestamp"},
			[]ldap.Control{pageControl},
		)

		response, err := conn.Search(userSearchRequest)

		if err != nil {
			klog.Errorln("search user", err)
			return nil, err
		}

		for _, entry := range response.Entries {

			uid := entry.GetAttributeValue("uid")
			email := entry.GetAttributeValue("mail")
			description := entry.GetAttributeValue("description")
			lang := entry.GetAttributeValue("preferredLanguage")
			createTimestamp, _ := time.Parse("20060102150405Z", entry.GetAttributeValue("createTimestamp"))

			user := User{Username: uid, Email: email, Description: description, Lang: lang, CreateTime: createTimestamp}

			if !im.shouldHidden(user) {
				users = append(users, user)
			}
		}

		updatedControl := ldap.FindControl(response.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pageControl.SetCookie(ctrl.Cookie)
			continue
		}

		break
	}

	sort.Slice(users, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		switch orderBy {
		case "username":
			return strings.Compare(users[i].Username, users[j].Username) <= 0
		case "createTime":
			fallthrough
		default:
			return users[i].CreateTime.Before(users[j].CreateTime)
		}
	})

	items := make([]interface{}, 0)

	for i, user := range users {

		if i >= offset && len(items) < limit {

			user.LastLoginTime = im.GetLastLoginTime(user.Username)
			clusterRole, err := im.GetUserClusterRole(user.Username)
			if err != nil {
				return nil, err
			}
			user.ClusterRole = clusterRole.Name
			items = append(items, user)
		}
	}

	return &models.PageableResponse{Items: items, TotalCount: len(users)}, nil
}

func (im *imOperator) shouldHidden(user User) bool {
	for _, initUser := range im.initUsers {
		if initUser.Username == user.Username {
			return initUser.Hidden
		}
	}
	return false
}

func (im *imOperator) DescribeUser(username string) (*User, error) {
	conn, err := im.ldap.NewConn()

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	filter := fmt.Sprintf("(&(objectClass=inetOrgPerson)(uid=%s))", username)

	usr := ldap.NewSearchRequest(
		im.ldap.UserSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 1, 0, false,
		filter,
		[]string{mailAttribute, descriptionAttribute, preferredLanguageAttribute, createTimestampAttribute},
		nil,
	)

	result, err := conn.Search(usr)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	if len(result.Entries) != 1 {
		return nil, ldap.NewError(ldap.LDAPResultNoSuchObject, fmt.Errorf("user %s does not exist", username))
	}

	email := result.Entries[0].GetAttributeValue(mailAttribute)
	description := result.Entries[0].GetAttributeValue(descriptionAttribute)
	lang := result.Entries[0].GetAttributeValue(preferredLanguageAttribute)
	createTimestamp, _ := time.Parse(dateTimeLayout, result.Entries[0].GetAttributeValue(createTimestampAttribute))
	user := &User{Username: username, Email: email, Description: description, Lang: lang, CreateTime: createTimestamp}

	return user, nil
}

func (im *imOperator) GetLastLoginTime(username string) string {
	cacheKey := fmt.Sprintf("kubesphere:users:%s:login-log", username)
	lastLogin, err := im.redis.LRange(cacheKey, -1, -1).Result()
	if err != nil {
		return ""
	}

	if len(lastLogin) > 0 {
		return strings.Split(lastLogin[0], ",")[0]
	}

	return ""
}

func (im *imOperator) DeleteUser(username string) error {
	conn, err := im.ldap.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	deleteRequest := ldap.NewDelRequest(fmt.Sprintf("uid=%s,%s", username, im.ldap.UserSearchBase()), nil)

	if err = conn.Del(deleteRequest); err != nil {
		klog.Errorln("delete user", err)
		return err
	}

	if err = im.deleteRoleBindings(username); err != nil {
		klog.Errorln("delete user role bindings failed", username, err)
	}

	if err := kubeconfig.DelKubeConfig(username); err != nil {
		klog.Errorln("delete user kubeconfig failed", username, err)
	}

	if err := kubectl.DelKubectlDeploy(username); err != nil {
		klog.Errorln("delete user terminal pod failed", username, err)
	}

	if err := im.deleteUserInDevOps(username); err != nil {
		klog.Errorln("delete user in devops failed", username, err)
	}
	return nil

}

// deleteUserInDevOps is used to clean up user data of devops, such as permission rules
func (im *imOperator) deleteUserInDevOps(username string) error {

	devopsDb, err := clientset.ClientSets().MySQL()
	if err != nil {
		if err == clientset.ErrClientSetNotEnabled {
			klog.Warning("mysql is not enable")
			return nil
		}
		return err
	}

	dp, err := clientset.ClientSets().Devops()
	if err != nil {
		if err == clientset.ErrClientSetNotEnabled {
			klog.Warning("devops client is not enable")
			return nil
		}
		return err
	}

	jenkinsClient := dp.Jenkins()

	_, err = devopsDb.DeleteFrom(devops.DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(devops.DevOpsProjectMembershipUsernameColumn, username),
		)).Exec()
	if err != nil {
		klog.Errorf("%+v", err)
		return err
	}

	err = jenkinsClient.DeleteUserInProject(username)
	if err != nil {
		klog.Errorf("%+v", err)
		return err
	}
	return nil
}

func (im *imOperator) deleteRoleBindings(username string) error {
	roleBindingLister := informers.SharedInformerFactory().Rbac().V1().RoleBindings().Lister()
	roleBindings, err := roleBindingLister.List(labels.Everything())

	if err != nil {
		return err
	}

	for _, roleBinding := range roleBindings {
		roleBinding = roleBinding.DeepCopy()
		length1 := len(roleBinding.Subjects)

		for index, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind && subject.Name == username {
				roleBinding.Subjects = append(roleBinding.Subjects[:index], roleBinding.Subjects[index+1:]...)
				index--
			}
		}

		length2 := len(roleBinding.Subjects)

		if length2 == 0 {
			deletePolicy := metav1.DeletePropagationBackground
			err = clientset.ClientSets().K8s().Kubernetes().RbacV1().RoleBindings(roleBinding.Namespace).Delete(roleBinding.Name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})

			if err != nil {
				klog.Errorf("delete role binding %s %s %s failed: %v", username, roleBinding.Namespace, roleBinding.Name, err)
			}
		} else if length2 < length1 {
			_, err = clientset.ClientSets().K8s().Kubernetes().RbacV1().RoleBindings(roleBinding.Namespace).Update(roleBinding)

			if err != nil {
				klog.Errorf("update role binding %s %s %s failed: %v", username, roleBinding.Namespace, roleBinding.Name, err)
			}
		}
	}

	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	for _, clusterRoleBinding := range clusterRoleBindings {
		clusterRoleBinding = clusterRoleBinding.DeepCopy()
		length1 := len(clusterRoleBinding.Subjects)

		for index, subject := range clusterRoleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind && subject.Name == username {
				clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects[:index], clusterRoleBinding.Subjects[index+1:]...)
				index--
			}
		}

		length2 := len(clusterRoleBinding.Subjects)
		if length2 == 0 {
			// delete if it's not workspace role binding
			if isWorkspaceRoleBinding(clusterRoleBinding) {
				_, err = clientset.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
			} else {
				deletePolicy := metav1.DeletePropagationBackground
				err = clientset.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Delete(clusterRoleBinding.Name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
			}
			if err != nil {
				klog.Errorf("update cluster role binding %s failed:%s", clusterRoleBinding.Name, err)
			}
		} else if length2 < length1 {
			_, err = clientset.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)

			if err != nil {
				klog.Errorf("update cluster role binding %s failed:%s", clusterRoleBinding.Name, err)
			}
		}

	}

	return nil
}

func (im *imOperator) isWorkspaceRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) bool {
	return k8sutil.IsControlledBy(clusterRoleBinding.OwnerReferences, "Workspace", "")
}

func (im *imOperator) UserCreateCheck(check string) (exist bool, err error) {

	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return false, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// search for the given username
	userSearchRequest := ldap.NewSearchRequest(
		im.ldap.UserSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", check, check),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		klog.Errorln("search user", err)
		return false, err
	}

	return len(result.Entries) > 0, nil
}

func (im *imOperator) CreateUser(user *User) (*User, error) {
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.Password = strings.TrimSpace(user.Password)
	user.Description = strings.TrimSpace(user.Description)

	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return nil, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(
		im.ldap.UserSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", user.Username, user.Email),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		klog.Errorln("search user", err)
		return nil, err
	}

	if len(result.Entries) > 0 {
		return nil, ldap.NewError(ldap.LDAPResultEntryAlreadyExists, fmt.Errorf("username or email already exists"))
	}

	maxUid, err := getMaxUid()

	if err != nil {
		klog.Errorln("get max uid", err)
		return nil, err
	}

	maxUid += 1

	userCreateRequest := ldap.NewAddRequest(fmt.Sprintf("uid=%s,%s", user.Username, im.ldap.UserSearchBase()), nil)
	userCreateRequest.Attribute("objectClass", []string{"inetOrgPerson", "posixAccount", "top"})
	userCreateRequest.Attribute("cn", []string{user.Username})                       // RFC4519: common name(s) for which the entity is known by
	userCreateRequest.Attribute("sn", []string{" "})                                 // RFC2256: last (family) name(s) for which the entity is known by
	userCreateRequest.Attribute("gidNumber", []string{"500"})                        // RFC2307: An integer uniquely identifying a group in an administrative domain
	userCreateRequest.Attribute("homeDirectory", []string{"/home/" + user.Username}) // The absolute path to the home directory
	userCreateRequest.Attribute("uid", []string{user.Username})                      // RFC4519: user identifier
	userCreateRequest.Attribute("uidNumber", []string{strconv.Itoa(maxUid)})         // RFC2307: An integer uniquely identifying a user in an administrative domain
	userCreateRequest.Attribute("mail", []string{user.Email})                        // RFC1274: RFC822 Mailbox
	userCreateRequest.Attribute("userPassword", []string{user.Password})             // RFC4519/2307: password of user
	if user.Lang != "" {
		userCreateRequest.Attribute("preferredLanguage", []string{user.Lang})
	}
	if user.Description != "" {
		userCreateRequest.Attribute("description", []string{user.Description}) // RFC4519: descriptive information
	}

	if generateKubeConfig {
		if err = kubeconfig.CreateKubeConfig(user.Username); err != nil {
			klog.Errorln("create user kubeconfig failed", user.Username, err)
			return nil, err
		}
	}

	err = conn.Add(userCreateRequest)

	if err != nil {
		klog.Errorln("create user", err)
		return nil, err
	}

	if user.ClusterRole != "" {
		err := CreateClusterRoleBinding(user.Username, user.ClusterRole)

		if err != nil {
			klog.Errorln("create cluster role binding filed", err)
			return nil, err
		}
	}

	return DescribeUser(user.Username)
}

func (im *imOperator) getMaxUid() (int, error) {
	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return 0, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(im.ldap.UserSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=inetOrgPerson))",
		[]string{"uidNumber"},
		nil)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		return 0, err
	}

	var maxUid int

	if len(result.Entries) == 0 {
		maxUid = 1000
	} else {
		for _, usr := range result.Entries {
			uid, _ := strconv.Atoi(usr.GetAttributeValue("uidNumber"))
			if uid > maxUid {
				maxUid = uid
			}
		}
	}

	return maxUid, nil
}

func (im *imOperator) getMaxGid() (int, error) {

	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return 0, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	groupSearchRequest := ldap.NewSearchRequest(client.GroupSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=posixGroup))",
		[]string{"gidNumber"},
		nil)

	result, err := conn.Search(groupSearchRequest)

	if err != nil {
		return 0, err
	}

	var maxGid int

	if len(result.Entries) == 0 {
		maxGid = 500
	} else {
		for _, group := range result.Entries {
			gid, _ := strconv.Atoi(group.GetAttributeValue("gidNumber"))
			if gid > maxGid {
				maxGid = gid
			}
		}
	}

	return maxGid, nil
}

func (im *imOperator) UpdateUser(user *User) (*User, error) {

	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return nil, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dn := fmt.Sprintf("uid=%s,%s", user.Username, im.ldap.UserSearchBase())
	userModifyRequest := ldap.NewModifyRequest(dn, nil)
	if user.Email != "" {
		userSearchRequest := ldap.NewSearchRequest(
			im.ldap.UserSearchBase(),
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&(objectClass=inetOrgPerson)(mail=%s))", user.Email),
			[]string{"uid", "mail"},
			nil,
		)
		result, err := conn.Search(userSearchRequest)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if len(result.Entries) > 1 {
			err = ldap.NewError(ldap.ErrorDebugging, fmt.Errorf("email is duplicated: %s", user.Email))
			klog.Error(err)
			return nil, err
		}
		if len(result.Entries) == 1 && result.Entries[0].GetAttributeValue("uid") != user.Username {
			err = ldap.NewError(ldap.LDAPResultEntryAlreadyExists, fmt.Errorf("email is duplicated: %s", user.Email))
			klog.Error(err)
			return nil, err
		}
		userModifyRequest.Replace("mail", []string{user.Email})
	}
	if user.Description != "" {
		userModifyRequest.Replace("description", []string{user.Description})
	}

	if user.Lang != "" {
		userModifyRequest.Replace("preferredLanguage", []string{user.Lang})
	}

	if user.Password != "" {
		userModifyRequest.Replace("userPassword", []string{user.Password})
	}

	err = conn.Modify(userModifyRequest)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if user.ClusterRole != "" {
		err = CreateClusterRoleBinding(user.Username, user.ClusterRole)

		if err != nil {
			klog.Errorln(err)
			return nil, err
		}
	}

	// clear auth failed record
	if user.Password != "" {
		redisClient, err := clientset.ClientSets().Redis()
		if err != nil {
			return nil, err
		}

		records, err := im.redis.Keys(fmt.Sprintf("kubesphere:authfailed:%s:*", user.Username)).Result()

		if err == nil {
			im.redis.Del(records...)
		}
	}

	return GetUserInfo(user.Username)
}
func (im *imOperator) DeleteGroup(path string) error {

	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	searchBase, cn := splitPath(path)

	groupDeleteRequest := ldap.NewDelRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase), nil)
	err = conn.Del(groupDeleteRequest)

	if err != nil {
		klog.Errorln("delete user group", err)
		return err
	}

	return nil
}

func (im *imOperator) CreateGroup(group *models.Group) (*models.Group, error) {

	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return nil, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	maxGid, err := getMaxGid()

	if err != nil {
		klog.Errorln("get max gid", err)
		return nil, err
	}

	maxGid += 1

	if group.Path == "" {
		group.Path = group.Name
	}

	searchBase, cn := splitPath(group.Path)

	groupCreateRequest := ldap.NewAddRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase), nil)
	groupCreateRequest.Attribute("objectClass", []string{"posixGroup", "top"})
	groupCreateRequest.Attribute("cn", []string{cn})
	groupCreateRequest.Attribute("gidNumber", []string{strconv.Itoa(maxGid)})

	if group.Description != "" {
		groupCreateRequest.Attribute("description", []string{group.Description})
	}

	if group.Members != nil {
		groupCreateRequest.Attribute("memberUid", group.Members)
	}

	err = conn.Add(groupCreateRequest)

	if err != nil {
		klog.Errorln("create group", err)
		return nil, err
	}

	group.Gid = strconv.Itoa(maxGid)

	return DescribeGroup(group.Path)
}

func (im *imOperator) UpdateGroup(group *models.Group) (*models.Group, error) {

	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return nil, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	old, err := DescribeGroup(group.Path)

	if err != nil {
		return nil, err
	}

	searchBase, cn := splitPath(group.Path)

	groupUpdateRequest := ldap.NewModifyRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase), nil)

	if old.Description == "" {
		if group.Description != "" {
			groupUpdateRequest.Add("description", []string{group.Description})
		}
	} else {
		if group.Description != "" {
			groupUpdateRequest.Replace("description", []string{group.Description})
		} else {
			groupUpdateRequest.Delete("description", []string{})
		}
	}

	if group.Members != nil {
		groupUpdateRequest.Replace("memberUid", group.Members)
	}

	err = conn.Modify(groupUpdateRequest)

	if err != nil {
		klog.Errorln("update group", err)
		return nil, err
	}

	return group, nil
}

func (im *imOperator) ChildList(path string) ([]models.Group, error) {

	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return nil, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var groupSearchRequest *ldap.SearchRequest
	if path == "" {
		groupSearchRequest = ldap.NewSearchRequest(client.GroupSearchBase(),
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
			"(&(objectClass=posixGroup))",
			[]string{"cn", "gidNumber", "memberUid", "description"},
			nil)
	} else {
		searchBase, cn := splitPath(path)
		groupSearchRequest = ldap.NewSearchRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase),
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
			"(&(objectClass=posixGroup))",
			[]string{"cn", "gidNumber", "memberUid", "description"},
			nil)
	}

	result, err := conn.Search(groupSearchRequest)

	if err != nil {
		return nil, err
	}

	groups := make([]models.Group, 0)

	for _, v := range result.Entries {
		dn := v.DN
		cn := v.GetAttributeValue("cn")
		gid := v.GetAttributeValue("gidNumber")
		members := v.GetAttributeValues("memberUid")
		description := v.GetAttributeValue("description")

		group := models.Group{Path: convertDNToPath(dn), Name: cn, Gid: gid, Members: members, Description: description}

		childSearchRequest := ldap.NewSearchRequest(dn,
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
			"(&(objectClass=posixGroup))",
			[]string{""},
			nil)

		result, err = conn.Search(childSearchRequest)

		if err != nil {
			return nil, err
		}

		childGroups := make([]string, 0)

		for _, v := range result.Entries {
			child := convertDNToPath(v.DN)
			childGroups = append(childGroups, child)
		}

		group.ChildGroups = childGroups

		groups = append(groups, group)
	}

	return groups, nil
}

func (im *imOperator) DescribeGroup(path string) (*models.Group, error) {
	client, err := clientset.ClientSets().Ldap()
	if err != nil {
		return nil, err
	}
	conn, err := im.ldap.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	searchBase, cn := splitPath(path)

	groupSearchRequest := ldap.NewSearchRequest(searchBase,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=posixGroup)(cn=%s))", cn),
		[]string{"cn", "gidNumber", "memberUid", "description"},
		nil)

	result, err := conn.Search(groupSearchRequest)

	if err != nil {
		klog.Errorln("search group", err)
		return nil, err
	}

	if len(result.Entries) != 1 {
		return nil, ldap.NewError(ldap.LDAPResultNoSuchObject, fmt.Errorf("group %s does not exist", path))
	}

	dn := result.Entries[0].DN
	cn = result.Entries[0].GetAttributeValue("cn")
	gid := result.Entries[0].GetAttributeValue("gidNumber")
	members := result.Entries[0].GetAttributeValues("memberUid")
	description := result.Entries[0].GetAttributeValue("description")

	group := models.Group{Path: convertDNToPath(dn), Name: cn, Gid: gid, Members: members, Description: description}

	childGroups := make([]string, 0)

	group.ChildGroups = childGroups

	return &group, nil

}

func (im *imOperator) WorkspaceUsersTotalCount(workspace string) (int, error) {
	workspaceRoleBindings, err := GetWorkspaceRoleBindings(workspace)

	if err != nil {
		return 0, err
	}

	users := make([]string, 0)

	for _, roleBinding := range workspaceRoleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
				users = append(users, subject.Name)
			}
		}
	}

	return len(users), nil
}

func (im *imOperator) ListWorkspaceUsers(workspace string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	workspaceRoleBindings, err := GetWorkspaceRoleBindings(workspace)

	if err != nil {
		return nil, err
	}

	users := make([]*User, 0)

	for _, roleBinding := range workspaceRoleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
				user, err := GetUserInfo(subject.Name)
				if err != nil {
					return nil, err
				}
				prefix := fmt.Sprintf("workspace:%s:", workspace)
				user.WorkspaceRole = fmt.Sprintf("workspace-%s", strings.TrimPrefix(roleBinding.Name, prefix))
				if matchConditions(conditions, user) {
					users = append(users, user)
				}
			}
		}
	}

	// order & reverse
	sort.Slice(users, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		switch orderBy {
		default:
			fallthrough
		case "name":
			return strings.Compare(users[i].Username, users[j].Username) <= 0
		}
	})

	result := make([]interface{}, 0)

	for i, d := range users {
		if i >= offset && (limit == -1 || len(result) < limit) {
			result = append(result, d)
		}
	}

	return &models.PageableResponse{Items: result, TotalCount: len(users)}, nil
}

func matchConditions(conditions *params.Conditions, user *User) bool {
	for k, v := range conditions.Match {
		switch k {
		case "keyword":
			if !strings.Contains(user.Username, v) &&
				!strings.Contains(user.Email, v) &&
				!strings.Contains(user.Description, v) {
				return false
			}
		case "name":
			names := strings.Split(v, "|")
			if !sliceutil.HasString(names, user.Username) {
				return false
			}
		case "email":
			email := strings.Split(v, "|")
			if !sliceutil.HasString(email, user.Email) {
				return false
			}
		case "role":
			if user.WorkspaceRole != v {
				return false
			}
		}
	}
	return true
}

type User struct {
	Username        string            `json:"username"`
	Email           string            `json:"email"`
	Lang            string            `json:"lang,omitempty"`
	Description     string            `json:"description"`
	CreateTime      time.Time         `json:"create_time"`
	Groups          []string          `json:"groups,omitempty"`
	Password        string            `json:"password,omitempty"`
	CurrentPassword string            `json:"current_password,omitempty"`
	AvatarUrl       string            `json:"avatar_url"`
	LastLoginTime   string            `json:"last_login_time"`
	Status          int               `json:"status"`
	ClusterRole     string            `json:"cluster_role"`
	Roles           map[string]string `json:"roles,omitempty"`
	Role            string            `json:"role,omitempty"`
	RoleBinding     string            `json:"role_binding,omitempty"`
	RoleBindTime    *time.Time        `json:"role_bind_time,omitempty"`
	WorkspaceRole   string            `json:"workspace_role,omitempty"`
}
