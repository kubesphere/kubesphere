/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package namespace

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/kubesphere/pkg/constants"
)

var _ = Describe("Namespace", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	workspace := &tenantv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
	}
	BeforeEach(func() {
		// Create workspace
		Expect(k8sClient.Create(context.Background(), workspace)).Should(Succeed())
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Namespace Controller", func() {
		It("Should create successfully", func() {
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-namespace",
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: workspace.Name},
				},
			}

			// Create namespace
			Expect(k8sClient.Create(context.Background(), namespace)).Should(Succeed())

			By("Expecting to create namespace successfully")
			Eventually(func() bool {
				k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace.Name}, namespace)
				return !namespace.CreationTimestamp.IsZero()
			}, timeout, interval).Should(BeTrue())

			By("Expecting to set owner reference successfully")
			Eventually(func() bool {
				k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace.Name}, namespace)
				return len(namespace.OwnerReferences) > 0
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace)).Should(Succeed())

			controlled := true
			expectedOwnerReference := metav1.OwnerReference{
				Kind:               workspace.Kind,
				APIVersion:         workspace.APIVersion,
				UID:                workspace.UID,
				Name:               workspace.Name,
				Controller:         &controlled,
				BlockOwnerDeletion: &controlled,
			}

			By("Expecting to bind workspace successfully")
			Expect(namespace.OwnerReferences).To(ContainElement(expectedOwnerReference))
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace)).Should(Succeed())

			By("Expecting to update namespace successfully")
			updated := namespace.DeepCopy()
			updated.Labels[constants.WorkspaceLabelKey] = "workspace-not-exist"
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			By("Expecting to unbind workspace successfully")
			Eventually(func() bool {
				k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace.Name}, namespace)
				return len(namespace.OwnerReferences) == 0
			}, timeout, interval).Should(BeTrue())
		})
	})
})
