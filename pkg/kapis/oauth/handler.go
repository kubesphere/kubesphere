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
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/auth"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	kindTokenReview            = "TokenReview"
	passwordGrantType          = "password"
	authorizationCodeGrantType = "authorization_code"
	refreshTokenGrantType      = "refresh_token"
	responseTypeToken          = "token"
	responseTypeIDToken        = "id_token"
	responseTypeCode           = "code"

	errInvalidRequest          = "invalid_request"
	errUnauthorizedClient      = "unauthorized_client"
	errAccessDenied            = "access_denied"
	errUnsupportedResponseType = "unsupported_response_type"
	errRequestNotSupported     = "request_not_supported"
	errInvalidScope            = "invalid_scope"
	errServerError             = "server_error"
	errTemporarilyUnavailable  = "temporarily_unavailable"
	errUnsupportedGrantType    = "unsupported_grant_type"
	errInvalidGrant            = "invalid_grant"
	errInvalidClient           = "invalid_client"
	// Request a refresh token.
	scopeOfflineAccess = "offline_access"
	scopeOpenID        = "openid"
	scopeGroups        = "groups"
	scopeEmail         = "email"
	scopeProfile       = "profile"
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
	options               *authoptions.AuthenticationOptions
	tokenOperator         auth.TokenManagementInterface
	passwordAuthenticator auth.PasswordAuthenticator
	oauth2Authenticator   auth.OAuthAuthenticator
	loginRecorder         auth.LoginRecorder
}

func newHandler(im im.IdentityManagementInterface,
	tokenOperator auth.TokenManagementInterface,
	passwordAuthenticator auth.PasswordAuthenticator,
	oauth2Authenticator auth.OAuthAuthenticator,
	loginRecorder auth.LoginRecorder,
	options *authoptions.AuthenticationOptions) *handler {
	return &handler{im: im,
		tokenOperator:         tokenOperator,
		passwordAuthenticator: passwordAuthenticator,
		oauth2Authenticator:   oauth2Authenticator,
		loginRecorder:         loginRecorder,
		options:               options}
}

// Implement webhook authentication interface
// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
func (h *handler) TokenReview(req *restful.Request, resp *restful.Response) {
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

	claims, err := h.tokenOperator.Verify(tokenReview.Spec.Token)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	success := TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: kindTokenReview,
		Status: &Status{
			Authenticated: true,
			User:          map[string]interface{}{"username": claims.Username, "uid": claims.Subject},
		},
	}

	resp.WriteEntity(success)
}

func (h *handler) authError(error, description, state, redirectURI string, r *restful.Request, w *restful.Response) {
	v := url.Values{}
	v.Add("state", state)
	v.Add("error", error)
	if description != "" {
		v.Add("error_description", description)
	}
	if strings.Contains(redirectURI, "?") {
		redirectURI = redirectURI + "&" + v.Encode()
	} else {
		redirectURI = redirectURI + "?" + v.Encode()
	}
	http.Redirect(w, r.Request, redirectURI, http.StatusFound)
}

func (h *handler) tokenError(statusCode int, error, description string, r *restful.Request, w *restful.Response) {
	// https://tools.ietf.org/html/rfc6749#section-5.2
	err := struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description,omitempty"`
	}{
		Error:            error,
		ErrorDescription: description,
	}
	w.WriteHeaderAndEntity(statusCode, err)
}

