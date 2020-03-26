/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package oauth

import (
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/auth"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	"kubesphere.io/kubesphere/pkg/constants"
	"net/http"
)

// ks-apiserver includes a built-in OAuth server. Users obtain OAuth access tokens to authenticate themselves to the API.
// The OAuth server supports standard authorization code grant and the implicit grant OAuth authorization flows.
// All requests for OAuth tokens involve a request to <ks-apiserver>/oauth/authorize.
// Most authentication integrations place an authenticating proxy in front of this endpoint, or configure ks-apiserver
// to validate credentials against a backing identity provider.
// Requests to <ks-apiserver>/oauth/authorize can come from user-agents that cannot display interactive login pages, such as the CLI.
func AddToContainer(c *restful.Container, issuer token.Issuer, options *authoptions.AuthenticationOptions) error {
	ws := &restful.WebService{}
	ws.Path("/oauth").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	handler := newOAUTHHandler(issuer, options)

	// Implement webhook authentication interface
	// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
	ws.Route(ws.POST("/authenticate").
		Doc("TokenReview attempts to authenticate a token to a known user. Note: TokenReview requests may be cached by the webhook token authenticator plugin in the kube-apiserver.").
		Reads(auth.TokenReview{}).
		To(handler.TokenReviewHandler).
		Returns(http.StatusOK, api.StatusOK, auth.TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))

	// Only support implicit grant flow
	// https://tools.ietf.org/html/rfc6749#section-4.2
	ws.Route(ws.GET("/authorize").
		Doc("All requests for OAuth tokens involve a request to <ks-apiserver>/oauth/authorize.").
		To(handler.AuthorizeHandler))
	//ws.Route(ws.POST("/token"))

	// Authorization callback URL, where the end of the URL contains the identity provider name.
	// The provider name is also used to build the callback URL.
	ws.Route(ws.GET("/callback/{callback}").
		Doc("OAuth callback API, the path param callback is config by identity provider").
		To(handler.OAuthCallBackHandler).
		Returns(http.StatusOK, api.StatusOK, oauth.Token{}))

	c.Add(ws)

	return nil
}
