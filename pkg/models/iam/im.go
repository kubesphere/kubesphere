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
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-ldap/ldap"
	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/kubesphere/pkg/client"
	ldapPool "kubesphere.io/kubesphere/pkg/client/ldap"
	"kubesphere.io/kubesphere/pkg/constants"
	secret "kubesphere.io/kubesphere/pkg/iam/jwt"
	. "kubesphere.io/kubesphere/pkg/models"
)

var once sync.Once
var pool ldapPool.Pool
var counter Counter

func getPool() ldapPool.Pool {

	if pool == nil {
		once.Do(poolInit)
	}

	return pool
}

func poolInit() {
	var err error
	pool, err = ldapPool.NewChannelPool(8, 96, "ks-account", func(s string) (ldap.Client, error) {
		conn, err := ldap.Dial("tcp", constants.LdapServerHost)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}, []uint16{ldap.LDAPResultTimeLimitExceeded, ldap.ErrorNetwork})

	if err != nil {
		log.Panicln(err)
	}
}

func CheckAndInit() error {
	var conn ldap.Client
	var err error
	maxRetry := 5
	for retry := 0; retry < maxRetry; retry++ {
		conn, err = NewConnection()
		if err == nil {
			break
		} else if retry == maxRetry-1 {
			log.Printf("cannot connect to ldap server ,%s", err)
		} else {
			log.Printf("cannot connect to ldap server ,retry %d/%d\n after 2s", retry+1, maxRetry)
		}
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		return err
	}

	defer conn.Close()

	if err != nil {
		return err
	}

	// search for the given username
	userSearchRequest := ldap.NewSearchRequest(
		constants.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=inetOrgPerson))",
		nil,
		nil,
	)

	users, err := conn.Search(userSearchRequest)

	if err != nil {
		switch err.(type) {
		case *ldap.Error:
			if err.(*ldap.Error).ResultCode == 32 {
				err := createUserBaseDN()
				if err != nil {
					return fmt.Errorf("UserBaseDN %s create failed: %s\n", constants.UserSearchBase, err)
				} else {
					log.Printf("UserBaseDN %s create success\n", constants.UserSearchBase)
				}
			} else {
				return fmt.Errorf("UserBaseDN %s not exist: %s\n", constants.UserSearchBase, err)
			}
		default:
			return fmt.Errorf("UserBaseDN %s not exist: %s\n", constants.UserSearchBase, err)
		}
	}

	counter = NewCounter(len(users.Entries))

	if users == nil || len(users.Entries) == 0 {
		err := CreateUser(User{Username: constants.AdminUserName, Email: constants.AdminEmail, Password: constants.AdminPWD, Description: "Administrator account that was always created by default."})

		if err != nil {
			return fmt.Errorf("admin create failed: %s\n", err)
		}

		log.Println("admin init success")
	}

	// search user group
	groupSearchRequest := ldap.NewSearchRequest(
		constants.GroupSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=posixGroup))",
		nil,
		nil,
	)

	groups, err := conn.Search(groupSearchRequest)

	if err != nil {
		switch err.(type) {
		case *ldap.Error:
			if err.(*ldap.Error).ResultCode == 32 {
				err := createGroupsBaseDN()
				if err != nil {
					return fmt.Errorf("GroupBaseDN %s create failed: %s\n", constants.GroupSearchBase, err)
				} else {
					log.Printf("GroupBaseDN %s create success\n", constants.GroupSearchBase)
				}
			} else {
				return fmt.Errorf("GroupBaseDN %s not exist: %s\n", constants.GroupSearchBase, err)
			}
		default:
			return fmt.Errorf("GroupBaseDN %s not exist: %s\n", constants.GroupSearchBase, err)
		}
	}

	if groups == nil || len(groups.Entries) == 0 {
		systemGroup := Group{Path: constants.SystemWorkspace, Name: constants.SystemWorkspace, Creator: constants.AdminUserName, Description: "system workspace"}

		_, err = CreateGroup(systemGroup)

		if err != nil {
			return fmt.Errorf("system-group create failed: %s\n", err)
		}

		log.Println("system-workspace init success")
	}

	return nil
}

