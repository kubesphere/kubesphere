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
	"fmt"
	"net/http"

	authuser "k8s.io/apiserver/pkg/authentication/user"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
)

type oauthAuthenticator struct {
	client                 runtimeclient.Client
	userGetter             *userMapper
	idpConfigurationGetter identityprovider.ConfigurationGetter
}

func NewOAuthAuthenticator(cacheClient runtimeclient.Client) OAuthAuthenticator {
	authenticator := &oauthAuthenticator{
		client:                 cacheClient,
		userGetter:             &userMapper{cache: cacheClient},
		idpConfigurationGetter: identityprovider.NewConfigurationGetter(cacheClient),
	}
	return authenticator
}

func (o *oauthAuthenticator) Authenticate(ctx context.Context, provider string, req *http.Request) (authuser.Info, error) {
	providerConfig, err := o.idpConfigurationGetter.GetConfiguration(ctx, provider)
	// identity provider not registered
	if err != nil {
		return nil, fmt.Errorf("failed to get identity provider configuration for %s, error: %v", provider, err)
	}

	oauthIdentityProvider, exist := identityprovider.SharedIdentityProviderController.GetOAuthProvider(provider)
	if !exist {
		return nil, fmt.Errorf("identity provider %s not exist", provider)
	}

	identity, err := oauthIdentityProvider.IdentityExchangeCallback(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange identity for %s, error: %v", provider, err)
	}

	mappedUser, err := o.userGetter.FindMappedUser(ctx, providerConfig.Name, identity.GetUserID())
	if err != nil {
		return nil, fmt.Errorf("failed to find mapped user for %s, error: %v", provider, err)
	}

	if mappedUser == nil {
		if providerConfig.MappingMethod == identityprovider.MappingMethodLookup {
			return nil, fmt.Errorf("failed to find mapped user: %s", identity.GetUserID())
		}

		if providerConfig.MappingMethod == identityprovider.MappingMethodManual {
			return newRreRegistrationUser(providerConfig.Name, identity), nil
		}

		if providerConfig.MappingMethod == identityprovider.MappingMethodAuto {
			mappedUser = newMappedUser(providerConfig.Name, identity)

			if err = o.client.Create(ctx, mappedUser); err != nil {
				return nil, err
			}

			return &authuser.DefaultInfo{Name: mappedUser.GetName()}, nil
		}

		return nil, fmt.Errorf("invalid mapping method found %s", providerConfig.MappingMethod)
	}

	if mappedUser.Status.State == iamv1beta1.UserDisabled {
		return nil, AccountIsNotActiveError
	}

	return &authuser.DefaultInfo{Name: mappedUser.GetName()}, nil
}
