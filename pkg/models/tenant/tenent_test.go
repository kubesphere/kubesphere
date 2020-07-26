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

package tenant

import (
	"github.com/google/go-cmp/cmp"
	fakeistio "istio.io/client-go/pkg/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"reflect"
	fakeapp "sigs.k8s.io/application/pkg/client/clientset/versioned/fake"
	"testing"
)

func TestTenantOperator_ListWorkspaces(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		result      *api.ListResult
		username    string
		expectError error
	}{
		{
			username: admin.Name,
			result: &api.ListResult{
				Items:      workspaceTemplates,
				TotalItems: len(workspaceTemplates),
			},
		},
		{
			username: tester2.Name,
			result: &api.ListResult{
				Items:      []interface{}{systemWorkspaceTmpl},
				TotalItems: 1,
			},
		},
	}

	for i, test := range tests {
		result, err := tenantOperator.ListWorkspaces(&user.DefaultInfo{Name: test.username}, query.New())

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("got %#v, expected %#v", err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Errorf("case %d,%s", i, diff)
		}
	}
}

func TestTenantOperator_ListNamespaces(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		result      *api.ListResult
		username    string
		workspace   string
		expectError error
	}{
		{
			workspace: systemWorkspace.Name,
			username:  admin.Name,
			result: &api.ListResult{
				Items:      []interface{}{kubesphereSystem, defaultNamespace},
				TotalItems: 2,
			},
		},
		{
			workspace: systemWorkspace.Name,
			username:  tester1.Name,
			result: &api.ListResult{
				Items:      []interface{}{},
				TotalItems: 0,
			},
		},
		{
			workspace: testWorkspace.Name,
			username:  tester2.Name,
			result: &api.ListResult{
				Items:      []interface{}{testNamespace},
				TotalItems: 1,
			},
		},
	}

	for i, test := range tests {
		result, err := tenantOperator.ListNamespaces(&user.DefaultInfo{Name: test.username}, test.workspace, query.New())

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("got %#v, expected %#v", err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Errorf("case %d, %s", i, diff)
		}
	}
}

func TestTenantOperator_DescribeNamespace(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		result      *corev1.Namespace
		username    string
		workspace   string
		namespace   string
		expectError error
	}{
		{
			result:      testNamespace,
			username:    tester2.Name,
			workspace:   testWorkspace.Name,
			namespace:   testNamespace.Name,
			expectError: nil,
		},
		{
			result:      testNamespace,
			username:    tester2.Name,
			workspace:   systemWorkspace.Name,
			namespace:   testNamespace.Name,
			expectError: errors.NewNotFound(corev1.Resource("namespace"), testNamespace.Name),
		},
	}

	for _, test := range tests {
		result, err := tenantOperator.DescribeNamespace(test.workspace, test.namespace)

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("got %#v, expected %#v", err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Error(diff)
		}
	}
}

func TestTenantOperator_CreateNamespace(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		result      *corev1.Namespace
		workspace   string
		namespace   *corev1.Namespace
		expectError error
	}{
		{
			result: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: testWorkspace.Name},
				},
			},
			workspace: testWorkspace.Name,
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expectError: nil,
		},
	}

	for i, test := range tests {
		result, err := tenantOperator.CreateNamespace(test.workspace, test.namespace)

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("case %d, got %#v, expected %#v", i, err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Error(diff)
		}
	}
}

func TestTenantOperator_DeleteNamespace(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		workspace   string
		namespace   string
		expectError error
	}{
		{
			workspace:   testWorkspace.Name,
			namespace:   kubesphereSystem.Name,
			expectError: errors.NewNotFound(corev1.Resource("namespace"), kubesphereSystem.Name),
		},
		{
			workspace:   testWorkspace.Name,
			namespace:   testNamespace.Name,
			expectError: nil,
		},
	}

	for i, test := range tests {
		err := tenantOperator.DeleteNamespace(test.workspace, test.namespace)
		if err != nil {
			if test.expectError != nil && test.expectError.Error() == err.Error() {
				continue
			} else {
				if !reflect.DeepEqual(err, test.expectError) {
					t.Errorf("case %d, got %#v, expected %#v", i, err, test.expectError)
				}
			}
		}
	}
}

func TestTenantOperator_UpdateNamespace(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		result      *corev1.Namespace
		workspace   string
		namespace   *corev1.Namespace
		expectError error
	}{
		{
			result: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-namespace",
					Annotations: map[string]string{"test": "test"},
					Labels:      map[string]string{tenantv1alpha1.WorkspaceLabel: testWorkspace.Name},
				},
			},
			workspace: testWorkspace.Name,
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-namespace",
					Annotations: map[string]string{"test": "test"},
					Labels:      map[string]string{tenantv1alpha1.WorkspaceLabel: testWorkspace.Name},
				},
			},
			expectError: nil,
		},
	}

	for i, test := range tests {
		result, err := tenantOperator.UpdateNamespace(test.workspace, test.namespace)

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("case %d, got %#v, expected %#v", i, err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Error(diff)
		}
	}
}

