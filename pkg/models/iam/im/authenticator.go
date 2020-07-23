/*

 Copyright 2020 The KubeSphere Authors.

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

package im

import (
	"fmt"
	"github.com/go-ldap/ldap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	"net/mail"
)

var (
	AuthRateLimitExceeded             = fmt.Errorf("auth rate limit exceeded")
	AuthFailedIncorrectPassword       = fmt.Errorf("incorrect password")
	AuthFailedAccountIsNotActive      = fmt.Errorf("account is not active")
	AuthFailedIdentityMappingNotMatch = fmt.Errorf("identity mapping not match")
)

type PasswordAuthenticator interface {
	Authenticate(username, password string) (authuser.Info, error)
}

type passwordAuthenticator struct {
	ksClient   kubesphere.Interface
	userLister iamv1alpha2listers.UserLister
	options    *authoptions.AuthenticationOptions
}

func NewPasswordAuthenticator(ksClient kubesphere.Interface,
	userLister iamv1alpha2listers.UserLister,
	options *authoptions.AuthenticationOptions) PasswordAuthenticator {
	return &passwordAuthenticator{
		ksClient:   ksClient,
		userLister: userLister,
		options:    options}
}

func (im *passwordAuthenticator) Authenticate(username, password string) (authuser.Info, error) {

	user, err := im.searchUser(username)
	if err != nil {
		// internal error
		if !errors.IsNotFound(err) {
			klog.Error(err)
			return nil, err
		}
	}

	providerOptions, ldapProvider := im.getLdapProvider()

	// no identity provider
	// even auth failed, still return username to record login attempt
	if user == nil && (providerOptions == nil || providerOptions.MappingMethod != oauth.MappingMethodAuto) {
		return nil, AuthFailedIncorrectPassword
	}

	if user != nil && user.Status.State != iamv1alpha2.UserActive {
		if user.Status.State == iamv1alpha2.UserAuthLimitExceeded {
			klog.Errorf("%s, username: %s", AuthRateLimitExceeded, username)
			return nil, AuthRateLimitExceeded
		} else {
			klog.Errorf("%s, username: %s", AuthFailedAccountIsNotActive, username)
			return nil, AuthFailedAccountIsNotActive
		}
	}

	// able to login using the locally principal admin account and password in case of a disruption of LDAP services.
	if ldapProvider != nil && username != constants.AdminUserName {
		if providerOptions.MappingMethod == oauth.MappingMethodLookup &&
			(user == nil || user.Labels[iamv1alpha2.IdentifyProviderLabel] != providerOptions.Name) {
			klog.Error(AuthFailedIdentityMappingNotMatch)
			return nil, AuthFailedIdentityMappingNotMatch
		}
		if providerOptions.MappingMethod == oauth.MappingMethodAuto &&
			user != nil && user.Labels[iamv1alpha2.IdentifyProviderLabel] != providerOptions.Name {
			klog.Error(AuthFailedIdentityMappingNotMatch)
			return nil, AuthFailedIdentityMappingNotMatch
		}

		authenticated, err := ldapProvider.Authenticate(username, password)
		if err != nil {
			klog.Error(err)
			if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) || ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
				return nil, AuthFailedIncorrectPassword
			} else {
				return nil, err
			}
		}

		if authenticated != nil && user == nil {
			authenticated.Labels = map[string]string{iamv1alpha2.IdentifyProviderLabel: providerOptions.Name}
			if authenticated, err = im.ksClient.IamV1alpha2().Users().Create(authenticated); err != nil {
				klog.Error(err)
				return nil, err
			}
		}

		if authenticated != nil {
			return &authuser.DefaultInfo{
				Name: authenticated.Name,
				UID:  string(authenticated.UID),
			}, nil
		}
	}

	if checkPasswordHash(password, user.Spec.EncryptedPassword) {
		return &authuser.DefaultInfo{
			Name: user.Name,
			UID:  string(user.UID),
		}, nil
	}

	return nil, AuthFailedIncorrectPassword
}

func (im *passwordAuthenticator) searchUser(username string) (*iamv1alpha2.User, error) {

	if _, err := mail.ParseAddress(username); err != nil {
		return im.userLister.Get(username)
	} else {
		users, err := im.userLister.List(labels.Everything())
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		for _, find := range users {
			if find.Spec.Email == username {
				return find, nil
			}
		}
	}

	return nil, errors.NewNotFound(iamv1alpha2.Resource("user"), username)
}

func (im *passwordAuthenticator) getLdapProvider() (*oauth.IdentityProviderOptions, identityprovider.LdapProvider) {
	for _, options := range im.options.OAuthOptions.IdentityProviders {
		if options.Type == identityprovider.LdapIdentityProvider {
			if provider, err := identityprovider.NewLdapProvider(options.Provider); err != nil {
				klog.Error(err)
			} else {
				return &options, provider
			}
		}
	}
	return nil, nil
}
