/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */
package iam

import (
	"github.com/go-ldap/ldap"
	"github.com/go-redis/redis"
	"golang.org/x/oauth2"
	"k8s.io/klog"
	ldappool "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"time"
)

type IdentityManagementInterface interface {
	CreateUser(user *User) (*User, error)
	DescribeUser(username string) (*User, error)
	Login(username, password, ip string) (*oauth2.Token, error)
}

type Config struct {
	authRateLimit    string
	maxAuthFailed    int
	authTimeInterval time.Duration
	tokenIdleTimeout time.Duration
	enableMultiLogin bool
}

type imOperator struct {
	config Config
	ldap   ldappool.Client
	redis  redis.Client
}

const (
	authRateLimitRegex         = `(\d+)/(\d+[s|m|h])`
	defaultMaxAuthFailed       = 5
	defaultAuthTimeInterval    = 30 * time.Minute
	mailAttribute              = "mail"
	uidAttribute               = "uid"
	descriptionAttribute       = "description"
	preferredLanguageAttribute = "preferredLanguage"
	createTimestampAttribute   = "createTimestampAttribute"
	dateTimeLayout             = "20060102150405Z"
)

func newIMOperator(ldap ldappool.Client, config Config) *imOperator {

	imOperator := &imOperator{ldap: ldap, config: config}
	return imOperator
}

func (im *imOperator) Init() error {

	userSearchBase := &ldap.AddRequest{
		DN: im.ldap.UserSearchBase(),
		Attributes: []ldap.Attribute{{
			Type: "objectClass",
			Vals: []string{"organizationalUnit", "top"},
		}, {
			Type: "ou",
			Vals: []string{"Users"},
		}},
		Controls: nil,
	}

	err := im.createIfNotExists(userSearchBase)

	if err != nil {
		return err
	}

	groupSearchBase := &ldap.AddRequest{
		DN: im.ldap.GroupSearchBase(),
		Attributes: []ldap.Attribute{{
			Type: "objectClass",
			Vals: []string{"organizationalUnit", "top"},
		}, {
			Type: "ou",
			Vals: []string{"Groups"},
		}},
		Controls: nil,
	}

	err = im.createIfNotExists(groupSearchBase)

	if err != nil {
		return err
	}

	return nil
}

func (im *imOperator) createIfNotExists(createRequest *ldap.AddRequest) error {
	conn, err := im.ldap.NewConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	searchRequest := ldap.NewSearchRequest(
		createRequest.DN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",
		nil,
		nil,
	)

	_, err = conn.Search(searchRequest)

	if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
		err = conn.Add(createRequest)
	}

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

