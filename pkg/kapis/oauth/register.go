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

package oauth

import (
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/auth"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"net/http"
)

// ks-apiserver includes a built-in OAuth server. Users obtain OAuth access tokens to authenticate themselves to the API.
// The OAuth server supports standard authorization code grant and the implicit grant OAuth authorization flows.
// All requests for OAuth tokens involve a request to <ks-apiserver>/oauth/authorize.
// Most authentication integrations place an authenticating proxy in front of this endpoint, or configure ks-apiserver
// to validate credentials against a backing identity provider.
// Requests to <ks-apiserver>/oauth/authorize can come from user-agents that cannot display interactive login pages, such as the CLI.
func AddToContainer(c *restful.Container, im im.IdentityManagementInterface, tokenOperator im.TokenManagementInterface, authenticator im.PasswordAuthenticator, loginRecorder im.LoginRecorder, options *authoptions.AuthenticationOptions) error {
	ws := &restful.WebService{}
	ws.Path("/oauth").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	handler := newHandler(im, tokenOperator, authenticator, loginRecorder, options)

	// Implement webhook authentication interface
	// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
	ws.Route(ws.POST("/authenticate").
		Doc("TokenReview attempts to authenticate a token to a known user. Note: TokenReview requests may be "+
			"cached by the webhook token authenticator plugin in the kube-apiserver.").
		Reads(auth.TokenReview{}).
		To(handler.TokenReview).
		Returns(http.StatusOK, api.StatusOK, auth.TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))

	// Only support implicit grant flow
	// https://tools.ietf.org/html/rfc6749#section-4.2
	ws.Route(ws.GET("/authorize").
		Doc("All requests for OAuth tokens involve a request to <ks-apiserver>/oauth/authorize.").
		Param(ws.QueryParameter("response_type", "The value MUST be one of \"code\" for requesting an "+
			"authorization code as described by [RFC6749] Section 4.1.1, \"token\" for requesting an access token (implicit grant)"+
			" as described by [RFC6749] Section 4.2.2.").Required(true)).
		Param(ws.QueryParameter("client_id", "The client identifier issued to the client during the "+
			"registration process described by [RFC6749] Section 2.2.").Required(true)).
		Param(ws.QueryParameter("redirect_uri", "After completing its interaction with the resource owner, "+
			"the authorization server directs the resource owner's user-agent back to the client.The redirection endpoint "+
			"URI MUST be an absolute URI as defined by [RFC3986] Section 4.3.").Required(false)).
		To(handler.Authorize))
	// Resource Owner Password Credentials Grant
	// https://tools.ietf.org/html/rfc6749#section-4.3
	ws.Route(ws.POST("/token").
		Consumes("application/x-www-form-urlencoded").
		Doc("The resource owner password credentials grant type is suitable in\n" +
			"cases where the resource owner has a trust relationship with the\n" +
			"client, such as the device operating system or a highly privileged application.").
		To(handler.Token))

	// Authorization callback URL, where the end of the URL contains the identity provider name.
	// The provider name is also used to build the callback URL.
	ws.Route(ws.GET("/callback/{callback}").
		Doc("OAuth callback API, the path param callback is config by identity provider").
		Param(ws.QueryParameter("access_token", "The access token issued by the authorization server.").
			Required(true)).
		Param(ws.QueryParameter("token_type", "The type of the token issued as described in [RFC6479] Section 7.1. "+
			"Value is case insensitive.").Required(true)).
		Param(ws.QueryParameter("expires_in", "The lifetime in seconds of the access token.  For "+
			"example, the value \"3600\" denotes that the access token will "+
			"expire in one hour from the time the response was generated."+
			"If omitted, the authorization server SHOULD provide the "+
			"expiration time via other means or document the default value.")).
		Param(ws.QueryParameter("scope", "if identical to the scope requested by the client;"+
			"otherwise, REQUIRED.  The scope of the access token as described by [RFC6479] Section 3.3.").Required(false)).
		Param(ws.QueryParameter("state", "if the \"state\" parameter was present in the client authorization request."+
			"The exact value received from the client.").Required(true)).
		To(handler.oAuthCallBack).
		Returns(http.StatusOK, api.StatusOK, oauth.Token{}))

	c.Add(ws)

	// legacy auth API
	legacy := &restful.WebService{}
	legacy.Path("/kapis/iam.kubesphere.io/v1alpha2/login").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	legacy.Route(legacy.POST("").
		To(handler.Login).
		Deprecate().
		Doc("KubeSphere APIs support token-based authentication via the Authtoken request header. The POST Login API is used to retrieve the authentication token. After the authentication token is obtained, it must be inserted into the Authtoken header for all requests.").
		Reads(auth.LoginRequest{}).
		Returns(http.StatusOK, api.StatusOK, oauth.Token{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))

	c.Add(legacy)

	return nil
}
