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
	"reflect"
	"strings"

	apinetworkingv1alpha3 "istio.io/api/networking/v1alpha3"
	clientgonetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	log "k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice/util"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
	istioinformers "istio.io/client-go/pkg/informers/externalversions/networking/v1alpha3"
	istiolisters "istio.io/client-go/pkg/listers/networking/v1alpha3"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	servicemeshclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	servicemeshinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/servicemesh/v1alpha2"
	servicemeshlisters "kubesphere.io/kubesphere/pkg/client/listers/servicemesh/v1alpha2"

	"time"
)

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

type VirtualServiceController struct {
	client clientset.Interface

	virtualServiceClient istioclient.Interface
	servicemeshClient    servicemeshclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	serviceLister corelisters.ServiceLister
	serviceSynced cache.InformerSynced

	virtualServiceLister istiolisters.VirtualServiceLister
	virtualServiceSynced cache.InformerSynced

	destinationRuleLister istiolisters.DestinationRuleLister
	destinationRuleSynced cache.InformerSynced

	strategyLister servicemeshlisters.StrategyLister
	strategySynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewVirtualServiceController(serviceInformer coreinformers.ServiceInformer,
	virtualServiceInformer istioinformers.VirtualServiceInformer,
	destinationRuleInformer istioinformers.DestinationRuleInformer,
	strategyInformer servicemeshinformers.StrategyInformer,
	client clientset.Interface,
	virtualServiceClient istioclient.Interface,
	servicemeshClient servicemeshclient.Interface) *VirtualServiceController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		log.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "virtualservice-controller"})

	v := &VirtualServiceController{
		client:               client,
		virtualServiceClient: virtualServiceClient,
		servicemeshClient:    servicemeshClient,
		queue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "virtualservice"),
		workerLoopPeriod:     time.Second,
	}

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.enqueueService,
		DeleteFunc: v.enqueueService,
		UpdateFunc: func(old, cur interface{}) {
			// TODO(jeff): need a more robust mechanism, because user may change labels
			v.enqueueService(cur)
		},
	})

	v.serviceLister = serviceInformer.Lister()
	v.serviceSynced = serviceInformer.Informer().HasSynced

	v.strategyLister = strategyInformer.Lister()
	v.strategySynced = strategyInformer.Informer().HasSynced

	strategyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: v.addStrategy,
		AddFunc:    v.addStrategy,
		UpdateFunc: func(old, cur interface{}) {
			v.addStrategy(cur)
		},
	})

	v.destinationRuleLister = destinationRuleInformer.Lister()
	v.destinationRuleSynced = destinationRuleInformer.Informer().HasSynced

	destinationRuleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.addDestinationRule,
		UpdateFunc: func(old, cur interface{}) {
			v.addDestinationRule(cur)
		},
	})

	v.virtualServiceLister = virtualServiceInformer.Lister()
	v.virtualServiceSynced = virtualServiceInformer.Informer().HasSynced

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	return v

}

func (v *VirtualServiceController) Start(stopCh <-chan struct{}) error {
	return v.Run(5, stopCh)
}

func (v *VirtualServiceController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer v.queue.ShutDown()

	log.V(0).Info("starting virtualservice controller")
	defer log.Info("shutting down virtualservice controller")

	if !cache.WaitForCacheSync(stopCh, v.serviceSynced, v.virtualServiceSynced, v.destinationRuleSynced, v.strategySynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(v.worker, v.workerLoopPeriod, stopCh)
	}

	<-stopCh
	return nil
}

func (v *VirtualServiceController) enqueueService(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	v.queue.Add(key)
}

func (v *VirtualServiceController) worker() {

	for v.processNextWorkItem() {
	}
}

func (v *VirtualServiceController) processNextWorkItem() bool {
	eKey, quit := v.queue.Get()
	if quit {
		return false
	}

	defer v.queue.Done(eKey)

	err := v.syncService(eKey.(string))
	v.handleErr(err, eKey)

	return true
}

