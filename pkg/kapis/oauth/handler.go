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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/auth"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"net/http"
)

type handler struct {
	im            im.IdentityManagementInterface
	options       *authoptions.AuthenticationOptions
	tokenOperator im.TokenManagementInterface
	authenticator im.PasswordAuthenticator
	loginRecorder im.LoginRecorder
}

func newHandler(im im.IdentityManagementInterface, tokenOperator im.TokenManagementInterface, authenticator im.PasswordAuthenticator, loginRecorder im.LoginRecorder, options *authoptions.AuthenticationOptions) *handler {
	return &handler{im: im, tokenOperator: tokenOperator, authenticator: authenticator, loginRecorder: loginRecorder, options: options}
}

// Implement webhook authentication interface
// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
func (h *handler) TokenReview(req *restful.Request, resp *restful.Response) {
	var tokenReview auth.TokenReview

	err := req.ReadEntity(&tokenReview)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	if err = tokenReview.Validate(); err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	authenticated, err := h.tokenOperator.Verify(tokenReview.Spec.Token)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, req, err)
		return
	}

	success := auth.TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: auth.KindTokenReview,
		Status: &auth.Status{
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
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	if responseType != "token" {
		err := apierrors.NewBadRequest(fmt.Sprintf("Response type %s is not supported", responseType))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	if !ok {
		err := apierrors.NewUnauthorized("Unauthorized")
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	token, err := h.tokenOperator.IssueTo(authenticated)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	redirectURL, err := conf.ResolveRedirectURL(redirectURI)

	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	redirectURL = fmt.Sprintf("%s#access_token=%s&token_type=Bearer", redirectURL, token.AccessToken)

	if token.ExpiresIn > 0 {
		redirectURL = fmt.Sprintf("%s&expires_in=%v", redirectURL, token.ExpiresIn)
	}
	resp.Header().Set("Content-Type", "text/plain")
	http.Redirect(resp, req.Request, redirectURL, http.StatusFound)
}

func (h *handler) oAuthCallBack(req *restful.Request, resp *restful.Response) {

	code := req.QueryParameter("code")
	name := req.PathParameter("callback")

	if code == "" {
		err := apierrors.NewUnauthorized("Unauthorized: missing code")
		resp.WriteError(http.StatusUnauthorized, err)
	}

	providerOptions, err := h.options.OAuthOptions.IdentityProviderOptions(name)

	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.WriteError(http.StatusUnauthorized, err)
	}

	oauthIdentityProvider, err := identityprovider.GetOAuthProvider(providerOptions.Type, providerOptions.Provider)

	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	identity, err := oauthIdentityProvider.IdentityExchange(code)

	if err != nil {
		err = apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	authenticated, err := h.im.DescribeUser(identity.GetName())
	if err != nil {
		// create user if not exist
		if (oauth.MappingMethodAuto == providerOptions.MappingMethod ||
			oauth.MappingMethodMixed == providerOptions.MappingMethod) &&
			apierrors.IsNotFound(err) {
			create := &iamv1alpha2.User{
				ObjectMeta: v1.ObjectMeta{Name: identity.GetName(),
					Annotations: map[string]string{iamv1alpha2.IdentifyProviderLabel: providerOptions.Name}},
				Spec: iamv1alpha2.UserSpec{Email: identity.GetEmail()},
			}
			if authenticated, err = h.im.CreateUser(create); err != nil {
				klog.Error(err)
				api.HandleInternalError(resp, req, err)
				return
			}
		} else {
			klog.Error(err)
			api.HandleInternalError(resp, req, err)
			return
		}
	}

	if oauth.MappingMethodLookup == providerOptions.MappingMethod &&
		authenticated == nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("user %s cannot bound to this identify provider", identity.GetName()))
		klog.Error(err)
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	// oauth.MappingMethodAuto
	// Fails if a user with that user name is already mapped to another identity.
	if providerOptions.MappingMethod == oauth.MappingMethodAuto && authenticated.Annotations[iamv1alpha2.IdentifyProviderLabel] != providerOptions.Name {
		err := apierrors.NewUnauthorized(fmt.Sprintf("user %s is already bound to other identify provider", identity.GetName()))
		klog.Error(err)
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	result, err := h.tokenOperator.IssueTo(&user.DefaultInfo{
		Name: authenticated.Name,
		UID:  string(authenticated.UID),
	})

	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	if err = h.loginRecorder.RecordLogin(authenticated.Name, iamv1alpha2.OAuth, providerOptions.Name, nil, req.Request); err != nil {
		klog.Error(err)
		err := apierrors.NewInternalError(err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *handler) Login(request *restful.Request, response *restful.Response) {
	var loginRequest auth.LoginRequest
	err := request.ReadEntity(&loginRequest)
	if err != nil || loginRequest.Username == "" || loginRequest.Password == "" {
		response.WriteHeaderAndEntity(http.StatusUnauthorized, fmt.Errorf("empty username or password"))
		return
	}
	h.passwordGrant(loginRequest.Username, loginRequest.Password, request, response)
}

func (h *handler) Token(req *restful.Request, response *restful.Response) {
	grantType, err := req.BodyParameter("grant_type")
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, req, err)
		return
	}
	switch grantType {
	case "password":
		username, err := req.BodyParameter("username")
		if err != nil {
			klog.Error(err)
			api.HandleBadRequest(response, req, err)
			return
		}
		password, err := req.BodyParameter("password")
		if err != nil {
			klog.Error(err)
			api.HandleBadRequest(response, req, err)
			return
		}
		h.passwordGrant(username, password, req, response)
		break
	case "refresh_token":
		h.refreshTokenGrant(req, response)
		break
	default:
		err := apierrors.NewBadRequest(fmt.Sprintf("Grant type %s is not supported", grantType))
		response.WriteError(http.StatusBadRequest, err)
	}
}

func (h *handler) passwordGrant(username string, password string, req *restful.Request, response *restful.Response) {
	authenticated, err := h.authenticator.Authenticate(username, password)
	if err != nil {
		klog.Error(err)
		switch err {
		case im.AuthFailedIncorrectPassword:
			if err := h.loginRecorder.RecordLogin(username, iamv1alpha2.Token, "", err, req.Request); err != nil {
				klog.Error(err)
				response.WriteError(http.StatusInternalServerError, apierrors.NewInternalError(err))
				return
			}
			response.WriteError(http.StatusUnauthorized, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
			return
		case im.AuthFailedIdentityMappingNotMatch:
			response.WriteError(http.StatusUnauthorized, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
			return
		case im.AuthRateLimitExceeded:
			response.WriteError(http.StatusTooManyRequests, apierrors.NewTooManyRequests(fmt.Sprintf("Unauthorized: %s", err), 60))
			return
		default:
			response.WriteError(http.StatusInternalServerError, apierrors.NewInternalError(err))
			return
		}
	}

	result, err := h.tokenOperator.IssueTo(authenticated)
	if err != nil {
		klog.Error(err)
		response.WriteError(http.StatusInternalServerError, apierrors.NewInternalError(err))
		return
	}

	if err = h.loginRecorder.RecordLogin(authenticated.GetName(), iamv1alpha2.Token, "", nil, req.Request); err != nil {
		klog.Error(err)
		response.WriteError(http.StatusInternalServerError, apierrors.NewInternalError(err))
		return
	}

	response.WriteEntity(result)
}

func (h *handler) refreshTokenGrant(req *restful.Request, response *restful.Response) {
	refreshToken, err := req.BodyParameter("refresh_token")
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, req, err)
		return
	}

	authenticated, err := h.tokenOperator.Verify(refreshToken)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		response.WriteError(http.StatusUnauthorized, err)
		return
	}

	result, err := h.tokenOperator.IssueTo(authenticated)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		response.WriteError(http.StatusUnauthorized, err)
		return
	}

	response.WriteEntity(result)
}
