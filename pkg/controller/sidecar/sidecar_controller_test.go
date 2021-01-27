/*
Copyright 2021 The KubeSphere Authors.

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

package sidecar

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	"time"
)

var replicas = int32(1)
var _ = Describe("Sidecar", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Sidecar Controller", func() {
		It("Should create successfully", func() {

			deploy := newDeployment("test-deploy-1")

			Expect(k8sClient.Create(context.Background(), deploy)).Should(Succeed())

			By("Expecting to Inject Sidecar successfully")
			Eventually(func() bool {
				f := &v1.Deployment{}
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, f)
				return f.Spec.Template.Annotations[servicemesh.SidecarInjectAnnotation] == "true"
			}, timeout, interval).Should(BeTrue())

			// Update
			updated := &v1.Deployment{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, updated)).Should(Succeed())
			updated.Annotations[servicemesh.ServiceMeshEnabledAnnotation] = "false"
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())
			Eventually(func() bool {
				u := &v1.Deployment{}
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, u)
				return u.Spec.Template.Annotations[servicemesh.SidecarInjectAnnotation] == "false"
			})
		})
	})
})

func newDeployment(name string) *v1.Deployment {
	lbs := map[string]string{"foo-1": "bar-1"}
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   metav1.NamespaceDefault,
			Annotations: map[string]string{servicemesh.ServiceMeshEnabledAnnotation: "true"},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: lbs,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: lbs,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "c1",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "https",
									ContainerPort: 443,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "mysql",
									ContainerPort: 3306,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
		Status: v1.DeploymentStatus{
			AvailableReplicas: replicas,
			ReadyReplicas:     replicas,
			Replicas:          replicas,
		},
	}

	return deployment
}
