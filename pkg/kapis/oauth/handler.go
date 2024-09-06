/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package oauth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang-jwt/jwt/v4"
	"gopkg.in/square/go-jose.v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/models/auth"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	serverrors "kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	KindTokenReview            = "TokenReview"
	internalServerErrorMessage = "An internal server error occurred while processing the request."
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

// ProviderMetadata https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
type ProviderMetadata struct {
	// URL using the https scheme with no query or fragment component that the OP
	// asserts as its Issuer Identifier.
	Issuer string `json:"issuer"`
	// URL of the OP's OAuth 2.0 Authorization Endpoint.
	Auth string `json:"authorization_endpoint"`
	// URL of the OP's OAuth 2.0 Token Endpoint.
	Token string `json:"token_endpoint"`
	// URL of the OP's UserInfo Endpoint
	UserInfo string `json:"userinfo_endpoint"`
	// URL of the OP's JSON Web Key Set [JWK] document.
	Keys string `json:"jwks_uri"`
	// JSON array containing a list of the OAuth 2.0 Grant Type values that this OP supports.
	GrantTypes []string `json:"grant_types_supported"`
	// JSON array containing a list of the OAuth 2.0 response_type values that this OP supports.
	ResponseTypes []string `json:"response_types_supported"`
	// JSON array containing a list of the Subject Identifier types that this OP supports.
	Subjects []string `json:"subject_types_supported"`
	// JSON array containing a list of the JWS signing algorithms (alg values) supported by
	// the OP for the ID Token to encode the Claims in a JWT [JWT].
	IDTokenAlgs []string `json:"id_token_signing_alg_values_supported"`
	// JSON array containing a list of Proof Key for Code
	// Exchange (PKCE) [RFC7636] code challenge methods supported by this authorization server.
	CodeChallengeAlgs []string `json:"code_challenge_methods_supported"`
	// JSON array containing a list of the OAuth 2.0 [RFC6749] scope values that this server supports.
	Scopes []string `json:"scopes_supported"`
	// JSON array containing a list of Client Authentication methods supported by this Token Endpoint.
	AuthMethods []string `json:"token_endpoint_auth_methods_supported"`
	// JSON array containing a list of the Claim Names of the Claims that the OpenID Provider
	// MAY be able to supply values for.
	Claims []string `json:"claims_supported"`
}

type handler struct {
	im                    im.IdentityManagementInterface
	options               *authentication.Options
	tokenOperator         auth.TokenManagementInterface
	passwordAuthenticator auth.PasswordAuthenticator
	oauthAuthenticator    auth.OAuthAuthenticator
	loginRecorder         auth.LoginRecorder
	clientGetter          oauth.ClientGetter
}

func NewHandler(im im.IdentityManagementInterface,
	tokenOperator auth.TokenManagementInterface,
	passwordAuthenticator auth.PasswordAuthenticator,
	oauth2Authenticator auth.OAuthAuthenticator,
	loginRecorder auth.LoginRecorder,
	options *authentication.Options,
	oauthOperator oauth.ClientGetter) rest.Handler {
	handler := &handler{im: im,
		tokenOperator:         tokenOperator,
		passwordAuthenticator: passwordAuthenticator,
		oauthAuthenticator:    oauth2Authenticator,
		loginRecorder:         loginRecorder,
		options:               options,
		clientGetter:          oauthOperator}
	return handler
}

func FakeHandler() rest.Handler {
	handler := &handler{}
	return handler
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

	_ = resp.WriteEntity(success)
}

func (h *handler) discovery(_ *restful.Request, response *restful.Response) {
	result := ProviderMetadata{
		Issuer:            h.options.Issuer.URL,
		Auth:              h.options.Issuer.URL + root + "/authorize",
		Token:             h.options.Issuer.URL + root + "/token",
		Keys:              h.options.Issuer.URL + root + "/keys",
		UserInfo:          h.options.Issuer.URL + root + "/userinfo",
		Subjects:          []string{"public"},
		GrantTypes:        []string{oauth.GrantTypeAuthorizationCode, oauth.GrantTypeRefreshToken},
		IDTokenAlgs:       []string{string(jose.RS256)},
		CodeChallengeAlgs: []string{"plain", "S256"},
		Scopes:            []string{oauth.ScopeOpenID, oauth.ScopeEmail, oauth.ScopeProfile},
		AuthMethods:       []string{"client_secret_post"},
		Claims: []string{
			"iss", "sub", "aud", "iat", "exp", "email", "locale", "preferred_username",
		},
		ResponseTypes: []string{
			oauth.ResponseTypeCode,
			oauth.ResponseTypeIDToken,
		},
	}

	_ = response.WriteAsJson(result)
}

