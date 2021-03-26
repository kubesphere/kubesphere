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

package workspacetemplate

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha2"
)

var _ = Describe("WorkspaceTemplate", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {
		workspaceAdmin := newWorkspaceAdmin()

		err := k8sClient.Create(context.Background(), &workspaceAdmin)
		Expect(err).NotTo(HaveOccurred())

		admin := iamv1alpha2.User{ObjectMeta: metav1.ObjectMeta{Name: "admin"}}
		err = k8sClient.Create(context.Background(), &admin)
		Expect(err).NotTo(HaveOccurred())
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("WorkspaceTemplate Controller", func() {
		It("Should create successfully", func() {
			key := types.NamespacedName{
				Name: "workspace-template",
			}

			created := &tenantv1alpha2.WorkspaceTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting to create workspace template successfully")
			Eventually(func() bool {
				f := &tenantv1alpha2.WorkspaceTemplate{}
				k8sClient.Get(context.Background(), key, f)
				return !f.CreationTimestamp.IsZero()
			}, timeout, interval).Should(BeTrue())

			By("Expecting to create workspace successfully")
			Eventually(func() bool {
				f := &tenantv1alpha1.Workspace{}
				k8sClient.Get(context.Background(), key, f)
				return !f.CreationTimestamp.IsZero()
			}, timeout, interval).Should(BeTrue())

			// List workspace roles
			By("Expecting to create workspace role successfully")
			Eventually(func() bool {
				f := &iamv1alpha2.WorkspaceRoleList{}
				k8sClient.List(context.Background(), f, &client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{tenantv1alpha1.WorkspaceLabel: key.Name})})
				return len(f.Items) == 1
			}, timeout, interval).Should(BeTrue())

			// Update
			updated := &tenantv1alpha2.WorkspaceTemplate{}
			Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())
			updated.Spec.Template.Spec.Manager = "admin"
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			// List workspace role bindings
			By("Expecting to create workspace manager role binding successfully")
			Eventually(func() bool {
				f := &iamv1alpha2.WorkspaceRoleBindingList{}
				k8sClient.List(context.Background(), f, &client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{tenantv1alpha1.WorkspaceLabel: key.Name})})
				return len(f.Items) == 1
			}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to delete workspace successfully")
			Eventually(func() error {
				f := &tenantv1alpha2.WorkspaceTemplate{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete workspace finish")
			Eventually(func() error {
				f := &tenantv1alpha2.WorkspaceTemplate{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})

func newWorkspaceAdmin() iamv1alpha2.RoleBase {
	return iamv1alpha2.RoleBase{
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
