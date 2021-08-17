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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"

	"kubesphere.io/kubesphere/pkg/server/errors"

	"github.com/emicklei/go-restful"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/auth"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
)

const (
	KindTokenReview       = "TokenReview"
	grantTypePassword     = "password"
	grantTypeRefreshToken = "refresh_token"
	grantTypeCode         = "code"
)

type Spec struct {
	Token string `json:"token" description:"access token"`
}

type Status struct {
	Authenticated bool                   `json:"authenticated" description:"is authenticated"`
	User          map[string]interface{} `json:"user,omitempty" description:"user info"`
}

type TokenReview struct {
	APIVersion string  `json:"apiVersion" description:"Kubernetes API version"`
	Kind       string  `json:"kind" description:"kind of the API object"`
	Spec       *Spec   `json:"spec,omitempty"`
	Status     *Status `json:"status,omitempty" description:"token review status"`
}

type LoginRequest struct {
	Username string `json:"username" description:"username"`
	Password string `json:"password" description:"password"`
}

func (request *TokenReview) Validate() error {
	if request.Spec == nil || request.Spec.Token == "" {
		return fmt.Errorf("token must not be null")
	}
	return nil
}

type handler struct {
	im                    im.IdentityManagementInterface
	options               *authentication.Options
	tokenOperator         auth.TokenManagementInterface
	passwordAuthenticator auth.PasswordAuthenticator
	oauthAuthenticator    auth.OAuthAuthenticator
	loginRecorder         auth.LoginRecorder
}

func newHandler(im im.IdentityManagementInterface,
	tokenOperator auth.TokenManagementInterface,
	passwordAuthenticator auth.PasswordAuthenticator,
	oauthAuthenticator auth.OAuthAuthenticator,
	loginRecorder auth.LoginRecorder,
	options *authentication.Options) *handler {
	return &handler{im: im,
		tokenOperator:         tokenOperator,
		passwordAuthenticator: passwordAuthenticator,
		oauthAuthenticator:    oauthAuthenticator,
		loginRecorder:         loginRecorder,
		options:               options}
}

// tokenReview Implement webhook authentication interface
// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
func (h *handler) tokenReview(req *restful.Request, resp *restful.Response) {
	var tokenReview TokenReview

	err := req.ReadEntity(&tokenReview)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	if err = tokenReview.Validate(); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	verified, err := h.tokenOperator.Verify(tokenReview.Spec.Token)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	authenticated := verified.User
	success := TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: KindTokenReview,
		Status: &Status{
			Authenticated: true,
			User:          map[string]interface{}{"username": authenticated.GetName(), "uid": authenticated.GetUID()},
		},
	}

	resp.WriteEntity(success)
}