func (h *handler) keys(_ *restful.Request, response *restful.Response) {
	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{*h.tokenOperator.Keys().SigningKeyPub},
	}
	_ = response.WriteEntity(jwks)
}

// The Authorization Endpoint performs Authentication of the End-User.
func (h *handler) authorize(req *restful.Request, response *restful.Response) {
	var scope, responseType, clientID, redirectURI, state, nonce, prompt string
	scope = req.QueryParameter("scope")
	clientID = req.QueryParameter("client_id")
	redirectURI = req.QueryParameter("redirect_uri")
	responseType = req.QueryParameter("response_type")
	state = req.QueryParameter("state")
	nonce = req.QueryParameter("nonce")
	prompt = req.QueryParameter("prompt")
	// Authorization Servers MUST support the use of the HTTP GET and POST methods
	// defined in RFC 2616 [RFC2616] at the Authorization Endpoint.
	if req.Request.Method == http.MethodPost {
		scope, _ = req.BodyParameter("scope")
		clientID, _ = req.BodyParameter("client_id")
		redirectURI, _ = req.BodyParameter("redirect_uri")
		responseType, _ = req.BodyParameter("response_type")
		state, _ = req.BodyParameter("state")
		nonce, _ = req.BodyParameter("nonce")
		prompt, _ = req.BodyParameter("prompt")
	}

	client, err := h.clientGetter.GetOAuthClient(req.Request.Context(), clientID)
	if err != nil {
		if errors.Is(err, oauth.ErrorClientNotFound) {
			_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidClient("The provided client_id is invalid or does not exist."))
			return
		}
		klog.Errorf("failed to get oauth client: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	redirectURL, err := client.ResolveRedirectURL(redirectURI)
	if err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest("Redirect URI is not allowed."))
		return
	}

	// Unless the Redirection URI is invalid, the Authorization Server returns the Client to the Redirection URI
	// specified in the Authorization Request with the appropriate error and state parameters.
	// Other parameters SHOULD NOT be returned.
	// The authorization server informs the client by adding the following
	// parameters to the query component of the redirection URI using the
	// "application/x-www-form-urlencoded" format
	informsError := func(err *oauth.Error) {
		values := make(url.Values)
		values.Add("error", string(err.Type))
		if err.Description != "" {
			values.Add("error_description", err.Description)
		}
		if state != "" {
			values.Add("state", state)
		}
		redirectURL.RawQuery = values.Encode()
		http.Redirect(response.ResponseWriter, req.Request, redirectURL.String(), http.StatusFound)
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

	if !client.IsValidScope(scope) {
		informsError(oauth.NewInvalidScope("The requested scope is invalid or not supported."))
		return
	}

	// Hybrid flow is not supported now
	if len(responseTypes) > 1 || !oauth.IsValidResponseTypes(responseTypes) {
		informsError(oauth.NewError(oauth.UnsupportedResponseType, fmt.Sprintf("The provided response_type %s is not supported by the authorization server.", responseType)))
		return
	}

	if client.GrantMethod == oauth.GrantMethodDeny {
		informsError(oauth.NewInvalidGrant("The resource owner or authorization server denied the request."))
		return
	}

	authenticated, _ := request.UserFrom(req.Request.Context())
	if authenticated == nil || authenticated.GetName() == user.Anonymous {
		if prompt == "none" {
			informsError(oauth.NewError(oauth.LoginRequired, "Not authenticated."))
			return
		}
		// TODO redirect to login page with refer
		http.Redirect(response.ResponseWriter, req.Request, h.options.Issuer.URL, http.StatusFound)
		return
	}

	approved := client.GrantMethod == oauth.GrantMethodAuto
	if prompt == "none" && !approved {
		informsError(oauth.NewError(oauth.InteractionRequired, "Consent is required before proceeding with the request."))
		return
	}

	// TODO oauth.GrantMethodPrompt

	// oauth.GrantMethodAuto
	switch responseType {
	case oauth.ResponseTypeCode:
		h.handleAuthorizationCodeRequest(req, response, authCodeRequest{
			authenticated: authenticated,
			clientID:      clientID,
			nonce:         nonce,
			scopes:        scopes,
			redirectURL:   redirectURL,
			state:         state,
		})
	case oauth.ResponseTypeIDToken:
		h.handleAuthIDTokenRequest(req, response, &authIDTokenRequest{
			idTokenRequest: &idTokenRequest{
				authenticated: authenticated,
				client:        client,
				nonce:         nonce,
				scopes:        scopes,
			},
			redirectURL: redirectURL,
			state:       state,
		})
	default:
		informsError(oauth.NewError(oauth.UnsupportedResponseType, "The provided response_type is not supported by the authorization server."))
	}
}

