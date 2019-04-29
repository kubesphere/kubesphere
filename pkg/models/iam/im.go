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
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/models/kubectl"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/redis"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-ldap/ldap"
	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"

	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/utils/jwtutil"
)

var (
	adminEmail      string
	adminPassword   string
	tokenExpireTime time.Duration
	initUsers       []initUser
)

type initUser struct {
	models.User
	Hidden bool `json:"hidden"`
}

const (
	userInitFile = "/etc/ks-iam/users.json"
)

func Init(email, password string, t time.Duration) error {
	adminEmail = email
	adminPassword = password
	tokenExpireTime = t

	conn, err := ldapclient.Client()

	if err != nil {
		return err
	}

	defer conn.Close()

	err = checkAndCreateDefaultUser(conn)

	if err != nil {
		glog.Errorln("create default users", err)
		return err
	}

	err = checkAndCreateDefaultGroup(conn)

	if err != nil {
		glog.Errorln("create default groups", err)
		return err
	}

	return nil
}

func checkAndCreateDefaultGroup(conn ldap.Client) error {

	groupSearchRequest := ldap.NewSearchRequest(
		ldapclient.GroupSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=posixGroup))",
		nil,
		nil,
	)

	_, err := conn.Search(groupSearchRequest)

	if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
		err = createGroupsBaseDN(conn)
		if err != nil {
			return fmt.Errorf("GroupBaseDN %s create failed: %s\n", ldapclient.GroupSearchBase, err)
		}
	}

	if err != nil {
		return fmt.Errorf("iam database init failed: %s\n", err)
	}

	return nil
}