// The Authorization Endpoint performs Authentication of the End-User.
func (h *handler) authorize(req *restful.Request, response *restful.Response) {
	var scope, responseType, clientID, redirectURI, state, nonce string
	scope = req.QueryParameter("scope")
	clientID = req.QueryParameter("client_id")
	redirectURI = req.QueryParameter("redirect_uri")
	//prompt = req.QueryParameter("prompt")
	responseType = req.QueryParameter("response_type")
	state = req.QueryParameter("state")
	nonce = req.QueryParameter("nonce")

	// Authorization Servers MUST support the use of the HTTP GET and POST methods
	// defined in RFC 2616 [RFC2616] at the Authorization Endpoint.
	if req.Request.Method == http.MethodPost {
		scope, _ = req.BodyParameter("scope")
		clientID, _ = req.BodyParameter("client_id")
		redirectURI, _ = req.BodyParameter("redirect_uri")
		responseType, _ = req.BodyParameter("response_type")
		state, _ = req.BodyParameter("state")
		nonce, _ = req.BodyParameter("nonce")
	}

	oauthClient, err := h.options.OAuthOptions.OAuthClient(clientID)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidClient(err))
		return
	}

	redirectURL, err := oauthClient.ResolveRedirectURL(redirectURI)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest(err))
		return
	}

	authenticated, _ := request.UserFrom(req.Request.Context())
	if authenticated == nil || authenticated.GetName() == user.Anonymous {
		response.Header().Add("WWW-Authenticate", "Basic")
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauth.ErrorLoginRequired)
		return
	}

	// If no openid scope value is present, the request may still be a valid OAuth 2.0 request,
	// but is not an OpenID Connect request.
	var scopes []string
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}
	var responseTypes []string
	if responseType != "" {
		responseTypes = strings.Split(responseType, " ")
	}

	// If the resource owner denies the access request or if the request
	// fails for reasons other than a missing or invalid redirection URI,
	// the authorization server informs the client by adding the following
	// parameters to the query component of the redirection URI using the
	// "application/x-www-form-urlencoded" format
	informsError := func(err oauth.Error) {
		values := make(url.Values, 0)
		values.Add("error", err.Type)
		if err.Description != "" {
			values.Add("error_description", err.Description)
		}
		if state != "" {
			values.Add("state", state)
		}
		redirectURL.RawQuery = values.Encode()
		http.Redirect(response.ResponseWriter, req.Request, redirectURL.String(), http.StatusFound)
		return
	}

	// Other scope values MAY be present.
	// Scope values used that are not understood by an implementation SHOULD be ignored.
	if !oauth.IsValidScopes(scopes) {
		klog.Warningf("Some requested scopes were invalid: %v", scopes)
	}

	if !oauth.IsValidResponseTypes(responseTypes) {
		err := fmt.Errorf("Some requested response types were invalid")
		informsError(oauth.NewInvalidRequest(err))
		return
	}

	// TODO(hongming) support Hybrid Flow
	// Authorization Code Flow
	if responseType == oauth.ResponseCode {
		code, err := h.tokenOperator.IssueTo(&token.IssueRequest{
			User: authenticated,
			Claims: token.Claims{
				StandardClaims: jwt.StandardClaims{
					Audience: []string{clientID},
				},
				TokenType: token.AuthorizationCode,
				Nonce:     nonce,
			},
			// A maximum authorization code lifetime of 10 minutes is
			ExpiresIn: 10 * time.Minute,
		})
		if err != nil {
			response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
			return
		}
		values := redirectURL.Query()
		values.Add("code", code)
		redirectURL.RawQuery = values.Encode()
		http.Redirect(response, req.Request, redirectURL.String(), http.StatusFound)
	}

	// Implicit Flow
	if responseType != oauth.ResponseToken {
		informsError(oauth.ErrorUnsupportedResponseType)
		return
	}

	result, err := h.issueTokenTo(authenticated)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
		return
	}

	values := make(url.Values, 0)
	values.Add("access_token", result.AccessToken)
	values.Add("refresh_token", result.RefreshToken)
	values.Add("token_type", result.TokenType)
	values.Add("expires_in", fmt.Sprint(result.ExpiresIn))
	redirectURL.Fragment = values.Encode()
	http.Redirect(response, req.Request, redirectURL.String(), http.StatusFound)
}

func (h *handler) oauthCallback(req *restful.Request, response *restful.Response) {
	provider := req.PathParameter("callback")
	authenticated, provider, err := h.oauthAuthenticator.Authenticate(req.Request.Context(), provider, req.Request)
	if err != nil {
		api.HandleUnauthorized(response, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
		return
	}

	result, err := h.issueTokenTo(authenticated)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
		return
	}

	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(authenticated.GetName(), iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", authenticated.GetName(), err)
	}

	response.WriteEntity(result)
}

func (h *handler) login(request *restful.Request, response *restful.Response) {
	var loginRequest LoginRequest
	err := request.ReadEntity(&loginRequest)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	h.passwordGrant(loginRequest.Username, loginRequest.Password, request, response)
}

