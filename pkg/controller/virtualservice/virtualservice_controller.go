package virtualservice

import (
	"fmt"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/util/metrics"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice/util"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	istioclient "github.com/knative/pkg/client/clientset/versioned"
	istioinformers "github.com/knative/pkg/client/informers/externalversions/istio/v1alpha3"
	istiolisters "github.com/knative/pkg/client/listers/istio/v1alpha3"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
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

var log = logf.Log.WithName("virtualservice-controller")

type VirtualServiceController struct {
	client clientset.Interface

	virtualServiceClient istioclient.Interface

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
	virtualServiceClient istioclient.Interface) *VirtualServiceController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(log.Info)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "virtualservice-controller"})

	if client != nil && client.CoreV1().RESTClient().GetRateLimiter() != nil {
		metrics.RegisterMetricAndTrackRateLimiterUsage("virtualservice_controller", client.CoreV1().RESTClient().GetRateLimiter())
	}

	v := &VirtualServiceController{
		client:               client,
		virtualServiceClient: virtualServiceClient,
		queue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "virtualservice"),
		workerLoopPeriod:     time.Second,
	}

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.enqueueService,
		DeleteFunc: v.enqueueService,
		UpdateFunc: func(old, cur interface{}) {
			v.enqueueService(cur)
		},
	})

	v.serviceLister = serviceInformer.Lister()
	v.serviceSynced = serviceInformer.Informer().HasSynced

	v.strategyLister = strategyInformer.Lister()
	v.strategySynced = strategyInformer.Informer().HasSynced

	strategyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: v.deleteStrategy,
	})

	v.destinationRuleLister = destinationRuleInformer.Lister()
	v.destinationRuleSynced = destinationRuleInformer.Informer().HasSynced

	destinationRuleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.addDestinationRule,
	})

	v.virtualServiceLister = virtualServiceInformer.Lister()
	v.virtualServiceSynced = virtualServiceInformer.Informer().HasSynced

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	return v

}

func (v *VirtualServiceController) Start(stopCh <-chan struct{}) error {
	v.Run(1, stopCh)
	return nil
}

func (v *VirtualServiceController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer v.queue.ShutDown()

	log.Info("starting virtualservice controller")
	defer log.Info("shutting down virtualservice controller")

	if !controller.WaitForCacheSync("virtualservice-controller", stopCh, v.serviceSynced, v.virtualServiceSynced, v.destinationRuleSynced, v.strategySynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(v.worker, v.workerLoopPeriod, stopCh)
	}

	<-stopCh
}

func (v *VirtualServiceController) enqueueService(obj interface{}) {
	key, err := controller.KeyFunc(obj)
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

func (v *VirtualServiceController) syncService(key string) error {
	startTime := time.Now()
	defer func() {
		log.V(4).Info("Finished syncing service virtualservice. ", "service", key, "duration", time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	service, err := v.serviceLister.Services(namespace).Get(name)
	if err != nil {
		// Delete the corresponding virtualservice, as the service has been deleted.
		err = v.virtualServiceClient.NetworkingV1alpha3().VirtualServices(namespace).Delete(service.Name, nil)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		return nil
	}

	if len(service.Labels) < len(util.ApplicationLabels) || !util.IsApplicationComponent(&service.ObjectMeta) ||
		len(service.Spec.Ports) == 0 {
		// services don't have enough labels to create a virtualservice
		// or they don't have necessary labels
		// or they don't have any ports defined
		return nil
	}

	vs, err := v.virtualServiceLister.VirtualServices(namespace).Get(name)
	if err == nil {
		// there already is virtual service there, no need to create another one
		return nil
	}

	destinationRule, err := v.destinationRuleLister.DestinationRules(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// there is no destinationrule for this service
			// maybe corresponding workloads are not created yet
			return nil
		}
		log.Error(err, "Couldn't get destinationrule for service.", "service", types.NamespacedName{Name: service.Name, Namespace: service.Namespace}.String())
		return err
	}

	subsets := destinationRule.Spec.Subsets
	if len(subsets) == 0 {
		// destination rule with no subsets, not possibly
		err = fmt.Errorf("find destinationrule with no subsets for service %s", name)
		log.Error(err, "Find destinationrule with no subsets for service", "service", service.String())
		return err
	} else {
		vs = &v1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    util.ExtractApplicationLabels(&service.ObjectMeta),
			},
			Spec: v1alpha3.VirtualServiceSpec{
				Hosts: []string{name},
			},
		}

		// check if service has TCP protocol ports
		for _, port := range service.Spec.Ports {
			var route v1alpha3.DestinationWeight
			if port.Protocol == v1.ProtocolTCP {
				route = v1alpha3.DestinationWeight{
					Destination: v1alpha3.Destination{
						Host:   name,
						Subset: subsets[0].Name,
						Port: v1alpha3.PortSelector{
							Number: uint32(port.Port),
						},
					},
					Weight: 100,
				}

				// a http port, add to HTTPRoute
				if len(port.Name) > 0 && (port.Name == "http" || strings.HasPrefix(port.Name, "http-")) {
					vs.Spec.Http = []v1alpha3.HTTPRoute{{Route: []v1alpha3.DestinationWeight{route}}}
					break
				}

				// everything else treated as TCPRoute
				vs.Spec.Tcp = []v1alpha3.TCPRoute{{Route: []v1alpha3.DestinationWeight{route}}}
			}
		}

		if len(vs.Spec.Http) > 0 || len(vs.Spec.Tcp) > 0 {
			_, err := v.virtualServiceClient.NetworkingV1alpha3().VirtualServices(namespace).Create(vs)
			if err != nil {
				v.eventRecorder.Eventf(vs, v1.EventTypeWarning, "FailedToCreateVirtualService", "Failed to create virtualservice for service %v/%v: %v", service.Namespace, service.Name, err)
				log.Error(err, "create virtualservice for service failed.", "service", service)
				return err
			}
		} else {
			log.Info("service doesn't have a tcp port.")
			return nil
		}
	}

	return nil
}

// When a destinationrule is added, figure out which service it will be used
// and enqueue it. obj must have *v1alpha3.DestinationRule type
func (v *VirtualServiceController) addDestinationRule(obj interface{}) {
	dr := obj.(*v1alpha3.DestinationRule)
	service, err := v.serviceLister.Services(dr.Namespace).Get(dr.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(0).Info("service not created yet", "key", dr.Name)
			return
		}
		utilruntime.HandleError(fmt.Errorf("unable to get service with name %s/%s", dr.Namespace, dr.Name))
		return
	}

	_, err = v.virtualServiceLister.VirtualServices(dr.Namespace).Get(dr.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			key, err := controller.KeyFunc(service)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("get service %s/%s key failed", service.Namespace, service.Name))
				return
			}

			v.queue.Add(key)
		}
	} else {
		// Already have a virtualservice created.
	}

	return
}

func (v *VirtualServiceController) deleteStrategy(obj interface{}) {
	// nothing to do right now
}

func (v *VirtualServiceController) handleErr(err error, key interface{}) {
	if err != nil {
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
