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
package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/errors"
	secret "kubesphere.io/kubesphere/pkg/iam/jwt"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/utils"
)

type Token struct {
	Token string `json:"access_token"`
}

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
		errors.HandleError(errors.New(errors.Unauthorized, "incorrect username or password"), resp)
		return
	}

	ip := utils.RemoteIp(req.Request)

	token, err := iam.Login(loginRequest.Username, loginRequest.Password, ip)

	if errors.HandleError(err, resp) {
		return
	}

	resp.WriteEntity(Token{Token: token})
}

// k8s token review
func TokenReviewHandler(req *restful.Request, resp *restful.Response) {
	var tokenReview TokenReview

	err := req.ReadEntity(&tokenReview)

	if err != nil {
		errors.HandleError(errors.New(errors.InvalidArgument, err.Error()), resp)
		return
	} else if tokenReview.Spec != nil {
		uToken := tokenReview.Spec.Token

		token, err := secret.ValidateToken(uToken)

		if err == nil {
			claims := token.Claims.(jwt.MapClaims)

			username := claims["username"].(string)

			conn, err := iam.NewConnection()

			if errors.HandleError(err, resp) {
				return
			}

			defer conn.Close()

			user, err := iam.UserDetail(username, conn)

			if errors.HandleError(err, resp) {
				return
			} else {
				success := TokenReview{APIVersion: APIVersion,
					Kind: KindTokenReview,
					Status: &Status{
						Authenticated: true,
						User:          map[string]interface{}{"username": user.Username, "uid": user.Username, "groups": user.Groups},
					},
				}
				resp.WriteEntity(success)
				return
			}
		} else {
			failed := TokenReview{APIVersion: APIVersion,
				Kind: KindTokenReview,
				Status: &Status{
					Authenticated: false,
				},
			}

			resp.WriteEntity(failed)
			return
		}
	} else {
		errors.HandleError(errors.New(errors.InvalidArgument, "token must not be null"), resp)
		return
	}
}
