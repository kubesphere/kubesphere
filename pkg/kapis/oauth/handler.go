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
	"github.com/emicklei/go-restful"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/auth"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"net/http"
)

const (
	KindTokenReview       = "TokenReview"
	passwordGrantType     = "password"
	refreshTokenGrantType = "refresh_token"
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

	authenticated, err := h.tokenOperator.Verify(tokenReview.Spec.Token)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	success := TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: KindTokenReview,
		Status: &Status{
			Authenticated: true,
			User:          map[string]interface{}{"username": authenticated.GetName(), "uid": authenticated.GetUID()},
		},
	}

	resp.WriteEntity(success)
}

func (h *handler) Authorize(req *restful.Request, resp *restful.Response) {
	authenticated, ok := request.UserFrom(req.Request.Context())
	clientId := req.QueryParameter("client_id")
	responseType := req.QueryParameter("response_type")
	redirectURI := req.QueryParameter("redirect_uri")

	conf, err := h.options.OAuthOptions.OAuthClient(clientId)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		api.HandleError(resp, req, err)
		return
	}

	if responseType != "token" {
		err := apierrors.NewBadRequest(fmt.Sprintf("Response type %s is not supported", responseType))
		api.HandleError(resp, req, err)
		return
	}

	if !ok {
		err := apierrors.NewUnauthorized("Unauthorized")
		api.HandleError(resp, req, err)
		return
	}

	token, err := h.tokenOperator.IssueTo(authenticated)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		api.HandleError(resp, req, err)
		return
	}

	redirectURL, err := conf.ResolveRedirectURL(redirectURI)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		api.HandleError(resp, req, err)
		return
	}

	redirectURL = fmt.Sprintf("%s#access_token=%s&token_type=Bearer", redirectURL, token.AccessToken)

	if token.ExpiresIn > 0 {
		redirectURL = fmt.Sprintf("%s&expires_in=%v", redirectURL, token.ExpiresIn)
	}
	resp.Header().Set("Content-Type", "text/plain")
	http.Redirect(resp, req.Request, redirectURL, http.StatusFound)
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
		err := apierrors.NewUnauthorized("Unauthorized: missing code")
		api.HandleError(resp, req, err)
		return
	}

	authenticated, provider, err := h.oauth2Authenticator.Authenticate(provider, code)
	if err != nil {
		api.HandleUnauthorized(resp, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
		return
	}

	result, err := h.tokenOperator.IssueTo(authenticated)
	if err != nil {
		api.HandleInternalError(resp, req, apierrors.NewInternalError(err))
		return
	}

	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(authenticated.GetName(), iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", authenticated.GetName(), err)
	}

	resp.WriteEntity(result)
}

func (h *handler) Login(request *restful.Request, response *restful.Response) {
	var loginRequest LoginRequest
	err := request.ReadEntity(&loginRequest)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	h.passwordGrant(loginRequest.Username, loginRequest.Password, request, response)
}

func (h *handler) Token(req *restful.Request, response *restful.Response) {
	grantType, err := req.BodyParameter("grant_type")
	if err != nil {
		api.HandleBadRequest(response, req, err)
		return
	}
	switch grantType {
	case passwordGrantType:
		username, _ := req.BodyParameter("username")
		password, _ := req.BodyParameter("password")
		h.passwordGrant(username, password, req, response)
		break
	case refreshTokenGrantType:
		h.refreshTokenGrant(req, response)
		break
	default:
		err := apierrors.NewBadRequest(fmt.Sprintf("Grant type %s is not supported", grantType))
		api.HandleBadRequest(response, req, err)
	}
}

func (h *handler) passwordGrant(username string, password string, req *restful.Request, response *restful.Response) {
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

	result, err := h.tokenOperator.IssueTo(authenticated)
	if err != nil {
		api.HandleInternalError(response, req, apierrors.NewInternalError(err))
		return
	}

	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(authenticated.GetName(), iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", username, err)
	}

	response.WriteEntity(result)
}

func (h *handler) refreshTokenGrant(req *restful.Request, response *restful.Response) {
	refreshToken, err := req.BodyParameter("refresh_token")
	if err != nil {
		api.HandleBadRequest(response, req, apierrors.NewBadRequest(err.Error()))
		return
	}

	authenticated, err := h.tokenOperator.Verify(refreshToken)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		api.HandleUnauthorized(response, req, apierrors.NewUnauthorized(err.Error()))
		return
	}

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
			api.HandleInternalError(response, req, apierrors.NewInternalError(err))
			return
		}
		if len(result.Items) != 1 {
			err := apierrors.NewUnauthorized("authenticated user does not exist")
			api.HandleUnauthorized(response, req, apierrors.NewUnauthorized(err.Error()))
			return
		}
		authenticated = &user.DefaultInfo{Name: result.Items[0].(*iamv1alpha2.User).Name}
	}

	result, err := h.tokenOperator.IssueTo(authenticated)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		api.HandleUnauthorized(response, req, apierrors.NewUnauthorized(err.Error()))
		return
	}

	response.WriteEntity(result)
}