// To obtain an Access Token, an ID Token, and optionally a Refresh Token,
// the RP (Client) sends a Token Request to the Token Endpoint to obtain a Token Response,
// as described in Section 3.2 of OAuth 2.0 [RFC6749], when using the Authorization Code Flow.
// Communication with the Token Endpoint MUST utilize TLS.
func (h *handler) token(req *restful.Request, response *restful.Response) {
	// TODO(hongming) support basic auth
	// https://datatracker.ietf.org/doc/html/rfc6749#section-2.3
	clientID, err := req.BodyParameter("client_id")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauth.NewInvalidClient(err))
		return
	}
	clientSecret, err := req.BodyParameter("client_secret")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauth.NewInvalidClient(err))
		return
	}

	client, err := h.options.OAuthOptions.OAuthClient(clientID)
	if err != nil {
		oauthError := oauth.NewInvalidClient(err)
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauthError)
		return
	}

	if client.Secret != clientSecret {
		oauthError := oauth.NewInvalidClient(fmt.Errorf("invalid client credential"))
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauthError)
		return
	}

	grantType, err := req.BodyParameter("grant_type")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest(err))
		return
	}

	switch grantType {
	case grantTypePassword:
		username, _ := req.BodyParameter("username")
		password, _ := req.BodyParameter("password")
		h.passwordGrant(username, password, req, response)
		return
	case grantTypeRefreshToken:
		h.refreshTokenGrant(req, response)
		return
	case grantTypeCode:
		h.codeGrant(req, response)
		return
	default:
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.ErrorUnsupportedGrantType)
		return
	}
}

// passwordGrant handle Resource Owner Password Credentials Grant
// for more details: https://datatracker.ietf.org/doc/html/rfc6749#section-4.3
// The resource owner password credentials grant type is suitable in
// cases where the resource owner has a trust relationship with the client,
// such as the device operating system or a highly privileged application.
// The authorization server should take special care when enabling this
// grant type and only allow it when other flows are not viable.
func (h *handler) passwordGrant(username string, password string, req *restful.Request, response *restful.Response) {
	authenticated, provider, err := h.passwordAuthenticator.Authenticate(req.Request.Context(), username, password)
	if err != nil {
		switch err {
		case auth.IncorrectPasswordError:
			requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
			if err := h.loginRecorder.RecordLogin(username, iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, err); err != nil {
				klog.Errorf("Failed to record unsuccessful login attempt for user %s, error: %v", username, err)
			}
			response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
			return
		case auth.RateLimitExceededError:
			response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
			return
		default:
			response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
			return
		}
	}

	result, err := h.issueTokenTo(authenticated)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
		return
	}

	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(authenticated.GetName(), iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", authenticated.GetName(), err)
	}

	response.WriteEntity(result)
}

func (h *handler) issueTokenTo(authenticated user.Info) (*oauth.Token, error) {
	accessToken, err := h.tokenOperator.IssueTo(&token.IssueRequest{
		User:      authenticated,
		Claims:    token.Claims{TokenType: token.AccessToken},
		ExpiresIn: h.options.OAuthOptions.AccessTokenMaxAge,
	})
	if err != nil {
		return nil, err
	}
	refreshToken, err := h.tokenOperator.IssueTo(&token.IssueRequest{
		User:      authenticated,
		Claims:    token.Claims{TokenType: token.RefreshToken},
		ExpiresIn: h.options.OAuthOptions.AccessTokenMaxAge + h.options.OAuthOptions.AccessTokenInactivityTimeout,
	})
	if err != nil {
		return nil, err
	}

	result := oauth.Token{
		AccessToken: accessToken,
		// The OAuth 2.0 token_type response parameter value MUST be Bearer,
		// as specified in OAuth 2.0 Bearer Token Usage [RFC6750]
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		ExpiresIn:    int(h.options.OAuthOptions.AccessTokenMaxAge.Seconds()),
	}
	return &result, nil
}

