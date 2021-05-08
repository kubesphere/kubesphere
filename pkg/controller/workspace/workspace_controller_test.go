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

package workspace

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
)

var _ = Describe("Workspace", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Workspace Controller", func() {
		It("Should create successfully", func() {

			workspace := &tenantv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "workspace-test",
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), workspace)).Should(Succeed())

			By("Expecting to create workspace successfully")
			Eventually(func() bool {
				f := &tenantv1alpha1.Workspace{}
				k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, f)
				return len(f.Finalizers) > 0
			}, timeout, interval).Should(BeTrue())

			// Update
			updated := &tenantv1alpha1.Workspace{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, updated)).Should(Succeed())
			updated.Spec.Manager = "admin"
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			// List workspace role bindings
			By("Expecting to update workspace successfully")
			Eventually(func() bool {
				k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, workspace)
				return workspace.Spec.Manager == "admin"
			}, timeout, interval).Should(BeTrue())

			// Delete
			By("Expecting to delete workspace successfully")
			Eventually(func() error {
				f := &tenantv1alpha1.Workspace{}
				k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete workspace finish")
			Eventually(func() error {
				f := &tenantv1alpha1.Workspace{}
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: workspace.Name}, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