func createUserBaseDN() error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	groupsCreateRequest := ldap.NewAddRequest(constants.UserSearchBase, nil)
	groupsCreateRequest.Attribute("objectClass", []string{"organizationalUnit", "top"})
	groupsCreateRequest.Attribute("ou", []string{"Users"})
	return conn.Add(groupsCreateRequest)
}

func createGroupsBaseDN() error {

	conn, err := NewConnection()

	if err != nil {
		return err
	}

	defer conn.Close()

	groupsCreateRequest := ldap.NewAddRequest(constants.GroupSearchBase, nil)
	groupsCreateRequest.Attribute("objectClass", []string{"organizationalUnit", "top"})
	groupsCreateRequest.Attribute("ou", []string{"Groups"})
	return conn.Add(groupsCreateRequest)
}

func NewConnection() (ldap.Client, error) {
	conn, err := getPool().Get()
	//conn, err := ldap.Dial("tcp", constants.LdapServerHost)
	if err != nil {
		return nil, err
	}
	err = conn.Bind(constants.RootDN, constants.RootPWD)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// User login
func Login(username string, password string, ip string) (string, error) {

	conn, err := NewConnection()

	if err != nil {
		return "", err
	}

	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(
		constants.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", username, username),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		return "", err
	}

	if len(result.Entries) != 1 {
		return "", ldap.NewError(ldap.LDAPResultInvalidCredentials, errors.New("incorrect password"))
	}

	uid := result.Entries[0].GetAttributeValue("uid")
	email := result.Entries[0].GetAttributeValue("mail")
	dn := result.Entries[0].DN

	user := User{Username: uid, Email: email}

	// bind as the user to verify their password
	err = conn.Bind(dn, password)

	if err != nil {
		return "", err
	}

	if ip != "" {
		redisClient := client.RedisClient()
		redisClient.RPush(fmt.Sprintf("kubesphere:users:%s:login-log", uid), fmt.Sprintf("%s,%s", time.Now().UTC().Format("2006-01-02T15:04:05Z"), ip))
		redisClient.LTrim(fmt.Sprintf("kubesphere:users:%s:login-log", uid), -10, -1)
	}

	claims := jwt.MapClaims{}

	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	claims["username"] = user.Username
	claims["email"] = user.Email

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	uToken, _ := token.SignedString(secret.Secret)

	return uToken, nil
}

func UserList(limit int, offset int) (int, []User, error) {

	conn, err := NewConnection()

	if err != nil {
		return 0, nil, err
	}

	defer conn.Close()

	users := make([]User, 0)

	pageControl := ldap.NewControlPaging(1000)

	entries := make([]*ldap.Entry, 0)

	cursor := 0
l1:
	for {

		userSearchRequest := ldap.NewSearchRequest(
			constants.UserSearchBase,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			"(&(objectClass=inetOrgPerson))",
			[]string{"uid", "mail", "description"},
			[]ldap.Control{pageControl},
		)

		response, err := conn.Search(userSearchRequest)

		if err != nil {
			return 0, nil, err
		}

		for _, entry := range response.Entries {
			cursor++
			if cursor > offset {
				if len(entries) < limit {
					entries = append(entries, entry)
				} else {
					break l1
				}
			}
		}

		updatedControl := ldap.FindControl(response.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pageControl.SetCookie(ctrl.Cookie)
			continue
		}

		break
	}

	redisClient := client.RedisClient()

	for _, v := range entries {

		uid := v.GetAttributeValue("uid")
		email := v.GetAttributeValue("mail")
		description := v.GetAttributeValue("description")
		user := User{Username: uid, Email: email, Description: description}

		avatar, err := redisClient.HMGet("kubesphere:users:avatar", uid).Result()

		if err != nil {
			return 0, nil, err
		}

		if len(avatar) > 0 {
			if url, ok := avatar[0].(string); ok {
				user.AvatarUrl = url
			}
		}

		lastLogin, err := redisClient.LRange(fmt.Sprintf("kubesphere:users:%s:login-log", uid), -1, -1).Result()

		if err != nil {
			return 0, nil, err
		}

		if len(lastLogin) > 0 {
			user.LastLoginTime = strings.Split(lastLogin[0], ",")[0]
		}

		user.ClusterRules = make([]SimpleRule, 0)

		users = append(users, user)
	}

	return counter.Get(), users, nil
}

func LoginLog(username string) ([]string, error) {
	redisClient := client.RedisClient()

	data, err := redisClient.LRange(fmt.Sprintf("kubesphere:users:%s:login-log", username), -10, -1).Result()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func Search(keyword string, limit int, offset int) (int, []User, error) {

	conn, err := NewConnection()

	if err != nil {
		return 0, nil, err
	}

	defer conn.Close()

	users := make([]User, 0)

	pageControl := ldap.NewControlPaging(80)

	entries := make([]*ldap.Entry, 0)

	cursor := 0
l1:
	for {
		userSearchRequest := ldap.NewSearchRequest(
			constants.UserSearchBase,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=*%s*)(mail=*%s*)(description=*%s*)))", keyword, keyword, keyword),
			[]string{"uid", "mail", "description"},
			[]ldap.Control{pageControl},
		)

		response, err := conn.Search(userSearchRequest)

		if err != nil {
			return 0, nil, err
		}

		for _, entry := range response.Entries {
			cursor++
			if cursor > offset {
				if len(entries) < limit {
					entries = append(entries, entry)
				} else {
					break l1
				}
			}
		}

		updatedControl := ldap.FindControl(response.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pageControl.SetCookie(ctrl.Cookie)
			continue
		}

		break
	}

	redisClient := client.RedisClient()

	for _, v := range entries {

		uid := v.GetAttributeValue("uid")
		email := v.GetAttributeValue("mail")
		description := v.GetAttributeValue("description")
		user := User{Username: uid, Email: email, Description: description}

		avatar, err := redisClient.HMGet("kubesphere:users:avatar", uid).Result()

		if err != nil {
			return 0, nil, err
		}

		if len(avatar) > 0 {
			if url, ok := avatar[0].(string); ok {
				user.AvatarUrl = url
			}
		}

		lastLogin, err := redisClient.LRange(fmt.Sprintf("kubesphere:users:%s:login-log", uid), -1, -1).Result()

		if err != nil {
			return 0, nil, err
		}

		if len(lastLogin) > 0 {
			user.LastLoginTime = strings.Split(lastLogin[0], ",")[0]
		}

		user.ClusterRules = make([]SimpleRule, 0)

		users = append(users, user)
	}

	return counter.Get(), users, nil
}

