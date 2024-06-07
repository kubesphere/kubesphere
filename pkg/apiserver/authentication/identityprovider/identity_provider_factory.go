/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package identityprovider

var (
	oauthProviderFactories   = make(map[string]OAuthProviderFactory)
	genericProviderFactories = make(map[string]GenericProviderFactory)
)

// RegisterOAuthProviderFactory register OAuthProviderFactory with the specified type
func RegisterOAuthProviderFactory(factory OAuthProviderFactory) {
	oauthProviderFactories[factory.Type()] = factory
}

// RegisterGenericProviderFactory registers GenericProviderFactory with the specified type
func RegisterGenericProviderFactory(factory GenericProviderFactory) {
	genericProviderFactories[factory.Type()] = factory
}
