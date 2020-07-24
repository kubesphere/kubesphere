/*
Copyright 2014 The Kubernetes Authors.

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

// Following code copied from k8s.io/apiserver/pkg/authorization/authorizerfactory to avoid import collision

package basictoken

import (
	"errors"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"net/http"
)

type Authenticator struct {
	auth authenticator.Password
}

func New(auth authenticator.Password) *Authenticator {
	return &Authenticator{auth}
}

var invalidToken = errors.New("invalid basic token")

func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {

	username, password, ok := req.BasicAuth()

	if !ok {
		return nil, false, nil
	}

	resp, ok, err := a.auth.AuthenticatePassword(req.Context(), username, password)

	// If the token authenticator didn't error, provide a default error
	if !ok && err == nil {
		err = invalidToken
	}

	return resp, ok, err
}