func checkAndCreateDefaultUser(conn ldap.Client) error {

	userSearchRequest := ldap.NewSearchRequest(
		ldapclient.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=inetOrgPerson))",
		[]string{"uid"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
		err = createUserBaseDN(conn)
		if err != nil {
			return fmt.Errorf("UserBaseDN %s create failed: %s\n", ldapclient.UserSearchBase, err)
		}
	}

	if err != nil {
		return fmt.Errorf("iam database init failed: %s\n", err)
	}

	data, err := ioutil.ReadFile(userInitFile)
	if err == nil {
		json.Unmarshal(data, &initUsers)
	}
	initUsers = append(initUsers, initUser{User: models.User{Username: constants.AdminUserName, Email: adminEmail, Password: adminPassword, Description: "Administrator account that was always created by default.", ClusterRole: constants.ClusterAdmin}})

	for _, user := range initUsers {
		if result == nil || !containsUser(result.Entries, user) {
			_, err = CreateUser(&user.User)
			if err != nil && !ldap.IsErrorWithCode(err, ldap.LDAPResultEntryAlreadyExists) {
				glog.Errorln("user init failed", user.Username, err)
				return fmt.Errorf("user %s init failed: %s\n", user.Username, err)
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

func createUserBaseDN(conn ldap.Client) error {

	conn, err := ldapclient.Client()
	if err != nil {
		return err
	}
	defer conn.Close()

	groupsCreateRequest := ldap.NewAddRequest(ldapclient.UserSearchBase, nil)
	groupsCreateRequest.Attribute("objectClass", []string{"organizationalUnit", "top"})
	groupsCreateRequest.Attribute("ou", []string{"Users"})
	return conn.Add(groupsCreateRequest)
}

func createGroupsBaseDN(conn ldap.Client) error {
	groupsCreateRequest := ldap.NewAddRequest(ldapclient.GroupSearchBase, nil)
	groupsCreateRequest.Attribute("objectClass", []string{"organizationalUnit", "top"})
	groupsCreateRequest.Attribute("ou", []string{"Groups"})
	return conn.Add(groupsCreateRequest)
}

// User login
func Login(username string, password string, ip string) (*models.Token, error) {

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(
		ldapclient.UserSearchBase,
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
		glog.Errorln("auth error", username, err)
		return nil, err
	}

	claims := jwt.MapClaims{}

	claims["exp"] = time.Now().Add(tokenExpireTime).Unix()
	claims["username"] = uid
	claims["email"] = email

	token := jwtutil.MustSigned(claims)

	loginLog(uid, ip)

	return &models.Token{Token: token}, nil
}

func loginLog(uid, ip string) {
	if ip != "" {
		redisClient := redis.Client()
		redisClient.RPush(fmt.Sprintf("kubesphere:users:%s:login-log", uid), fmt.Sprintf("%s,%s", time.Now().UTC().Format("2006-01-02T15:04:05Z"), ip))
		redisClient.LTrim(fmt.Sprintf("kubesphere:users:%s:login-log", uid), -10, -1)
	}
}

func LoginLog(username string) ([]string, error) {
	redisClient := redis.Client()

	data, err := redisClient.LRange(fmt.Sprintf("kubesphere:users:%s:login-log", username), -10, -1).Result()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func ListUsers(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	pageControl := ldap.NewControlPaging(80)

	users := make([]models.User, 0)

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
			ldapclient.UserSearchBase,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			filter,
			[]string{"uid", "mail", "description", "preferredLanguage", "createTimestamp"},
			[]ldap.Control{pageControl},
		)

		response, err := conn.Search(userSearchRequest)

		if err != nil {
			glog.Errorln("search user", err)
			return nil, err
		}

		for _, entry := range response.Entries {

			uid := entry.GetAttributeValue("uid")
			email := entry.GetAttributeValue("mail")
			description := entry.GetAttributeValue("description")
			lang := entry.GetAttributeValue("preferredLanguage")
			createTimestamp, _ := time.Parse("20060102150405Z", entry.GetAttributeValue("createTimestamp"))

			user := models.User{Username: uid, Email: email, Description: description, Lang: lang, CreateTime: createTimestamp}

			if !shouldHidden(user) {
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
			fallthrough
		case "createTime":
			return users[i].CreateTime.Before(users[j].CreateTime)
		default:
			return strings.Compare(users[i].Username, users[j].Username) <= 0
		}
	})

	items := make([]interface{}, 0)

	for i, user := range users {

		if i >= offset && len(items) < limit {

			user.AvatarUrl = getAvatar(user.Username)
			user.LastLoginTime = getLastLoginTime(user.Username)
			clusterRole, err := GetUserClusterRole(user.Username)
			if err != nil {
				return nil, err
			}
			user.ClusterRole = clusterRole.Name
			items = append(items, user)
		}
	}

	return &models.PageableResponse{Items: items, TotalCount: len(users)}, nil
}

func shouldHidden(user models.User) bool {
	for _, initUser := range initUsers {
		if initUser.Username == user.Username {
			return initUser.Hidden
		}
	}
	return false
}

func DescribeUser(username string) (*models.User, error) {

	user, err := GetUserInfo(username)

	if err != nil {
		return nil, err
	}

	groups, err := GetUserGroups(username)

	if err == nil {
		user.Groups = groups
	}

	user.AvatarUrl = getAvatar(username)

	return user, nil
}

// Get user info only included email description & lang
func GetUserInfo(username string) (*models.User, error) {

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	userSearchRequest := ldap.NewSearchRequest(
		ldapclient.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(uid=%s))", username),
		[]string{"mail", "description", "preferredLanguage", "createTimestamp"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		glog.Errorln("search user", err)
		return nil, err
	}

	if len(result.Entries) != 1 {
		return nil, ldap.NewError(ldap.LDAPResultNoSuchObject, fmt.Errorf("user %s does not exist", username))
	}

	email := result.Entries[0].GetAttributeValue("mail")
	description := result.Entries[0].GetAttributeValue("description")
	lang := result.Entries[0].GetAttributeValue("preferredLanguage")
	createTimestamp, _ := time.Parse("20060102150405Z", result.Entries[0].GetAttributeValue("createTimestamp"))
	user := &models.User{Username: username, Email: email, Description: description, Lang: lang, CreateTime: createTimestamp}

	user.LastLoginTime = getLastLoginTime(username)

	return user, nil
}

func GetUserGroups(username string) ([]string, error) {
	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	groupSearchRequest := ldap.NewSearchRequest(
		ldapclient.GroupSearchBase,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=posixGroup)(memberUid=%s))", username),
		nil,
		nil,
	)

	result, err := conn.Search(groupSearchRequest)

	if err != nil {
		return nil, err
	}

	groups := make([]string, 0)

	for _, group := range result.Entries {
		groupName := convertDNToPath(group.DN)
		groups = append(groups, groupName)
	}

	return groups, nil
}

func getLastLoginTime(username string) string {
	lastLogin, err := redis.Client().LRange(fmt.Sprintf("kubesphere:users:%s:login-log", username), -1, -1).Result()

	if err != nil {
		return ""
	}

	if len(lastLogin) > 0 {
		return strings.Split(lastLogin[0], ",")[0]
	}

	return ""
}

func setAvatar(username, avatar string) error {
	_, err := redis.Client().HMSet("kubesphere:users:avatar", map[string]interface{}{"username": avatar}).Result()
	return err
}

