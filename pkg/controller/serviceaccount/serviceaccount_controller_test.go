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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
				Annotations: map[string]string{iamv1alpha2.RoleAnnotation: saRole},
			},
		}
		req = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: saNamespace,
				Name:      saName,
			},
		}
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("ServiceAccount Controller", func() {
		It("Should create ServiceAccount successfully and create a rolebinding", func() {
			ctx := context.Background()

			reconciler := &Reconciler{
				Client:   fake.NewFakeClientWithScheme(scheme.Scheme),
				logger:   ctrl.Log.WithName("controllers").WithName("acrpullbinding-controller"),
				scheme:   scheme.Scheme,
				recorder: record.NewFakeRecorder(5),
			}

			Expect(reconciler.Create(ctx, sa)).Should(Succeed())
			Expect(reconciler.Create(ctx, role)).Should(Succeed())

			_, err := reconciler.Reconcile(req)
			Expect(err).To(BeNil())

			By("Expecting to bind role successfully")
			rolebindings := &rbacv1.RoleBindingList{}
			Expect(func() bool {
				reconciler.List(ctx, rolebindings, client.InNamespace(sa.Namespace), client.MatchingLabels{iamv1alpha2.ServiceAccountReferenceLabel: sa.Name})
				return len(rolebindings.Items) == 1 && k8sutil.IsControlledBy(rolebindings.Items[0].OwnerReferences, "ServiceAccount", saName)
			}()).Should(BeTrue())
		})

		It("Should report NotFound error when role doesn't exist", func() {
			ctx := context.Background()

			reconciler := &Reconciler{
				Client:   fake.NewFakeClientWithScheme(scheme.Scheme),
				logger:   ctrl.Log.WithName("controllers").WithName("acrpullbinding-controller"),
				scheme:   scheme.Scheme,
				recorder: record.NewFakeRecorder(5),
			}

			Expect(reconciler.Create(ctx, sa)).Should(Succeed())
			_, err := reconciler.Reconcile(req)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
	})
})
