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

package tenant

import (
	"github.com/google/go-cmp/cmp"
	fakeapp "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned/fake"
	fakeistio "istio.io/client-go/pkg/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"testing"
)

func TestTenantOperator_ListWorkspaces(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		name        string
		result      *api.ListResult
		username    string
		expectError error
	}{
		{
			name:     "list workspace",
			username: "admin",
			result: &api.ListResult{
				Items:      workspaces,
				TotalItems: len(workspaces),
			},
		},
		{
			name:     "list workspaces",
			username: "regular",
			result: &api.ListResult{
				Items:      []interface{}{workspaceBar},
				TotalItems: 1,
			},
		},
	}

	for _, test := range tests {
		result, err := tenantOperator.ListWorkspaces(&user.DefaultInfo{Name: test.username}, query.New())

		if err != nil {
			if test.expectError != err {
				t.Error(err)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Error(diff)
		}
	}
}

func TestTenantOperator_ListNamespaces(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		name        string
		result      *api.ListResult
		username    string
		workspace   string
		expectError error
	}{
		{
			name:      "list namespaces",
			workspace: "foo",
			username:  "admin",
			result: &api.ListResult{
				Items:      []interface{}{foo2, foo1},
				TotalItems: 2,
			},
		},
		{
			name:      "list namespaces",
			workspace: "foo",
			username:  "regular",
			result: &api.ListResult{
				Items:      []interface{}{},
				TotalItems: 0,
			},
		},
		{
			name:      "list namespaces",
			workspace: "bar",
			username:  "regular",
			result: &api.ListResult{
				Items:      []interface{}{bar1},
				TotalItems: 1,
			},
		},
	}

	for _, test := range tests {
		result, err := tenantOperator.ListNamespaces(&user.DefaultInfo{Name: test.username}, test.workspace, query.New())

		if err != nil {
			if test.expectError != err {
				t.Error(err)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Error(diff)
		}
	}
}

var (
	foo1 = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "foo1",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: "foo"},
		},
	}

	foo2 = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "foo2",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: "foo"},
		},
	}
	bar1 = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "bar1",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: "bar"},
		},
	}
	adminGlobalRole = &iamv1alpha2.GlobalRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "global-admin",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"*"},
			},
		},
		AggregationRule: nil,
	}
	regularGlobalRole = &iamv1alpha2.GlobalRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "regular",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{},
				APIGroups: []string{},
				Resources: []string{},
			},
		},
		AggregationRule: nil,
	}
	reguarWorksapceRole = &iamv1alpha2.WorkspaceRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "workspace-regular",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: "bar"},
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{},
				APIGroups: []string{},
				Resources: []string{},
			},
		},
		AggregationRule: nil,
	}
	adminGlobalRoleBinding = &iamv1alpha2.GlobalRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "global-admin",
		},
		Subjects: []rbacv1.Subject{{
			Kind: "User",
			Name: "admin",
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     "global-admin",
		},
	}
	regularGlobalRoleBinding = &iamv1alpha2.GlobalRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "regular",
		},
		Subjects: []rbacv1.Subject{{
			Kind: "User",
			Name: "regular",
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     "regular",
		},
	}

	regularWorkspaceRoleBinding = &iamv1alpha2.WorkspaceRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "workspace-regular",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: "bar"},
		},
		Subjects: []rbacv1.Subject{{
			Kind: "User",
			Name: "regular",
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     "workspace-regular",
		},
	}
	bar1NamespaceRole = &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "admin",
			Namespace: "bar1",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"*"},
			},
		},
	}
	bar1NamespaceRoleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "admin",
			Namespace: "bar1",
		},
		Subjects: []rbacv1.Subject{{
			Kind: "User",
			Name: "regular",
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.String(),
			Kind:     "Role",
			Name:     "admin",
		},
	}
	workspaceFoo = &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	workspaceBar = &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar",
		},
	}

	workspaces            = []interface{}{workspaceFoo, workspaceBar}
	namespaces            = []interface{}{foo1, foo2, bar1}
	globalRoles           = []interface{}{adminGlobalRole, regularGlobalRole}
	globalRoleBindings    = []interface{}{adminGlobalRoleBinding, regularGlobalRoleBinding}
	workspaceRoles        = []interface{}{regularGlobalRole}
	workspaceRoleBindings = []interface{}{regularWorkspaceRoleBinding}
	namespaceRoles        = []interface{}{bar1NamespaceRole}
	namespaceRoleBindings = []interface{}{bar1NamespaceRoleBinding}
)

func prepare() Interface {
	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	istioClient := fakeistio.NewSimpleClientset()
	appClient := fakeapp.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, istioClient, appClient)

	for _, workspace := range workspaces {
		fakeInformerFactory.KubeSphereSharedInformerFactory().Tenant().V1alpha1().
			Workspaces().Informer().GetIndexer().Add(workspace)
	}

	for _, namespace := range namespaces {
		fakeInformerFactory.KubernetesSharedInformerFactory().Core().V1().
			Namespaces().Informer().GetIndexer().Add(namespace)
	}

	for _, globalRole := range globalRoles {
		fakeInformerFactory.KubeSphereSharedInformerFactory().Iam().V1alpha2().
			GlobalRoles().Informer().GetIndexer().Add(globalRole)
	}

	for _, globalRoleBinding := range globalRoleBindings {
		fakeInformerFactory.KubeSphereSharedInformerFactory().Iam().V1alpha2().
			GlobalRoleBindings().Informer().GetIndexer().Add(globalRoleBinding)
	}

	for _, workspaceRole := range workspaceRoles {
		fakeInformerFactory.KubeSphereSharedInformerFactory().Iam().V1alpha2().
			WorkspaceRoles().Informer().GetIndexer().Add(workspaceRole)
	}

	for _, workspaceRoleBinding := range workspaceRoleBindings {
		fakeInformerFactory.KubeSphereSharedInformerFactory().Iam().V1alpha2().
			WorkspaceRoleBindings().Informer().GetIndexer().Add(workspaceRoleBinding)
	}

	for _, role := range namespaceRoles {
		fakeInformerFactory.KubernetesSharedInformerFactory().Rbac().V1().
			Roles().Informer().GetIndexer().Add(role)
	}

	for _, roleBinding := range namespaceRoleBindings {
		fakeInformerFactory.KubernetesSharedInformerFactory().Rbac().V1().
			RoleBindings().Informer().GetIndexer().Add(roleBinding)
	}

	return New(fakeInformerFactory)
}