//func parseAuthRateLimit(authRateLimit string) (int, time.Duration) {
//	regex := regexp.MustCompile(authRateLimitRegex)
//	groups := regex.FindStringSubmatch(authRateLimit)
//
//	maxCount := defaultMaxAuthFailed
//	timeInterval := defaultAuthTimeInterval
//
//	if len(groups) == 3 {
//		maxCount, _ = strconv.Atoi(groups[1])
//		timeInterval, _ = time.ParseDuration(groups[2])
//	} else {
//		klog.Warning("invalid auth rate limit", authRateLimit)
//	}
//
//	return maxCount, timeInterval
//}
//
//func (im *imOperator) Login(username, password, ip string) (*oauth2.Token, error) {
//
//	records, err := im.redis.Keys(fmt.Sprintf("kubesphere:authfailed:%s:*", username)).Result()
//
//	if err != nil {
//		klog.Error(err)
//		return nil, err
//	}
//
//	if len(records) >= im.config.maxAuthFailed {
//		return nil, restful.NewError(http.StatusTooManyRequests, "auth rate limit exceeded")
//	}
//
//	user, err := im.DescribeUser(&User{Username: username, Email: username})
//
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		klog.Error(err)
//		return nil, err
//	}
//	defer conn.Close()
//
//	dn := fmt.Sprintf("%s=%s,%s", uidAttribute, user.Username, im.ldap.UserSearchBase())
//
//	// bind as the user to verify their password
//	err = conn.Bind(dn, password)
//
//	if err != nil {
//		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
//			authFailedCacheKey := fmt.Sprintf("kubesphere:authfailed:%s:%d", user.Username, time.Now().UnixNano())
//			im.redis.Set(authFailedCacheKey, "", im.config.authTimeInterval)
//		}
//		return nil, err
//	}
//
//	claims := jwt.MapClaims{}
//
//	loginTime := time.Now()
//	// token without expiration time will auto sliding
//	claims["username"] = user.Username
//	claims["email"] = user.Email
//	claims["iat"] = loginTime.Unix()
//
//	token := jwtutil.MustSigned(claims)
//
//	if !im.config.enableMultiLogin {
//		// multi login not allowed, remove the previous token
//		sessionCacheKey := fmt.Sprintf("kubesphere:users:%s:token:*", user.Username)
//		sessions, err := im.redis.Keys(sessionCacheKey).Result()
//
//		if err != nil {
//			klog.Errorln(err)
//			return nil, err
//		}
//
//		if len(sessions) > 0 {
//			klog.V(4).Infoln("revoke token", sessions)
//			err = im.redis.Del(sessions...).Err()
//			if err != nil {
//				klog.Errorln(err)
//				return nil, err
//			}
//		}
//	}
//
//	// cache token with expiration time
//	sessionCacheKey := fmt.Sprintf("kubesphere:users:%s:token:%s", user.Username, token)
//	if err = im.redis.Set(sessionCacheKey, token, im.config.tokenIdleTimeout).Err(); err != nil {
//		klog.Errorln(err)
//		return nil, err
//	}
//
//	im.loginRecord(user.Username, ip, loginTime)
//
//	return &oauth2.Token{AccessToken: token}, nil
//}
//
//func (im *imOperator) loginRecord(username, ip string, loginTime time.Time) {
//	if ip != "" {
//		im.redis.RPush(fmt.Sprintf("kubesphere:users:%s:login-log", username), fmt.Sprintf("%s,%s", loginTime.UTC().Format("2006-01-02T15:04:05Z"), ip))
//		im.redis.LTrim(fmt.Sprintf("kubesphere:users:%s:login-log", username), -10, -1)
//	}
//}
//
//func (im *imOperator) LoginHistory(username string) ([]string, error) {
//	data, err := im.redis.LRange(fmt.Sprintf("kubesphere:users:%s:login-log", username), -10, -1).Result()
//
//	if err != nil {
//		return nil, err
//	}
//
//	return data, nil
//}
//
//func (im *imOperator) ListUsers(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		return nil, err
//	}
//	defer conn.Close()
//
//	pageControl := ldap.NewControlPaging(1000)
//
//	users := make([]User, 0)
//
//	filter := "(&(objectClass=inetOrgPerson))"
//
//	if keyword := conditions.Match["keyword"]; keyword != "" {
//		filter = fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=*%s*)(mail=*%s*)(description=*%s*)))", keyword, keyword, keyword)
//	}
//
//	if username := conditions.Match["username"]; username != "" {
//		uidFilter := ""
//		for _, username := range strings.Split(username, "|") {
//			uidFilter += fmt.Sprintf("(uid=%s)", username)
//		}
//		filter = fmt.Sprintf("(&(objectClass=inetOrgPerson)(|%s))", uidFilter)
//	}
//
//	if email := conditions.Match["email"]; email != "" {
//		emailFilter := ""
//		for _, username := range strings.Split(email, "|") {
//			emailFilter += fmt.Sprintf("(mail=%s)", username)
//		}
//		filter = fmt.Sprintf("(&(objectClass=inetOrgPerson)(|%s))", emailFilter)
//	}
//
//	for {
//		userSearchRequest := ldap.NewSearchRequest(
//			im.ldap.UserSearchBase(),
//			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
//			filter,
//			[]string{"uid", "mail", "description", "preferredLanguage", "createTimestamp"},
//			[]ldap.Control{pageControl},
//		)
//
//		response, err := conn.Search(userSearchRequest)
//
//		if err != nil {
//			klog.Errorln("search user", err)
//			return nil, err
//		}
//
//		for _, entry := range response.Entries {
//
//			uid := entry.GetAttributeValue("uid")
//			email := entry.GetAttributeValue("mail")
//			description := entry.GetAttributeValue("description")
//			lang := entry.GetAttributeValue("preferredLanguage")
//			createTimestamp, _ := time.Parse("20060102150405Z", entry.GetAttributeValue("createTimestamp"))
//
//			user := User{Username: uid, Email: email, Description: description, Lang: lang, CreateTime: createTimestamp}
//
//			if !im.shouldHidden(user) {
//				users = append(users, user)
//			}
//		}
//
//		updatedControl := ldap.FindControl(response.Controls, ldap.ControlTypePaging)
//		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
//			pageControl.SetCookie(ctrl.Cookie)
//			continue
//		}
//
//		break
//	}
//
//	sort.Slice(users, func(i, j int) bool {
//		if reverse {
//			tmp := i
//			i = j
//			j = tmp
//		}
//		switch orderBy {
//		case "username":
//			return strings.Compare(users[i].Username, users[j].Username) <= 0
//		case "createTime":
//			fallthrough
//		default:
//			return users[i].CreateTime.Before(users[j].CreateTime)
//		}
//	})
//
//	items := make([]interface{}, 0)
//
//	for i, user := range users {
//
//		if i >= offset && len(items) < limit {
//
//			user.LastLoginTime = im.GetLastLoginTime(user.Username)
//			clusterRole, err := im.GetUserClusterRole(user.Username)
//			if err != nil {
//				return nil, err
//			}
//			user.ClusterRole = clusterRole.Name
//			items = append(items, user)
//		}
//	}
//
//	return &models.PageableResponse{Items: items, TotalCount: len(users)}, nil
//}
//
//func (im *imOperator) shouldHidden(user User) bool {
//	for _, initUser := range im.initUsers {
//		if initUser.Username == user.Username {
//			return initUser.Hidden
//		}
//	}
//	return false
//}
//
//func (im *imOperator) DescribeUser(user *User) (*User, error) {
//	conn, err := im.ldap.NewConn()
//
//	if err != nil {
//		klog.Errorln(err)
//		return nil, err
//	}
//	defer conn.Close()
//
//	filter := fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", user.Username, user.Email)
//
//	searchRequest := ldap.NewSearchRequest(
//		im.ldap.UserSearchBase(),
//		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 1, 0, false,
//		filter,
//		[]string{mailAttribute, descriptionAttribute, preferredLanguageAttribute, createTimestampAttribute},
//		nil,
//	)
//
//	result, err := conn.Search(searchRequest)
//
//	if err != nil {
//		klog.Errorln(err)
//		return nil, err
//	}
//
//	if len(result.Entries) != 1 {
//		return nil, ldap.NewError(ldap.LDAPResultNoSuchObject, errors.New("user does not exist"))
//	}
//
//	entry := result.Entries[0]
//
//	return convertLdapEntryToUser(entry), nil
//}
//
//func convertLdapEntryToUser(entry *ldap.Entry) *User {
//	username := entry.GetAttributeValue(uidAttribute)
//	email := entry.GetAttributeValue(mailAttribute)
//	description := entry.GetAttributeValue(descriptionAttribute)
//	lang := entry.GetAttributeValue(preferredLanguageAttribute)
//	createTimestamp, _ := time.Parse(dateTimeLayout, entry.GetAttributeValue(createTimestampAttribute))
//	return &User{Username: username, Email: email, Description: description, Lang: lang, CreateTime: createTimestamp}
//}
//
//func (im *imOperator) GetLastLoginTime(username string) string {
//	cacheKey := fmt.Sprintf("kubesphere:users:%s:login-log", username)
//	lastLogin, err := im.redis.LRange(cacheKey, -1, -1).Result()
//	if err != nil {
//		return ""
//	}
//
//	if len(lastLogin) > 0 {
//		return strings.Split(lastLogin[0], ",")[0]
//	}
//
//	return ""
//}
//
//func (im *imOperator) DeleteUser(username string) error {
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		return err
//	}
//	defer conn.Close()
//
//	deleteRequest := ldap.NewDelRequest(fmt.Sprintf("uid=%s,%s", username, im.ldap.UserSearchBase()), nil)
//
//	if err = conn.Del(deleteRequest); err != nil {
//		klog.Errorln("delete user", err)
//		return err
//	}
//
//	if err = im.deleteRoleBindings(username); err != nil {
//		klog.Errorln("delete user role bindings failed", username, err)
//	}
//
//	if err := kubeconfig.DelKubeConfig(username); err != nil {
//		klog.Errorln("delete user kubeconfig failed", username, err)
//	}
//
//	if err := kubectl.DelKubectlDeploy(username); err != nil {
//		klog.Errorln("delete user terminal pod failed", username, err)
//	}
//
//	if err := im.deleteUserInDevOps(username); err != nil {
//		klog.Errorln("delete user in devops failed", username, err)
//	}
//	return nil
//
//}
//
//// deleteUserInDevOps is used to clean up user data of devops, such as permission rules
//func (im *imOperator) deleteUserInDevOps(username string) error {
//
//	devopsDb, err := clientset.ClientSets().MySQL()
//	if err != nil {
//		if err == clientset.ErrClientSetNotEnabled {
//			klog.Warning("mysql is not enable")
//			return nil
//		}
//		return err
//	}
//
//	dp, err := clientset.ClientSets().Devops()
//	if err != nil {
//		if err == clientset.ErrClientSetNotEnabled {
//			klog.Warning("devops client is not enable")
//			return nil
//		}
//		return err
//	}
//
//	jenkinsClient := dp.Jenkins()
//
//	_, err = devopsDb.DeleteFrom(devops.DevOpsProjectMembershipTableName).
//		Where(db.And(
//			db.Eq(devops.DevOpsProjectMembershipUsernameColumn, username),
//		)).Exec()
//	if err != nil {
//		klog.Errorf("%+v", err)
//		return err
//	}
//
//	err = jenkinsClient.DeleteUserInProject(username)
//	if err != nil {
//		klog.Errorf("%+v", err)
//		return err
//	}
//	return nil
//}
//
//func (im *imOperator) deleteRoleBindings(username string) error {
//	roleBindingLister := informers.SharedInformerFactory().Rbac().V1().RoleBindings().Lister()
//	roleBindings, err := roleBindingLister.List(labels.Everything())
//
//	if err != nil {
//		return err
//	}
//
//	for _, roleBinding := range roleBindings {
//		roleBinding = roleBinding.DeepCopy()
//		length1 := len(roleBinding.Subjects)
//
//		for index, subject := range roleBinding.Subjects {
//			if subject.Kind == rbacv1.UserKind && subject.Name == username {
//				roleBinding.Subjects = append(roleBinding.Subjects[:index], roleBinding.Subjects[index+1:]...)
//				index--
//			}
//		}
//
//		length2 := len(roleBinding.Subjects)
//
//		if length2 == 0 {
//			deletePolicy := metav1.DeletePropagationBackground
//			err = clientset.ClientSets().K8s().Kubernetes().RbacV1().RoleBindings(roleBinding.Namespace).Delete(roleBinding.Name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
//
//			if err != nil {
//				klog.Errorf("delete role binding %s %s %s failed: %v", username, roleBinding.Namespace, roleBinding.Name, err)
//			}
//		} else if length2 < length1 {
//			_, err = clientset.ClientSets().K8s().Kubernetes().RbacV1().RoleBindings(roleBinding.Namespace).Update(roleBinding)
//
//			if err != nil {
//				klog.Errorf("update role binding %s %s %s failed: %v", username, roleBinding.Namespace, roleBinding.Name, err)
//			}
//		}
//	}
//
//	clusterRoleBindingLister := informers.SharedInformerFactory().Rbac().V1().ClusterRoleBindings().Lister()
//	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())
//
//	for _, clusterRoleBinding := range clusterRoleBindings {
//		clusterRoleBinding = clusterRoleBinding.DeepCopy()
//		length1 := len(clusterRoleBinding.Subjects)
//
//		for index, subject := range clusterRoleBinding.Subjects {
//			if subject.Kind == rbacv1.UserKind && subject.Name == username {
//				clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects[:index], clusterRoleBinding.Subjects[index+1:]...)
//				index--
//			}
//		}
//
//		length2 := len(clusterRoleBinding.Subjects)
//		if length2 == 0 {
//			// delete if it's not workspace role binding
//			if isWorkspaceRoleBinding(clusterRoleBinding) {
//				_, err = clientset.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
//			} else {
//				deletePolicy := metav1.DeletePropagationBackground
//				err = clientset.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Delete(clusterRoleBinding.Name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
//			}
//			if err != nil {
//				klog.Errorf("update cluster role binding %s failed:%s", clusterRoleBinding.Name, err)
//			}
//		} else if length2 < length1 {
//			_, err = clientset.ClientSets().K8s().Kubernetes().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
//
//			if err != nil {
//				klog.Errorf("update cluster role binding %s failed:%s", clusterRoleBinding.Name, err)
//			}
//		}
//
//	}
//
//	return nil
//}
//
//func (im *imOperator) isWorkspaceRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) bool {
//	return k8sutil.IsControlledBy(clusterRoleBinding.OwnerReferences, "Workspace", "")
//}
//
//func (im *imOperator) CreateUser(user *User) (*User, error) {
//	user.Username = strings.TrimSpace(user.Username)
//	user.Email = strings.TrimSpace(user.Email)
//	user.Password = strings.TrimSpace(user.Password)
//	user.Description = strings.TrimSpace(user.Description)
//
//	existed, err := im.DescribeUser(user)
//
//	if err != nil {
//		klog.Errorln(err)
//		return nil, err
//	}
//
//	if existed != nil {
//		return nil, ldap.NewError(ldap.LDAPResultEntryAlreadyExists, errors.New("username or email already exists"))
//	}
//
//	uidNumber := im.uidNumberNext()
//
//	createRequest := ldap.NewAddRequest(fmt.Sprintf("uid=%s,%s", user.Username, im.ldap.UserSearchBase()), nil)
//	createRequest.Attribute("objectClass", []string{"inetOrgPerson", "posixAccount", "top"})
//	createRequest.Attribute("cn", []string{user.Username})                       // RFC4519: common name(s) for which the entity is known by
//	createRequest.Attribute("sn", []string{" "})                                 // RFC2256: last (family) name(s) for which the entity is known by
//	createRequest.Attribute("gidNumber", []string{"500"})                        // RFC2307: An integer uniquely identifying a group in an administrative domain
//	createRequest.Attribute("homeDirectory", []string{"/home/" + user.Username}) // The absolute path to the home directory
//	createRequest.Attribute("uid", []string{user.Username})                      // RFC4519: user identifier
//	createRequest.Attribute("uidNumber", []string{strconv.Itoa(uidNumber)})      // RFC2307: An integer uniquely identifying a user in an administrative domain
//	createRequest.Attribute("mail", []string{user.Email})                        // RFC1274: RFC822 Mailbox
//	createRequest.Attribute("userPassword", []string{user.Password})             // RFC4519/2307: password of user
//	if user.Lang != "" {
//		createRequest.Attribute("preferredLanguage", []string{user.Lang})
//	}
//	if user.Description != "" {
//		createRequest.Attribute("description", []string{user.Description}) // RFC4519: descriptive information
//	}
//
//	conn, err := im.ldap.NewConn()
//
//	if err != nil {
//		klog.Errorln(err)
//		return nil, err
//	}
//
//	err = conn.Add(createRequest)
//
//	if err != nil {
//		klog.Errorln(err)
//		return nil, err
//	}
//
//	return user, nil
//}
//
//func (im *imOperator) UpdateUser(user *User) (*User, error) {
//
//	client, err := clientset.ClientSets().Ldap()
//	if err != nil {
//		return nil, err
//	}
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		return nil, err
//	}
//	defer conn.Close()
//
//	dn := fmt.Sprintf("uid=%s,%s", user.Username, im.ldap.UserSearchBase())
//	userModifyRequest := ldap.NewModifyRequest(dn, nil)
//	if user.Email != "" {
//		userSearchRequest := ldap.NewSearchRequest(
//			im.ldap.UserSearchBase(),
//			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
//			fmt.Sprintf("(&(objectClass=inetOrgPerson)(mail=%s))", user.Email),
//			[]string{"uid", "mail"},
//			nil,
//		)
//		result, err := conn.Search(userSearchRequest)
//		if err != nil {
//			klog.Error(err)
//			return nil, err
//		}
//		if len(result.Entries) > 1 {
//			err = ldap.NewError(ldap.ErrorDebugging, fmt.Errorf("email is duplicated: %s", user.Email))
//			klog.Error(err)
//			return nil, err
//		}
//		if len(result.Entries) == 1 && result.Entries[0].GetAttributeValue("uid") != user.Username {
//			err = ldap.NewError(ldap.LDAPResultEntryAlreadyExists, fmt.Errorf("email is duplicated: %s", user.Email))
//			klog.Error(err)
//			return nil, err
//		}
//		userModifyRequest.Replace("mail", []string{user.Email})
//	}
//	if user.Description != "" {
//		userModifyRequest.Replace("description", []string{user.Description})
//	}
//
//	if user.Lang != "" {
//		userModifyRequest.Replace("preferredLanguage", []string{user.Lang})
//	}
//
//	if user.Password != "" {
//		userModifyRequest.Replace("userPassword", []string{user.Password})
//	}
//
//	err = conn.Modify(userModifyRequest)
//
//	if err != nil {
//		klog.Error(err)
//		return nil, err
//	}
//
//	if user.ClusterRole != "" {
//		err = CreateClusterRoleBinding(user.Username, user.ClusterRole)
//
//		if err != nil {
//			klog.Errorln(err)
//			return nil, err
//		}
//	}
//
//	// clear auth failed record
//	if user.Password != "" {
//		redisClient, err := clientset.ClientSets().Redis()
//		if err != nil {
//			return nil, err
//		}
//
//		records, err := im.redis.Keys(fmt.Sprintf("kubesphere:authfailed:%s:*", user.Username)).Result()
//
//		if err == nil {
//			im.redis.Del(records...)
//		}
//	}
//
//	return GetUserInfo(user.Username)
//}
//func (im *imOperator) DeleteGroup(path string) error {
//
//	client, err := clientset.ClientSets().Ldap()
//	if err != nil {
//		return err
//	}
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		return err
//	}
//	defer conn.Close()
//
//	searchBase, cn := splitPath(path)
//
//	groupDeleteRequest := ldap.NewDelRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase), nil)
//	err = conn.Del(groupDeleteRequest)
//
//	if err != nil {
//		klog.Errorln("delete user group", err)
//		return err
//	}
//
//	return nil
//}
//
//func (im *imOperator) CreateGroup(group *models.Group) (*models.Group, error) {
//
//	client, err := clientset.ClientSets().Ldap()
//	if err != nil {
//		return nil, err
//	}
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		return nil, err
//	}
//	defer conn.Close()
//
//	maxGid, err := getMaxGid()
//
//	if err != nil {
//		klog.Errorln("get max gid", err)
//		return nil, err
//	}
//
//	maxGid += 1
//
//	if group.Path == "" {
//		group.Path = group.Name
//	}
//
//	searchBase, cn := splitPath(group.Path)
//
//	groupCreateRequest := ldap.NewAddRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase), nil)
//	groupCreateRequest.Attribute("objectClass", []string{"posixGroup", "top"})
//	groupCreateRequest.Attribute("cn", []string{cn})
//	groupCreateRequest.Attribute("gidNumber", []string{strconv.Itoa(maxGid)})
//
//	if group.Description != "" {
//		groupCreateRequest.Attribute("description", []string{group.Description})
//	}
//
//	if group.Members != nil {
//		groupCreateRequest.Attribute("memberUid", group.Members)
//	}
//
//	err = conn.Add(groupCreateRequest)
//
//	if err != nil {
//		klog.Errorln("create group", err)
//		return nil, err
//	}
//
//	group.Gid = strconv.Itoa(maxGid)
//
//	return DescribeGroup(group.Path)
//}
//
//func (im *imOperator) UpdateGroup(group *models.Group) (*models.Group, error) {
//
//	client, err := clientset.ClientSets().Ldap()
//	if err != nil {
//		return nil, err
//	}
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		return nil, err
//	}
//	defer conn.Close()
//
//	old, err := DescribeGroup(group.Path)
//
//	if err != nil {
//		return nil, err
//	}
//
//	searchBase, cn := splitPath(group.Path)
//
//	groupUpdateRequest := ldap.NewModifyRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase), nil)
//
//	if old.Description == "" {
//		if group.Description != "" {
//			groupUpdateRequest.Add("description", []string{group.Description})
//		}
//	} else {
//		if group.Description != "" {
//			groupUpdateRequest.Replace("description", []string{group.Description})
//		} else {
//			groupUpdateRequest.Delete("description", []string{})
//		}
//	}
//
//	if group.Members != nil {
//		groupUpdateRequest.Replace("memberUid", group.Members)
//	}
//
//	err = conn.Modify(groupUpdateRequest)
//
//	if err != nil {
//		klog.Errorln("update group", err)
//		return nil, err
//	}
//
//	return group, nil
//}
//
//func (im *imOperator) ChildList(path string) ([]models.Group, error) {
//
//	client, err := clientset.ClientSets().Ldap()
//	if err != nil {
//		return nil, err
//	}
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		return nil, err
//	}
//	defer conn.Close()
//
//	var groupSearchRequest *ldap.SearchRequest
//	if path == "" {
//		groupSearchRequest = ldap.NewSearchRequest(client.GroupSearchBase(),
//			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
//			"(&(objectClass=posixGroup))",
//			[]string{"cn", "gidNumber", "memberUid", "description"},
//			nil)
//	} else {
//		searchBase, cn := splitPath(path)
//		groupSearchRequest = ldap.NewSearchRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase),
//			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
//			"(&(objectClass=posixGroup))",
//			[]string{"cn", "gidNumber", "memberUid", "description"},
//			nil)
//	}
//
//	result, err := conn.Search(groupSearchRequest)
//
//	if err != nil {
//		return nil, err
//	}
//
//	groups := make([]models.Group, 0)
//
//	for _, v := range result.Entries {
//		dn := v.DN
//		cn := v.GetAttributeValue("cn")
//		gid := v.GetAttributeValue("gidNumber")
//		members := v.GetAttributeValues("memberUid")
//		description := v.GetAttributeValue("description")
//
//		group := models.Group{Path: convertDNToPath(dn), Name: cn, Gid: gid, Members: members, Description: description}
//
//		childSearchRequest := ldap.NewSearchRequest(dn,
//			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
//			"(&(objectClass=posixGroup))",
//			[]string{""},
//			nil)
//
//		result, err = conn.Search(childSearchRequest)
//
//		if err != nil {
//			return nil, err
//		}
//
//		childGroups := make([]string, 0)
//
//		for _, v := range result.Entries {
//			child := convertDNToPath(v.DN)
//			childGroups = append(childGroups, child)
//		}
//
//		group.ChildGroups = childGroups
//
//		groups = append(groups, group)
//	}
//
//	return groups, nil
//}
//
//func (im *imOperator) DescribeGroup(path string) (*models.Group, error) {
//	client, err := clientset.ClientSets().Ldap()
//	if err != nil {
//		return nil, err
//	}
//	conn, err := im.ldap.NewConn()
//	if err != nil {
//		return nil, err
//	}
//	defer conn.Close()
//
//	searchBase, cn := splitPath(path)
//
//	groupSearchRequest := ldap.NewSearchRequest(searchBase,
//		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
//		fmt.Sprintf("(&(objectClass=posixGroup)(cn=%s))", cn),
//		[]string{"cn", "gidNumber", "memberUid", "description"},
//		nil)
//
//	result, err := conn.Search(groupSearchRequest)
//
//	if err != nil {
//		klog.Errorln("search group", err)
//		return nil, err
//	}
//
//	if len(result.Entries) != 1 {
//		return nil, ldap.NewError(ldap.LDAPResultNoSuchObject, fmt.Errorf("group %s does not exist", path))
//	}
//
//	dn := result.Entries[0].DN
//	cn = result.Entries[0].GetAttributeValue("cn")
//	gid := result.Entries[0].GetAttributeValue("gidNumber")
//	members := result.Entries[0].GetAttributeValues("memberUid")
//	description := result.Entries[0].GetAttributeValue("description")
//
//	group := models.Group{Path: convertDNToPath(dn), Name: cn, Gid: gid, Members: members, Description: description}
//
//	childGroups := make([]string, 0)
//
//	group.ChildGroups = childGroups
//
//	return &group, nil
//
//}
//
//func (im *imOperator) WorkspaceUsersTotalCount(workspace string) (int, error) {
//	workspaceRoleBindings, err := GetWorkspaceRoleBindings(workspace)
//
//	if err != nil {
//		return 0, err
//	}
//
//	users := make([]string, 0)
//
//	for _, roleBinding := range workspaceRoleBindings {
//		for _, subject := range roleBinding.Subjects {
//			if subject.Kind == rbacv1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
//				users = append(users, subject.Name)
//			}
//		}
//	}
//
//	return len(users), nil
//}
//
//func (im *imOperator) ListWorkspaceUsers(workspace string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
//
//	workspaceRoleBindings, err := GetWorkspaceRoleBindings(workspace)
//
//	if err != nil {
//		return nil, err
//	}
//
//	users := make([]*User, 0)
//
//	for _, roleBinding := range workspaceRoleBindings {
//		for _, subject := range roleBinding.Subjects {
//			if subject.Kind == rbacv1.UserKind && !k8sutil.ContainsUser(users, subject.Name) {
//				user, err := GetUserInfo(subject.Name)
//				if err != nil {
//					return nil, err
//				}
//				prefix := fmt.Sprintf("workspace:%s:", workspace)
//				user.WorkspaceRole = fmt.Sprintf("workspace-%s", strings.TrimPrefix(roleBinding.Name, prefix))
//				if matchConditions(conditions, user) {
//					users = append(users, user)
//				}
//			}
//		}
//	}
//
//	// order & reverse
//	sort.Slice(users, func(i, j int) bool {
//		if reverse {
//			tmp := i
//			i = j
//			j = tmp
//		}
//		switch orderBy {
//		default:
//			fallthrough
//		case "name":
//			return strings.Compare(users[i].Username, users[j].Username) <= 0
//		}
//	})
//
//	result := make([]interface{}, 0)
//
//	for i, d := range users {
//		if i >= offset && (limit == -1 || len(result) < limit) {
//			result = append(result, d)
//		}
//	}
//
//	return &models.PageableResponse{Items: result, TotalCount: len(users)}, nil
//}
//
//func (im *imOperator) uidNumberNext() int {
//	return 0
//}
//
//func matchConditions(conditions *params.Conditions, user *User) bool {
//	for k, v := range conditions.Match {
//		switch k {
//		case "keyword":
//			if !strings.Contains(user.Username, v) &&
//				!strings.Contains(user.Email, v) &&
//				!strings.Contains(user.Description, v) {
//				return false
//			}
//		case "name":
//			names := strings.Split(v, "|")
//			if !sliceutil.HasString(names, user.Username) {
//				return false
//			}
//		case "email":
//			email := strings.Split(v, "|")
//			if !sliceutil.HasString(email, user.Email) {
//				return false
//			}
//		case "role":
//			if user.WorkspaceRole != v {
//				return false
//			}
//		}
//	}
//	return true
//}
//
type User struct {
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Lang        string    `json:"lang,omitempty"`
	Description string    `json:"description"`
	CreateTime  time.Time `json:"create_time"`
	Groups      []string  `json:"groups,omitempty"`
	Password    string    `json:"password,omitempty"`
}
