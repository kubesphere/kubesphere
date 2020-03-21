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
	"kubesphere.io/kubesphere/pkg/api/auth/token"
	"kubesphere.io/kubesphere/pkg/constants"
	"net/http"
)

func AddToContainer(c *restful.Container, options *auth.AuthenticationOptions) error {
	ws := restful.WebService{}
	ws.Path("/oauth").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	handler := newOAUTHHandler(token.NewJwtTokenIssuer(token.DefaultIssuerName, []byte(options.JwtSecret)))

	// Implement webhook authentication interface
	// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
	ws.Route(ws.POST("/authenticate").
		Doc("TokenReview attempts to authenticate a token to a known user. Note: TokenReview requests may be cached by the webhook token authenticator plugin in the kube-apiserver.").
		Reads(auth.TokenReview{}).
		To(handler.TokenReviewHandler).
		Returns(http.StatusOK, api.StatusOK, auth.TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))

	// TODO Built-in oauth2 server (provider)
	// Low priority
	c.Add(ws.Route(ws.POST("/authorize")))

	// web console use 'Resource Owner Password Credentials Grant' or 'Client Credentials Grant' request for an OAuth token
	// https://tools.ietf.org/html/rfc6749#section-4.3
	// https://tools.ietf.org/html/rfc6749#section-4.4
	c.Add(ws.Route(ws.POST("/token")))

	// oauth2 client callback
	c.Add(ws.Route(ws.POST("/callback/{callback}")))

	return nil
}
