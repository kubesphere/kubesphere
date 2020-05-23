/*
Copyright (C) 2020 The KubeSphere Authors.

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
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

var (
	ErrorIdentityProviderNotFound = errors.New("the identity provider was not found")
	ErrorAlreadyRegistered        = errors.New("the identity provider was not found")
	oauthProviderCodecs           = map[string]OAuthProviderCodec{}
)

type OAuthProvider interface {
	IdentityExchange(code string) (user.Info, error)
}

type OAuthProviderCodec interface {
	Type() string
	Decode(options *oauth.DynamicOptions) (OAuthProvider, error)
	Encode(provider OAuthProvider) (*oauth.DynamicOptions, error)
}

func ResolveOAuthProvider(providerType string, options *oauth.DynamicOptions) (OAuthProvider, error) {
	if codec, ok := oauthProviderCodecs[providerType]; ok {
		return codec.Decode(options)
	}
	return nil, ErrorIdentityProviderNotFound
}

func ResolveOAuthOptions(providerType string, provider OAuthProvider) (*oauth.DynamicOptions, error) {
	if codec, ok := oauthProviderCodecs[providerType]; ok {
		return codec.Encode(provider)
	}
	return nil, ErrorIdentityProviderNotFound
}

func RegisterOAuthProviderCodec(codec OAuthProviderCodec) error {
	if _, ok := oauthProviderCodecs[codec.Type()]; ok {
		return ErrorAlreadyRegistered
	}
	oauthProviderCodecs[codec.Type()] = codec
	return nil
}