// created virtualservice's name are same as the service name, same
// as the destinationrule name
// labels:
//      servicemesh.kubernetes.io/enabled: ""
//      app.kubernetes.io/name: bookinfo
//      app: reviews
// are used to bind them together.
// syncService are the main part of reconcile function body, it takes
// service, destinationrule, strategy as input to create a virtualservice
// for service.
func (v *VirtualServiceController) syncService(key string) error {
	startTime := time.Now()
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Error(err, "not a valid controller key", "key", key)
		return err
	}

	// default component name to service name
	appName := name

	defer func() {
		log.V(4).Infof("Finished syncing service virtualservice %s/%s in %s.", namespace, name, time.Since(startTime))
	}()

	service, err := v.serviceLister.Services(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Delete the corresponding virtualservice, as the service has been deleted.
			err = v.virtualServiceClient.NetworkingV1alpha3().VirtualServices(namespace).Delete(name, nil)
			if err != nil && !errors.IsNotFound(err) {
				log.Error(err, "delete orphan virtualservice failed", "namespace", namespace, "name", service.Name)
				return err
			}

			// delete the orphan strategy if there is any
			err = v.servicemeshClient.ServicemeshV1alpha2().Strategies(namespace).Delete(name, nil)
			if err != nil && !errors.IsNotFound(err) {
				log.Error(err, "delete orphan strategy failed", "namespace", namespace, "name", service.Name)
				return err
			}

			return nil
		}
		log.Error(err, "get service failed", "namespace", namespace, "name", name)
		return err
	}

	if len(service.Labels) < len(util.ApplicationLabels) ||
		!util.IsApplicationComponent(service.Labels) ||
		!util.IsServicemeshEnabled(service.Annotations) ||
		len(service.Spec.Ports) == 0 {
		// services don't have enough labels to create a virtualservice
		// or they don't have necessary labels
		// or they don't have any ports defined
		return nil
	}
	// get real component name, i.e label app value
	appName = util.GetComponentName(&service.ObjectMeta)

	destinationRule, err := v.destinationRuleLister.DestinationRules(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// there is no destinationrule for this service
			// maybe corresponding workloads are not created yet
			log.Info("destination rules for service not found, retrying.", "namespace", namespace, "name", name)
			return fmt.Errorf("destination rule for service %s/%s not found", namespace, name)
		}
		log.Error(err, "Couldn't get destinationrule for service.", "service", types.NamespacedName{Name: service.Name, Namespace: service.Namespace}.String())
		return err
	}

	subsets := destinationRule.Spec.Subsets
	if len(subsets) == 0 {
		// destination rule with no subsets, not possibly
		return nil
	}

	// fetch all strategies applied to service
	strategies, err := v.strategyLister.Strategies(namespace).List(labels.SelectorFromSet(map[string]string{util.AppLabel: appName}))
	if err != nil {
		log.Error(err, "list strategies for service failed", "namespace", namespace, "name", appName)
		return err
	} else if len(strategies) > 1 {
		// more than one strategies are not allowed, it will cause collision
		err = fmt.Errorf("more than one strategies applied to service %s/%s is forbbiden", namespace, appName)
		log.Error(err, "")
		return err
	}

	// get current virtual service
	currentVirtualService, err := v.virtualServiceLister.VirtualServices(namespace).Get(appName)
	if err != nil {
		if errors.IsNotFound(err) {
			currentVirtualService = &clientgonetworkingv1alpha3.VirtualService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: namespace,
					Labels:    util.ExtractApplicationLabels(&service.ObjectMeta),
				},
			}
		} else {
			log.Error(err, "cannot get virtualservice ", "namespace", namespace, "name", appName)
			return err
		}
	}
	vs := currentVirtualService.DeepCopy()

	// create a whole new virtualservice

	// TODO(jeff): use FQDN to replace service name
	vs.Spec.Hosts = []string{name}

	// check if service has TCP protocol ports
	for _, port := range service.Spec.Ports {
		var route apinetworkingv1alpha3.HTTPRouteDestination
		if port.Protocol == v1.ProtocolTCP {
			route = apinetworkingv1alpha3.HTTPRouteDestination{
				Destination: &apinetworkingv1alpha3.Destination{
					Host:   name,
					Subset: subsets[0].Name,
					Port: &apinetworkingv1alpha3.PortSelector{
						Number: uint32(port.Port),
					},
				},
				Weight: 100,
			}

			// a http port, add to HTTPRoute
			if len(port.Name) > 0 && (port.Name == "http" || strings.HasPrefix(port.Name, "http-")) {
				vs.Spec.Http = []*apinetworkingv1alpha3.HTTPRoute{{Route: []*apinetworkingv1alpha3.HTTPRouteDestination{&route}}}
				break
			}

			// everything else treated as TCPRoute
			tcpRoute := apinetworkingv1alpha3.TCPRoute{
				Route: []*apinetworkingv1alpha3.RouteDestination{
					{
						Destination: route.Destination,
						Weight:      route.Weight,
					},
				},
			}
			vs.Spec.Tcp = []*apinetworkingv1alpha3.TCPRoute{&tcpRoute}
		}
	}

	if len(strategies) > 0 {
		// apply strategy spec to virtualservice

		switch strategies[0].Spec.StrategyPolicy {
		case servicemeshv1alpha2.PolicyPause:
			break
		case servicemeshv1alpha2.PolicyWaitForWorkloadReady:
			set := v.getSubsets(strategies[0])

			setNames := sets.String{}
			for i := range subsets {
				setNames.Insert(subsets[i].Name)
			}

			nonExist := false
			for k := range set {
				if !setNames.Has(k) {
					nonExist = true
				}
			}
			// strategy has subset that are not ready
			if nonExist {
				break
			} else {
				vs.Spec = v.generateVirtualServiceSpec(strategies[0], service).Spec
			}
		case servicemeshv1alpha2.PolicyImmediately:
			vs.Spec = v.generateVirtualServiceSpec(strategies[0], service).Spec
		default:
			vs.Spec = v.generateVirtualServiceSpec(strategies[0], service).Spec
		}

	}

	createVirtualService := len(currentVirtualService.ResourceVersion) == 0

	if !createVirtualService &&
		reflect.DeepEqual(vs.Spec, currentVirtualService.Spec) &&
		reflect.DeepEqual(service.Labels, currentVirtualService.Labels) {
		log.V(4).Info("virtual service are equal, skipping update ")
		return nil
	}

	newVirtualService := currentVirtualService.DeepCopy()
	newVirtualService.Labels = service.Labels
	newVirtualService.Spec = vs.Spec
	if newVirtualService.Annotations == nil {
		newVirtualService.Annotations = make(map[string]string)
	}

	if len(newVirtualService.Spec.Http) == 0 && len(newVirtualService.Spec.Tcp) == 0 && len(newVirtualService.Spec.Tls) == 0 {
		err = fmt.Errorf("service %s/%s doesn't have a valid port spec", namespace, name)
		log.Error(err, "")
		return err
	}

	if createVirtualService {
		_, err = v.virtualServiceClient.NetworkingV1alpha3().VirtualServices(namespace).Create(newVirtualService)
	} else {
		_, err = v.virtualServiceClient.NetworkingV1alpha3().VirtualServices(namespace).Update(newVirtualService)
	}

	if err != nil {
		if createVirtualService {
			v.eventRecorder.Event(newVirtualService, v1.EventTypeWarning, "FailedToCreateVirtualService", fmt.Sprintf("Failed to create virtualservice for service %v/%v: %v", namespace, name, err))
		} else {
			v.eventRecorder.Event(newVirtualService, v1.EventTypeWarning, "FailedToUpdateVirtualService", fmt.Sprintf("Failed to update virtualservice for service %v/%v: %v", namespace, name, err))
		}

		return err
	}

	return nil
}

