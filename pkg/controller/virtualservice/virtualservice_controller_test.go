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
	"fmt"
	apiv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	alwaysReady     = func() bool { return true }
	serviceName     = "foo"
	applicationName = "bookinfo"
	namespace       = metav1.NamespaceDefault
	subsets         = []string{"v1", "v2"}
)

type fixture struct {
	t testing.TB

	kubeClient        *kubefake.Clientset
	istioClient       *istiofake.Clientset
	servicemeshClient *fake.Clientset

	serviceLister  []*v1.Service
	vrLister       []*v1alpha3.VirtualService
	drLister       []*v1alpha3.DestinationRule
	strategyLister []*v1alpha2.Strategy

	kubeObjects        []runtime.Object
	istioObjects       []runtime.Object
	servicemeshObjects []runtime.Object
}

type Labels map[string]string

func NewLabels() Labels {
	m := make(map[string]string)
	return m
}

func (l Labels) WithApp(name string) Labels {
	l["app"] = name
	return l
}

func (l Labels) WithVersion(version string) Labels {
	l["version"] = version
	return l
}

func (l Labels) WithApplication(name string) Labels {
	l["app.kubernetes.io/name"] = name
	l["app.kubernetes.io/version"] = ""
	return l
}

func (l Labels) WithServiceMeshEnabled(enabled bool) Labels {
	if enabled {
		l[util.ServiceMeshEnabledAnnotation] = "true"
	}

	return l
}

func newFixture(t testing.TB) *fixture {
	f := &fixture{}
	f.t = t
	f.kubeObjects = []runtime.Object{}
	f.istioObjects = []runtime.Object{}
	f.servicemeshObjects = []runtime.Object{}
	return f
}

func newVirtualService(name string, host string, labels map[string]string) *v1alpha3.VirtualService {
	vr := v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: make(map[string]string),
		},
		Spec: apiv1alpha3.VirtualService{
			Hosts: []string{host},
			Http:  nil,
			Tls:   nil,
			Tcp:   nil,
		},
	}

	return &vr
}

func newService(name string, labels map[string]string, selector map[string]string, protocol string, port int) *v1.Service {
	svc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       int32(port),
					Name:       fmt.Sprintf("%s-aaa", protocol),
					TargetPort: intstr.FromInt(port),
				},
			},
			Selector: selector,
			Type:     v1.ServiceTypeClusterIP,
		},
	}

	return &svc
}

func newDestinationRule(name string, host string, labels map[string]string, subsets ...string) *v1alpha3.DestinationRule {
	dr := v1alpha3.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
			Labels:    labels,
		},
		Spec: apiv1alpha3.DestinationRule{
			Host: host,
		},
	}
	dr.Spec.Subsets = []*apiv1alpha3.Subset{}
	for _, subset := range subsets {
		dr.Spec.Subsets = append(dr.Spec.Subsets, &apiv1alpha3.Subset{
			Name:   subset,
			Labels: labels,
		})
	}

	return &dr
}

func newStrategy(name string, service *v1.Service, principalVersion string) *v1alpha2.Strategy {
	st := v1alpha2.Strategy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   service.Namespace,
			Labels:      NewLabels().WithApp(""),
			Annotations: nil,
		},
		Spec: v1alpha2.StrategySpec{
			Type:             v1alpha2.CanaryType,
			PrincipalVersion: principalVersion,
			GovernorVersion:  "",
			Template: v1alpha2.VirtualServiceTemplateSpec{
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
			StrategyPolicy: v1alpha2.PolicyImmediately,
		},
	}

	return &st
}

func toHost(service *v1.Service) string {
	return fmt.Sprintf("%s.%s.svc", service.Name, service.Namespace)
}