func UserDetail(username string, conn ldap.Client) (*User, error) {

	userSearchRequest := ldap.NewSearchRequest(
		constants.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(uid=%s))", username),
		[]string{"mail", "description", "preferredLanguage"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		return nil, err
	}

	if len(result.Entries) != 1 {
		return nil, ldap.NewError(ldap.LDAPResultNoSuchObject, fmt.Errorf("user %s does not exist", username))
	}

	email := result.Entries[0].GetAttributeValue("mail")
	description := result.Entries[0].GetAttributeValue("description")
	lang := result.Entries[0].GetAttributeValue("preferredLanguage")
	user := User{Username: username, Email: email, Description: description, Lang: lang}

	groupSearchRequest := ldap.NewSearchRequest(
		constants.GroupSearchBase,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=posixGroup)(memberUid=%s))", username),
		nil,
		nil,
	)

	result, err = conn.Search(groupSearchRequest)

	if err != nil {
		return nil, err

	}

	groups := make([]string, 0)

	for _, group := range result.Entries {
		groupName := convertDNToPath(group.DN)
		groups = append(groups, groupName)
	}

	user.Groups = groups

	redisClient := client.RedisClient()

	avatar, err := redisClient.HMGet("kubesphere:users:avatar", username).Result()

	if err != nil {
		return nil, err
	}

	if len(avatar) > 0 {
		if url, ok := avatar[0].(string); ok {
			user.AvatarUrl = url
		}
	}

	user.Status = 0

	lastLogin, err := redisClient.LRange(fmt.Sprintf("kubesphere:users:%s:login-log", username), -1, -1).Result()

	if err != nil {
		return nil, err
	}

	if len(lastLogin) > 0 {
		user.LastLoginTime = strings.Split(lastLogin[0], ",")[0]
	}

	return &user, nil
}

func DeleteUser(username string) error {

	// bind root DN
	conn, err := NewConnection()
	if err != nil {
		return err
	}

	defer conn.Close()

	deleteRequest := ldap.NewDelRequest(fmt.Sprintf("uid=%s,%s", username, constants.UserSearchBase), nil)

	err = conn.Del(deleteRequest)

	if err != nil {
		return err
	}

	err = deleteRoleBindings(username)

	if err != nil {
		return err
	}

	counter.Sub(1)

	return nil
}

