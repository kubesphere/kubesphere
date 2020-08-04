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
	"fmt"
	apiv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice/util"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"testing"
)

var (
	alwaysReady = func() bool { return true }
	replicas    = int32(2)
)

func newDeployments(service *corev1.Service, version string) *appsv1.Deployment {
	lbs := service.Labels
	lbs["version"] = version

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", service.Name, version),
			Namespace:   metav1.NamespaceDefault,
			Labels:      lbs,
			Annotations: service.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
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
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: replicas,
			ReadyReplicas:     replicas,
			Replicas:          replicas,
		},
	}

	return deployment
}

func newDestinationRule(service *corev1.Service, deployments ...*appsv1.Deployment) *v1alpha3.DestinationRule {
	dr := &v1alpha3.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   service.Namespace,
			Labels:      service.Labels,
			Annotations: make(map[string]string),
		},
		Spec: apiv1alpha3.DestinationRule{
			Host: service.Name,
		},
	}

	dr.Spec.Subsets = []*apiv1alpha3.Subset{}
	for _, deployment := range deployments {
		subset := &apiv1alpha3.Subset{
			Name: util.GetComponentVersion(&deployment.ObjectMeta),
			Labels: map[string]string{
				"version": util.GetComponentVersion(&deployment.ObjectMeta),
			},
		}
		dr.Spec.Subsets = append(dr.Spec.Subsets, subset)
	}

	return dr
}

func newService(name string) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				"app.kubernetes.io/name":    "bookinfo",
				"app.kubernetes.io/version": "1",
				"app":                       "foo",
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

func newServicePolicy(name string, service *corev1.Service, deployments ...*appsv1.Deployment) *v1alpha2.ServicePolicy {
	sp := &v1alpha2.ServicePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   metav1.NamespaceDefault,
			Labels:      service.Labels,
			Annotations: service.Annotations,
		},
		Spec: v1alpha2.ServicePolicySpec{
			Template: v1alpha2.DestinationRuleSpecTemplate{
				Spec: apiv1alpha3.DestinationRule{
					Host: service.Name,
				},
			},
		},
	}

	sp.Spec.Template.Spec.Subsets = []*apiv1alpha3.Subset{}
	for _, deployment := range deployments {
		subset := &apiv1alpha3.Subset{
			Name: util.GetComponentVersion(&deployment.ObjectMeta),
			Labels: map[string]string{
				"version": util.GetComponentVersion(&deployment.ObjectMeta),
			},
		}
		sp.Spec.Template.Spec.Subsets = append(sp.Spec.Template.Spec.Subsets, subset)
	}

	return sp
}

type fixture struct {
	t testing.TB

	kubeClient        *kubefake.Clientset
	istioClient       *istiofake.Clientset
	servicemeshClient *fake.Clientset

	serviceLister    []*corev1.Service
	deploymentLister []*appsv1.Deployment
	drLister         []*v1alpha3.DestinationRule
	spLister         []*v1alpha2.ServicePolicy

	kubeObjects        []runtime.Object
	istioObjects       []runtime.Object
	servicemeshObjects []runtime.Object
}

func newFixture(t testing.TB) *fixture {
	f := &fixture{}
	f.t = t
	f.kubeObjects = []runtime.Object{}
	f.istioObjects = []runtime.Object{}
	f.servicemeshObjects = []runtime.Object{}
	return f
}

func (f *fixture) newController() (*DestinationRuleController, kubeinformers.SharedInformerFactory, istioinformers.SharedInformerFactory, informers.SharedInformerFactory, error) {
	f.kubeClient = kubefake.NewSimpleClientset(f.kubeObjects...)
	f.servicemeshClient = fake.NewSimpleClientset(f.servicemeshObjects...)
	f.istioClient = istiofake.NewSimpleClientset(f.istioObjects...)
	kubeInformers := kubeinformers.NewSharedInformerFactory(f.kubeClient, 0)
	istioInformers := istioinformers.NewSharedInformerFactory(f.istioClient, 0)
	servicemeshInformers := informers.NewSharedInformerFactory(f.servicemeshClient, 0)

	c := NewDestinationRuleController(kubeInformers.Apps().V1().Deployments(),
		istioInformers.Networking().V1alpha3().DestinationRules(),
		kubeInformers.Core().V1().Services(),
		servicemeshInformers.Servicemesh().V1alpha2().ServicePolicies(),
		f.kubeClient,
		f.istioClient,
		f.servicemeshClient)
	c.eventRecorder = &record.FakeRecorder{}
	c.destinationRuleSynced = alwaysReady
	c.deploymentSynced = alwaysReady
	c.servicePolicySynced = alwaysReady
	c.serviceSynced = alwaysReady

	for _, s := range f.serviceLister {
		kubeInformers.Core().V1().Services().Informer().GetIndexer().Add(s)
	}

	for _, d := range f.drLister {
		istioInformers.Networking().V1alpha3().DestinationRules().Informer().GetIndexer().Add(d)
	}

	for _, d := range f.deploymentLister {
		kubeInformers.Apps().V1().Deployments().Informer().GetIndexer().Add(d)
	}

	for _, s := range f.spLister {
		servicemeshInformers.Servicemesh().V1alpha2().ServicePolicies().Informer().GetIndexer().Add(s)
	}

	return c, kubeInformers, istioInformers, servicemeshInformers, nil
}

