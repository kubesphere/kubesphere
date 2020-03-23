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
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	"kubesphere.io/kubesphere/pkg/constants"
	"net/http"
)

func AddToContainer(c *restful.Container, issuer token.Issuer, configuration oauth.Configuration) error {
	ws := &restful.WebService{}
	ws.Path("/oauth").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	handler := newOAUTHHandler(issuer, configuration)

	// Implement webhook authentication interface
	// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
	ws.Route(ws.POST("/authenticate").
		Doc("TokenReview attempts to authenticate a token to a known user. Note: TokenReview requests may be cached by the webhook token authenticator plugin in the kube-apiserver.").
		Reads(auth.TokenReview{}).
		To(handler.TokenReviewHandler).
		Returns(http.StatusOK, api.StatusOK, auth.TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))

	// TODO Built-in oauth2 server (provider)
	// web console use 'Resource Owner Password Credentials Grant' or 'Client Credentials Grant' request for an OAuth token
	// https://tools.ietf.org/html/rfc6749#section-4.3
	// https://tools.ietf.org/html/rfc6749#section-4.4

	// curl -u admin:P@88w0rd 'http://ks-apiserver.kubesphere-system.svc/oauth/authorize?client_id=kubesphere-console-client&response_type=token' -v
	ws.Route(ws.GET("/authorize").
		To(handler.AuthorizeHandler))
	//ws.Route(ws.POST("/token"))
	//ws.Route(ws.POST("/callback/{callback}"))

	c.Add(ws)

	return nil
}
