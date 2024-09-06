/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspace

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
)

var _ = Describe("Workspace", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Workspace Controller", func() {
		It("Should create successfully", func() {

			workspace := &tenantv1beta1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "workspace-test",
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), workspace)).Should(Succeed())

			By("Expecting to create workspace successfully")
			Eventually(func() bool {
				f := &tenantv1beta1.Workspace{}
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, f)
				return len(f.Finalizers) > 0
			}, timeout, interval).Should(BeTrue())

			// Update
			updated := &tenantv1beta1.Workspace{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, updated)).Should(Succeed())
			updated.Spec.Manager = "admin"
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			// List workspace role bindings
			By("Expecting to update workspace successfully")
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace)
				return workspace.Spec.Manager == "admin"
			}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to delete workspace successfully")
			Eventually(func() error {
				f := &tenantv1beta1.Workspace{}
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete workspace finish")
			Eventually(func() error {
				f := &tenantv1beta1.Workspace{}
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
