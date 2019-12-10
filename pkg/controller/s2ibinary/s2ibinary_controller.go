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
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"time"

	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	devopsclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha1"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha1"
)

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
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "s2ibinary-controller"})

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
		AddFunc: v.enqueueS2iBinary,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*devopsv1alpha1.S2iBinary)
			new := newObj.(*devopsv1alpha1.S2iBinary)
			if old.ResourceVersion == new.ResourceVersion {
				return
			}
			v.enqueueS2iBinary(newObj)
		},
		DeleteFunc: v.enqueueS2iBinary,
	})
	return v
}

// enqueueS2iBinary takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work workqueue. This method should *not* be
// passed resources of any type other than S2iBinary.
func (c *S2iBinaryController) enqueueS2iBinary(obj interface{}) {
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
		klog.Error(err, "could not reconcile s2ibinary")
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

	klog.Info("starting s2ibinary controller")
	defer klog.Info("shutting down s2ibinary controller")

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
		klog.Error(err, fmt.Sprintf("could not split s2ibin meta %s ", key))
		return nil
	}
	s2ibin, err := c.s2iBinaryLister.S2iBinaries(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("s2ibin '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Error(err, fmt.Sprintf("could not get s2ibin %s ", key))
		return err
	}
	if s2ibin.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(s2ibin.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName) {
			s2ibin.ObjectMeta.Finalizers = append(s2ibin.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName)
			_, err := c.devopsClient.DevopsV1alpha1().S2iBinaries(namespace).Update(s2ibin)
			if err != nil {
				klog.Error(err, fmt.Sprintf("failed to update s2ibin %s ", key))
				return err
			}
		}

	} else {
		if sliceutil.HasString(s2ibin.ObjectMeta.Finalizers, devopsv1alpha1.S2iBinaryFinalizerName) {
			if err := c.deleteBinaryInS3(s2ibin); err != nil {
				klog.Error(err, fmt.Sprintf("failed to delete resource %s in s3", key))
				return err
			}
			s2ibin.ObjectMeta.Finalizers = sliceutil.RemoveString(s2ibin.ObjectMeta.Finalizers, func(item string) bool {
				return item == devopsv1alpha1.S2iBinaryFinalizerName
			})
			_, err := c.devopsClient.DevopsV1alpha1().S2iBinaries(namespace).Update(s2ibin)
			if err != nil {
				klog.Error(err, fmt.Sprintf("failed to update s2ibin %s ", key))
				return err
			}
		}
	}

	return nil
}

func (c *S2iBinaryController) deleteBinaryInS3(s2ibin *devopsv1alpha1.S2iBinary) error {
	s3Client, err := client.ClientSets().S3()
	if err != nil {
		return err
	}

	input := &s3.DeleteObjectInput{
		Bucket: s3Client.Bucket(),
		Key:    aws.String(fmt.Sprintf("%s-%s", s2ibin.Namespace, s2ibin.Name)),
	}
	_, err = s3Client.Client().DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil
			default:
				klog.Error(err, fmt.Sprintf("failed to delete s2ibin %s/%s in s3", s2ibin.Namespace, s2ibin.Name))
				return err
			}
		} else {
			klog.Error(err, fmt.Sprintf("failed to delete s2ibin %s/%s in s3", s2ibin.Namespace, s2ibin.Name))
			return err
		}
	}
	return nil
}
