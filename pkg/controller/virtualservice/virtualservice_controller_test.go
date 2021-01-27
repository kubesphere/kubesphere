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

package virtualservice

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiv1alpha3 "istio.io/api/networking/v1alpha3"
	apisnetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	"reflect"

	"time"
)

var replicas = int32(2)
var _ = Describe("Virtualservice", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	ctx := context.TODO()
	service := newService("productpage")
	deployments := []*v1.Deployment{newDeployments(service, "v1"), newDeployments(service, "v2")}

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
	Context("Virtualservice Controller", func() {
		It("Should create successfully", func() {

			By("Expecting to create VirtualService successfully")
			dr := &apisnetworkingv1alpha3.VirtualService{}
			// Virtualservice should be created automatically by controller
			Eventually(func() bool {
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: metav1.NamespaceDefault}, dr)
				return dr.Name == service.Name
			}, timeout, interval).Should(BeTrue())

			expectStrategy := newStrategy(service)

			// Create actualStrategy
			Expect(k8sClient.Create(ctx, expectStrategy)).Should(Succeed())

			By("Expecting to create Strategy successfully")
			expectedStrategySpec := expectStrategy.Spec.Template.Spec.Http[0].Route[0].Destination.Host
			Eventually(func() bool {
				actualStrategy := &servicemeshv1alpha2.Strategy{}
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: metav1.NamespaceDefault}, actualStrategy)
				actualStrategySpec := actualStrategy.Spec.Template.Spec.Http[0].Route[0].Destination.Host
				if actualStrategy.Labels[servicemesh.AppLabel] != service.Name {
					klog.Errorf("strategy name not match, actual: %s, expect: %s", actualStrategy.Labels[servicemesh.AppLabel], service.Name)
					return false
				}
				if !reflect.DeepEqual(actualStrategySpec, expectedStrategySpec) {
					klog.Errorf("strategy not match, actual: %s, expect: %s", actualStrategySpec, expectedStrategySpec)
				}

				return true
			}, timeout, interval).Should(BeTrue())

			By("Expecting to reconcile virtualservice successfully")
			Eventually(func() bool {
				actualVituralservice := &apisnetworkingv1alpha3.VirtualService{}
				_ = k8sClient.Get(context.Background(), types.NamespacedName{Name: service.Name, Namespace: metav1.NamespaceDefault}, actualVituralservice)
				actualVirtualserviceSpec := actualVituralservice.Spec.Http[0].Route[0].Destination.Host
				if actualVituralservice.Name != service.Name {
					klog.Errorf("virtualservice name not match, actual: %s, expect: %s", actualVituralservice.Name, service.Name)
					return false
				}
				if !reflect.DeepEqual(actualVirtualserviceSpec, expectedStrategySpec) {
					klog.Errorf("virtualservice not match, actual: %s, expect: %s", actualVirtualserviceSpec, expectedStrategySpec)
					return true
				}
				return true
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

func newStrategy(service *corev1.Service) *servicemeshv1alpha2.Strategy {
	st := servicemeshv1alpha2.Strategy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   service.Namespace,
			Labels:      service.Labels,
			Annotations: service.Annotations,
		},
		Spec: servicemeshv1alpha2.StrategySpec{
			Type:             servicemeshv1alpha2.CanaryType,
			PrincipalVersion: "v1",
			GovernorVersion:  "",
			Template: servicemeshv1alpha2.VirtualServiceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: apiv1alpha3.VirtualService{
					Hosts: []string{service.Name},
					Http: []*apiv1alpha3.HTTPRoute{
						{
							Route: []*apiv1alpha3.HTTPRouteDestination{
								{
									Destination: &apiv1alpha3.Destination{
										Host:   service.Name,
										Subset: "",
									},
								},
							},
						},
					},
				},
			},
			StrategyPolicy: servicemeshv1alpha2.PolicyImmediately,
		},
	}

	return &st
}