func (h *handler) refreshTokenGrant(req *restful.Request, response *restful.Response) {
	refreshToken, err := req.BodyParameter("refresh_token")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest(err))
		return
	}

	verified, err := h.tokenOperator.Verify(refreshToken)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
		return
	}

	if verified.TokenType != token.RefreshToken {
		err = fmt.Errorf("ivalid token type %v want %v", verified.TokenType, token.RefreshToken)
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
		return
	}

	authenticated := verified.User
	// update token after registration
	if authenticated.GetName() == iamv1alpha2.PreRegistrationUser &&
		authenticated.GetExtra() != nil &&
		len(authenticated.GetExtra()[iamv1alpha2.ExtraIdentityProvider]) > 0 &&
		len(authenticated.GetExtra()[iamv1alpha2.ExtraUID]) > 0 {

		idp := authenticated.GetExtra()[iamv1alpha2.ExtraIdentityProvider][0]
		uid := authenticated.GetExtra()[iamv1alpha2.ExtraUID][0]
		queryParam := query.New()
		queryParam.LabelSelector = labels.SelectorFromSet(labels.Set{
			iamv1alpha2.IdentifyProviderLabel: idp,
			iamv1alpha2.OriginUIDLabel:        uid}).String()
		result, err := h.im.ListUsers(queryParam)
		if err != nil {
			response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
			return
		}
		if len(result.Items) != 1 {
			response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(fmt.Errorf("authenticated user does not exist")))
			return
		}

		authenticated = &user.DefaultInfo{Name: result.Items[0].(*iamv1alpha2.User).Name}
	}

	result, err := h.issueTokenTo(authenticated)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
		return
	}

	response.WriteEntity(result)
}

func (h *handler) codeGrant(req *restful.Request, response *restful.Response) {
	code, err := req.BodyParameter("code")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest(err))
		return
	}

	authorizeContext, err := h.tokenOperator.Verify(code)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
		return
	}
	if authorizeContext.TokenType != token.AuthorizationCode {
		err = fmt.Errorf("ivalid token type %v want %v", authorizeContext.TokenType, token.AuthorizationCode)
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
		return
	}

	// The client MUST NOT use the authorization code more than once.
	err = h.tokenOperator.Revoke(code)
	if err != nil {
		klog.Warningf("grant: failed to revoke authorization code: %v", err)
	}

	authenticated := authorizeContext.User
	result, err := h.issueTokenTo(authenticated)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
		return
	}

	idToken, err := h.tokenOperator.IssueTo(&token.IssueRequest{
		User: authenticated,
		Claims: token.Claims{
			StandardClaims: jwt.StandardClaims{
				Audience: authorizeContext.Audience,
			},
			Nonce:     authorizeContext.Nonce,
			TokenType: token.IDToken,
		},
		ExpiresIn: h.options.OAuthOptions.AccessTokenMaxAge + h.options.OAuthOptions.AccessTokenInactivityTimeout,
	})
	result.IDToken = idToken
	response.WriteEntity(result)
}

func (h *handler) logout(req *restful.Request, resp *restful.Response) {
	authenticated, ok := request.UserFrom(req.Request.Context())
	if ok {
		if err := h.tokenOperator.RevokeAllUserTokens(authenticated.GetName()); err != nil {
			api.HandleInternalError(resp, req, apierrors.NewInternalError(err))
			return
		}
	}

	postLogoutRedirectURI := req.QueryParameter("post_logout_redirect_uri")
	if postLogoutRedirectURI == "" {
		resp.WriteAsJson(errors.None)
		return
	}

	redirectURL, err := url.Parse(postLogoutRedirectURI)
	if err != nil {
		api.HandleBadRequest(resp, req, fmt.Errorf("invalid logout redirect URI: %s", err))
		return
	}

	state := req.QueryParameter("state")
	if state != "" {
		redirectURL.Query().Add("state", state)
	}

	resp.Header().Set("Content-Type", "text/plain")
	http.Redirect(resp, req.Request, redirectURL.String(), http.StatusFound)
}