func (h *handler) Authorize(req *restful.Request, resp *restful.Response) {
	clientID := req.QueryParameter("client_id")
	scopes := strings.Fields(req.QueryParameter("scope"))
	responseTypes := strings.Fields(req.QueryParameter("response_type"))
	redirectURI := req.QueryParameter("redirect_uri")
	state := req.QueryParameter("state")
	nonce := req.QueryParameter("nonce")

	authenticated, ok := request.UserFrom(req.Request.Context())
	if !ok || authenticated.GetName() == user.Anonymous {
		api.HandleUnauthorized(resp, req, errors.New("Login required"))
		return
	}

	// validate oauth client
	oauthClient, err := h.options.OAuthOptions.OAuthClient(clientID)
	if err != nil {
		description := fmt.Sprintf("Invalid client_id (%q).", clientID)
		h.authError(errInvalidClient, description, state, redirectURI, req, resp)
		return
	}

	// validate redirect URI
	redirectURL, err := oauthClient.ValidateRedirectURI(redirectURI)
	if err != nil {
		description := fmt.Sprintf("Unregistered redirect_uri (%q).", redirectURI)
		h.authError(errInvalidRequest, description, state, redirectURI, req, resp)
		return
	}

	// doesn't support request parameter and must return request_not_supported error
	// https://openid.net/specs/openid-connect-core-1_0.html#6.1
	if req.QueryParameter("request") != "" {
		h.authError(errRequestNotSupported, "Server does not support request parameter.", state, redirectURI, req, resp)
		return
	}

	// Scope values used that are not understood by an implementation SHOULD be ignored.
	// Scope filter not implemented
	if !sliceutil.HasString(scopes, scopeOpenID) {
		h.authError(errInvalidScope, `Missing required scope(s) ["openid"].`, state, redirectURI, req, resp)
		return
	}

	if len(responseTypes) == 0 {
		h.authError(errInvalidRequest, "No response_type provided", state, redirectURI, req, resp)
		return
	}

	// https://tools.ietf.org/html/rfc6749#section-4.1.2.1
	var hasToken, hasIDToken, hasCode bool
	for _, responseType := range responseTypes {
		switch responseType {
		case responseTypeCode:
			hasCode = true
		case responseTypeIDToken:
			hasIDToken = true
		case responseTypeToken:
			hasToken = true
		default:
			h.authError(errUnsupportedResponseType, fmt.Sprintf("Unsupported response type %s", responseType), state, redirectURI, req, resp)
			return
		}
	}
	if hasToken && !hasIDToken && !hasCode {
		// "token" can't be provided by its own.
		// https://openid.net/specs/openid-connect-core-1_0.html#Authentication
		h.authError(errInvalidRequest, "Response type 'token' must be provided with type 'id_token' and/or 'code'", state, redirectURI, req, resp)
	}

	if hasIDToken && nonce == "" {
		// Either "id_token token" or "id_token" has been provided which implies the implicit flow.
		// Implicit flow requires a nonce value.
		// https://openid.net/specs/openid-connect-core-1_0.html#ImplicitAuthRequest
		h.authError(errInvalidRequest, "Response type 'token' requires a 'nonce' value.", redirectURI, state, req, resp)
		return
	}

	values := make(url.Values, 0)
	if state != "" {
		values.Add("state", state)
	}
	if hasCode {
		claims := token.NewClaims(authenticated)
		claims.Nonce = nonce
		claims.Audience = clientID
		claims.TokenType = token.AuthorizationCode
		authorizationCode, err := h.tokenOperator.IssueCodeTo(claims)
		if err != nil {
			api.HandleUnauthorized(resp, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
			return
		}
		values.Add("code", authorizationCode)
		redirectURL.RawQuery = values.Encode()
	}
	if hasToken || hasIDToken {
		authToken, err := h.tokenOperator.IssueTo(token.NewClaims(authenticated))
		if err != nil {
			api.HandleUnauthorized(resp, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
			return
		}
		if hasToken {
			values.Add("access_token", authToken.AccessToken)
			values.Add("token_type", "Bearer")
			if authToken.ExpiresIn > 0 {
				values.Add("expires_in", strconv.Itoa(authToken.ExpiresIn))
			}
		}
		// https://openid.net/specs/openid-connect-core-1_0.html#ImplicitAuthResponse
		if hasIDToken {
			values.Add("id_token", authToken.IDToken)
		}
		// When using the Implicit Flow, all response parameters are added to the fragment component of the Redirection URI.
		// https://openid.net/specs/openid-connect-core-1_0.html#ImplicitAuthResponse
		redirectURL.Fragment = values.Encode()
	} else {
		// When using the Authorization Code Flow, the Authorization Response MUST return the parameters
		// defined in Section 4.1.2 of OAuth 2.0 [RFC6749] by adding them as query parameters
		// to the redirect_uri specified in the Authorization Request using the application/x-www-form-urlencoded format,
		// unless a different Response Mode was specified.
		// https://openid.net/specs/openid-connect-core-1_0.html#AuthResponse
		redirectURL.RawQuery = values.Encode()
	}

	resp.Header().Set("Content-Type", "text/plain")
	http.Redirect(resp, req.Request, redirectURL.String(), http.StatusFound)
}

func (h *handler) oauthCallback(req *restful.Request, resp *restful.Response) {
	provider := req.PathParameter("callback")
	// OAuth2 callback, see also https://tools.ietf.org/html/rfc6749#section-4.1.2
	code := req.QueryParameter("code")
	// CAS callback, see also https://apereo.github.io/cas/6.3.x/protocol/CAS-Protocol-V2-Specification.html#25-servicevalidate-cas-20
	if code == "" {
		code = req.QueryParameter("ticket")
	}
	if code == "" {
		api.HandleError(resp, req, apierrors.NewUnauthorized("Unauthorized: missing code"))
		return
	}
	authenticated, provider, err := h.oauth2Authenticator.Authenticate(provider, code)
	if err != nil {
		api.HandleUnauthorized(resp, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
		return
	}

	authToken, err := h.tokenOperator.IssueTo(token.NewClaims(authenticated))
	if err != nil {
		api.HandleInternalError(resp, req, apierrors.NewInternalError(err))
		return
	}

	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(authenticated.GetName(), iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", authenticated.GetName(), err)
	}

	resp.WriteEntity(authToken)
}

func (h *handler) Login(request *restful.Request, response *restful.Response) {
	var loginRequest LoginRequest
	err := request.ReadEntity(&loginRequest)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	h.passwordGrant(oauth.Client{Name: "default"}, loginRequest.Username, loginRequest.Password, request, response)
}

func (h *handler) Token(req *restful.Request, resp *restful.Response) {
	grantType, _ := req.BodyParameter("grant_type")
	clientID, _ := req.BodyParameter("client_id")
	// https://tools.ietf.org/html/rfc6749#section-2.3.1
	clientSecret, _ := req.BodyParameter("client_secret")
	if grantType == "" || clientID == "" || clientSecret == "" {
		h.tokenError(http.StatusBadRequest, errInvalidRequest, "required parameter is missing", req, resp)
		return
	}

	oauthClient, err := h.options.OAuthOptions.OAuthClient(clientID)
	if err != nil || clientSecret != oauthClient.Secret {
		api.HandleUnauthorized(resp, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", "client authentication failed")))
		return
	}

	switch grantType {
	// https://tools.ietf.org/html/rfc6749#section-4.1.3
	case authorizationCodeGrantType:
		h.authorizationCodeGrant(oauthClient, req, resp)
	case passwordGrantType:
		username, _ := req.BodyParameter("username")
		password, _ := req.BodyParameter("password")
		h.passwordGrant(oauthClient, username, password, req, resp)
	case refreshTokenGrantType:
		h.refreshTokenGrant(oauthClient, req, resp)
	default:
		h.tokenError(http.StatusBadRequest, errInvalidGrant, fmt.Sprintf("Grant type %s is not supported", grantType), req, resp)
	}
}

func (h *handler) authorizationCodeGrant(oauthClient oauth.Client, req *restful.Request, resp *restful.Response) {
	authorizationCode, _ := req.BodyParameter("code")
	redirectURI, _ := req.BodyParameter("redirect_uri")

	if authorizationCode == "" || redirectURI == "" {
		h.tokenError(http.StatusBadRequest, errInvalidRequest, "required parameter is missing", req, resp)
		return
	}

	claims, err := h.tokenOperator.Verify(authorizationCode)
	if err != nil || claims.TokenType != token.AuthorizationCode || claims.Audience != oauthClient.Name {
		h.tokenError(http.StatusBadRequest, errInvalidGrant, "The provided authorization grant code is invalid or was issued to another client", req, resp)
		return
	}

	userInfo, err := h.im.DescribeUser(claims.Username)
	if err != nil {
		if apierrors.IsNotFound(err) {
			h.tokenError(http.StatusBadRequest, errInvalidGrant, "Grant code has been revoked", req, resp)
			return
		}
		api.HandleInternalError(resp, req, err)
		return
	}

	authToken, err := h.tokenOperator.IssueTo(claims)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	claims.PreferredUsername = userInfo.Name
	claims.Subject = string(userInfo.UID)
	claims.Name = userInfo.Name
	claims.Email = userInfo.Spec.Email
	verified := false
	claims.EmailVerified = &verified
	idToken, err := h.tokenOperator.IssueIDTokenTo(claims, oauthClient.Secret)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	authToken.IDToken = idToken

	resp.WriteEntity(authToken)
}

func (h *handler) passwordGrant(oauthClient oauth.Client, username string, password string, req *restful.Request, response *restful.Response) {
	authenticated, provider, err := h.passwordAuthenticator.Authenticate(username, password)
	if err != nil {
		switch err {
		case auth.IncorrectPasswordError:
			requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
			if err := h.loginRecorder.RecordLogin(username, iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, err); err != nil {
				klog.Errorf("Failed to record unsuccessful login attempt for user %s, error: %v", username, err)
			}
			api.HandleUnauthorized(response, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
			return
		case auth.RateLimitExceededError:
			api.HandleTooManyRequests(response, req, apierrors.NewTooManyRequestsError(fmt.Sprintf("Unauthorized: %s", err)))
			return
		default:
			api.HandleInternalError(response, req, apierrors.NewInternalError(err))
			return
		}
	}
	authToken, err := h.tokenOperator.IssueTo(token.NewClaims(authenticated))
	if err != nil {
		api.HandleInternalError(response, req, apierrors.NewInternalError(err))
		return
	}

	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(authenticated.GetName(), iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", username, err)
	}

	response.WriteEntity(authToken)
}

func (h *handler) refreshTokenGrant(oauthClient oauth.Client, req *restful.Request, resp *restful.Response) {
	refreshToken, err := req.BodyParameter("refresh_token")
	if err != nil {
		h.tokenError(http.StatusBadRequest, errInvalidRequest, "required parameter is missing", req, resp)
		return
	}

	claims, err := h.tokenOperator.Verify(refreshToken)
	if err != nil {
		api.HandleUnauthorized(resp, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
		return
	}

	if claims.TokenType != token.RefreshToken {
		h.tokenError(http.StatusBadRequest, errInvalidRequest, "invalid token type", req, resp)
		return
	}

	// update token after registration
	if claims.Username == iamv1alpha2.PreRegistrationUser &&
		claims.Extra != nil &&
		len(claims.Extra[iamv1alpha2.ExtraIdentityProvider]) > 0 &&
		len(claims.Extra[iamv1alpha2.ExtraUID]) > 0 {

		idp := claims.Extra[iamv1alpha2.ExtraIdentityProvider][0]
		uid := claims.Extra[iamv1alpha2.ExtraUID][0]
		queryParam := query.New()
		queryParam.LabelSelector = labels.SelectorFromSet(labels.Set{
			iamv1alpha2.IdentifyProviderLabel: idp,
			iamv1alpha2.OriginUIDLabel:        uid}).String()
		result, err := h.im.ListUsers(queryParam)
		if err != nil {
			api.HandleInternalError(resp, req, apierrors.NewInternalError(err))
			return
		}
		if len(result.Items) != 1 {
			api.HandleUnauthorized(resp, req, apierrors.NewUnauthorized("authenticated user does not exist"))
			return
		}
		userInfo := result.Items[0].(*iamv1alpha2.User)
		claims = &token.Claims{Username: userInfo.Name}
	}

	authToken, err := h.tokenOperator.IssueTo(claims)
	if err != nil {
		api.HandleInternalError(resp, req, apierrors.NewInternalError(err))
		return
	}

	resp.WriteEntity(authToken)
}
