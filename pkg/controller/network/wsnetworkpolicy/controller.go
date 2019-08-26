package wsnetworkpolicy

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8snetwork "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1informer "k8s.io/client-go/informers/core/v1"
	k8snetworkinformer "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	k8snetworklister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	workspaceapi "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	kubespherescheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	networkinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/network/v1alpha1"
	workspaceinformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/tenant/v1alpha1"
	networklister "kubesphere.io/kubesphere/pkg/client/listers/network/v1alpha1"
	workspacelister "kubesphere.io/kubesphere/pkg/client/listers/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/controller/network/controllerapi"
)

const controllerAgentName = "wsnp-controller"

var (
	log      = klogr.New().WithName("Controller").WithValues(controllerAgentName)
	errCount = 0
)

type controller struct {
	kubeClientset       kubernetes.Interface
	kubesphereClientset kubesphereclient.Interface

	wsnpInformer networkinformer.WorkspaceNetworkPolicyInformer
	wsnpLister   networklister.WorkspaceNetworkPolicyLister
	wsnpSynced   cache.InformerSynced

	networkPolicyInformer k8snetworkinformer.NetworkPolicyInformer
	networkPolicyLister   k8snetworklister.NetworkPolicyLister
	networkPolicySynced   cache.InformerSynced

	namespaceLister   corev1lister.NamespaceLister
	namespaceInformer corev1informer.NamespaceInformer
	namespaceSynced   cache.InformerSynced

	workspaceLister   workspacelister.WorkspaceLister
	workspaceInformer workspaceinformer.WorkspaceInformer
	workspaceSynced   cache.InformerSynced
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewController(kubeclientset kubernetes.Interface,
	kubesphereclientset kubesphereclient.Interface,
	wsnpInformer networkinformer.WorkspaceNetworkPolicyInformer,
	networkPolicyInformer k8snetworkinformer.NetworkPolicyInformer,
	namespaceInformer corev1informer.NamespaceInformer,
	workspaceInformer workspaceinformer.WorkspaceInformer) controllerapi.Controller {
	utilruntime.Must(kubespherescheme.AddToScheme(scheme.Scheme))
	log.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	ctl := &controller{
		kubeClientset:         kubeclientset,
		kubesphereClientset:   kubesphereclientset,
		wsnpInformer:          wsnpInformer,
		wsnpLister:            wsnpInformer.Lister(),
		wsnpSynced:            wsnpInformer.Informer().HasSynced,
		networkPolicyInformer: networkPolicyInformer,
		networkPolicyLister:   networkPolicyInformer.Lister(),
		networkPolicySynced:   networkPolicyInformer.Informer().HasSynced,
		namespaceInformer:     namespaceInformer,
		namespaceLister:       namespaceInformer.Lister(),
		namespaceSynced:       namespaceInformer.Informer().HasSynced,
		workspaceInformer:     workspaceInformer,
		workspaceLister:       workspaceInformer.Lister(),
		workspaceSynced:       workspaceInformer.Informer().HasSynced,

		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "WorkspaceNetworkPolicies"),
		recorder:  recorder,
	}
	log.Info("Setting up event handlers")
	wsnpInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueWSNP,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueWSNP(new)
		},
	})
	networkPolicyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.handleNP,
		UpdateFunc: func(old, new interface{}) {
			newNP := new.(*k8snetwork.NetworkPolicy)
			oldNP := old.(*k8snetwork.NetworkPolicy)
			if newNP.ResourceVersion == oldNP.ResourceVersion {
				return
			}
			ctl.handleNP(new)
		},
		DeleteFunc: ctl.handleNP,
	})
	workspaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.handleWS,
		UpdateFunc: func(old, new interface{}) {
			newNP := new.(*workspaceapi.Workspace)
			oldNP := old.(*workspaceapi.Workspace)
			if newNP.ResourceVersion == oldNP.ResourceVersion {
				return
			}
			ctl.handleWS(new)
		},
		DeleteFunc: ctl.handleNP,
	})
	return ctl
}

func (c *controller) handleWS(obj interface{}) {
	ws := obj.(*workspaceapi.Workspace)
	wsnps, err := c.wsnpLister.List(labels.Everything())
	if err != nil {
		log.Error(err, "Failed to get WSNP when a workspace changed ")
		return
	}
	for _, wsnp := range wsnps {
		log.V(4).Info("Enqueue wsnp because a workspace being  changed", "obj", ws.Name)
		c.enqueueWSNP(wsnp)
	}
	return
}

func (c *controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Info("Starting WSNP controller")

	// Wait for the caches to be synced before starting workers
	log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.wsnpSynced, c.namespaceSynced, c.networkPolicySynced, c.workspaceSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	log.Info("Shutting down workers")

	return nil
}

func (c *controller) enqueueWSNP(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *controller) handleNP(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		log.V(4).Info("Recovered deleted object from tombstone", "name", object.GetName())
	}
	log.V(4).Info("Processing object:", "name", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "WorkspaceNetworkPol" {
			return
		}

		wsnp, err := c.wsnpLister.Get(ownerRef.Name)
		if err != nil {
			log.V(4).Info("ignoring orphaned object", "link", object.GetSelfLink(), "name", ownerRef.Name)
			return
		}
		c.enqueueWSNP(wsnp)
		return
	}
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
		log.Info("Successfully synced", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *controller) handleError(err error) {
	log.Error(err, "Error in handling")
	errCount++
}
