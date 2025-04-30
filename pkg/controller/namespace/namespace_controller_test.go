/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package namespace

import (
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/constants"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
)

const timeout = time.Second * 30
const interval = time.Second * 1

var _ = Describe("Namespace", func() {
	workspace := &tenantv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
	}
	role := iamv1beta1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "iam.kubesphere.io/v1beta1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "admin",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}

	jsonData, _ := json.Marshal(role)
	builtinRole := &iamv1beta1.BuiltinRole{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"iam.kubesphere.io/scope": "namespace",
			},
			Name: "project-admin",
		},
		Role: runtime.RawExtension{
			Raw: jsonData,
		},
	}

	BeforeEach(func() {
		// Create workspace
		Expect(k8sClient.Create(ctx, workspace)).Should(Succeed())
		Expect(k8sClient.Create(ctx, builtinRole)).Should(Succeed())
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Namespace Controller", func() {
		It("Should create successfully", func() {
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Labels: map[string]string{
						tenantv1beta1.WorkspaceLabel:     workspace.Name,
						constants.KubeSphereManagedLabel: "true",
					},
					Annotations: map[string]string{constants.CreatorAnnotationKey: "admin"},
				},
			}

			// Create namespace
			Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())
			By("Expecting to create namespace successfully")
			Eventually(func() bool {
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: namespace.Name}, namespace)
				return len(namespace.Finalizers) > 0
			}, timeout, interval).Should(BeTrue())

			By("Expecting to create namespace builtin roles successfully")
			Eventually(func() bool {
				roles := iamv1beta1.RoleList{}
				_ = k8sClient.List(ctx, &roles, runtimeclient.InNamespace(namespace.Name))
				return len(roles.Items) > 0
			}, timeout, interval).Should(BeTrue())

			By("Expecting to create creator role binding successfully")
			Eventually(func() bool {
				roleBindings := iamv1beta1.RoleBindingList{}
				_ = k8sClient.List(ctx, &roleBindings, runtimeclient.InNamespace(namespace.Name))
				return len(roleBindings.Items) > 0
			}, timeout, interval).Should(BeTrue())
		})
	})
})