// When a destinationrule is added, figure out which service it will be used
// and enqueue it. obj must have *v1alpha3.DestinationRule type
func (v *VirtualServiceController) addDestinationRule(obj interface{}) {
	dr := obj.(*clientgonetworkingv1alpha3.DestinationRule)
	service, err := v.serviceLister.Services(dr.Namespace).Get(dr.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(3).Info("service not created yet", "namespace", dr.Namespace, "service", dr.Name)
			return
		}
		utilruntime.HandleError(fmt.Errorf("unable to get service with name %s/%s", dr.Namespace, dr.Name))
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(service)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("get service %s/%s key failed", service.Namespace, service.Name))
		return
	}

	v.queue.Add(key)
}

// when a strategy created
func (v *VirtualServiceController) addStrategy(obj interface{}) {
	strategy := obj.(*servicemeshv1alpha2.Strategy)

	lbs := util.ExtractApplicationLabels(&strategy.ObjectMeta)
	if len(lbs) == 0 {
		err := fmt.Errorf("invalid strategy %s/%s labels %s, not have required labels", strategy.Namespace, strategy.Name, strategy.Labels)
		log.Error(err, "")
		utilruntime.HandleError(err)
		return
	}

	allServices, err := v.serviceLister.Services(strategy.Namespace).List(labels.SelectorFromSet(lbs))
	if err != nil {
		log.Error(err, "list services failed")
		utilruntime.HandleError(err)
		return
	}

	// avoid insert a key multiple times
	set := sets.String{}

	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil || len(service.Spec.Ports) == 0 {
			// services with nil selectors match nothing, not everything.
			continue
		}

		key, err := cache.MetaNamespaceKeyFunc(service)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}
		set.Insert(key)
	}

	for key := range set {
		v.queue.Add(key)
	}
}