func (f *fixture) newController() (*VirtualServiceController, kubeinformers.SharedInformerFactory, istioinformers.SharedInformerFactory, informers.SharedInformerFactory, error) {
	f.kubeClient = kubefake.NewSimpleClientset(f.kubeObjects...)
	f.servicemeshClient = fake.NewSimpleClientset(f.servicemeshObjects...)
	f.istioClient = istiofake.NewSimpleClientset(f.istioObjects...)
	kubeInformers := kubeinformers.NewSharedInformerFactory(f.kubeClient, 0)
	istioInformers := istioinformers.NewSharedInformerFactory(f.istioClient, 0)
	servicemeshInformers := informers.NewSharedInformerFactory(f.servicemeshClient, 0)

	c := NewVirtualServiceController(kubeInformers.Core().V1().Services(),
		istioInformers.Networking().V1alpha3().VirtualServices(),
		istioInformers.Networking().V1alpha3().DestinationRules(),
		servicemeshInformers.Servicemesh().V1alpha2().Strategies(),
		f.kubeClient,
		f.istioClient,
		f.servicemeshClient)
	c.eventRecorder = &record.FakeRecorder{}
	c.destinationRuleSynced = alwaysReady
	c.virtualServiceSynced = alwaysReady
	c.strategySynced = alwaysReady
	c.serviceSynced = alwaysReady

	for _, s := range f.serviceLister {
		kubeInformers.Core().V1().Services().Informer().GetIndexer().Add(s)
	}

	for _, d := range f.drLister {
		istioInformers.Networking().V1alpha3().DestinationRules().Informer().GetIndexer().Add(d)
	}

	for _, v := range f.vrLister {
		istioInformers.Networking().V1alpha3().VirtualServices().Informer().GetIndexer().Add(v)
	}

	for _, s := range f.strategyLister {
		servicemeshInformers.Servicemesh().V1alpha2().Strategies().Informer().GetIndexer().Add(s)
	}

	return c, kubeInformers, istioInformers, servicemeshInformers, nil
}

func (f *fixture) run(serviceKey string, expectedVirtualService *v1alpha3.VirtualService) {
	f.run_(serviceKey, expectedVirtualService, true, false)
}

func (f *fixture) run_(serviceKey string, expectedVS *v1alpha3.VirtualService, startInformers bool, expectError bool) {
	namespace, name, err := cache.SplitMetaNamespaceKey(serviceKey)
	if err != nil {
		f.t.Fatalf("service key %s is not valid", serviceKey)
	}

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

	err = c.syncService(serviceKey)
	if !expectError && err != nil {
		f.t.Errorf("error syncing service: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing service, got nil")
	}

	if expectedVS != nil {
		got, err := c.virtualServiceClient.NetworkingV1alpha3().VirtualServices(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			f.t.Errorf("error getting virtualservice: %v", err)
			return
		}

		if unequals := reflectutils.Equal(got, expectedVS); len(unequals) != 0 {
			f.t.Errorf("didn't get expected result, got %#v, unequal fields:", got)
			for _, unequal := range unequals {
				f.t.Errorf("%s", unequal)
			}
		}
	}
}

func TestInitialStrategyCreate(t *testing.T) {
	f := newFixture(t)

	svc := newService("foo", NewLabels().WithApplication(applicationName).WithApp(serviceName), NewLabels().WithApplication(serviceName).WithApp(applicationName), "http", 80)
	dr := newDestinationRule(svc.Name, toHost(svc), NewLabels().WithApp("foo").WithApplication(applicationName), subsets[0])
	svc.Annotations = NewLabels().WithServiceMeshEnabled(true)

	f.kubeObjects = append(f.kubeObjects, svc)
	f.serviceLister = append(f.serviceLister, svc)
	f.istioObjects = append(f.istioObjects, dr)
	f.drLister = append(f.drLister, dr)

	vs := newVirtualService(svc.Name, "foo", NewLabels().WithApplication("bookinfo").WithApp(svc.Name))
	vs.Annotations = make(map[string]string)
	vs.Spec.Http = []*apiv1alpha3.HTTPRoute{
		{
			Route: []*apiv1alpha3.HTTPRouteDestination{
				{
					Destination: &apiv1alpha3.Destination{
						Host:   svc.Name,
						Subset: "v1",
						Port: &apiv1alpha3.PortSelector{
							Number: uint32(svc.Spec.Ports[0].Port),
						},
					},
					Weight: 100,
				},
			},
		},
	}

	key, err := cache.MetaNamespaceKeyFunc(svc)
	if err != nil {
		t.Fatal(err)
	}
	f.run(key, vs)
}

func runStrategy(t *testing.T, svc *v1.Service, dr *v1alpha3.DestinationRule, strategy *v1alpha2.Strategy, expectedVS *v1alpha3.VirtualService) {
	key, err := cache.MetaNamespaceKeyFunc(svc)
	if err != nil {
		t.Fatal(err)
	}

	f := newFixture(t)

	f.kubeObjects = append(f.kubeObjects, svc)
	f.serviceLister = append(f.serviceLister, svc)
	f.istioObjects = append(f.istioObjects, dr)
	f.drLister = append(f.drLister, dr)
	f.servicemeshObjects = append(f.servicemeshObjects, strategy)
	f.strategyLister = append(f.strategyLister, strategy)

	f.run(key, expectedVS)
}

