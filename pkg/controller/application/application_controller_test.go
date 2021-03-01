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
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kubesphere.io/kubesphere/pkg/controller/utils/servicemesh"
	"sigs.k8s.io/application/api/v1beta1"
)

const (
	applicationName = "bookinfo"
	serviceName     = "productpage"
	timeout         = time.Second * 30
	interval        = time.Second * 2
)

var replicas = int32(2)

var _ = Context("Inside of a new namespace", func() {
	ctx := context.TODO()
	ns := SetupTest(ctx)

	Describe("Application", func() {
		applicationLabels := map[string]string{
			"app.kubernetes.io/name":    "bookinfo",
			"app.kubernetes.io/version": "1",
		}

		BeforeEach(func() {
			By("create deployment,service,application objects")
			service := newService(serviceName, ns.Name, applicationLabels)
			deployments := []*v1.Deployment{newDeployments(serviceName, ns.Name, applicationLabels, "v1")}
			app := newApplication(applicationName, ns.Name, applicationLabels)

			Expect(k8sClient.Create(ctx, service.DeepCopy())).Should(Succeed())
			for i := range deployments {
				deployment := deployments[i]
				Expect(k8sClient.Create(ctx, deployment.DeepCopy())).Should(Succeed())
			}
			Expect(k8sClient.Create(ctx, app)).Should(Succeed())
		})

		Context("Application Controller", func() {
			It("Should not reconcile application", func() {
				By("update application labels")
				application := &v1beta1.Application{}

				err := k8sClient.Get(ctx, types.NamespacedName{Name: applicationName, Namespace: ns.Name}, application)
				Expect(err).Should(Succeed())

				updateApplication := func(object interface{}) {
					newApp := object.(*v1beta1.Application)
					newApp.Labels["kubesphere.io/creator"] = ""
				}

				updated, err := updateWithRetries(k8sClient, ctx, application.Namespace, applicationName, updateApplication, 1*time.Second, 5*time.Second)
				Expect(updated).Should(BeTrue())

				Eventually(func() bool {

					err = k8sClient.Get(ctx, types.NamespacedName{Name: applicationName, Namespace: ns.Name}, application)

					// application status field should not be populated with selected deployments and services
					return len(application.Status.ComponentList.Objects) == 0
				}, timeout, interval).Should(BeTrue())

			})

			It("Should reconcile application successfully", func() {

				By("check if application status been updated by controller")
				application := &v1beta1.Application{}

				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{Name: applicationName, Namespace: ns.Name}, application)
					Expect(err).Should(Succeed())

					// application status field should be populated by controller
					return len(application.Status.ComponentList.Objects) > 0
				}, timeout, interval).Should(BeTrue())

			})
		})
	})
})

type UpdateObjectFunc func(obj interface{})

func updateWithRetries(client client.Client, ctx context.Context, namespace, name string, updateFunc UpdateObjectFunc, interval, timeout time.Duration) (bool, error) {
	var updateErr error

	pollErr := wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		app := &v1beta1.Application{}
		if err = client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, app); err != nil {
			return false, err
		}

		updateFunc(app)
		if err = client.Update(ctx, app); err == nil {
			return true, nil
		}

		updateErr = err
		return false, nil
	})

	if pollErr == wait.ErrWaitTimeout {
		pollErr = fmt.Errorf("couldn't apply the provided update to object %q: %v", name, updateErr)
		return false, pollErr
	}
	return true, nil
}

func newDeployments(deploymentName, namespace string, labels map[string]string, version string) *v1.Deployment {
	labels["app"] = deploymentName
	labels["version"] = version

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", deploymentName, version),
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{servicemesh.ServiceMeshEnabledAnnotation: "true"},
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
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

func newService(serviceName, namesapce string, labels map[string]string) *corev1.Service {
	labels["app"] = serviceName

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namesapce,
			Labels:    labels,
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
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
		},
		Status: corev1.ServiceStatus{},
	}
	return svc
}

func newApplication(applicationName, namespace string, labels map[string]string) *v1beta1.Application {
	app := &v1beta1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:        applicationName,
			Namespace:   namespace,
			Labels:      labels,
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
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			AddOwnerRef: true,
		},
	}
	return app
}
