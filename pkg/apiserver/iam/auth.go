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
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/models"
	"net/http"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/utils"
	jwtutils "kubesphere.io/kubesphere/pkg/utils/jwt"
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
		resp.WriteHeaderAndEntity(http.StatusUnauthorized, errors.Wrap(fmt.Errorf("incorrect username or password")))
		return
	}

	ip := utils.RemoteIp(req.Request)

	token, err := iam.Login(loginRequest.Username, loginRequest.Password, ip)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusUnauthorized, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(models.Token{Token: token})
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
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(fmt.Errorf("token must not be null")))
		return
	}

	uToken := tokenReview.Spec.Token

	token, err := jwtutils.ValidateToken(uToken)

	if err != nil {
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

	username := claims["username"].(string)

	conn, err := iam.NewConnection()

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	defer conn.Close()

	user, err := iam.UserDetail(username, conn)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

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