func deleteRoleBindings(username string) error {

	roleBindings, err := roleBindingLister.List(labels.Everything())

	if err != nil {
		return err
	}

	for _, roleBinding := range roleBindings {

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
			err = client.K8sClient().RbacV1().RoleBindings(roleBinding.Namespace).Delete(roleBinding.Name, &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})

			if err != nil {
				glog.Errorf("delete role binding %s %s failed:%s", username, roleBinding.Namespace, roleBinding.Name, err)
			}
		} else if length2 < length1 {
			_, err = client.K8sClient().RbacV1().RoleBindings(roleBinding.Namespace).Update(roleBinding)

			if err != nil {
				glog.Errorf("update role binding %s %s failed:%s", username, roleBinding.Namespace, roleBinding.Name, err)
			}
		}
	}

	clusterRoleBindings, err := clusterRoleBindingLister.List(labels.Everything())

	for _, clusterRoleBinding := range clusterRoleBindings {
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
				_, err = client.K8sClient().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
			} else {
				deletePolicy := meta_v1.DeletePropagationForeground
				err = client.K8sClient().RbacV1().ClusterRoleBindings().Delete(clusterRoleBinding.Name, &meta_v1.DeleteOptions{PropagationPolicy: &deletePolicy})
			}
			if err != nil {
				glog.Errorf("update cluster role binding %s failed:%s", clusterRoleBinding.Name, err)
			}
		} else if length2 < length1 {
			_, err = client.K8sClient().RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)

			if err != nil {
				glog.Errorf("update cluster role binding %s failed:%s", clusterRoleBinding.Name, err)
			}
		}

	}

	return nil
}

func UserCreateCheck(check string) (exist bool, err error) {

	// bind root DN
	conn, err := NewConnection()

	if err != nil {
		return false, err
	}

	defer conn.Close()

	// search for the given username
	userSearchRequest := ldap.NewSearchRequest(
		constants.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", check, check),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		return false, err
	}

	if len(result.Entries) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func CreateUser(user User) error {
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.Password = strings.TrimSpace(user.Password)
	user.Description = strings.TrimSpace(user.Description)

	conn, err := NewConnection()

	if err != nil {
		return err
	}

	defer conn.Close()

	userSearchRequest := ldap.NewSearchRequest(
		constants.UserSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", user.Username, user.Email),
		[]string{"uid", "mail"},
		nil,
	)

	result, err := conn.Search(userSearchRequest)

	if err != nil {
		return err
	}

	if len(result.Entries) > 0 {
		return errors.New("username or email already exists")
	}

	maxUid, err := getMaxUid(conn)

	if err != nil {
		return err
	}

	maxUid += 1

	userCreateRequest := ldap.NewAddRequest(fmt.Sprintf("uid=%s,%s", user.Username, constants.UserSearchBase), nil)
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
		userCreateRequest.Attribute("preferredLanguage", []string{user.Lang}) // RFC4519/2307: password of user
	}
	if user.Description != "" {
		userCreateRequest.Attribute("description", []string{user.Description}) // RFC4519: descriptive information
	}

	err = conn.Add(userCreateRequest)

	if err != nil {
		return err
	}

	counter.Add(1)

	if user.ClusterRole != "" {
		CreateClusterRoleBinding(user.Username, user.ClusterRole)
	}

	return nil
}

