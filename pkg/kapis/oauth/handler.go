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
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/auth"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"net/http"
)

type oauthHandler struct {
	issuer  token.Issuer
	im      im.IdentityManagementInterface
	options *authoptions.AuthenticationOptions
}

func newOAUTHHandler(im im.IdentityManagementInterface, issuer token.Issuer, options *authoptions.AuthenticationOptions) *oauthHandler {
	return &oauthHandler{im: im, issuer: issuer, options: options}
}

// Implement webhook authentication interface
// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
func (h *oauthHandler) TokenReviewHandler(req *restful.Request, resp *restful.Response) {
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

	user, err := h.issuer.Verify(tokenReview.Spec.Token)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, req, err)
		return
	}

	success := auth.TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: auth.KindTokenReview,
		Status: &auth.Status{
			Authenticated: true,
			User:          map[string]interface{}{"username": user.GetName(), "uid": user.GetUID()},
		},
	}

	resp.WriteEntity(success)
}

func (h *oauthHandler) AuthorizeHandler(req *restful.Request, resp *restful.Response) {
	user, ok := request.UserFrom(req.Request.Context())
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
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: response type %s is not supported", responseType))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	if !ok {
		err := apierrors.NewUnauthorized("Unauthorized")
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	token, err := h.issueTo(user.GetName())
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

func (h *oauthHandler) OAuthCallBackHandler(req *restful.Request, resp *restful.Response) {

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

	user, err := oauthIdentityProvider.IdentityExchange(code)

	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	existed, err := h.im.DescribeUser(user.GetName())
	if err != nil {
		// create user if not exist
		if apierrors.IsNotFound(err) && oauth.MappingMethodAuto == providerOptions.MappingMethod {
			create := &iamv1alpha2.User{
				ObjectMeta: v1.ObjectMeta{Name: user.GetName(),
					Annotations: map[string]string{iamv1alpha2.IdentifyProviderLabel: providerOptions.Name}},
				Spec: iamv1alpha2.UserSpec{Email: user.GetEmail()},
			}
			if existed, err = h.im.CreateUser(create); err != nil {
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

	// oauth.MappingMethodLookup
	if existed == nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("user %s cannot bound to this identify provider", user.GetName()))
		klog.Error(err)
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	// oauth.MappingMethodAuto
	// Fails if a user with that user name is already mapped to another identity.
	if providerOptions.MappingMethod == oauth.MappingMethodMixed || existed.Annotations[iamv1alpha2.IdentifyProviderLabel] != providerOptions.Name {
		err := apierrors.NewUnauthorized(fmt.Sprintf("user %s is already bound to other identify provider", user.GetName()))
		klog.Error(err)
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}

	result, err := h.issueTo(user.GetName())
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		resp.WriteError(http.StatusUnauthorized, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *oauthHandler) Login(request *restful.Request, response *restful.Response) {
	var loginRequest auth.LoginRequest

	err := request.ReadEntity(&loginRequest)
	if err != nil || loginRequest.Username == "" || loginRequest.Password == "" {
		response.WriteHeaderAndEntity(http.StatusUnauthorized, fmt.Errorf("empty username or password"))
		return
	}

	user, err := h.im.Authenticate(loginRequest.Username, loginRequest.Password)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		response.WriteError(http.StatusUnauthorized, err)
		return
	}

	result, err := h.issueTo(user.Name)
	if err != nil {
		err := apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err))
		response.WriteError(http.StatusUnauthorized, err)
		return
	}
	response.WriteEntity(result)
}

func (h *oauthHandler) issueTo(username string) (*oauth.Token, error) {
	expiresIn := h.options.OAuthOptions.AccessTokenMaxAge

	accessToken, err := h.issuer.IssueTo(&authuser.DefaultInfo{
		Name: username,
	}, expiresIn)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	result := &oauth.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(expiresIn.Seconds()),
	}

	return result, nil
}
