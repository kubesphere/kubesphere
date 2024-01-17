package kubectl

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
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
	return true
}
func (c *Controller) reconcile(key string) error {
	return nil
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
