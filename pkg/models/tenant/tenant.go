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
package tenant

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
)

type Interface interface {
	ListWorkspaces(username string) (*api.ListResult, error)
	ListNamespaces(username, workspace string) (*api.ListResult, error)
}

type tenantOperator struct {
	informers informers.InformerFactory
	am        am.AccessManagementInterface
}

func New(k8sClient k8s.Client, informers informers.InformerFactory) Interface {
	return &tenantOperator{
		informers: informers,
		am:        am.NewAMOperator(k8sClient.KubeSphere(), informers.KubeSphereSharedInformerFactory()),
	}
}

func (t *tenantOperator) ListWorkspaces(username string) (*api.ListResult, error) {

	workspaceRoles, err := t.am.ListRolesOfUser(iamv1alpha2.WorkspaceScope, username)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	workspaces := make([]*tenantv1alpha1.Workspace, 0)

	for _, role := range workspaceRoles {

		// all workspaces are allowed
		if role.Target.Name == iamv1alpha2.TargetAll {
			workspaces, err = t.informers.KubeSphereSharedInformerFactory().
				Tenant().V1alpha1().Workspaces().Lister().List(labels.Everything())
			break
		}
		workspace, err := t.informers.KubeSphereSharedInformerFactory().
			Tenant().V1alpha1().Workspaces().Lister().Get(role.Target.Name)

		if errors.IsNotFound(err) {
			klog.Warningf("workspace role: %s found but workspace not exist", role.Target)
			continue
		}
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if !containsWorkspace(workspaces, workspace) {
			workspaces = append(workspaces, workspace)
		}
	}

	return &api.ListResult{
		TotalItems: len(workspaces),
		Items:      workspacesToInterfaces(workspaces),
	}, nil
}

func (t *tenantOperator) ListNamespaces(username, workspace string) (*api.ListResult, error) {

	namespaceRoles, err := t.am.ListRolesOfUser(iamv1alpha2.NamespaceScope, username)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	namespaces := make([]*corev1.Namespace, 0)

	for _, role := range namespaceRoles {

		// all workspaces are allowed
		if role.Target.Name == iamv1alpha2.TargetAll {
			namespaces, err = t.informers.KubernetesSharedInformerFactory().
				Core().V1().Namespaces().Lister().List(labels.Everything())
			break
		}

		namespace, err := t.informers.KubernetesSharedInformerFactory().
			Core().V1().Namespaces().Lister().Get(role.Target.Name)

		if errors.IsNotFound(err) {
			klog.Warningf("workspace role: %s found but workspace not exist", role.Target)
			continue
		}
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if !containsNamespace(namespaces, namespace) {
			namespaces = append(namespaces, namespace)
		}
	}

	return &api.ListResult{
		TotalItems: len(namespaces),
		Items:      namespacesToInterfaces(namespaces),
	}, nil
}

func containsWorkspace(workspaces []*tenantv1alpha1.Workspace, workspace *tenantv1alpha1.Workspace) bool {
	for _, item := range workspaces {
		if item.Name == workspace.Name {
			return true
		}
	}
	return false
}

func containsNamespace(namespaces []*corev1.Namespace, namespace *corev1.Namespace) bool {
	for _, item := range namespaces {
		if item.Name == namespace.Name {
			return true
		}
	}
	return false
}

func workspacesToInterfaces(workspaces []*tenantv1alpha1.Workspace) []interface{} {
	ret := make([]interface{}, len(workspaces))
	for index, v := range workspaces {
		ret[index] = v
	}
	return ret
}

func namespacesToInterfaces(namespaces []*corev1.Namespace) []interface{} {
	ret := make([]interface{}, len(namespaces))
	for index, v := range namespaces {
		ret[index] = v
	}
	return ret
}
