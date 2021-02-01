/*
Copyright 2020 KubeSphere Authors

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

package application

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	"sigs.k8s.io/application/api/v1beta1"
	"time"
)

var replicas = int32(2)
var _ = Describe("Application", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	ctx := context.TODO()

	service := newService("productpage")
	app := newAppliation(service)
	deployments := []*v1.Deployment{newDeployments(service, "v1")}

	BeforeEach(func() {

		// Create application service and deployment
		Expect(k8sClient.Create(ctx, app)).Should(Succeed())
		Expect(k8sClient.Create(ctx, service)).Should(Succeed())
		for i := range deployments {
			deployment := deployments[i]
			Expect(k8sClient.Create(ctx, deployment)).Should(Succeed())
		}
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Application Controller", func() {
		It("Should create successfully", func() {

			By("Reconcile Application successfully")
			// application should have "kubesphere.io/last-updated" annotation
			Eventually(func() bool {
				app := &v1beta1.Application{}
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: service.Labels[servicemesh.ApplicationNameLabel], Namespace: metav1.NamespaceDefault}, app)
				time, ok := app.Annotations["kubesphere.io/last-updated"]
				return len(time) > 0 && ok
			}, timeout, interval).Should(BeTrue())
		})
	})
})

func newDeployments(service *corev1.Service, version string) *v1.Deployment {
	lbs := service.Labels
	lbs["version"] = version

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", service.Name, version),
			Namespace:   metav1.NamespaceDefault,
			Labels:      lbs,
			Annotations: map[string]string{servicemesh.ServiceMeshEnabledAnnotation: "true"},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: lbs,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      lbs,
					Annotations: service.Annotations,
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

func newService(name string) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				"app.kubernetes.io/name":    "bookinfo",
				"app.kubernetes.io/version": "1",
				"app":                       name,
			},
			Annotations: map[string]string{
				"servicemesh.kubesphere.io/enabled": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     80,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "https",
					Port:     443,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "mysql",
					Port:     3306,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name":    "bookinfo",
				"app.kubernetes.io/version": "1",
				"app":                       "foo",
			},
			Type: corev1.ServiceTypeClusterIP,
		},
		Status: corev1.ServiceStatus{},
	}
	return svc
}

func newAppliation(service *corev1.Service) *v1beta1.Application {
	app := &v1beta1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Labels[servicemesh.ApplicationNameLabel],
			Namespace:   metav1.NamespaceDefault,
			Labels:      service.Labels,
			Annotations: map[string]string{servicemesh.ServiceMeshEnabledAnnotation: "true"},
		},
		Spec: v1beta1.ApplicationSpec{
			ComponentGroupKinds: []metav1.GroupKind{
				{
					Group: "",
					Kind:  "Service",
				},
				{
					Group: "apps",
					Kind:  "Deployment",
				},
			},
			AddOwnerRef: true,
		},
	}
	return app
}
