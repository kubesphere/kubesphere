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
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-ldap/ldap"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	ldappool "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/utils/jwtutil"
	"strconv"
	"strings"
	"time"
)

type IdentityManagementInterface interface {
	CreateUser(user *User) (*User, error)
	DeleteUser(username string) error
	DescribeUser(username string) (*User, error)
	Login(username, password, ip string) (*oauth2.Token, error)
	ModifyUser(user *User) (*User, error)
	ListUsers(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	GetUserRoles(username string) ([]*rbacv1.Role, error)
	GetUserRole(namespace string, username string) (*rbacv1.Role, error)
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

var (
	AuthRateLimitExceeded = errors.New("user auth rate limit exceeded")
	UserAlreadyExists     = errors.New("user already exists")
	UserNotExists         = errors.New("user not exists")
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

func (im *imOperator) ModifyUser(user *User) (*User, error) {
	conn, err := im.ldap.NewConn()

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	defer conn.Close()

	dn := fmt.Sprintf("uid=%s,%s", user.Username, im.ldap.UserSearchBase())
	userModifyRequest := ldap.NewModifyRequest(dn, nil)

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

	// clear auth failed record
	if user.Password != "" {

		records, err := im.redis.Keys(fmt.Sprintf("kubesphere:authfailed:%s:*", user.Username)).Result()

		if err == nil {
			im.redis.Del(records...)
		}
	}

	return im.DescribeUser(user.Username)
}

func (im *imOperator) Login(username, password, ip string) (*oauth2.Token, error) {

	records, err := im.redis.Keys(fmt.Sprintf("kubesphere:authfailed:%s:*", username)).Result()

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(records) >= im.config.maxAuthFailed {
		return nil, AuthRateLimitExceeded
	}

	user, err := im.DescribeUser(username)

	conn, err := im.ldap.NewConn()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer conn.Close()

	dn := fmt.Sprintf("uid=%s,%s", user.Username, im.ldap.UserSearchBase())

	// bind as the user to verify their password
	err = conn.Bind(dn, password)

	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			cacheKey := fmt.Sprintf("kubesphere:authfailed:%s:%d", user.Username, time.Now().UnixNano())
			im.redis.Set(cacheKey, "", im.config.authTimeInterval)
		}
		return nil, err
	}

	loginTime := time.Now()
	// token without expiration time will auto sliding
	claims := jwt.MapClaims{
		"iat":      loginTime.Unix(),
		"username": user.Username,
		"email":    user.Email,
	}
	token := jwtutil.MustSigned(claims)

	if !im.config.enableMultiLogin {
		// multi login not allowed, remove the previous token
		cacheKey := fmt.Sprintf("kubesphere:users:%s:token:*", user.Username)
		sessions, err := im.redis.Keys(cacheKey).Result()

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
	cacheKey := fmt.Sprintf("kubesphere:users:%s:token:%s", user.Username, token)
	if err = im.redis.Set(cacheKey, token, im.config.tokenIdleTimeout).Err(); err != nil {
		klog.Errorln(err)
		return nil, err
	}

	im.loginRecord(user.Username, ip, loginTime)

	return &oauth2.Token{AccessToken: token}, nil
}

func (im *imOperator) loginRecord(username, ip string, loginTime time.Time) {
	if ip != "" {
		im.redis.RPush(fmt.Sprintf("kubesphere:users:%s:login-log", username), fmt.Sprintf("%s,%s", loginTime.UTC().Format("2006-01-02T15:04:05Z"), ip))
		im.redis.LTrim(fmt.Sprintf("kubesphere:users:%s:login-log", username), -10, -1)
	}
}

func (im *imOperator) LoginHistory(username string) ([]string, error) {
	data, err := im.redis.LRange(fmt.Sprintf("kubesphere:users:%s:login-log", username), -10, -1).Result()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (im *imOperator) ListUsers(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	panic("implement me")
}

func (im *imOperator) DescribeUser(username string) (*User, error) {
	conn, err := im.ldap.NewConn()

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	defer conn.Close()

	filter := fmt.Sprintf("(&(objectClass=inetOrgPerson)(|(uid=%s)(mail=%s)))", username, username)

	searchRequest := ldap.NewSearchRequest(
		im.ldap.UserSearchBase(),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{mailAttribute, descriptionAttribute, preferredLanguageAttribute, createTimestampAttribute},
		nil,
	)

	result, err := conn.Search(searchRequest)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	if len(result.Entries) != 1 {
		return nil, UserNotExists
	}

	entry := result.Entries[0]

	return convertLdapEntryToUser(entry), nil
}

func convertLdapEntryToUser(entry *ldap.Entry) *User {
	username := entry.GetAttributeValue(uidAttribute)
	email := entry.GetAttributeValue(mailAttribute)
	description := entry.GetAttributeValue(descriptionAttribute)
	lang := entry.GetAttributeValue(preferredLanguageAttribute)
	createTimestamp, err := time.Parse(dateTimeLayout, entry.GetAttributeValue(createTimestampAttribute))
	if err != nil {
		klog.Errorln(err)
	}
	return &User{Username: username, Email: email, Description: description, Lang: lang, CreateTime: createTimestamp}
}

func (im *imOperator) getLastLoginTime(username string) string {
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
		klog.Errorln(err)
		return err
	}

	defer conn.Close()

	deleteRequest := ldap.NewDelRequest(fmt.Sprintf("uid=%s,%s", username, im.ldap.UserSearchBase()), nil)

	if err = conn.Del(deleteRequest); err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (im *imOperator) CreateUser(user *User) (*User, error) {
	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.Password = strings.TrimSpace(user.Password)
	user.Description = strings.TrimSpace(user.Description)

	existed, err := im.DescribeUser(user.Username)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	if existed != nil {
		return nil, UserAlreadyExists
	}

	uidNumber := im.uidNumberNext()

	createRequest := ldap.NewAddRequest(fmt.Sprintf("uid=%s,%s", user.Username, im.ldap.UserSearchBase()), nil)
	createRequest.Attribute("objectClass", []string{"inetOrgPerson", "posixAccount", "top"})
	createRequest.Attribute("cn", []string{user.Username})                       // RFC4519: common name(s) for which the entity is known by
	createRequest.Attribute("sn", []string{" "})                                 // RFC2256: last (family) name(s) for which the entity is known by
	createRequest.Attribute("gidNumber", []string{"500"})                        // RFC2307: An integer uniquely identifying a group in an administrative domain
	createRequest.Attribute("homeDirectory", []string{"/home/" + user.Username}) // The absolute path to the home directory
	createRequest.Attribute("uid", []string{user.Username})                      // RFC4519: user identifier
	createRequest.Attribute("uidNumber", []string{strconv.Itoa(uidNumber)})      // RFC2307: An integer uniquely identifying a user in an administrative domain
	createRequest.Attribute("mail", []string{user.Email})                        // RFC1274: RFC822 Mailbox
	createRequest.Attribute("userPassword", []string{user.Password})             // RFC4519/2307: password of user
	if user.Lang != "" {
		createRequest.Attribute("preferredLanguage", []string{user.Lang})
	}
	if user.Description != "" {
		createRequest.Attribute("description", []string{user.Description}) // RFC4519: descriptive information
	}

	conn, err := im.ldap.NewConn()

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	err = conn.Add(createRequest)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	return user, nil
}

func (im *imOperator) uidNumberNext() int {
	// TODO fix me
	return 0
}
func (im *imOperator) GetUserRoles(username string) ([]*rbacv1.Role, error) {
	panic("implement me")
}

func (im *imOperator) GetUserRole(namespace string, username string) (*rbacv1.Role, error) {
	panic("implement me")
}
