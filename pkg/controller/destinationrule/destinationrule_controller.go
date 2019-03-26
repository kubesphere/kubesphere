package destinationrule

import (
	"fmt"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/util/metrics"
	"kubesphere.io/kubesphere/pkg/controller/virtualservice/util"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	istioclientset "github.com/knative/pkg/client/clientset/versioned"
	istioinformers "github.com/knative/pkg/client/informers/externalversions/istio/v1alpha3"
	istiolisters "github.com/knative/pkg/client/listers/istio/v1alpha3"
	informersv1 "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
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

var log = logf.Log.WithName("destinationrule-controller")

type DestinationRuleController struct {
	client clientset.Interface

	destinationRuleClient istioclientset.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	serviceLister corelisters.ServiceLister
	serviceSynced cache.InformerSynced

	deploymentLister listersv1.DeploymentLister
	deploymentSynced cache.InformerSynced

	destinationRuleLister istiolisters.DestinationRuleLister
	destinationRuleSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewDestinationRuleController(deploymentInformer informersv1.DeploymentInformer,
	destinationRuleInformer istioinformers.DestinationRuleInformer,
	serviceInformer coreinformers.ServiceInformer,
	client clientset.Interface,
	destinationRuleClient istioclientset.Interface) *DestinationRuleController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		log.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "destinationrule-controller"})

	if client != nil && client.CoreV1().RESTClient().GetRateLimiter() != nil {
		metrics.RegisterMetricAndTrackRateLimiterUsage("virtualservice_controller", client.CoreV1().RESTClient().GetRateLimiter())
	}

	v := &DestinationRuleController{
		client:                client,
		destinationRuleClient: destinationRuleClient,
		queue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "destinationrule"),
		workerLoopPeriod:      time.Second,
	}

	v.deploymentLister = deploymentInformer.Lister()
	v.deploymentSynced = deploymentInformer.Informer().HasSynced

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.addDeployment,
		DeleteFunc: v.deleteDeployment,
		UpdateFunc: func(old, cur interface{}) {
			v.addDeployment(cur)
		},
	})

	v.serviceLister = serviceInformer.Lister()
	v.serviceSynced = serviceInformer.Informer().HasSynced

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    v.enqueueService,
		DeleteFunc: v.enqueueService,
		UpdateFunc: func(old, cur interface{}) {
			v.enqueueService(cur)
		},
	})

	v.destinationRuleLister = destinationRuleInformer.Lister()
	v.destinationRuleSynced = destinationRuleInformer.Informer().HasSynced

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	return v

}

func (v *DestinationRuleController) Start(stopCh <-chan struct{}) error {
	v.Run(5, stopCh)

	return nil
}

func (v *DestinationRuleController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer v.queue.ShutDown()

	log.Info("starting destinationrule controller")
	defer log.Info("shutting down destinationrule controller")

	if !controller.WaitForCacheSync("destinationrule-controller", stopCh, v.serviceSynced, v.destinationRuleSynced, v.deploymentSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(v.worker, v.workerLoopPeriod, stopCh)
	}

	<-stopCh
}

func (v *DestinationRuleController) enqueueService(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	v.queue.Add(key)
}

func (v *DestinationRuleController) worker() {
	for v.processNextWorkItem() {

	}
}

func (v *DestinationRuleController) processNextWorkItem() bool {
	eKey, quit := v.queue.Get()
	if quit {
		return false
	}

	defer v.queue.Done(eKey)

	err := v.syncService(eKey.(string))
	v.handleErr(err, eKey)

	return true
}