func TestTenantOperator_PatchNamespace(t *testing.T) {
	tenantOperator := prepare()
	tests := []struct {
		result      *corev1.Namespace
		workspace   string
		patch       *corev1.Namespace
		expectError error
	}{
		{
			result: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-namespace",
					Annotations: map[string]string{"test": "test2"},
					Labels:      map[string]string{tenantv1alpha1.WorkspaceLabel: testWorkspace.Name},
				},
			},
			workspace: testWorkspace.Name,
			patch: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-namespace",
					Annotations: map[string]string{"test": "test2"},
				},
			},
			expectError: nil,
		},
	}

	for i, test := range tests {
		result, err := tenantOperator.PatchNamespace(test.workspace, test.patch)

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("case %d, got %#v, expected %#v", i, err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Error(diff)
		}
	}
}

var (
	admin = user.DefaultInfo{
		Name: "admin",
	}
	tester1 = user.DefaultInfo{
		Name: "tester1",
	}
	tester2 = user.DefaultInfo{
		Name: "tester2",
	}
	systemWorkspaceTmpl = &tenantv1alpha2.WorkspaceTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-workspace",
		},
	}
	testWorkspaceTmpl = &tenantv1alpha2.WorkspaceTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
	}
	systemWorkspace = &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-workspace",
		},
	}
	testWorkspace = &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
	}
	kubesphereSystem = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "kubesphere-system",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: systemWorkspace.Name},
		},
	}
	defaultNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "default",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: systemWorkspace.Name},
		},
	}
	testNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-namespace",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: testWorkspace.Name},
		},
	}
	platformAdmin = &iamv1alpha2.GlobalRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "platform-admin",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"*"},
			},
		},
	}
	platformRegular = &iamv1alpha2.GlobalRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "platform-regular",
		},
		Rules: []rbacv1.PolicyRule{},
	}
	systemWorkspaceRegular = &iamv1alpha2.WorkspaceRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "system-workspace-regular",
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: systemWorkspace.Name},
		},
		Rules: []rbacv1.PolicyRule{},
	}
	adminGlobalRoleBinding = &iamv1alpha2.GlobalRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: admin.Name + "-" + platformAdmin.Name,
		},
		Subjects: []rbacv1.Subject{{
			Kind: "User",
			Name: admin.Name,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     platformAdmin.Name,
		},
	}
	regularGlobalRoleBinding = &iamv1alpha2.GlobalRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: tester1.Name + "-" + platformRegular.Name,
		},
		Subjects: []rbacv1.Subject{{
			Kind: "User",
			Name: tester1.Name,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     platformRegular.Name,
		},
	}

	systemWorkspaceRegularRoleBinding = &iamv1alpha2.WorkspaceRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tester2.Name + "-" + systemWorkspaceRegular.Name,
			Labels: map[string]string{tenantv1alpha1.WorkspaceLabel: systemWorkspace.Name},
		},
		Subjects: []rbacv1.Subject{{
			Kind: "User",
			Name: tester2.Name,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     systemWorkspaceRegular.Name,
		},
	}
	testNamespaceAdmin = &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "admin",
			Namespace: testNamespace.Name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"*"},
				APIGroups: []string{"*"},
				Resources: []string{"*"},
			},
		},
	}
	testNamespaceAdminRoleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "admin",
			Namespace: testNamespace.Name,
		},
		Subjects: []rbacv1.Subject{{
			Kind: "User",
			Name: tester2.Name,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.String(),
			Kind:     "Role",
			Name:     testNamespaceAdmin.Name,
		},
	}

	workspaces            = []interface{}{systemWorkspace, testWorkspace}
	workspaceTemplates    = []interface{}{testWorkspaceTmpl, systemWorkspaceTmpl}
	namespaces            = []interface{}{kubesphereSystem, defaultNamespace, testNamespace}
	globalRoles           = []interface{}{platformAdmin, platformRegular}
	globalRoleBindings    = []interface{}{adminGlobalRoleBinding, regularGlobalRoleBinding}
	workspaceRoles        = []interface{}{systemWorkspaceRegular}
	workspaceRoleBindings = []interface{}{systemWorkspaceRegularRoleBinding}
	namespaceRoles        = []interface{}{testNamespaceAdmin}
	namespaceRoleBindings = []interface{}{testNamespaceAdminRoleBinding}
)

func prepare() Interface {
	ksClient := fakeks.NewSimpleClientset([]runtime.Object{testWorkspace, systemWorkspace}...)
	k8sClient := fakek8s.NewSimpleClientset([]runtime.Object{testNamespace, kubesphereSystem}...)
	istioClient := fakeistio.NewSimpleClientset()
	appClient := fakeapp.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, istioClient, appClient, nil, nil)

	for _, workspace := range workspaces {
		fakeInformerFactory.KubeSphereSharedInformerFactory().Tenant().V1alpha1().
			Workspaces().Informer().GetIndexer().Add(workspace)
	}

	for _, workspaceTmpl := range workspaceTemplates {
		fakeInformerFactory.KubeSphereSharedInformerFactory().Tenant().V1alpha2().
			WorkspaceTemplates().Informer().GetIndexer().Add(workspaceTmpl)
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

	return New(fakeInformerFactory, k8sClient, ksClient, nil, nil, nil)
}