func (h *handler) oauthCallback(req *restful.Request, response *restful.Response) {
	provider := req.PathParameter("callback")
	authenticated, err := h.oauthAuthenticator.Authenticate(req.Request.Context(), provider, req.Request)
	if err != nil {
		api.HandleUnauthorized(response, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
		return
	}

	// TODO(@hongming) using the really client configuration
	result, err := h.issueTokenTo(authenticated, nil)
	if err != nil {
		klog.Errorf("failed to issue token: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(req.Request.Context(), authenticated.GetName(), iamv1beta1.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", authenticated.GetName(), err)
	}

	_ = response.WriteEntity(result)
}

// token handles the Token Request to obtain an Access Token, an ID Token, and optionally a Refresh Token.
// This is used in the Authorization Code Flow, where the RP (Client) sends a Token Request to the Token Endpoint
// (described in Section 3.2 of OAuth 2.0 [RFC6749]) to obtain a Token Response.
// Communication with the Token Endpoint is required to utilize TLS for security.
func (h *handler) token(req *restful.Request, response *restful.Response) {
	clientID, _ := req.BodyParameter("client_id")
	clientSecret, _ := req.BodyParameter("client_secret")
	grantType, _ := req.BodyParameter("grant_type")

	// All Token Responses containing sensitive information MUST include the following HTTP response header fields and values:
	// Cache-Control: no-store
	// Pragma: no-cache
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("Pragma", "no-cache")

	// Retrieve the OAuth client associated with the provided client_id.
	client, err := h.clientGetter.GetOAuthClient(req.Request.Context(), clientID)
	if err != nil {
		if errors.Is(err, oauth.ErrorClientNotFound) {
			klog.Warningf("The provided client_id %s is invalid or does not exist.", clientID)
			_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidClient("The provided client_id is invalid or does not exist."))
			return
		}
		klog.Errorf("failed to get oauth client: %v", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	// Check if the client_secret matches the one associated with the retrieved client.
	if client.Secret != clientSecret {
		klog.Warningf("Invalid client credential for client_id %s", clientID)
		_ = response.WriteHeaderAndEntity(http.StatusUnauthorized, oauth.NewError(oauth.UnauthorizedClient, "Invalid client credential."))
		return
	}

	unsupportedGrantType := oauth.NewError(oauth.UnsupportedGrantType, "The provided grant_type is not supported.")

	switch grantType {
	case oauth.GrantTypePassword:
		if client.Trusted {
			h.passwordGrant(req, response, client)
			return
		}
		klog.Warningf("The client %s is not trusted.", client.Name)
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, unsupportedGrantType)
	case oauth.GrantTypeRefreshToken:
		h.refreshTokenGrant(req, response, client)
	case oauth.GrantTypeCode, oauth.GrantTypeAuthorizationCode:
		h.codeGrant(req, response, client)
	default:
		klog.Warningf("The provided grant_type %s is not supported.", grantType)
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, unsupportedGrantType)
	}
}

