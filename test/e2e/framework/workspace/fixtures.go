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

package workspace

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
	fedb1 "kubesphere.io/api/types/v1beta1"

	"context"

	"kubesphere.io/client-go/client"
)

var URLOptions = client.URLOptions{
	Group:   "tenant.kubesphere.io",
	Version: "v1alpha2",
}

// NewWorkspaceTemplate returns a WorkspaceTemplate spec with the specified argument.
func NewWorkspaceTemplate(name string, manager string, hosts ...string) *tenantv1alpha2.WorkspaceTemplate {

	clusters := []fedb1.GenericClusterReference{}

	if hosts != nil {
		for _, h := range hosts {
			clusters = append(clusters, fedb1.GenericClusterReference{Name: h})
		}
	}

	return &tenantv1alpha2.WorkspaceTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: fedb1.FederatedWorkspaceSpec{
			Placement: fedb1.GenericPlacementFields{
				Clusters: clusters,
			},
			Template: fedb1.WorkspaceTemplate{
				Spec: tenantv1alpha1.WorkspaceSpec{
					Manager: manager,
				},
			},
		},
	}
}

// CreateWorkSpace uses c to create Workspace. If the returned error is nil, the returned Workspace is valid and has
// been created.
func CreateWorkspace(c client.Client, w *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha2.WorkspaceTemplate, error) {
	opts := &client.URLOptions{
		AbsPath: "kapis/tenant.kubesphere.io/v1alpha2/workspaces",
	}
	err := c.Create(context.TODO(), w, opts)
	return w, err
}

// GetJob uses c to get the Workspace by name. If the returned error is nil, the returned Workspace is valid.
func GetWorkspace(c client.Client, name string) (*tenantv1alpha1.Workspace, error) {
	wsp := &tenantv1alpha1.Workspace{}

	err := c.Get(context.TODO(), client.ObjectKey{Name: name}, wsp, &URLOptions)
	if err != nil {
		return nil, err
	}
	return wsp, nil
}

// DeleteWorkspace uses c to delete the Workspace by name. If the returned error is nil, the returned Workspace is valid.
func DeleteWorkspace(c client.Client, name string, opts ...client.DeleteOption) (*tenantv1alpha1.Workspace, error) {
	wsp := &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	opts = append(opts, &URLOptions)
	err := c.Delete(context.TODO(), wsp, opts...)
	if err != nil {
		return nil, err
	}
	return wsp, nil
}
