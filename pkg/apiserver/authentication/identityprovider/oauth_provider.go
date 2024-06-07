/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package identityprovider

import (
	"net/http"

	"kubesphere.io/kubesphere/pkg/server/options"
)

type OAuthProvider interface {
	// IdentityExchangeCallback handle oauth callback, exchange identity from remote server
	IdentityExchangeCallback(req *http.Request) (Identity, error)
}

type OAuthProviderFactory interface {
	// Type unique type of the provider
	Type() string
	// Create Apply the dynamic options
	Create(options options.DynamicOptions) (OAuthProvider, error)
}