func (f *fixture) run(service *corev1.Service, expected *v1alpha3.DestinationRule, startInformers bool, expectedError bool) {
	c, kubeInformers, istioInformers, servicemeshInformers, err := f.newController()
	if err != nil {
		f.t.Fatal(err)
	}

	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		kubeInformers.Start(stopCh)
		istioInformers.Start(stopCh)
		servicemeshInformers.Start(stopCh)
	}

	key, err := cache.MetaNamespaceKeyFunc(service)
	if err != nil {
		f.t.Fatal(err)
	}

	err = c.syncService(key)
	if !expectedError && err != nil {
		f.t.Fatalf("error syncing service: %v", err)
	} else if expectedError && err == nil {
		f.t.Fatal("expected error syncing service, got nil")
	}

	got, err := c.destinationRuleClient.NetworkingV1alpha3().DestinationRules(service.Namespace).Get(service.Name, metav1.GetOptions{})
	if err != nil {
		f.t.Fatal(err)
	}

	if unequals := reflectutils.Equal(got, expected); len(unequals) != 0 {
		f.t.Errorf("expected %#v, got %#v, unequal fields:", expected, got)
		for _, unequal := range unequals {
			f.t.Error(unequal)
		}
	}

}

func runServicePolicy(t *testing.T, service *corev1.Service, sp *v1alpha2.ServicePolicy, expected *v1alpha3.DestinationRule, expectedError bool, deployments ...*appsv1.Deployment) {
	f := newFixture(t)

	f.kubeObjects = append(f.kubeObjects, service)
	f.serviceLister = append(f.serviceLister, service)
	for _, deployment := range deployments {
		f.kubeObjects = append(f.kubeObjects, deployment)
		f.deploymentLister = append(f.deploymentLister, deployment)
	}
	if sp != nil {
		f.servicemeshObjects = append(f.servicemeshObjects, sp)
		f.spLister = append(f.spLister, sp)
	}

	f.run(service, expected, true, expectedError)
}

func TestServicePolicy(t *testing.T) {
	defaultService := newService("foo")
	defaultDeploymentV1 := newDeployments(defaultService, "v1")
	defaultDeploymentV2 := newDeployments(defaultService, "v2")
	defaultServicePolicy := newServicePolicy("foo", defaultService, defaultDeploymentV1, defaultDeploymentV2)
	defaultExpected := newDestinationRule(defaultService, defaultDeploymentV1, defaultDeploymentV2)

	t.Run("should create default destination rule", func(t *testing.T) {
		runServicePolicy(t, defaultService, nil, defaultExpected, false, defaultDeploymentV1, defaultDeploymentV2)
	})

	t.Run("should create destination rule only to v1", func(t *testing.T) {
		deploymentV2 := defaultDeploymentV2.DeepCopy()
		deploymentV2.Status.AvailableReplicas = 0
		deploymentV2.Status.ReadyReplicas = 0

		expected := defaultExpected.DeepCopy()
		expected.Spec.Subsets = expected.Spec.Subsets[:1]
		runServicePolicy(t, defaultService, nil, expected, false, defaultDeploymentV1, deploymentV2)
	})

	t.Run("should create destination rule match service policy", func(t *testing.T) {
		sp := defaultServicePolicy.DeepCopy()
		sp.Spec.Template.Spec.TrafficPolicy = &apiv1alpha3.TrafficPolicy{
			LoadBalancer: &apiv1alpha3.LoadBalancerSettings{
				LbPolicy: &apiv1alpha3.LoadBalancerSettings_Simple{
					Simple: apiv1alpha3.LoadBalancerSettings_ROUND_ROBIN,
				},
			},
			ConnectionPool: &apiv1alpha3.ConnectionPoolSettings{
				Http: &apiv1alpha3.ConnectionPoolSettings_HTTPSettings{
					Http1MaxPendingRequests:  10,
					Http2MaxRequests:         20,
					MaxRequestsPerConnection: 5,
					MaxRetries:               4,
				},
			},
			OutlierDetection: &apiv1alpha3.OutlierDetection{
				ConsecutiveErrors:  5,
				MaxEjectionPercent: 10,
				MinHealthPercent:   20,
			},
		}

		expected := defaultExpected.DeepCopy()
		expected.Spec.TrafficPolicy = sp.Spec.Template.Spec.TrafficPolicy
		runServicePolicy(t, defaultService, sp, expected, false, defaultDeploymentV1, defaultDeploymentV2)
	})
}
