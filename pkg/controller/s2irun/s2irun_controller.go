package s2irun

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"time"

	s2iv1alpha1 "github.com/kubesphere/s2ioperator/pkg/apis/devops/v1alpha1"
	s2iclient "github.com/kubesphere/s2ioperator/pkg/client/clientset/versioned"
	s2iinformers "github.com/kubesphere/s2ioperator/pkg/client/informers/externalversions/devops/v1alpha1"
	s2ilisters "github.com/kubesphere/s2ioperator/pkg/client/listers/devops/v1alpha1"
	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	devopsclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha1"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha1"
)

type S2iRunController struct {
	client    clientset.Interface
	s2iClient s2iclient.Interface

	devopsClient devopsclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	s2iRunLister s2ilisters.S2iRunLister
	s2iRunSynced cache.InformerSynced

	s2iBinaryLister devopslisters.S2iBinaryLister
	s2iBinarySynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewController(devopsclientset devopsclient.Interface, s2iclientset s2iclient.Interface,
	client clientset.Interface,
	s2ibinInformer devopsinformers.S2iBinaryInformer, s2iRunInformer s2iinformers.S2iRunInformer) *S2iRunController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "s2irun-controller"})

	v := &S2iRunController{
		client:           client,
		devopsClient:     devopsclientset,
		s2iClient:        s2iclientset,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "s2irun"),
		s2iBinaryLister:  s2ibinInformer.Lister(),
		s2iBinarySynced:  s2ibinInformer.Informer().HasSynced,
		s2iRunLister:     s2iRunInformer.Lister(),
		s2iRunSynced:     s2iRunInformer.Informer().HasSynced,
		workerLoopPeriod: time.Second,
	}

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	s2iRunInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.enqueueS2iRun,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*s2iv1alpha1.S2iRun)
			new := newObj.(*s2iv1alpha1.S2iRun)
			if old.ResourceVersion == new.ResourceVersion {
				return
			}
			v.enqueueS2iRun(newObj)
		},
		DeleteFunc: v.enqueueS2iRun,
	})
	return v
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work workqueue. This method should *not* be
// passed resources of any type other than Foo.
func (c *S2iRunController) enqueueS2iRun(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *S2iRunController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		klog.V(5).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		klog.Error(err, "could not reconcile s2irun")
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *S2iRunController) worker() {

	for c.processNextWorkItem() {
	}
}

func (c *S2iRunController) Start(stopCh <-chan struct{}) error {
	return c.Run(1, stopCh)
}

func (c *S2iRunController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting s2irun controller")
	defer klog.Info("shutting down s2irun controller")

	if !cache.WaitForCacheSync(stopCh, c.s2iBinarySynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	<-stopCh
	return nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *S2iRunController) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Error(err, fmt.Sprintf("could not split s2irun meta %s ", key))
		return nil
	}
	s2irun, err := c.s2iRunLister.S2iRuns(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("s2irun '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Error(err, fmt.Sprintf("could not get s2irun %s ", key))
		return err
	}
	if s2irun.Labels != nil {
		_, ok := s2irun.Labels[devopsv1alpha1.S2iBinaryLabelKey]
		if ok {
			if s2irun.ObjectMeta.DeletionTimestamp.IsZero() {
				if !sliceutil.HasString(s2irun.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName) {
					s2irun.ObjectMeta.Finalizers = append(s2irun.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName)
					_, err := c.s2iClient.DevopsV1alpha1().S2iRuns(namespace).Update(s2irun)
					if err != nil {
						klog.Error(err, fmt.Sprintf("failed to update s2irun %s", key))
						return err
					}
				}

			} else {
				if sliceutil.HasString(s2irun.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName) {
					if err := c.DeleteS2iBinary(s2irun); err != nil {
						klog.Error(err, fmt.Sprintf("failed to delete s2ibin %s in", key))
						return err
					}
					s2irun.ObjectMeta.Finalizers = sliceutil.RemoveString(s2irun.ObjectMeta.Finalizers, func(item string) bool {
						return item == devopsv1alpha1.S2iBinaryFinalizerName
					})
					_, err := c.s2iClient.DevopsV1alpha1().S2iRuns(namespace).Update(s2irun)
					if err != nil {
						klog.Error(err, fmt.Sprintf("failed to update s2irun %s ", key))
						return err
					}
				}
			}
		}
	}

	return nil
}

func (c *S2iRunController) DeleteS2iBinary(s2irun *s2iv1alpha1.S2iRun) error {
	s2iBinName := s2irun.Labels[devopsv1alpha1.S2iBinaryLabelKey]
	s2iBin, err := c.s2iBinaryLister.S2iBinaries(s2irun.Namespace).Get(s2iBinName)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("s2ibin '%s/%s' has been delted ", s2irun.Namespace, s2iBinName))
			return nil
		}
		klog.Error(err, fmt.Sprintf("failed to get s2ibin %s/%s ", s2irun.Namespace, s2iBinName))
		return err
	}
	err = c.devopsClient.DevopsV1alpha1().S2iBinaries(s2iBin.Namespace).Delete(s2iBinName, nil)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("s2ibin '%s/%s' has been delted ", s2irun.Namespace, s2iBinName))
			return nil
		}
		klog.Error(err, fmt.Sprintf("failed to delete s2ibin %s/%s ", s2irun.Namespace, s2iBinName))
		return err
	}
	if err = c.cleanOtherS2iBinary(s2irun.Namespace); err != nil {
		klog.Error(err, fmt.Sprintf("failed to clean s2ibinary in %s", s2irun.Namespace))
	}

	return nil
}

// cleanOtherS2iBinary clean up s2ibinary created for more than 24 hours without associated s2irun
func (c *S2iRunController) cleanOtherS2iBinary(namespace string) error {
	s2iBins, err := c.s2iBinaryLister.S2iBinaries(namespace).List(nil)
	if err != nil {
		klog.Error(err, fmt.Sprintf("failed to list s2ibin in %s ", namespace))
		return err
	}
	now := time.Now()
	dayBefore := metav1.NewTime(now.Add(time.Hour * 24 * -1))
	for _, s2iBin := range s2iBins {
		if s2iBin.Status.Phase != devopsv1alpha1.StatusReady && s2iBin.CreationTimestamp.Before(&dayBefore) {

			runs, err := c.s2iRunLister.S2iRuns(namespace).List(labels.SelectorFromSet(labels.Set{devopsv1alpha1.S2iBinaryLabelKey: s2iBin.Name}))
			if err != nil {
				klog.Error(err, fmt.Sprintf("failed to list s2irun in  %s ", namespace))
				return err
			}
			if len(runs) == 0 {
				err = c.devopsClient.DevopsV1alpha1().S2iBinaries(namespace).Delete(s2iBin.Name, nil)
				if err != nil {
					if errors.IsNotFound(err) {
						klog.Info(fmt.Sprintf("s2ibin '%s/%s' has been deleted ", namespace, s2iBin.Name))
						return nil
					}
					klog.Error(err, fmt.Sprintf("failed to delete s2ibin %s/%s ", namespace, s2iBin.Name))
					return err
				}
			}
		}
	}
	return nil
}
