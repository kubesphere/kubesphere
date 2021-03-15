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

package v1alpha2

import (
	"context"

	resty "github.com/go-resty/resty/v2"
)

type RoleBindingsGetter interface {
	RoleBindings() RoleBindingInterface
}

type RoleBindingInterface interface {
	CreateRoleBinding(ctx context.Context, namespace, role, group string) (string, error)
	CreateWorkspaceRoleBinding(ctx context.Context, namespace, role, group string) (string, error)
}

type rolebindings struct {
	client *resty.Client
}

func newRoleBindings(c *IamV1alpha2Client) *rolebindings {
	return &rolebindings{
		client: c.client,
	}
}

// CreateRoleBinding assembling of a rolebinding object and creates it. Returns the server's response and an error, if there is any.
func (c *rolebindings) CreateRoleBinding(ctx context.Context, namespace, role, group string) (result string, err error) {

	roles := []map[string]interface{}{{
		"subjects": []map[string]interface{}{
			{
				"kind":     "Group",
				"apiGroup": "rbac.authorization.k8s.io",
				"name":     group,
			},
		},
		"roleRef": map[string]interface{}{
			"apiGroup": "rbac.authorization.k8s.io",
			"kind":     "Role",
			"name":     role,
		},
	}}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(roles).
		SetPathParams(map[string]string{
			"namespace": namespace,
		}).
		Post("/kapis/iam.kubesphere.io/v1alpha2/namespaces/{namespace}/rolebindings")
	return resp.String(), err
}

// CreateWorkspaceRoleBinding assembling of a workspacerolebinding object and creates it. Returns the server's response, and an error, if there is any.
func (c *rolebindings) CreateWorkspaceRoleBinding(ctx context.Context, workspace, role, group string) (result string, err error) {

	roles := []map[string]interface{}{{
		"subjects": []map[string]interface{}{
			{
				"kind":     "Group",
				"apiGroup": "rbac.authorization.k8s.io",
				"name":     group,
			},
		},
		"roleRef": map[string]interface{}{
			"apiGroup": "iam.kubesphere.io/v1alpha2",
			"kind":     "WorkspaceRoleBinding",
			"name":     role,
		},
	}}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(roles).
		SetPathParams(map[string]string{
			"workspace": workspace,
		}).
		Post("/kapis/iam.kubesphere.io/v1alpha2/workspaces/{workspace}/workspacerolebindings/")
	return resp.String(), err
}
