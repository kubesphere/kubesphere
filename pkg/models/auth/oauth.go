/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auth

import (
	"context"
	"fmt"
	"net/http"

	authuser "k8s.io/apiserver/pkg/authentication/user"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
)

type oauthAuthenticator struct {
	client                 runtimeclient.Client
	userMapper             UserMapper
	idpConfigurationGetter identityprovider.ConfigurationGetter
}

func NewOAuthAuthenticator(client runtimeclient.Client) OAuthAuthenticator {
	authenticator := &oauthAuthenticator{
		client:                 client,
		userMapper:             &userMapper{cache: client},
		idpConfigurationGetter: identityprovider.NewConfigurationGetter(client),
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

	return authByIdentityProvider(ctx, o.client, o.userMapper, providerConfig, identity)
}
