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
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/iam"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/utils/jwtutil"
	"time"
)

type IdentityManagementInterface interface {
	CreateUser(user *iam.User) (*iam.User, error)
	DeleteUser(username string) error
	DescribeUser(username string) (*iam.User, error)
	Login(username, password, ip string) (*oauth2.Token, error)
	ModifyUser(user *iam.User) (*iam.User, error)
	ListUsers(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	GetUserRoles(username string) ([]*rbacv1.Role, error)
	GetUserRole(namespace string, username string) (*rbacv1.Role, error)
}

type imOperator struct {
	authenticateOptions *iam.AuthenticationOptions
	ldapClient          ldap.Interface
	cacheClient         cache.Interface
}

var (
	AuthRateLimitExceeded = errors.New("user auth rate limit exceeded")
	UserAlreadyExists     = errors.New("user already exists")
	UserNotExists         = errors.New("user not exists")
)

func NewIMOperator(ldapClient ldap.Interface, cacheClient cache.Interface, options *iam.AuthenticationOptions) *imOperator {
	return &imOperator{ldapClient: ldapClient, cacheClient: cacheClient, authenticateOptions: options}

}

func (im *imOperator) ModifyUser(user *iam.User) (*iam.User, error) {
	err := im.ldapClient.Update(user)
	if err != nil {
		return nil, err
	}

	// clear auth failed record
	if user.Password != "" {
		records, err := im.cacheClient.Keys(authenticationFailedKeyForUsername(user.Username, "*"))
		if err == nil {
			im.cacheClient.Del(records...)
		}
	}

	return im.ldapClient.Get(user.Username)
}

func authenticationFailedKeyForUsername(username, failedTimestamp string) string {
	return fmt.Sprintf("kubesphere:authfailed:%s:%s", username, failedTimestamp)
}

func tokenKeyForUsername(username, token string) string {
	return fmt.Sprintf("kubesphere:users:%s:token:%s", username, token)
}

func loginKeyForUsername(username, loginTimestamp, ip string) string {
	return fmt.Sprintf("kubesphere:users:%s:login-log:%s:%s", username, loginTimestamp, ip)
}

func (im *imOperator) Login(username, password, ip string) (*oauth2.Token, error) {

	records, err := im.cacheClient.Keys(authenticationFailedKeyForUsername(username, "*"))
	if err != nil {
		return nil, err
	}

	if len(records) >= im.authenticateOptions.MaxAuthenticateRetries {
		return nil, AuthRateLimitExceeded
	}

	user, err := im.ldapClient.Get(username)
	if err != nil {
		return nil, err
	}

	err = im.ldapClient.Verify(user.Username, password)
	if err != nil {
		if err == ldap.ErrInvalidCredentials {
			im.cacheClient.Set(authenticationFailedKeyForUsername(username, fmt.Sprintf("%d", time.Now().UnixNano())), "", 30*time.Minute)
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

	tokenKey := tokenKeyForUsername(user.Username, "*")
	if !im.authenticateOptions.MultipleLogin {
		// multi login not allowed, remove the previous token
		sessions, err := im.cacheClient.Keys(tokenKey)
		if err != nil {
			return nil, err
		}

		if len(sessions) > 0 {
			klog.V(4).Infoln("revoke token", sessions)
			err = im.cacheClient.Del(sessions...)
			if err != nil {
				return nil, err
			}
		}
	}

	// cache token with expiration time
	if err = im.cacheClient.Set(tokenKey, token, im.authenticateOptions.TokenExpiration); err != nil {
		return nil, err
	}

	im.loginRecord(user.Username, ip, loginTime)

	return &oauth2.Token{AccessToken: token}, nil
}

func (im *imOperator) loginRecord(username, ip string, loginTime time.Time) {
	if ip != "" {
		_ = im.cacheClient.Set(loginKeyForUsername(username, loginTime.UTC().Format("2006-01-02T15:04:05Z"), ip), "", 30*24*time.Hour)
	}
}

func (im *imOperator) LoginHistory(username string) ([]string, error) {
	keys, err := im.cacheClient.Keys(loginKeyForUsername(username, "*", "*"))
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (im *imOperator) ListUsers(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	panic("implement me")
}

func (im *imOperator) DescribeUser(username string) (*iam.User, error) {
	return im.ldapClient.Get(username)
}

func (im *imOperator) getLastLoginTime(username string) string {

	return ""
}

func (im *imOperator) DeleteUser(username string) error {
	return im.ldapClient.Delete(username)
}

func (im *imOperator) CreateUser(user *iam.User) (*iam.User, error) {
	err := im.ldapClient.Create(user)
	if err != nil {
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
