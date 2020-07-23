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

package user

import (
	"context"
	"fmt"
	"kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"net/http"
	"net/mail"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type EmailValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (a *EmailValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	user := &v1alpha2.User{}
	err := a.decoder.Decode(req, user)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	allUsers := v1alpha2.UserList{}

	err = a.Client.List(ctx, &allUsers, &client.ListOptions{})

	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if _, err := mail.ParseAddress(user.Spec.Email); user.Spec.Email != "" && err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid email address:%s", user.Spec.Email))
	}

	alreadyExist := emailAlreadyExist(allUsers, user)

	if alreadyExist {
		return admission.Errored(http.StatusConflict, fmt.Errorf("user email: %s already exists", user.Spec.Email))
	}

	return admission.Allowed("")
}

func emailAlreadyExist(users v1alpha2.UserList, user *v1alpha2.User) bool {
	for _, exist := range users.Items {
		if exist.Spec.Email == user.Spec.Email && exist.Name != user.Name {
			return true
		}
	}
	return false
}

// InjectDecoder injects the decoder.
func (a *EmailValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
