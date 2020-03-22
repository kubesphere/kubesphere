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
	"encoding/json"
	"fmt"
	"k8s.io/apiserver/pkg/authentication/user"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

type FakeRole struct {
	Name string
	Rego string
}
type FakeOperator struct {
	cache cache.Interface
}

func (f FakeOperator) queryFakeRole(cacheKey string) (Role, error) {
	data, err := f.cache.Get(cacheKey)
	if err != nil {
		if err == cache.ErrNoSuchKey {
			return &FakeRole{
				Name: "DenyAll",
				Rego: "package authz\ndefault allow = false",
			}, nil
		}
		return nil, err
	}
	var role FakeRole
	err = json.Unmarshal([]byte(data), &role)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (f FakeOperator) saveFakeRole(cacheKey string, role FakeRole) error {
	data, err := json.Marshal(role)
	if err != nil {
		return err
	}
	return f.cache.Set(cacheKey, string(data), 0)
}

func (f FakeOperator) GetPlatformRole(username string) (Role, error) {
	return f.queryFakeRole(platformRoleCacheKey(username))
}

func (f FakeOperator) GetClusterRole(cluster, username string) (Role, error) {
	return f.queryFakeRole(clusterRoleCacheKey(cluster, username))
}

func (f FakeOperator) GetWorkspaceRole(workspace, username string) (Role, error) {
	return f.queryFakeRole(workspaceRoleCacheKey(workspace, username))
}

func (f FakeOperator) GetNamespaceRole(cluster, namespace, username string) (Role, error) {
	return f.queryFakeRole(namespaceRoleCacheKey(cluster, namespace, username))
}

func (f FakeOperator) Prepare(platformRoles map[string]FakeRole, clusterRoles map[string]map[string]FakeRole, workspaceRoles map[string]map[string]FakeRole, namespaceRoles map[string]map[string]map[string]FakeRole) {

	for username, role := range platformRoles {
		f.saveFakeRole(platformRoleCacheKey(username), role)
	}
	for cluster, roles := range clusterRoles {
		for username, role := range roles {
			f.saveFakeRole(clusterRoleCacheKey(cluster, username), role)
		}
	}

	for workspace, roles := range workspaceRoles {
		for username, role := range roles {
			f.saveFakeRole(workspaceRoleCacheKey(workspace, username), role)
		}
	}

	for cluster, nsRoles := range namespaceRoles {
		for namespace, roles := range nsRoles {
			for username, role := range roles {
				f.saveFakeRole(namespaceRoleCacheKey(cluster, namespace, username), role)
			}
		}
	}
}

func namespaceRoleCacheKey(cluster, namespace, username string) string {
	return fmt.Sprintf("cluster.%s.namespaces.%s.roles.%s", cluster, namespace, username)
}

func clusterRoleCacheKey(cluster, username string) string {
	return fmt.Sprintf("cluster.%s.roles.%s", cluster, username)
}
func workspaceRoleCacheKey(workspace, username string) string {
	return fmt.Sprintf("workspace.%s.roles.%s", workspace, username)
}

func platformRoleCacheKey(username string) string {
	return fmt.Sprintf("platform.roles.%s", username)
}

func (f FakeRole) GetName() string {
	return f.Name
}

func (f FakeRole) GetRego() string {
	return f.Rego
}

func NewFakeAMOperator() *FakeOperator {
	operator := &FakeOperator{cache: cache.NewSimpleCache()}
	operator.saveFakeRole(platformRoleCacheKey("admin"), FakeRole{
		Name: "admin",
		Rego: "package authz\ndefault allow = true",
	})
	operator.saveFakeRole(platformRoleCacheKey(user.Anonymous), FakeRole{
		Name: "admin",
		Rego: "package authz\ndefault allow = false",
	})
	return operator
}
