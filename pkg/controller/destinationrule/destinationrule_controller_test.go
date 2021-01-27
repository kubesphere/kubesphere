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

package destinationrule

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiv1alpha3 "istio.io/api/networking/v1alpha3"
	destinationrule "istio.io/client-go/pkg/apis/networking/v1alpha3"
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
var _ = Describe("Destinationrule", func() {

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
		dr := &destinationrule.DestinationRule{}
		// Destinationrule should be created automatically by controller
		Eventually(k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: metav1.NamespaceDefault}, dr))
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Desinationrule Controller", func() {
		It("Should create successfully", func() {

			// Create servicepolicy
			By("Expecting to create servicepolicy successfully")
			expectSp := newServicePolicy(service, deployments...)
			Expect(k8sClient.Create(ctx, expectSp)).Should(Succeed())

			expectedSubsets := expectSp.Spec.Template.Spec.Subsets
			expectedTraffic := expectSp.Spec.Template.Spec.TrafficPolicy.LoadBalancer
			Eventually(func() bool {
				actualSp := &servicemeshv1alpha2.ServicePolicy{}
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: metav1.NamespaceDefault}, actualSp)
				if len(actualSp.Name) == 0 {
					return false
				}
				actualSubsets := actualSp.Spec.Template.Spec.Subsets
				actualTraffic := actualSp.Spec.Template.Spec.TrafficPolicy.LoadBalancer
				if actualSp.Name != service.Name {
					klog.Errorf("serivcepolicy name not the same, actual: %s, expect: %s", actualSp.Name, service.Name)
					return false
				}

				if !reflect.DeepEqual(actualSubsets, expectedSubsets) {
					klog.Errorf("servicepolicy strategy not the same, actual: %s, expect: %s", actualSubsets, expectedSubsets)
					return false
				}

				if !reflect.DeepEqual(actualTraffic, expectedTraffic) {
					klog.Errorf("servicepolicy traffic not the same, actual: %s, expect: %s", actualTraffic, expectedTraffic)
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Expecting to reconcile destinationrule successfully")
			Eventually(func() bool {
				dr := &destinationrule.DestinationRule{}
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: metav1.NamespaceDefault}, dr)
				drActual := dr.Spec.TrafficPolicy
				if dr.Name != service.Name {
					klog.Errorf("destinationrule name not match, actual: %s, expect: %s", dr.Name, service.Name)
					return false
				}
				if reflect.DeepEqual(drActual, expectedTraffic) {
					klog.Errorf("destination traffic not match, actual: %s, expect: %s", drActual, expectedTraffic)
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})

func newServicePolicy(service *corev1.Service, deployments ...*v1.Deployment) *servicemeshv1alpha2.ServicePolicy {
	sp := &servicemeshv1alpha2.ServicePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   metav1.NamespaceDefault,
			Labels:      service.Labels,
			Annotations: service.Annotations,
		},
		Spec: servicemeshv1alpha2.ServicePolicySpec{
			Template: servicemeshv1alpha2.DestinationRuleSpecTemplate{
				Spec: apiv1alpha3.DestinationRule{
					Host: service.Name,
				},
			},
		},
	}

	sp.Spec.Template.Spec.Subsets = []*apiv1alpha3.Subset{}
	for _, deployment := range deployments {
		subset := &apiv1alpha3.Subset{
			Name: servicemesh.GetComponentVersion(&deployment.ObjectMeta),
			Labels: map[string]string{
				"version": servicemesh.GetComponentVersion(&deployment.ObjectMeta),
			},
		}
		sp.Spec.Template.Spec.Subsets = append(sp.Spec.Template.Spec.Subsets, subset)
	}

	sp.Spec.Template.Spec.TrafficPolicy = &apiv1alpha3.TrafficPolicy{
		LoadBalancer: &apiv1alpha3.LoadBalancerSettings{
			LbPolicy: &apiv1alpha3.LoadBalancerSettings_Simple{Simple: apiv1alpha3.LoadBalancerSettings_SimpleLB(1)},
		},
	}
	return sp
}

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
	service := &corev1.Service{
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
	return service
}
