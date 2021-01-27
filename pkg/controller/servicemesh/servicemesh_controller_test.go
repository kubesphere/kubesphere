/*
Copyright 2021 KubeSphere Authors

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

package servicemesh

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	"time"
)

var replicas = int32(2)
var _ = Describe("ServiceMesh Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	ctx := context.TODO()
	service := newService("productpage")
	deployments := []*v1.Deployment{newDeployments(service, "v2"), newDeployments(service, "v3")}

	BeforeEach(func() {

		// Create service and deployment
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
	Context("ServiceMesh Controller", func() {
		It("Should reconcile successfully", func() {

			By("Expecting to label and annotate deployments successfully")

			Eventually(func() bool {
				return CheckDeploymentReconciled(service)
			}, timeout, interval).Should(BeTrue())

			By("Change service label, reconcile deployment")

			service.Labels[servicemesh.ApplicationNameLabel] = "app-test"
			Expect(k8sClient.Update(ctx, service)).Should(Succeed())

			// S should be created automatically by controller
			Eventually(func() bool {
				return CheckDeploymentReconciled(service)
			}, timeout, interval).Should(BeTrue())
		})
	})
})

func CheckDeploymentReconciled(service *corev1.Service) bool {
	deploys := servicemesh.GetDeploymentsFromService(service, k8sClient)
	for i := range deploys {
		deploy := deploys[i]
		if len(deploy.Labels) < len(servicemesh.DeploymentLabels) {
			klog.Error("deployment labels num is not satisfied")
			return false
		}
		for _, i := range servicemesh.ApplicationLabels {
			if deploy.Labels[i] != service.Labels[i] {
				klog.Errorf("deployment label %s is not the same with service", deploy.Labels[i])
				return false
			}
		}
		if len(deploy.Labels[servicemesh.VersionLabel]) == 0 {
			klog.Error("deploy has no version label")
			return false
		}
		if deploy.Annotations[servicemesh.ServiceMeshEnabledAnnotation] != service.Annotations[servicemesh.ServiceMeshEnabledAnnotation] {
			klog.Error("deployment annotation is not the same")
			return false
		}
		klog.Info("test success")
	}
	return true
}

func newDeployments(service *corev1.Service, version string) *v1.Deployment {

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", service.Name, version),
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: service.Spec.Selector,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      service.Spec.Selector,
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