func TestStrategies(t *testing.T) {

	svc := newService(serviceName, NewLabels().WithApplication(applicationName).WithApp(serviceName), NewLabels().WithApplication(applicationName).WithApp(serviceName), "http", 80)
	defaultDr := newDestinationRule(svc.Name, toHost(svc), NewLabels().WithApp(serviceName).WithApplication(applicationName), subsets...)
	svc.Annotations = NewLabels().WithServiceMeshEnabled(true)
	defaultStrategy := &v1alpha2.Strategy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "foo",
			Namespace:   metav1.NamespaceDefault,
			Labels:      NewLabels().WithApp(serviceName).WithApplication(applicationName),
			Annotations: make(map[string]string),
		},
		Spec: v1alpha2.StrategySpec{
			Type:             v1alpha2.CanaryType,
			PrincipalVersion: "v1",
			Template: v1alpha2.VirtualServiceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: apiv1alpha3.VirtualService{
					Hosts: []string{serviceName},
					Http: []*apiv1alpha3.HTTPRoute{
						{
							Route: []*apiv1alpha3.HTTPRouteDestination{
								{
									Destination: &apiv1alpha3.Destination{
										Host:   serviceName,
										Subset: "v1",
										Port: &apiv1alpha3.PortSelector{
											Number: 0,
										},
									},
									Weight: 80,
								},
								{
									Destination: &apiv1alpha3.Destination{
										Host:   serviceName,
										Subset: "v2",
										Port: &apiv1alpha3.PortSelector{
											Number: 0,
										},
									},
									Weight: 20,
								},
							},
						},
					},
				},
			},
			StrategyPolicy: v1alpha2.PolicyImmediately,
		},
	}

	defaultExpected := newVirtualService(serviceName, serviceName, svc.Labels)
	defaultExpected.Spec.Http = []*apiv1alpha3.HTTPRoute{
		{
			Route: []*apiv1alpha3.HTTPRouteDestination{
				{
					Destination: &apiv1alpha3.Destination{
						Host:   svc.Name,
						Subset: "v1",
						Port: &apiv1alpha3.PortSelector{
							Number: uint32(svc.Spec.Ports[0].Port),
						},
					},
					Weight: 80,
				},
				{
					Destination: &apiv1alpha3.Destination{
						Host:   svc.Name,
						Subset: "v2",
						Port: &apiv1alpha3.PortSelector{
							Number: uint32(svc.Spec.Ports[0].Port),
						},
					},
					Weight: 20,
				},
			},
		},
	}

	t.Run("Canary: 80% v1 and 20% v2", func(t *testing.T) {
		runStrategy(t, svc, defaultDr, defaultStrategy, defaultExpected)
	})

	t.Run("Canary: 0% v1 and 100% v2", func(t *testing.T) {
		strategy := defaultStrategy.DeepCopy()
		strategy.Spec.Template.Spec.Http[0].Route[0].Weight = 0
		strategy.Spec.Template.Spec.Http[0].Route[1].Weight = 100

		expected := defaultExpected.DeepCopy()
		expected.Spec.Http[0].Route[0].Weight = 0
		expected.Spec.Http[0].Route[1].Weight = 100
		runStrategy(t, svc, defaultDr, strategy, expected)
	})

	t.Run("Canary: v2 is governing", func(t *testing.T) {
		strategy := defaultStrategy.DeepCopy()
		strategy.Spec.GovernorVersion = "v2"

		expected := defaultExpected.DeepCopy()
		expected.Spec.Http[0].Route[0].Weight = 100
		expected.Spec.Http[0].Route[0].Destination.Subset = "v2"
		expected.Spec.Http[0].Route = expected.Spec.Http[0].Route[:1]
		runStrategy(t, svc, defaultDr, strategy, expected)
	})

	t.Run("Canary: http match route", func(t *testing.T) {
		strategy := defaultStrategy.DeepCopy()
		strategy.Spec.Template.Spec.Http[0].Match = []*apiv1alpha3.HTTPMatchRequest{
			{
				Headers: map[string]*apiv1alpha3.StringMatch{
					"X-USER": {
						MatchType: &apiv1alpha3.StringMatch_Regex{Regex: "users"},
					},
				},
				Uri: &apiv1alpha3.StringMatch{
					MatchType: &apiv1alpha3.StringMatch_Prefix{Prefix: "/apis"},
				},
			},
		}

		expected := defaultExpected.DeepCopy()
		expected.Spec.Http[0].Match = []*apiv1alpha3.HTTPMatchRequest{
			{
				Headers: map[string]*apiv1alpha3.StringMatch{
					"X-USER": {
						MatchType: &apiv1alpha3.StringMatch_Regex{Regex: "users"},
					},
				},
				Uri: &apiv1alpha3.StringMatch{
					MatchType: &apiv1alpha3.StringMatch_Prefix{Prefix: "/apis"},
				},
			},
		}
		runStrategy(t, svc, defaultDr, strategy, expected)
	})

}
