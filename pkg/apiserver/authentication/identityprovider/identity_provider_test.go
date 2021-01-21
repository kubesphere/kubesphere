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
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"testing"
)

type emptyOAuthProviderFactory struct {
	typeName string
}

func (e emptyOAuthProviderFactory) Type() string {
	return e.typeName
}

type emptyOAuthProvider struct {
}

type emptyIdentity struct {
}

func (e emptyIdentity) GetUserID() string {
	return "test"
}

func (e emptyIdentity) GetUsername() string {
	return "test"
}

func (e emptyIdentity) GetEmail() string {
	return "test@test.com"
}

func (e emptyOAuthProvider) IdentityExchange(code string) (Identity, error) {
	return emptyIdentity{}, nil
}

func (e emptyOAuthProviderFactory) Create(options oauth.DynamicOptions) (OAuthProvider, error) {
	return emptyOAuthProvider{}, nil
}

type emptyGenericProviderFactory struct {
	typeName string
}

func (e emptyGenericProviderFactory) Type() string {
	return e.typeName
}

type emptyGenericProvider struct {
}

func (e emptyGenericProvider) Authenticate(username string, password string) (Identity, error) {
	return emptyIdentity{}, nil
}

func (e emptyGenericProviderFactory) Create(options oauth.DynamicOptions) (GenericProvider, error) {
	return emptyGenericProvider{}, nil
}

func TestSetupWith(t *testing.T) {
	RegisterOAuthProvider(emptyOAuthProviderFactory{typeName: "GitHubIdentityProvider"})
	RegisterOAuthProvider(emptyOAuthProviderFactory{typeName: "OIDCIdentityProvider"})
	RegisterGenericProvider(emptyGenericProviderFactory{typeName: "LDAPIdentityProvider"})
	type args struct {
		options []oauth.IdentityProviderOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ldap",
			args: args{options: []oauth.IdentityProviderOptions{
				{
					Name:          "ldap",
					MappingMethod: "auto",
					Type:          "LDAPIdentityProvider",
					Provider:      oauth.DynamicOptions{},
				},
			}},
			wantErr: false,
		},
		{
			name: "conflict",
			args: args{options: []oauth.IdentityProviderOptions{
				{
					Name:          "ldap",
					MappingMethod: "auto",
					Type:          "LDAPIdentityProvider",
					Provider:      oauth.DynamicOptions{},
				},
			}},
			wantErr: true,
		},
		{
			name: "not supported",
			args: args{options: []oauth.IdentityProviderOptions{
				{
					Name:          "test",
					MappingMethod: "auto",
					Type:          "NotSupported",
					Provider:      oauth.DynamicOptions{},
				},
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetupWithOptions(tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("SetupWithOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
