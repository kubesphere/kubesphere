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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/kubesphere/pkg/constants"
)

var _ = Describe("Workspace", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Workspace Controller", func() {
		It("DeletePropagationBackground", func() {
			workspace := &tenantv1beta1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "workspace-test1",
				},
			}
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "namespace-test1",
					Labels: map[string]string{
						tenantv1beta1.WorkspaceLabel: workspace.Name,
					},
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), workspace)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), namespace)).Should(Succeed())

			By("Expecting to create workspace successfully")
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace)
				return len(workspace.Finalizers) > 0
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

			// Update DeletionPropagation
			By("Expecting to delete workspace successfully")
			Eventually(func() error {
				if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace); err != nil {
					return err
				}
				if workspace.Annotations == nil {
					workspace.Annotations = make(map[string]string)
				}
				workspace.Annotations[constants.DeletionPropagationAnnotation] = string(metav1.DeletePropagationBackground)
				return k8sClient.Update(context.Background(), workspace)
			}, timeout, interval).Should(Succeed())

			// Delete workspace
			By("Expecting to delete workspace successfully")
			Eventually(func() error {
				return k8sClient.Delete(context.Background(), workspace)
			}, timeout, interval).Should(Succeed())

			By("Expecting to cascading deletion finish")
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace.Name}, namespace)
				// Deleting a namespace will seem to succeed, but the namespace will just be put in a Terminating state, and never actually be reclaimed.
				// Reference: https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
				return namespace.Status.Phase == corev1.NamespaceTerminating
			}, timeout, interval).Should(BeTrue())

			By("Expecting to delete workspace finish")
			Eventually(func() bool {
				return apierrors.IsNotFound(k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace))
			}, timeout, interval).Should(BeTrue())
		})
		It("DeletePropagationOrphan", func() {
			workspace := &tenantv1beta1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "workspace-test2",
				},
			}
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "namespace-test2",
					Labels: map[string]string{
						tenantv1beta1.WorkspaceLabel: workspace.Name,
					},
				},
			}
			// Create
			Expect(k8sClient.Create(context.Background(), workspace)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), namespace)).Should(Succeed())

			By("Expecting to create workspace successfully")
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace)
				return len(workspace.Finalizers) > 0
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

			// Delete workspace
			By("Expecting to delete workspace successfully")
			Eventually(func() error {
				return k8sClient.Delete(context.Background(), workspace)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete workspace finish")
			Eventually(func() bool {
				return apierrors.IsNotFound(k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace))
			}, timeout, interval).Should(BeTrue())

			By("Expecting to cascading deletion finish")
			Eventually(func() bool {
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: namespace.Name}, namespace)
				return namespace.Labels[tenantv1beta1.WorkspaceLabel] == "" && namespace.Status.Phase != corev1.NamespaceTerminating
			}, timeout, interval).Should(BeTrue())
		})
	})
})
