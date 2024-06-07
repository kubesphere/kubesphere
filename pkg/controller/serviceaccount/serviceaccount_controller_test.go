/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package serviceaccount

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

func TestServiceAccountController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KubeSphere e2e suite")
}

var _ = Describe("ServiceAccount", func() {
	const (
		saName      = "test-serviceaccount"
		saNamespace = "default"
		saRole      = "test-role"
	)
	var role *rbacv1.Role
	var sa *corev1.ServiceAccount
	var req ctrl.Request

	BeforeEach(func() {
		role = &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      saRole,
				Namespace: saNamespace,
			},
		}

		sa = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:        saName,
				Namespace:   saNamespace,
				Annotations: map[string]string{iamv1beta1.RoleAnnotation: saRole},
			},
		}
		req = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: saNamespace,
				Name:      saName,
			},
		}
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("ServiceAccount Controller", func() {
		It("Should create ServiceAccount successfully and create a rolebinding", func() {
			ctx := context.Background()

			reconciler := &Reconciler{
				//nolint:staticcheck
				Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).Build(),
				logger:   ctrl.Log.WithName("controllers").WithName("serviceaccount"),
				recorder: record.NewFakeRecorder(5),
			}

			Expect(reconciler.Create(ctx, sa)).Should(Succeed())
			Expect(reconciler.Create(ctx, role)).Should(Succeed())

			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(BeNil())

			By("Expecting to bind role successfully")
			rolebindings := &rbacv1.RoleBindingList{}
			Expect(func() bool {
				reconciler.List(ctx, rolebindings, client.InNamespace(sa.Namespace), client.MatchingLabels{iamv1beta1.ServiceAccountReferenceLabel: sa.Name})
				return len(rolebindings.Items) == 1 && k8sutil.IsControlledBy(rolebindings.Items[0].OwnerReferences, "ServiceAccount", saName)
			}()).Should(BeTrue())
		})

		It("Should report NotFound error when role doesn't exist", func() {
			ctx := context.Background()

			reconciler := &Reconciler{
				//nolint:staticcheck
				Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).Build(),
				logger:   ctrl.Log.WithName("controllers").WithName("serviceaccount"),
				recorder: record.NewFakeRecorder(5),
			}

			Expect(reconciler.Create(ctx, sa)).Should(Succeed())
			_, err := reconciler.Reconcile(ctx, req)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
	})
})