func (v *DestinationRuleController) syncService(key string) error {
	startTime := time.Now()
	defer func() {
		log.V(4).Info("Finished syncing service destinationrule.", "key", key, "duration", time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	service, err := v.serviceLister.Services(namespace).Get(name)
	if err != nil {
		// Delete the corresponding destinationrule, as the service has been deleted.
		err = v.destinationRuleClient.NetworkingV1alpha3().DestinationRules(namespace).Delete(name, nil)
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

	deployments, err := v.deploymentLister.Deployments(namespace).List(labels.Set(service.Spec.Selector).AsSelectorPreValidated())
	if err != nil {
		return err
	}

	subsets := []v1alpha3.Subset{}
	for _, deployment := range deployments {

		version := util.GetComponentVersion(&deployment.ObjectMeta)

		if len(version) == 0 {
			log.V(4).Info("Deployment doesn't have a version label", "key", types.NamespacedName{Namespace: deployment.Namespace, Name: deployment.Name}.String())
			continue
		}

		subset := v1alpha3.Subset{
			Name: util.NormalizeVersionName(version),
			Labels: map[string]string{
				util.VersionLabel: version,
			},
		}

		subsets = append(subsets, subset)
	}

	currentDestinationRule, err := v.destinationRuleLister.DestinationRules(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			currentDestinationRule = &v1alpha3.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:   service.Name,
					Labels: service.Labels,
				},
				Spec: v1alpha3.DestinationRuleSpec{
					Host: name,
				},
			}
		} else {
			log.Error(err, "Couldn't get destinationrule for service", "key", key)
			return err
		}

	}

	createDestinationRule := len(currentDestinationRule.ResourceVersion) == 0

	if !createDestinationRule && reflect.DeepEqual(currentDestinationRule.Spec.Subsets, subsets) &&
		reflect.DeepEqual(currentDestinationRule.Labels, service.Labels) {
		log.V(5).Info("destinationrule are equal, skipping update", "key", types.NamespacedName{Namespace: service.Namespace, Name: service.Name}.String())
		return nil
	}

	newDestinationRule := currentDestinationRule.DeepCopy()
	newDestinationRule.Spec.Subsets = subsets
	newDestinationRule.Labels = service.Labels
	if newDestinationRule.Annotations == nil {
		newDestinationRule.Annotations = make(map[string]string)
	}

	if createDestinationRule {
		_, err = v.destinationRuleClient.NetworkingV1alpha3().DestinationRules(namespace).Create(newDestinationRule)
	} else {
		_, err = v.destinationRuleClient.NetworkingV1alpha3().DestinationRules(namespace).Update(newDestinationRule)
	}

	if err != nil {
		if createDestinationRule && errors.IsForbidden(err) {
			// A request is forbidden primarily for two reasons:
			// 1. namespace is terminating, endpoint creation is not allowed by default.
			// 2. policy is misconfigured, in which case no service would function anywhere.
			// Given the frequency of 1, we log at a lower level.
			log.V(5).Info("Forbidden from creating endpoints", "error", err)
		}

		if createDestinationRule {
			v.eventRecorder.Event(newDestinationRule, v1.EventTypeWarning, "FailedToCreateDestinationRule", fmt.Sprintf("Failed to create destinationrule for service %v/%v: %v", service.Namespace, service.Name, err))
		} else {
			v.eventRecorder.Event(newDestinationRule, v1.EventTypeWarning, "FailedToUpdateDestinationRule", fmt.Sprintf("Failed to update destinationrule for service %v/%v: %v", service.Namespace, service.Name, err))
		}

		return err
	}

	return nil
}

func (v *DestinationRuleController) isApplicationComponent(meta *metav1.ObjectMeta) bool {
	if len(meta.Labels) >= len(util.ApplicationLabels) && util.IsApplicationComponent(meta) {
		return true
	}
	return false
}

// When a destinationrule is added, figure out which service it will be used
// and enqueue it. obj must have *appsv1.Deployment type
func (v *DestinationRuleController) addDeployment(obj interface{}) {
	deploy := obj.(*appsv1.Deployment)

	// not a application component
	if !v.isApplicationComponent(&deploy.ObjectMeta) {
		return
	}

	services, err := v.getDeploymentServiceMemberShip(deploy)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to get deployment %s/%s's service memberships", deploy.Namespace, deploy.Name))
		return
	}

	for key := range services {
		v.queue.Add(key)
	}

	return
}

func (v *DestinationRuleController) deleteDeployment(obj interface{}) {
	if _, ok := obj.(*appsv1.Deployment); ok {
		v.addDeployment(obj)
		return
	}

	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
		return
	}

	deploy, ok := tombstone.Obj.(*appsv1.Deployment)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a deployment %#v", obj))
		return
	}

	v.addDeployment(deploy)
}

func (v *DestinationRuleController) getDeploymentServiceMemberShip(deployment *appsv1.Deployment) (sets.String, error) {
	set := sets.String{}

	allServices, err := v.serviceLister.Services(deployment.Namespace).List(labels.Everything())
	if err != nil {
		return set, err
	}

	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil || !v.isApplicationComponent(&service.ObjectMeta) {
			// services with nil selectors match nothing, not everything.
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(deployment.Spec.Selector.MatchLabels)) {
			key, err := controller.KeyFunc(service)
			if err != nil {
				return nil, err
			}
			set.Insert(key)
		}
	}

	return set, nil
}

func (v *DestinationRuleController) handleErr(err error, key interface{}) {
	if err != nil {
		v.queue.Forget(key)
		return
	}

	if v.queue.NumRequeues(key) < maxRetries {
		log.V(2).Info("Error syncing virtualservice for service, retrying.", "key", key, "error", err)
		v.queue.AddRateLimited(key)
		return
	}

	log.V(0).Info("Dropping service out of the queue", "key", key, "error", err)
	v.queue.Forget(key)
	utilruntime.HandleError(err)
}
