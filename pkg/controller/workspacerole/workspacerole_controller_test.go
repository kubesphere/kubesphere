/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspacerole

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
)

var _ = Describe("WorkspaceRole", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	workspace := &tenantv1beta1.WorkspaceTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "workspace1",
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
	Context("WorkspaceRole Controller", func() {
		It("Should create successfully", func() {
			workspaceAdmin := &iamv1beta1.WorkspaceRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:   fmt.Sprintf("%s-admin", workspace.Name),
					Labels: map[string]string{tenantv1beta1.WorkspaceLabel: workspace.Name},
				},
				Rules: []rbacv1.PolicyRule{},
			}

			// Create workspace role
			Expect(k8sClient.Create(context.Background(), workspaceAdmin)).Should(Succeed())

			By("Expecting to create workspace role successfully")
			Eventually(func() bool {
				k8sClient.Get(context.Background(), types.NamespacedName{Name: workspaceAdmin.Name}, workspaceAdmin)
				return !workspaceAdmin.CreationTimestamp.IsZero()
			}, timeout, interval).Should(BeTrue())

			By("Expecting to set owner reference successfully")
			Eventually(func() bool {
				k8sClient.Get(context.Background(), types.NamespacedName{Name: workspaceAdmin.Name}, workspaceAdmin)
				return len(workspaceAdmin.OwnerReferences) > 0
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
			Expect(workspaceAdmin.OwnerReferences).To(ContainElement(expectedOwnerReference))
		})
	})
})
