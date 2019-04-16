/*

 Copyright 2019 The KubeSphere Authors.

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
package iam

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
	"kubesphere.io/kubesphere/pkg/utils/jwtutil"
	"net/http"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/iam"
)

type Spec struct {
	Token string `json:"token"`
}

type Status struct {
	Authenticated bool                   `json:"authenticated"`
	User          map[string]interface{} `json:"user,omitempty"`
}

type TokenReview struct {
	APIVersion string  `json:"apiVersion"`
	Kind       string  `json:"kind"`
	Spec       *Spec   `json:"spec,omitempty"`
	Status     *Status `json:"status,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	APIVersion      = "authentication.k8s.io/v1beta1"
	KindTokenReview = "TokenReview"
)

func LoginHandler(req *restful.Request, resp *restful.Response) {
	var loginRequest LoginRequest

	err := req.ReadEntity(&loginRequest)

	if err != nil || loginRequest.Username == "" || loginRequest.Password == "" {
		resp.WriteHeaderAndEntity(http.StatusUnauthorized, errors.New("incorrect username or password"))
		return
	}

	ip := iputil.RemoteIp(req.Request)

	token, err := iam.Login(loginRequest.Username, loginRequest.Password, ip)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusUnauthorized, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(token)
}

// k8s token review
func TokenReviewHandler(req *restful.Request, resp *restful.Response) {
	var tokenReview TokenReview

	err := req.ReadEntity(&tokenReview)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if tokenReview.Spec == nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New("token must not be null"))
		return
	}

	uToken := tokenReview.Spec.Token

	token, err := jwtutil.ValidateToken(uToken)

	if err != nil {
		glog.Errorln("token review failed", uToken, err)
		failed := TokenReview{APIVersion: APIVersion,
			Kind: KindTokenReview,
			Status: &Status{
				Authenticated: false,
			},
		}
		resp.WriteAsJson(failed)
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	username, ok := claims["username"].(string)

	if !ok {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.New("username not found"))
		return
	}

	user, err := iam.GetUserInfo(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	groups, err := iam.GetUserGroups(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	user.Groups = groups

	success := TokenReview{APIVersion: APIVersion,
		Kind: KindTokenReview,
		Status: &Status{
			Authenticated: true,
			User:          map[string]interface{}{"username": user.Username, "uid": user.Username, "groups": user.Groups},
		},
	}

	resp.WriteAsJson(success)
	return
}
