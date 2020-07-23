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

package identityprovider

import (
	"errors"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

var (
	oauthProviders                = make(map[string]OAuthProvider, 0)
	ErrorIdentityProviderNotFound = errors.New("the identity provider was not found")
)

type OAuthProvider interface {
	Type() string
	Setup(options *oauth.DynamicOptions) (OAuthProvider, error)
	IdentityExchange(code string) (Identity, error)
}

func GetOAuthProvider(providerType string, options *oauth.DynamicOptions) (OAuthProvider, error) {
	if provider, ok := oauthProviders[providerType]; ok {
		return provider.Setup(options)
	}
	return nil, ErrorIdentityProviderNotFound
}

func RegisterOAuthProvider(provider OAuthProvider) {
	oauthProviders[provider.Type()] = provider
}
