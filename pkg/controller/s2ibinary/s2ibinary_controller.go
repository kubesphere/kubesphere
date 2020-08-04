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

package s2ibinary

import (
	"fmt"
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
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"time"

	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	devopsclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha1"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha1"
)

/**
s2ibinary-controller used to handle s2ibinary's delete logic.
s2ibinary creation and file upload provided by kubesphere/kapis
*/
type Controller struct {
	client       clientset.Interface
	devopsClient devopsclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	s2iBinaryLister devopslisters.S2iBinaryLister
	s2iBinarySynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	s3Client s3.Interface
}

func NewController(client clientset.Interface,
	devopsclientset devopsclient.Interface,
	s2ibinInformer devopsinformers.S2iBinaryInformer,
	s3Client s3.Interface) *Controller {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "s2ibinary-controller"})

	v := &Controller{
		client:           client,
		devopsClient:     devopsclientset,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "s2ibinary"),
		s2iBinaryLister:  s2ibinInformer.Lister(),
		s2iBinarySynced:  s2ibinInformer.Informer().HasSynced,
		workerLoopPeriod: time.Second,
		s3Client:         s3Client,
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
func (c *Controller) enqueueS2iBinary(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) processNextWorkItem() bool {
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

func (c *Controller) worker() {

	for c.processNextWorkItem() {
	}
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(1, stopCh)
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
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
func (c *Controller) syncHandler(key string) error {
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

func (c *Controller) deleteBinaryInS3(s2ibin *devopsv1alpha1.S2iBinary) error {

	key := fmt.Sprintf("%s-%s", s2ibin.Namespace, s2ibin.Name)
	err := c.s3Client.Delete(key)
	if err != nil {
		klog.Errorf("error happened while deleting %s, %v", key, err)
	}

	return nil
}
