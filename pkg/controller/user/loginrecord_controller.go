/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package user

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	"time"
)

type LoginRecordController struct {
	k8sClient           kubernetes.Interface
	ksClient            kubesphere.Interface
	loginRecordInformer iamv1alpha2informers.LoginRecordInformer
	loginRecordLister   iamv1alpha2listers.LoginRecordLister
	loginRecordSynced   cache.InformerSynced
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder                    record.EventRecorder
	loginHistoryRetentionPeriod time.Duration
}

func NewLoginRecordController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface,
	loginRecordInformer iamv1alpha2informers.LoginRecordInformer,
	loginHistoryRetentionPeriod time.Duration) *LoginRecordController {

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "loginrecord-controller"})
	ctl := &LoginRecordController{
		k8sClient:                   k8sClient,
		ksClient:                    ksClient,
		loginRecordInformer:         loginRecordInformer,
		loginRecordLister:           loginRecordInformer.Lister(),
		loginRecordSynced:           loginRecordInformer.Informer().HasSynced,
		loginHistoryRetentionPeriod: loginHistoryRetentionPeriod,
		workqueue:                   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "loginrecord"),
		recorder:                    recorder,
	}
	return ctl
}

func (c *LoginRecordController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting LoginRecord controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(stopCh, c.loginRecordSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	go wait.Until(func() {
		if err := c.sync(); err != nil {
			klog.Errorf("Error periodically sync user status, %v", err)
		}
	}, time.Hour, stopCh)

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
	return nil
}

func (c *LoginRecordController) enqueueLoginRecord(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *LoginRecordController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *LoginRecordController) processNextWorkItem() bool {
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
		klog.Infof("Successfully synced %s:%s", "key", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *LoginRecordController) reconcile(key string) error {
	loginRecord, err := c.loginRecordLister.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("login record '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}

	now := time.Now()
	if loginRecord.CreationTimestamp.Add(c.loginHistoryRetentionPeriod).Before(now) { // login record beyonds retention period
		if err = c.ksClient.IamV1alpha2().LoginRecords().Delete(loginRecord.Name, metav1.NewDeleteOptions(0)); err != nil {
			klog.Error(err)
			return err
		}
	} else { // put item back into the queue
		c.workqueue.AddAfter(key, loginRecord.CreationTimestamp.Add(c.loginHistoryRetentionPeriod).Sub(now))
	}
	c.recorder.Event(loginRecord, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *LoginRecordController) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

func (c *LoginRecordController) sync() error {
	records, err := c.loginRecordLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, record := range records {
		key, err := cache.MetaNamespaceKeyFunc(record)
		if err != nil {
			return err
		}
		c.workqueue.AddRateLimited(key)
	}
	return nil
}
