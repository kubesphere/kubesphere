package nsnetworkpolicy

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	kubespherescheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	networkinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/network/v1alpha1"
	networklister "kubesphere.io/kubesphere/pkg/client/listers/network/v1alpha1"
	"kubesphere.io/kubesphere/pkg/controller/network/controllerapi"
	"kubesphere.io/kubesphere/pkg/controller/network/provider"
)

const controllerAgentName = "nsnp-controller"

type controller struct {
	kubeClientset       kubernetes.Interface
	kubesphereClientset kubesphereclient.Interface

	nsnpInformer networkinformer.NamespaceNetworkPolicyInformer
	nsnpLister   networklister.NamespaceNetworkPolicyLister
	nsnpSynced   cache.InformerSynced
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder                record.EventRecorder
	nsNetworkPolicyProvider provider.NsNetworkPolicyProvider
}

var (
	log      = klogr.New().WithName("Controller").WithValues("Component", controllerAgentName)
	errCount = 0
)

func NewController(kubeclientset kubernetes.Interface,
	kubesphereclientset kubesphereclient.Interface,
	nsnpInformer networkinformer.NamespaceNetworkPolicyInformer,
	nsNetworkPolicyProvider provider.NsNetworkPolicyProvider) controllerapi.Controller {
	utilruntime.Must(kubespherescheme.AddToScheme(scheme.Scheme))
	log.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	ctl := &controller{
		kubeClientset:           kubeclientset,
		kubesphereClientset:     kubesphereclientset,
		nsnpInformer:            nsnpInformer,
		nsnpLister:              nsnpInformer.Lister(),
		nsnpSynced:              nsnpInformer.Informer().HasSynced,
		nsNetworkPolicyProvider: nsNetworkPolicyProvider,

		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "NamespaceNetworkPolicies"),
		recorder:  recorder,
	}
	log.Info("Setting up event handlers")
	nsnpInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueNSNP,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueNSNP(new)
		},
		DeleteFunc: ctl.enqueueNSNP,
	})
	return ctl
}

func (c *controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	//init client

	// Start the informer factories to begin populating the informer caches
	log.V(1).Info("Starting WSNP controller")

	// Wait for the caches to be synced before starting workers
	log.V(2).Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.nsnpSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.V(2).Info("Started workers")
	<-stopCh
	log.V(2).Info("Shutting down workers")
	return nil
}

func (c *controller) enqueueNSNP(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the reconcile, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.reconcile(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		log.Info("Successfully synced", "key", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}