func (v *VirtualServiceController) handleErr(err error, key interface{}) {
	if err == nil {
		v.queue.Forget(key)
		return
	}

	if v.queue.NumRequeues(key) < maxRetries {
		log.V(2).Info("Error syncing virtualservice for service retrying.", "key", key, "error", err)
		v.queue.AddRateLimited(key)
		return
	}

	log.V(4).Info("Dropping service out of the queue.", "key", key, "error", err)
	v.queue.Forget(key)
	utilruntime.HandleError(err)
}

func (v *VirtualServiceController) getSubsets(strategy *servicemeshv1alpha2.Strategy) sets.String {
	set := sets.String{}

	for _, httpRoute := range strategy.Spec.Template.Spec.Http {
		for _, dw := range httpRoute.Route {
			set.Insert(dw.Destination.Subset)
		}

		if httpRoute.Mirror != nil {
			set.Insert(httpRoute.Mirror.Subset)
		}
	}

	for _, tcpRoute := range strategy.Spec.Template.Spec.Tcp {
		for _, dw := range tcpRoute.Route {
			set.Insert(dw.Destination.Subset)
		}
	}

	for _, tlsRoute := range strategy.Spec.Template.Spec.Tls {
		for _, dw := range tlsRoute.Route {
			set.Insert(dw.Destination.Subset)
		}
	}

	return set
}

func (v *VirtualServiceController) generateVirtualServiceSpec(strategy *servicemeshv1alpha2.Strategy, service *v1.Service) *clientgonetworkingv1alpha3.VirtualService {

	// Define VirtualService to be created
	vs := &clientgonetworkingv1alpha3.VirtualService{
		Spec: strategy.Spec.Template.Spec,
	}

	// one version rules them all
	if len(strategy.Spec.GovernorVersion) > 0 {
		governorDestinationWeight := apinetworkingv1alpha3.HTTPRouteDestination{
			Destination: &apinetworkingv1alpha3.Destination{
				Host:   service.Name,
				Subset: strategy.Spec.GovernorVersion,
			},
			Weight: 100,
		}

		if len(strategy.Spec.Template.Spec.Http) > 0 {
			governorRoute := apinetworkingv1alpha3.HTTPRoute{
				Route: []*apinetworkingv1alpha3.HTTPRouteDestination{&governorDestinationWeight},
			}

			vs.Spec.Http = []*apinetworkingv1alpha3.HTTPRoute{&governorRoute}
		} else if len(strategy.Spec.Template.Spec.Tcp) > 0 {
			tcpRoute := apinetworkingv1alpha3.TCPRoute{
				Route: []*apinetworkingv1alpha3.RouteDestination{
					{
						Destination: &apinetworkingv1alpha3.Destination{
							Host:   governorDestinationWeight.Destination.Host,
							Subset: governorDestinationWeight.Destination.Subset,
						},
						Weight: governorDestinationWeight.Weight,
					},
				},
			}

			//governorRoute := v1alpha3.TCPRoute{tcpRoute}
			vs.Spec.Tcp = []*apinetworkingv1alpha3.TCPRoute{&tcpRoute}
		}

	}

	util.FillDestinationPort(vs, service)
	return vs
}
