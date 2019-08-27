package s2ibinary

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/metrics"
	"kubesphere.io/kubesphere/pkg/simple/client/s2is3"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"time"

	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	devopsclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha1"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha1"
)

var log = logf.Log.WithName("s2ibinary-controller")

type S2iBinaryController struct {
	client       clientset.Interface
	devopsClient devopsclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	s2iBinaryLister devopslisters.S2iBinaryLister
	s2iBinarySynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewController(devopsclientset devopsclient.Interface,
	client clientset.Interface,
	s2ibinInformer devopsinformers.S2iBinaryInformer) *S2iBinaryController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		log.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "s2ibinary-controller"})

	if client != nil && client.CoreV1().RESTClient().GetRateLimiter() != nil {
		metrics.RegisterMetricAndTrackRateLimiterUsage("s2ibinary_controller", client.CoreV1().RESTClient().GetRateLimiter())
	}

	v := &S2iBinaryController{
		client:           client,
		devopsClient:     devopsclientset,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "s2ibinary"),
		s2iBinaryLister:  s2ibinInformer.Lister(),
		s2iBinarySynced:  s2ibinInformer.Informer().HasSynced,
		workerLoopPeriod: time.Second,
	}

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	s2ibinInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.enqueueFoo,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*devopsv1alpha1.S2iBinary)
			new := newObj.(*devopsv1alpha1.S2iBinary)
			if old.ResourceVersion == new.ResourceVersion {
				return
			}
			v.enqueueFoo(newObj)
		},
		DeleteFunc: v.enqueueFoo,
	})
	return v
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work workqueue. This method should *not* be
// passed resources of any type other than Foo.
func (c *S2iBinaryController) enqueueFoo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *S2iBinaryController) processNextWorkItem() bool {
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
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		log.Error(err, "could not reconcile s2ibinary")
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *S2iBinaryController) worker() {

	for c.processNextWorkItem() {
	}
}

func (c *S2iBinaryController) Start(stopCh <-chan struct{}) error {
	return c.Run(1, stopCh)
}

func (c *S2iBinaryController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	log.Info("starting s2ibinary controller")
	defer log.Info("shutting down s2ibinary controller")

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
func (c *S2iBinaryController) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Error(err, fmt.Sprintf("could not split s2ibin meta %s ", key))
		return nil
	}
	s2ibin, err := c.s2iBinaryLister.S2iBinaries(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("s2ibin '%s' in work queue no longer exists ", key))
			return nil
		}
		log.Error(err, fmt.Sprintf("could not get s2ibin %s ", key))
		return err
	}
	if s2ibin.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(s2ibin.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName) {
			s2ibin.ObjectMeta.Finalizers = append(s2ibin.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName)
			_, err := c.devopsClient.DevopsV1alpha1().S2iBinaries(namespace).Update(s2ibin)
			if err != nil {
				log.Error(err, fmt.Sprintf("failed to update s2ibin %s ", key))
				return err
			}
		}

	} else {
		if sliceutil.HasString(s2ibin.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName) {
			if err := c.DeleteBinaryInS3(s2ibin); err != nil {
				log.Error(err, fmt.Sprintf("failed to delete resource %s in s3", key))
				return err
			}
			s2ibin.ObjectMeta.Finalizers = sliceutil.RemoveString(s2ibin.ObjectMeta.Finalizers, func(item string) bool {
				return item == devopsv1alpha1.S2iBinaryFinalizerName
			})
			_, err := c.devopsClient.DevopsV1alpha1().S2iBinaries(namespace).Update(s2ibin)
			if err != nil {
				log.Error(err, fmt.Sprintf("failed to update s2ibin %s ", key))
				return err
			}
		}
	}

	return nil
}

func (c *S2iBinaryController) DeleteBinaryInS3(s2ibin *devopsv1alpha1.S2iBinary) error {
	s3client := s2is3.Client()
	input := &s3.DeleteObjectInput{
		Bucket: s2is3.Bucket(),
		Key:    aws.String(fmt.Sprintf("%s-%s", s2ibin.Namespace, s2ibin.Name)),
	}
	_, err := s3client.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil
			default:
				log.Error(err, fmt.Sprintf("failed to delete s2ibin %s/%s in s3", s2ibin.Namespace, s2ibin.Name))
				return err
			}
		} else {
			log.Error(err, fmt.Sprintf("failed to delete s2ibin %s/%s in s3", s2ibin.Namespace, s2ibin.Name))
			return err
		}
	}
	return nil
}
