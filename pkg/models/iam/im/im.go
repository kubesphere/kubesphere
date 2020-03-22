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
package im

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/auth"
	"kubesphere.io/kubesphere/pkg/api/auth/token"
	"kubesphere.io/kubesphere/pkg/api/iam"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"time"
)

type IdentityManagementInterface interface {
	CreateUser(user *iam.User) (*iam.User, error)
	DeleteUser(username string) error
	ModifyUser(user *iam.User) (*iam.User, error)
	DescribeUser(username string) (*iam.User, error)
	Login(username, password, ip string) (*oauth2.Token, error)
	ListUsers(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	GetUserRoles(username string) ([]*rbacv1.Role, error)
	GetUserRole(namespace string, username string) (*rbacv1.Role, error)
}

type imOperator struct {
	authenticateOptions *auth.AuthenticationOptions
	ldapClient          ldap.Interface
	cacheClient         cache.Interface
	issuer              token.Issuer
}

var (
	AuthRateLimitExceeded = errors.New("user auth rate limit exceeded")
	UserAlreadyExists     = errors.New("user already exists")
	UserNotExists         = errors.New("user not exists")
)

func NewIMOperator(ldapClient ldap.Interface, cacheClient cache.Interface, options *auth.AuthenticationOptions) *imOperator {
	return &imOperator{
		ldapClient:          ldapClient,
		cacheClient:         cacheClient,
		authenticateOptions: options,
		issuer:              token.NewJwtTokenIssuer(token.DefaultIssuerName, []byte(options.JwtSecret)),
	}

}

func (im *imOperator) ModifyUser(user *iam.User) (*iam.User, error) {
	err := im.ldapClient.Update(user)
	if err != nil {
		return nil, err
	}

	// clear auth failed record
	if user.Password != "" {
		records, err := im.cacheClient.Keys(authenticationFailedKeyForUsername(user.Name, "*"))
		if err == nil {
			im.cacheClient.Del(records...)
		}
	}

	return im.ldapClient.Get(user.Name)
}

func (im *imOperator) Login(username, password, ip string) (*oauth2.Token, error) {

	records, err := im.cacheClient.Keys(authenticationFailedKeyForUsername(username, "*"))
	if err != nil {
		return nil, err
	}

	if len(records) > im.authenticateOptions.MaxAuthenticateRetries {
		return nil, AuthRateLimitExceeded
	}

	user, err := im.ldapClient.Get(username)
	if err != nil {
		return nil, err
	}

	err = im.ldapClient.Verify(user.Name, password)
	if err != nil {
		if err == ldap.ErrInvalidCredentials {
			im.cacheClient.Set(authenticationFailedKeyForUsername(username, fmt.Sprintf("%d", time.Now().UnixNano())), "", 30*time.Minute)
		}
		return nil, err
	}

	issuedToken, err := im.issuer.IssueTo(user)
	if err != nil {
		return nil, err
	}

	// TODO: I think we should come up with a better strategy to prevent multiple login.
	tokenKey := tokenKeyForUsername(user.Name, issuedToken)
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

	// save token with expiration time
	if err = im.cacheClient.Set(tokenKey, issuedToken, im.authenticateOptions.TokenExpiration); err != nil {
		return nil, err
	}

	im.logLogin(user.Name, ip, time.Now())

	return &oauth2.Token{AccessToken: issuedToken}, nil
}

func (im *imOperator) logLogin(username, ip string, loginTime time.Time) {
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

func (im *imOperator) VerifyToken(tokenString string) (*iam.User, error) {
	providedUser, err := im.issuer.Verify(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := im.ldapClient.Get(providedUser.GetName())
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

func authenticationFailedKeyForUsername(username, failedTimestamp string) string {
	return fmt.Sprintf("kubesphere:authfailed:%s:%s", username, failedTimestamp)
}

func tokenKeyForUsername(username, token string) string {
	return fmt.Sprintf("kubesphere:users:%s:token:%s", username, token)
}

func loginKeyForUsername(username, loginTimestamp, ip string) string {
	return fmt.Sprintf("kubesphere:users:%s:login-log:%s:%s", username, loginTimestamp, ip)
}
