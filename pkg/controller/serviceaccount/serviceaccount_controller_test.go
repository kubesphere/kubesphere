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

package serviceaccount

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ServiceAccount", func() {
	const (
		saName      = "test-serviceaccount"
		saNamespace = "default"
		saRole      = "test-role"
		timeout     = time.Second * 30
		interval    = time.Second * 1
	)
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saRole,
			Namespace: saNamespace,
		},
	}
	BeforeEach(func() {
		// Create workspace
		Expect(k8sClient.Create(context.Background(), role)).Should(Succeed())
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("ServiceAccount Controller", func() {
		It("Should create successfully", func() {
			ctx := context.Background()
			sa := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:        saName,
					Namespace:   saNamespace,
					Annotations: map[string]string{iamv1alpha2.RoleAnnotation: saRole},
				},
			}

			By("Expecting to create serviceaccount successfully")
			Expect(k8sClient.Create(ctx, sa)).Should(Succeed())
			expectedSa := &corev1.ServiceAccount{}
			Eventually(func() bool {
				k8sClient.Get(ctx, types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, expectedSa)
				return !expectedSa.CreationTimestamp.IsZero()
			}, timeout, interval).Should(BeTrue())

			By("Expecting to bind role successfully")
			rolebindings := &rbacv1.RoleBindingList{}
			Eventually(func() bool {
				k8sClient.List(ctx, rolebindings, client.InNamespace(sa.Namespace), client.MatchingLabels{iamv1alpha2.ServiceAccountReferenceLabel: sa.Name})
				return len(rolebindings.Items) == 1
			}, timeout, interval).Should(BeTrue())
		})
	})
})
