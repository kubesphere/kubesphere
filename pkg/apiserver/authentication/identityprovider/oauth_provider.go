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
)

type OAuthProvider interface {
	// IdentityExchange exchange identity from remote server
	IdentityExchange(code string) (Identity, error)
}

type OAuthProviderFactory interface {
	// Type unique type of the provider
	Type() string
	// Apply the dynamic options from kubesphere-config
	Create(options oauth.DynamicOptions) (OAuthProvider, error)
}
