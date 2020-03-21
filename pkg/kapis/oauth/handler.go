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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/auth"
	"kubesphere.io/kubesphere/pkg/api/auth/token"
)

type oauthHandler struct {
	issuer token.Issuer
}

func newOAUTHHandler(issuer token.Issuer) *oauthHandler {
	return &oauthHandler{issuer: issuer}
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
			User:          map[string]interface{}{"username": user.GetName(), "uid": user.GetUID(), "groups": user.GetGroups()},
		},
	}

	resp.WriteEntity(success)
}
