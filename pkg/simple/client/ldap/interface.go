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

package ldap

import (
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

// Interface defines CRUD behaviors of manipulating users
type Interface interface {
	// Create create a new user in ldap
	Create(user *iamv1alpha2.User) error

	// Update updates a user information, return error if user not exists
	Update(user *iamv1alpha2.User) error

	// Delete deletes a user from ldap, return nil if user not exists
	Delete(name string) error

	// Get gets a user by its username from ldap, return ErrUserNotExists if user not exists
	Get(name string) (*iamv1alpha2.User, error)

	// Authenticate checks if (name, password) is valid, return ErrInvalidCredentials if not
	Authenticate(name string, password string) error

	List(query *query.Query) (*api.ListResult, error)
}
