/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package identityprovider

import (
	"kubesphere.io/kubesphere/pkg/server/options"
)

type GenericProvider interface {
	// Authenticate from remote server
	Authenticate(username string, password string) (Identity, error)
}

type GenericProviderFactory interface {
	// Type unique type of the provider
	Type() string
	// Create generic identity provider
	Create(options options.DynamicOptions) (GenericProvider, error)
}
