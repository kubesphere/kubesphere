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
	*userGetter
	options *authentication.Options
}

func NewOAuthAuthenticator(userLister iamv1alpha2listers.UserLister,
	options *authentication.Options) OAuthAuthenticator {
	authenticator := &oauthAuthenticator{
		userGetter: &userGetter{userLister: userLister},
		options:    options,
	}
	return authenticator
}

func (o oauthAuthenticator) Authenticate(ctx context.Context, provider string, req *http.Request) (authuser.Info, string, error) {
	options, err := o.options.OAuthOptions.IdentityProviderOptions(provider)
	// identity provider not registered
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}
	identityProvider, err := identityprovider.GetOAuthProvider(options.Name)
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}
	identity, err := identityProvider.IdentityExchangeCallback(req)
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}

	mappedUser, err := o.findMappedUser(options.Name, identity.GetUserID())
	if mappedUser == nil && options.MappingMethod == oauth.MappingMethodLookup {
		klog.Error(err)
		return nil, "", err
	}
	// the user will automatically create and mapping when login successful.
	if mappedUser == nil && options.MappingMethod == oauth.MappingMethodAuto {
		return preRegistrationUser(options.Name, identity), options.Name, nil
	}
	if mappedUser != nil {
		return &authuser.DefaultInfo{Name: mappedUser.GetName()}, options.Name, nil
	}

	return nil, "", errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularUser), identity.GetUsername())
}