func getAvatar(username string) string {

	avatar, err := redis.Client().HMGet("kubesphere:users:avatar", username).Result()

	if err != nil {
		return ""
	}

	if len(avatar) > 0 {
		if url, ok := avatar[0].(string); ok {
			return url
		}
	}

	return ""
}

func DeleteUser(username string) error {

	conn, err := ldapclient.Client()

	if err != nil {
		return err
	}

	defer conn.Close()

	deleteRequest := ldap.NewDelRequest(fmt.Sprintf("uid=%s,%s", username, ldapclient.UserSearchBase), nil)

	if err = conn.Del(deleteRequest); err != nil {
		glog.Errorln("delete user", err)
		return err
	}

	if err = deleteRoleBindings(username); err != nil {
		glog.Errorln("delete user role bindings failed", username, err)
	}

	if err := kubeconfig.DelKubeConfig(username); err != nil {
		glog.Errorln("delete user kubeconfig failed", username, err)
	}

	if err := kubectl.DelKubectlDeploy(username); err != nil {
		glog.Errorln("delete user terminal pod failed", username, err)
	}

	return nil
}

func deleteRoleBindings(username string) error {
	roleBindingLister := informers.SharedInformerFactory().Rbac().V1().RoleBindings().Lister()
	roleBindings, err := roleBindingLister.List(labels.Everything())

	if err != nil {
		return err
	}

	for _, roleBinding := range roleBindings {
		roleBinding = roleBinding.DeepCopy()
		length1 := len(roleBinding.Subjects)

		for index, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				roleBinding.Subjects = append(roleBinding.Subjects[:index], roleBinding.Subjects[index+1:]...)
				index--
			}
		}

		length2 := len(roleBinding.Subjects)

		if length2 == 0 {
			deletePolicy := meta_v1.DeletePropagationForeground
			err = k8s.Client().RbacV1().RoleBindings(roleBinding.Namespace).Delete(roleBinding.Name, &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})

			if err != nil {
				glog.Errorf("delete role binding %s %s %s failed: %v", username, roleBinding.Namespace, roleBinding.Name, err)
			}
		} else if length2 < length1 {
			_, err = k8s.Client().RbacV1().RoleBindings(roleBinding.Namespace).Update(roleBinding)

			if err != nil {
				glog.Errorf("update role binding %s %s %s failed: %v", username, roleBinding.Namespace, roleBinding.Name, err)
			}
		}
	}

	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	for _, clusterRoleBinding := range clusterRoleBindings {
		clusterRoleBinding = clusterRoleBinding.DeepCopy()
		length1 := len(clusterRoleBinding.Subjects)

		for index, subject := range clusterRoleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects[:index], clusterRoleBinding.Subjects[index+1:]...)
				index--
			}
		}

		length2 := len(clusterRoleBinding.Subjects)
		if length2 == 0 {
			if groups := regexp.MustCompile(fmt.Sprintf(`^system:(\S+):(%s)$`, strings.Join(constants.WorkSpaceRoles, "|"))).FindStringSubmatch(clusterRoleBinding.RoleRef.Name); len(groups) == 3 {
				_, err = k8s.Client().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
			} else {
				deletePolicy := meta_v1.DeletePropagationForeground
				err = k8s.Client().RbacV1().ClusterRoleBindings().Delete(clusterRoleBinding.Name, &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})
			}
			if err != nil {
				glog.Errorf("update cluster role binding %s failed:%s", clusterRoleBinding.Name, err)
			}
		} else if length2 < length1 {
			_, err = k8s.Client().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)

			if err != nil {
				glog.Errorf("update cluster role binding %s failed:%s", clusterRoleBinding.Name, err)
			}
		}

	}

	return nil
}

func UserCreateCheck(check string) (exist bool, err error) {

	// bind root DN
	conn, err := ldapclient.Client()

	if err != nil {
		return false, err
	}

	defer conn.Close()

	// search for the given username
	userSearchRequest := ldap.NewSearchRequest(
		ldapclient.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", check, check),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		glog.Errorln("search user", err)
		return false, err
	}

	return len(result.Entries) > 0, nil
}

