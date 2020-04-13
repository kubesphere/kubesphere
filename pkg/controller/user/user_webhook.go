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

package user

import (
	"context"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strconv"
)

const (
	encryptedAnnotation = "iam.kubesphere.io/password-encrypted"
)

type EmailValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

type PasswordCipher struct {
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

	alreadyExist := emailAlreadyExist(allUsers, user)

	if alreadyExist {
		return admission.Denied("user email already exists")
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

func (a *PasswordCipher) Handle(ctx context.Context, req admission.Request) admission.Response {
	user := &v1alpha2.User{}
	err := a.decoder.Decode(req, user)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	encrypted, err := strconv.ParseBool(user.Annotations[encryptedAnnotation])

	if err != nil || !encrypted {
		password, err := hashPassword(user.Spec.EncryptedPassword)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		user.Spec.EncryptedPassword = password
		user.Annotations[encryptedAnnotation] = "true"
	}

	marshaledUser, err := json.Marshal(user)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledUser)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(bytes), err
}

// InjectDecoder injects the decoder.
func (a *PasswordCipher) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

// InjectDecoder injects the decoder.
func (a *EmailValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
