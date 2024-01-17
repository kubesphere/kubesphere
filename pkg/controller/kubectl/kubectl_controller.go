package kubectl

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinfomers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	Finalizer   = "liks.lixiang.com/kubectl-controller"
	parallelism = 4
)

type Controller struct {
	podSynced cache.InformerSynced
	workqueue workqueue.RateLimitingInterface
	getPod    func(namespace, name string) (*v1.Pod, error)
	updatePod func(pod *v1.Pod) (*v1.Pod, error)
	deletePod func(namespace, name string) error
	sync      func(key string) error
}

func NewController(k8sClient kubernetes.Interface, podInformer coreinfomers.PodInformer) *Controller {
	klog.Infof("create kubectl controller")
	ctl := &Controller{
		podSynced: podInformer.Informer().HasSynced,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kubectl"),
		getPod: func(namespace, name string) (*v1.Pod, error) {
			return podInformer.Lister().Pods(namespace).Get(name)
		},
		updatePod: func(pod *v1.Pod) (*v1.Pod, error) {
			return k8sClient.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
		},
		deletePod: func(namespace, name string) error {
			return k8sClient.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		},
		sync: nil,
	}
	ctl.sync = ctl.reconcile
	klog.Infof("Setting up kubectl controller event handlers")
	podInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			// 1. Pod in specific namespace.
			// 2. No controller owner.
			// 3. Name will kubectl prefix
			if pod, ok := obj.(*v1.Pod); ok && pod.Namespace == "term.ns" && len(pod.OwnerReferences) == 0 && strings.HasPrefix(pod.Name, "term.has") {
				return true
			}
			return false
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: ctl.enqueuePod,
			UpdateFunc: func(oldObj, newObj interface{}) {
				ctl.enqueuePod(newObj)
			},
			DeleteFunc: ctl.enqueuePod,
		},
	})
	return nil
}

func (c *Controller) reconcile(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("invalid kubectl pod key: %s", key)
		return err
	}
	// get the pod with this name
	sharedPod, err := c.getPod(namespace, name)
	if err != nil {
		// the user may no longer exist,in which case we stop processing
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("pod '%s' is work queue no longer exists", key))
			return nil
		}
		klog.Errorf("get kubectl pod %s failed: %v", key, err)
		return err
	}
	// Make a copy so we don't mutate the shared cache
	pod := sharedPod.DeepCopy()
	if pod.Status.Phase == v1.PodSucceeded {
		// check finalizer
		if sets.NewString(pod.ObjectMeta.Finalizers...).Has(Finalizer) {
			// remove our pod finalizer
			finalizers := sets.NewString(pod.ObjectMeta.Finalizers...)
			finalizers.Delete(Finalizer)
			pod.ObjectMeta.Finalizers = finalizers.List()
			if _, err := c.updatePod(pod); err != nil {
				klog.Errorf("update pod %s failed: %v", key, err)
				return err
			}
		} else if err := c.deletePod(namespace, name); err != nil {
			klog.Errorf("delete kubectl pod %s failed: %v", key, err)
			return err
		} else if pod.ObjectMeta.DeletionTimestamp.IsZero() {

		}
		return nil
	}
	return nil
}

func (c *Controller) Start(ctx context.Context) error {
	return c.Run(parallelism, ctx.Done())
}
func (c *Controller) Run(threadLines int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Infof("Starting kubectl controller")

	klog.Infof("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.podSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	klog.Infof("Starting workers")
	for i := 0; i < threadLines; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	klog.Infof("Started workers")
	<-stopCh
	klog.Infof("Shutting down worker")
	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		// we call Done here so the work queue knows we have finished processing this item
		// we also must remember to call Forget if we  do not want this work item being re-queued.For example,we do
		// not call Forget if a transient error occurs,instead the item is put back on the work queue and attempted again
		// after a back-off period
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue  but got %#v", obj))
			return nil
		}
		// Run the reconcile, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.sync(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced key:%s", key)
		return nil
	}(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) enqueuePod(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