func getMaxUid(conn ldap.Client) (int, error) {
	userSearchRequest := ldap.NewSearchRequest(constants.UserSearchBase,
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

	groupSearchRequest := ldap.NewSearchRequest(constants.GroupSearchBase,
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

func UpdateUser(user User) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}

	defer conn.Close()

	dn := fmt.Sprintf("uid=%s,%s", user.Username, constants.UserSearchBase)
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

	err = conn.Modify(userModifyRequest)

	if err != nil {
		return err
	}

	err = CreateClusterRoleBinding(user.Username, user.ClusterRole)

	if err != nil {
		return err
	}

	return nil
}
func DeleteGroup(path string) error {

	// bind root DN
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	searchBase, cn := splitPath(path)

	groupDeleteRequest := ldap.NewDelRequest(fmt.Sprintf("cn=%s,%s", cn, searchBase), nil)
	err = conn.Del(groupDeleteRequest)

	if err != nil {
		return err
	}

	return nil
}

func CreateGroup(group Group) (*Group, error) {

	// bind root DN
	conn, err := NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	maxGid, err := getMaxGid(conn)

	if err != nil {
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

	groupCreateRequest.Attribute("memberUid", []string{group.Creator})

	err = conn.Add(groupCreateRequest)

	if err != nil {
		return nil, err
	}

	group.Gid = strconv.Itoa(maxGid)

	group.CreateTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	redisClient := client.RedisClient()

	if err := redisClient.HMSet("kubesphere:groups:create-time", map[string]interface{}{group.Name: group.CreateTime}).Err(); err != nil {
		return nil, err
	}
	if err := redisClient.HMSet("kubesphere:groups:creator", map[string]interface{}{group.Name: group.Creator}).Err(); err != nil {
		return nil, err
	}

	return &group, nil
}

func UpdateGroup(group *Group) (*Group, error) {

	// bind root DN
	conn, err := NewConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	old, err := GroupDetail(group.Path, conn)

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
		return nil, err
	}

	return group, nil
}

func CountChild(path string) (int, error) {
	// bind root DN
	conn, err := NewConnection()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var groupSearchRequest *ldap.SearchRequest
	if path == "" {
		groupSearchRequest = ldap.NewSearchRequest(constants.GroupSearchBase,
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
		return 0, err
	}

	return len(result.Entries), nil
}

func ChildList(path string) (*[]Group, error) {

	// bind root DN
	conn, err := NewConnection()

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var groupSearchRequest *ldap.SearchRequest
	if path == "" {
		groupSearchRequest = ldap.NewSearchRequest(constants.GroupSearchBase,
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

	groups := make([]Group, 0)

	for _, v := range result.Entries {
		dn := v.DN
		cn := v.GetAttributeValue("cn")
		gid := v.GetAttributeValue("gidNumber")
		members := v.GetAttributeValues("memberUid")
		description := v.GetAttributeValue("description")

		group := Group{Path: convertDNToPath(dn), Name: cn, Gid: gid, Members: members, Description: description}

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

		redisClient := client.RedisClient()

		createTime, _ := redisClient.HMGet("kubesphere:groups:create-time", group.Name).Result()

		if len(createTime) > 0 {
			if t, ok := createTime[0].(string); ok {
				group.CreateTime = t
			}
		}

		creator, _ := redisClient.HMGet("kubesphere:groups:creator", group.Name).Result()

		if len(creator) > 0 {
			if t, ok := creator[0].(string); ok {
				group.Creator = t
			}
		}

		groups = append(groups, group)
	}

	return &groups, nil
}

func GroupDetail(path string, conn ldap.Client) (*Group, error) {

	searchBase, cn := splitPath(path)

	groupSearchRequest := ldap.NewSearchRequest(searchBase,
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=posixGroup)(cn=%s))", cn),
		[]string{"cn", "gidNumber", "memberUid", "description"},
		nil)

	result, err := conn.Search(groupSearchRequest)

	if err != nil {
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

	group := Group{Path: convertDNToPath(dn), Name: cn, Gid: gid, Members: members, Description: description}

	//childSearchRequest := ldap.NewSearchRequest(dn,
	//	ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
	//	"(&(objectClass=posixGroup))",
	//	[]string{""},
	//	nil, )

	//result, err = conn.Search(childSearchRequest)

	//if err != nil {
	//	return nil, err
	//}

	childGroups := make([]string, 0)
	//for _, v := range result.Entries {
	//	child := convertDNToPath(v.DN)
	//	childGroups = append(childGroups, child)
	//}

	group.ChildGroups = childGroups

	redisClient := client.RedisClient()

	createTime, _ := redisClient.HMGet("kubesphere:groups:create-time", group.Name).Result()

	if len(createTime) > 0 {
		if t, ok := createTime[0].(string); ok {
			group.CreateTime = t
		}
	}

	creator, _ := redisClient.HMGet("kubesphere:groups:creator", group.Name).Result()

	if len(creator) > 0 {
		if t, ok := creator[0].(string); ok {
			group.Creator = t
		}
	}

	return &group, nil

}
