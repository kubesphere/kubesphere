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

package am

import (
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

type fakeRole struct {
	Name string
	Rego string
}
type fakeOperator struct {
	cache cache.Interface
}

func newFakeRole(username string) Role {
	if username == user.Anonymous {
		return &fakeRole{
			Name: "anonymous",
			Rego: "package authz\ndefault allow = false",
		}
	}
	return &fakeRole{
		Name: "admin",
		Rego: "package authz\ndefault allow = true",
	}
}

func (f fakeOperator) GetPlatformRole(username string) (Role, error) {
	return newFakeRole(username), nil
}

func (f fakeOperator) GetClusterRole(cluster, username string) (Role, error) {
	return newFakeRole(username), nil
}

func (f fakeOperator) GetWorkspaceRole(workspace, username string) (Role, error) {
	return newFakeRole(username), nil
}

func (f fakeOperator) GetNamespaceRole(namespace, username string) (Role, error) {
	return newFakeRole(username), nil
}

func (f fakeRole) GetName() string {
	return f.Name
}

func (f fakeRole) GetRego() string {
	return f.Rego
}

func NewFakeAMOperator(cache cache.Interface) AccessManagementInterface {
	return &fakeOperator{cache: cache}
}