// passwordGrant handles the Resource Owner Password Credentials Grant.
// For more details, refer to: https://datatracker.ietf.org/doc/html/rfc6749#section-4.3
//
// The resource owner password credentials grant type is suitable in cases where
// the resource owner has a trust relationship with the client, such as the device
// operating system or a highly privileged application. The authorization server should
// take special care when enabling this grant type and only allow it when other flows
// are not viable.
func (h *handler) passwordGrant(req *restful.Request, response *restful.Response, client *oauth.Client) {
	// Extracting parameters from the request body.
	username, _ := req.BodyParameter("username")
	password, _ := req.BodyParameter("password")
	provider, _ := req.BodyParameter("provider")

	// Authenticate the user credentials.
	authenticated, err := h.passwordAuthenticator.Authenticate(req.Request.Context(), provider, username, password)
	if err != nil {
		switch {
		case errors.Is(err, auth.AccountIsNotActiveError):
			// The Account is suspended.
			_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant("Account suspended."))
			return
		case errors.Is(err, auth.IncorrectPasswordError):
			// Record unsuccessful login attempt.
			requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
			if err := h.loginRecorder.RecordLogin(req.Request.Context(), username, iamv1beta1.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, err); err != nil {
				klog.Errorf("Failed to record unsuccessful login attempt for user %s, error: %v", username, err)
			}
			// Invalid username or password.
			_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant("Invalid username or password."))
			return
		case errors.Is(err, auth.RateLimitExceededError):
			// Rate limit exceeded.
			_ = response.WriteHeaderAndEntity(http.StatusTooManyRequests, oauth.NewInvalidGrant("Rate limit exceeded."))
			return
		default:
			// Authentication failed.
			klog.Errorf("Authentication failed: %s", err)
			_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
			return
		}
	}

	// Issue token to the authenticated user.
	result, err := h.issueTokenTo(authenticated, client)
	if err != nil {
		// Failed to issue token.
		klog.Errorf("Failed to issue token: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	// Record successful login.
	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(req.Request.Context(), authenticated.GetName(), iamv1beta1.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", authenticated.GetName(), err)
	}

	// Respond with the issued token.
	_ = response.WriteEntity(result)
}

func (h *handler) issueTokenTo(user user.Info, client *oauth.Client) (*oauth.Token, error) {
	accessTokenMaxAge := h.options.Issuer.AccessTokenMaxAge
	accessTokenInactivityTimeout := h.options.Issuer.AccessTokenInactivityTimeout
	if client != nil && client.AccessTokenMaxAgeSeconds > 0 && client.AccessTokenInactivityTimeoutSeconds > 0 {
		accessTokenMaxAge = time.Duration(client.AccessTokenMaxAgeSeconds) * time.Second
		accessTokenInactivityTimeout = time.Duration(client.AccessTokenInactivityTimeoutSeconds) * time.Second
	}

	if !h.options.MultipleLogin {
		if err := h.tokenOperator.RevokeAllUserTokens(user.GetName()); err != nil {
			return nil, err
		}
	}
	accessToken, err := h.tokenOperator.IssueTo(&token.IssueRequest{
		User:      user,
		Claims:    token.Claims{TokenType: token.AccessToken},
		ExpiresIn: accessTokenMaxAge,
	})
	if err != nil {
		return nil, err
	}
	refreshToken, err := h.tokenOperator.IssueTo(&token.IssueRequest{
		User:      user,
		Claims:    token.Claims{TokenType: token.RefreshToken},
		ExpiresIn: accessTokenMaxAge + accessTokenInactivityTimeout,
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
		ExpiresIn:    int(accessTokenMaxAge.Seconds()),
	}
	return &result, nil
}

func (h *handler) refreshTokenGrant(req *restful.Request, response *restful.Response, client *oauth.Client) {
	refreshToken, _ := req.BodyParameter("refresh_token")
	verified, err := h.tokenOperator.Verify(refreshToken)
	if err != nil || verified.TokenType != token.RefreshToken {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant("The refresh token is invalid or expired."))
		return
	}

	authenticated := verified.User
	// update token after registration
	if authenticated.GetName() == iamv1beta1.PreRegistrationUser &&
		authenticated.GetExtra() != nil &&
		len(authenticated.GetExtra()[iamv1beta1.ExtraIdentityProvider]) > 0 &&
		len(authenticated.GetExtra()[iamv1beta1.ExtraUID]) > 0 {

		idp := authenticated.GetExtra()[iamv1beta1.ExtraIdentityProvider][0]
		uid := authenticated.GetExtra()[iamv1beta1.ExtraUID][0]
		queryParam := query.New()
		queryParam.LabelSelector = labels.SelectorFromSet(labels.Set{iamv1beta1.IdentifyProviderLabel: idp, iamv1beta1.OriginUIDLabel: uid}).String()
		users, err := h.im.ListUsers(queryParam)
		if err != nil {
			klog.Errorf("failed to list users: %s", err)
			_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
			return
		}
		if len(users.Items) != 1 {
			if len(users.Items) > 1 {
				klog.Errorf("duplicate user IDs associated: %s/%s", idp, uid)
			}
			_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant("Authenticated user does not exist."))
			return
		}

		authenticated = &user.DefaultInfo{Name: users.Items[0].(*iamv1beta1.User).Name}
	}

	result, err := h.issueTokenTo(authenticated, client)
	if err != nil {
		klog.Errorf("failed to issue token: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	_ = response.WriteEntity(result)
}

func (h *handler) codeGrant(req *restful.Request, response *restful.Response, client *oauth.Client) {
	code, _ := req.BodyParameter("code")
	if code == "" {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest("The authorization code is empty or missing."))
		return
	}
	redirectURI, _ := req.BodyParameter("redirect_uri")
	if _, err := client.ResolveRedirectURL(redirectURI); err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest("Redirect URI is not allowed."))
		return
	}

	authorizeContext, err := h.tokenOperator.Verify(code)
	if err != nil || authorizeContext.TokenType != token.AuthorizationCode {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant("The authorization code is invalid or expired."))
		return
	}

	defer func() {
		// The client MUST NOT use the authorization code more than once.
		if err = h.tokenOperator.Revoke(code); err != nil {
			klog.Warningf("grant: failed to revoke authorization code: %v", err)
		}
	}()

	result, err := h.issueTokenTo(authorizeContext.User, client)
	if err != nil {
		klog.Errorf("failed to issue token: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	// If no openid scope value is present, the request may still be a valid OAuth 2.0 request,
	// but is not an OpenID Connect request.
	if !sliceutil.HasString(authorizeContext.Scopes, oauth.ScopeOpenID) {
		_ = response.WriteEntity(result)
		return
	}

	idTokenRequest, err := h.buildIDTokenIssueRequest(&idTokenRequest{
		authenticated: authorizeContext.User,
		client:        client,
		scopes:        authorizeContext.Scopes,
		nonce:         authorizeContext.Nonce,
	})

	if err != nil {
		klog.Errorf("failed to build id token request: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	idToken, err := h.tokenOperator.IssueTo(idTokenRequest)
	if err != nil {
		klog.Errorf("failed to issue id token: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	result.IDToken = idToken

	_ = response.WriteEntity(result)
}

func (h *handler) buildIDTokenIssueRequest(request *idTokenRequest) (*token.IssueRequest, error) {
	authenticated, err := h.im.DescribeUser(request.authenticated.GetName())
	if err != nil {
		return nil, err
	}

	accessTokenMaxAge := h.options.Issuer.AccessTokenMaxAge
	accessTokenInactivityTimeout := h.options.Issuer.AccessTokenInactivityTimeout
	if request.client != nil && request.client.AccessTokenMaxAgeSeconds > 0 && request.client.AccessTokenInactivityTimeoutSeconds > 0 {
		accessTokenMaxAge = time.Duration(request.client.AccessTokenMaxAgeSeconds) * time.Second
		accessTokenInactivityTimeout = time.Duration(request.client.AccessTokenInactivityTimeoutSeconds) * time.Second
	}

	idTokenRequest := &token.IssueRequest{
		User: request.authenticated,
		Claims: token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Audience: []string{request.client.Name},
			},
			Nonce:     request.nonce,
			TokenType: token.IDToken,
			Name:      request.authenticated.GetName(),
		},
		ExpiresIn: accessTokenMaxAge + accessTokenInactivityTimeout,
	}

	if sliceutil.HasString(request.scopes, oauth.ScopeProfile) {
		idTokenRequest.PreferredUsername = authenticated.Name
		idTokenRequest.Locale = authenticated.Spec.Lang
	}

	if sliceutil.HasString(request.scopes, oauth.ScopeEmail) {
		idTokenRequest.Email = authenticated.Spec.Email
	}
	return idTokenRequest, nil
}

func (h *handler) logout(req *restful.Request, resp *restful.Response) {
	authHeader := strings.TrimSpace(req.Request.Header.Get("Authorization"))
	if authHeader == "" {
		_ = resp.WriteAsJson(serverrors.None)
		return
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
		_ = resp.WriteAsJson(serverrors.None)
		return
	}

	accessToken := parts[1]
	if err := h.tokenOperator.Revoke(accessToken); err != nil {
		reason := fmt.Errorf("failed to revoke access token")
		klog.Errorf("%s: %s", reason, err)
		api.HandleInternalError(resp, req, reason)
		return
	}

	postLogoutRedirectURI := req.QueryParameter("post_logout_redirect_uri")
	if postLogoutRedirectURI == "" {
		_ = resp.WriteAsJson(serverrors.None)
		return
	}

	redirectURL, err := url.Parse(postLogoutRedirectURI)
	if err != nil {
		api.HandleBadRequest(resp, req, fmt.Errorf("invalid logout redirect URI: %s", err))
		return
	}

	state := req.QueryParameter("state")
	if state != "" {
		qry := redirectURL.Query()
		qry.Add("state", state)
		redirectURL.RawQuery = qry.Encode()
	}

	resp.Header().Set("Content-Type", "text/plain")
	http.Redirect(resp, req.Request, redirectURL.String(), http.StatusFound)
}

// userinfo Endpoint is an OAuth 2.0 Protected Resource that returns Claims about the authenticated End-User.
func (h *handler) userinfo(req *restful.Request, response *restful.Response) {
	authenticated, _ := request.UserFrom(req.Request.Context())
	if authenticated == nil || authenticated.GetName() == user.Anonymous {
		_ = response.WriteHeaderAndEntity(http.StatusUnauthorized, oauth.LoginRequired)
		return
	}
	userDetails, err := h.im.DescribeUser(authenticated.GetName())
	if err != nil {
		klog.Errorf("failed to get user details: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	result := token.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userDetails.Name,
		},
		Name:              userDetails.Name,
		Email:             userDetails.Spec.Email,
		Locale:            userDetails.Spec.Lang,
		PreferredUsername: userDetails.Name,
	}
	_ = response.WriteEntity(result)
}

type authCodeRequest struct {
	authenticated user.Info
	clientID      string
	nonce         string
	scopes        []string
	redirectURL   *url.URL
	state         string
}

func (h *handler) handleAuthorizationCodeRequest(req *restful.Request, response *restful.Response, authCodeRequest authCodeRequest) {
	code, err := h.tokenOperator.IssueTo(&token.IssueRequest{
		User: authCodeRequest.authenticated,
		Claims: token.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Audience: []string{authCodeRequest.clientID},
			},
			TokenType: token.AuthorizationCode,
			Nonce:     authCodeRequest.nonce,
			Scopes:    authCodeRequest.scopes,
		},
		// A maximum authorization code lifetime of 10 minutes is
		ExpiresIn: 10 * time.Minute,
	})
	if err != nil {
		klog.Errorf("failed to issue auth code: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}
	values := authCodeRequest.redirectURL.Query()
	values.Add("code", code)
	if authCodeRequest.state != "" {
		values.Add("state", authCodeRequest.state)
	}
	authCodeRequest.redirectURL.RawQuery = values.Encode()
	http.Redirect(response, req.Request, authCodeRequest.redirectURL.String(), http.StatusFound)
}

type idTokenRequest struct {
	authenticated user.Info
	client        *oauth.Client
	nonce         string
	scopes        []string
}

type authIDTokenRequest struct {
	*idTokenRequest
	state       string
	redirectURL *url.URL
}

func (h *handler) handleAuthIDTokenRequest(req *restful.Request, response *restful.Response, authIDTokenRequest *authIDTokenRequest) {
	if authIDTokenRequest.nonce == "" {
		return
	}

	idTokenRequest, err := h.buildIDTokenIssueRequest(authIDTokenRequest.idTokenRequest)
	if err != nil {
		klog.Errorf("failed to build id token request: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
		return
	}

	idToken, err := h.tokenOperator.IssueTo(idTokenRequest)
	if err != nil {
		klog.Errorf("failed to issue id token: %s", err)
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(internalServerErrorMessage))
	}

	values := make(url.Values)
	values.Add("id_token", idToken)
	if authIDTokenRequest.state != "" {
		values.Add("state", authIDTokenRequest.state)
	}
	authIDTokenRequest.redirectURL.Fragment = values.Encode()
	http.Redirect(response, req.Request, authIDTokenRequest.redirectURL.String(), http.StatusFound)
}
