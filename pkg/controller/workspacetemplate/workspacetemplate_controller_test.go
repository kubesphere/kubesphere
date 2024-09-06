/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspacetemplate

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
)

var reconciler *Reconciler
var _ = Describe("WorkspaceTemplate", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {

		reconciler = &Reconciler{
			//nolint:staticcheck
			Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).Build(),
			logger:   ctrl.Log.WithName("controllers").WithName("workspacetemplate"),
			recorder: record.NewFakeRecorder(5),
		}

		workspaceAdmin := newWorkspaceAdmin()

		err := reconciler.Create(context.Background(), &workspaceAdmin)
		Expect(err).NotTo(HaveOccurred())

		admin := iamv1beta1.User{ObjectMeta: metav1.ObjectMeta{Name: "admin"}}
		err = reconciler.Create(context.Background(), &admin)
		Expect(err).NotTo(HaveOccurred())
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("WorkspaceTemplate Controller", func() {
		It("Should create successfully", func() {
			key := types.NamespacedName{
				Name: "workspace-template",
			}

			created := &tenantv1beta1.WorkspaceTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
				},
			}

			// Create
			Expect(reconciler.Create(context.Background(), created)).Should(Succeed())

			req := ctrl.Request{
				NamespacedName: key,
			}
			_, err := reconciler.Reconcile(context.Background(), req)
			Expect(err).To(BeNil())

			By("Expecting to create workspace template successfully")
			Expect(func() *tenantv1beta1.WorkspaceTemplate {
				f := &tenantv1beta1.WorkspaceTemplate{}
				reconciler.Get(context.Background(), key, f)
				return f
			}()).ShouldNot(BeNil())

			By("Expecting to create workspace successfully")
			Expect(func() *tenantv1beta1.Workspace {
				f := &tenantv1beta1.Workspace{}
				reconciler.Get(context.Background(), key, f)
				return f
			}()).ShouldNot(BeNil())

			// List workspace roles
			By("Expecting to create workspace role successfully")
			Eventually(func() bool {
				f := &iamv1beta1.WorkspaceRoleList{}
				reconciler.List(context.Background(), f, &client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{tenantv1beta1.WorkspaceLabel: key.Name})})
				return len(f.Items) == 1
			}, timeout, interval).Should(BeTrue())

			// Update
			updated := &tenantv1beta1.WorkspaceTemplate{}
			Expect(reconciler.Get(context.Background(), key, updated)).Should(Succeed())
			updated.Spec.Template.Spec.Manager = "admin"
			Expect(reconciler.Update(context.Background(), updated)).Should(Succeed())

			_, err = reconciler.Reconcile(context.Background(), req)
			Expect(err).To(BeNil())

			// List workspace role bindings
			By("Expecting to create workspace manager role binding successfully")
			Eventually(func() bool {
				f := &iamv1beta1.WorkspaceRoleBindingList{}
				reconciler.List(context.Background(), f, &client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{tenantv1beta1.WorkspaceLabel: key.Name})})
				return len(f.Items) == 1
			}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to finalize workspace successfully")
			Eventually(func() error {
				f := &tenantv1beta1.WorkspaceTemplate{}
				reconciler.Get(context.Background(), key, f)
				now := metav1.NewTime(time.Now())
				f.DeletionTimestamp = &now
				return reconciler.Update(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			_, err = reconciler.Reconcile(context.Background(), req)
			Expect(err).To(BeNil())
		})
	})
})

func newWorkspaceAdmin() iamv1beta1.BuiltinRole {
	return iamv1beta1.BuiltinRole{
		ObjectMeta: metav1.ObjectMeta{Name: "workspace-admin"},
		Role: runtime.RawExtension{
			Raw: []byte(`{
  "apiVersion": "iam.kubesphere.io/v1alpha2",
  "kind": "WorkspaceRole",
  "metadata": {
    "name": "admin"
  },
  "rules": [
    {
      "apiGroups": [
        "*"
      ],
      "resources": [
        "*"
      ],
      "verbs": [
        "*"
      ]
    }
  ]
}`)}}
}