func CreateUser(user *models.User) (*models.User, error) {
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.Password = strings.TrimSpace(user.Password)
	user.Description = strings.TrimSpace(user.Description)

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(
		ldapclient.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", user.Username, user.Email),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		glog.Errorln("search user", err)
		return nil, err
	}

	if len(result.Entries) > 0 {
		return nil, ldap.NewError(ldap.LDAPResultEntryAlreadyExists, fmt.Errorf("username or email already exists"))
	}

	maxUid, err := getMaxUid(conn)

	if err != nil {
		glog.Errorln("get max uid", err)
		return nil, err
	}

	maxUid += 1

	userCreateRequest := ldap.NewAddRequest(fmt.Sprintf("uid=%s,%s", user.Username, ldapclient.UserSearchBase), nil)
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

	err = conn.Add(userCreateRequest)

	if err != nil {
		glog.Errorln("create user", err)
		return nil, err
	}

	if user.AvatarUrl != "" {
		setAvatar(user.Username, user.AvatarUrl)
	}

	if err := kubeconfig.CreateKubeConfig(user.Username); err != nil {
		glog.Errorln("create user kubeconfig failed", user.Username, err)
	}

	if user.ClusterRole != "" {
		err := CreateClusterRoleBinding(user.Username, user.ClusterRole)

		if err != nil {
			glog.Errorln("create cluster role binding filed", err)
			return nil, err
		}
	}

	return DescribeUser(user.Username)
}

func getMaxUid(conn ldap.Client) (int, error) {
	userSearchRequest := ldap.NewSearchRequest(ldapclient.UserSearchBase,
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

func getMaxGid(conn ldap.Client) (int, error) {

	groupSearchRequest := ldap.NewSearchRequest(ldapclient.GroupSearchBase,
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

func UpdateUser(user *models.User) (*models.User, error) {

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	dn := fmt.Sprintf("uid=%s,%s", user.Username, ldapclient.UserSearchBase)
	userModifyRequest := ldap.NewModifyRequest(dn, nil)
	if user.Email != "" {
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

	if user.AvatarUrl != "" {
		err = setAvatar(user.Username, user.AvatarUrl)
	}

	if err != nil {
		return nil, err
	}

	err = conn.Modify(userModifyRequest)

	if err != nil {
		return nil, err
	}

	err = CreateClusterRoleBinding(user.Username, user.ClusterRole)

	if err != nil {
		glog.Errorln("create cluster role binding filed", err)
		return nil, err
	}

	return GetUserInfo(user.Username)
}
func DeleteGroup(path string) error {

	// bind root DN
	conn, err := ldapclient.Client()
	if err != nil {
		return err
	}
	defer conn.Close()

	searchBase, cn := splitPath(path)

	groupDeleteRequest := ldap.NewDelRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase), nil)
	err = conn.Del(groupDeleteRequest)

	if err != nil {
		glog.Errorln("delete user group", err)
		return err
	}

	return nil
}

func CreateGroup(group *models.Group) (*models.Group, error) {

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	maxGid, err := getMaxGid(conn)

	if err != nil {
		glog.Errorln("get max gid", err)
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
		glog.Errorln("create group", err)
		return nil, err
	}

	group.Gid = strconv.Itoa(maxGid)

	return DescribeGroup(group.Path)
}

func UpdateGroup(group *models.Group) (*models.Group, error) {

	// bind root DN
	conn, err := ldapclient.Client()
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
		glog.Errorln("update group", err)
		return nil, err
	}

	return group, nil
}

func ChildList(path string) ([]models.Group, error) {

	// bind root DN
	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var groupSearchRequest *ldap.SearchRequest
	if path == "" {
		groupSearchRequest = ldap.NewSearchRequest(ldapclient.GroupSearchBase,
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

func DescribeGroup(path string) (*models.Group, error) {

	searchBase, cn := splitPath(path)

	conn, err := ldapclient.Client()

	if err != nil {
		return nil, err
	}

	groupSearchRequest := ldap.NewSearchRequest(searchBase,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=posixGroup)(cn=%s))", cn),
		[]string{"cn", "gidNumber", "memberUid", "description"},
		nil)

	result, err := conn.Search(groupSearchRequest)

	if err != nil {
		glog.Errorln("search group", err)
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

func WorkspaceUsersTotalCount(workspace string) (int, error) {
	workspaceRoleBindings, err := GetWorkspaceRoleBindings(workspace)

	if err != nil {
		return 0, err
	}

	users := make([]string, 0)

	for _, roleBinding := range workspaceRoleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
				users = append(users, subject.Name)
			}
		}
	}

	return len(users), nil
}

func ListWorkspaceUsers(workspace string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {

	workspaceRoleBindings, err := GetWorkspaceRoleBindings(workspace)

	if err != nil {
		return nil, err
	}

	users := make([]*models.User, 0)

	for _, roleBinding := range workspaceRoleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
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

func matchConditions(conditions *params.Conditions, user *models.User) bool {
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
