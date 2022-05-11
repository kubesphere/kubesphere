/*

 Copyright 2021 The KubeSphere Authors.

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

package auth

import (
	"context"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"

	"k8s.io/apimachinery/pkg/api/errors"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
)

type oauthAuthenticator struct {
	ksClient   kubesphere.Interface
	userGetter *userGetter
	options    *authentication.Options
}

func NewOAuthAuthenticator(ksClient kubesphere.Interface,
	userLister iamv1alpha2listers.UserLister,
	options *authentication.Options) OAuthAuthenticator {
	authenticator := &oauthAuthenticator{
		ksClient:   ksClient,
		userGetter: &userGetter{userLister: userLister},
		options:    options,
	}
	return authenticator
}

func (o *oauthAuthenticator) Authenticate(_ context.Context, provider string, req *http.Request) (authuser.Info, string, error) {
	providerOptions, err := o.options.OAuthOptions.IdentityProviderOptions(provider)
	// identity provider not registered
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}
	oauthIdentityProvider, err := identityprovider.GetOAuthProvider(providerOptions.Name)
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}
	authenticated, err := oauthIdentityProvider.IdentityExchangeCallback(req)
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}

	user, err := o.userGetter.findMappedUser(providerOptions.Name, authenticated.GetUserID())
	if user == nil && providerOptions.MappingMethod == oauth.MappingMethodLookup {
		klog.Error(err)
		return nil, "", err
	}

	// the user will automatically create and mapping when login successful.
	if user == nil && providerOptions.MappingMethod == oauth.MappingMethodAuto {
		if !providerOptions.DisableLoginConfirmation {
			return preRegistrationUser(providerOptions.Name, authenticated), providerOptions.Name, nil
		}
		user, err = o.ksClient.IamV1alpha2().Users().Create(context.Background(), mappedUser(providerOptions.Name, authenticated), metav1.CreateOptions{})
		if err != nil {
			return nil, providerOptions.Name, err
		}
	}

	if user != nil {
		if user.Status.State == iamv1alpha2.UserDisabled {
			// state not active
			return nil, "", AccountIsNotActiveError
		}
		return &authuser.DefaultInfo{Name: user.GetName()}, providerOptions.Name, nil
	}

	return nil, "", errors.NewNotFound(iamv1alpha2.Resource("user"), authenticated.GetUsername())
}
