/*
Copyright 2020 KubeSphere Authors

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
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/client-go/client"
)

var URLOptions = client.URLOptions{
	Group:   "iam.kubesphere.io",
	Version: "v1alpha2",
}

// NewUser returns a User spec with the specified argument.
func NewUser(name, globelRole string) *iamv1alpha2.User {
	return &iamv1alpha2.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"iam.kubesphere.io/globalrole": globelRole,
			},
		},
		Spec: iamv1alpha2.UserSpec{
			Email:             fmt.Sprintf("%s@kubesphere.io", name),
			EncryptedPassword: "P@88w0rd",
		},
	}
}

// CreateUser uses c to create User. If the returned error is nil, the returned User is valid and has
// been created.
func CreateUser(c client.Client, u *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	err := c.Create(context.TODO(), u, &URLOptions)
	return u, err
}

// GetUser uses c to get the User by name. If the returned error is nil, the returned User is valid.
func GetUser(c client.Client, name string) (*iamv1alpha2.User, error) {
	u := &iamv1alpha2.User{}

	err := c.Get(context.TODO(), client.ObjectKey{Name: name}, u, &URLOptions)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// NewGroup returns a Group spec with the specified argument.
func NewGroup(name, workspace string) *iamv1alpha2.Group {
	return &iamv1alpha2.Group{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
		},
	}
}

// CreateGroup uses c to create Group. If the returned error is nil, the returned Group is valid and has
// been created.
func CreateGroup(c client.Client, u *iamv1alpha2.Group, workspace string) (*iamv1alpha2.Group, error) {
	err := c.Create(context.TODO(), u, &URLOptions, &client.WorkspaceOptions{Name: workspace})
	return u, err
}

// GetGroup uses c to get the User by name. If the returned error is nil, the returned User is valid.
func GetGroup(c client.Client, name, workspace string) (*iamv1alpha2.Group, error) {
	u := &iamv1alpha2.Group{}

	err := c.Get(context.TODO(), client.ObjectKey{Name: name}, u, &URLOptions, &client.WorkspaceOptions{Name: workspace})
	if err != nil {
		return nil, err
	}
	return u, nil
}
