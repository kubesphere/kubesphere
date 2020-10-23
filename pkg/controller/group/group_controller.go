/*
Copyright 2020 The KubeSphere Authors.

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

package group

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	iam1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	iamv1alpha1listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced = "Group synced successfully"
	controllerName        = "groupbinding-controller"
	finalizer             = "finalizers.kubesphere.io/groups"
)

type Controller struct {
	scheme        *runtime.Scheme
	k8sClient     kubernetes.Interface
	ksClient      kubesphere.Interface
	groupInformer iamv1alpha2informers.GroupInformer
	groupLister   iamv1alpha1listers.GroupLister
	groupSynced   cache.InformerSynced
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

// NewController creates Group Controller instance
func NewController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface, groupInformer iamv1alpha2informers.GroupInformer) *Controller {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	ctl := &Controller{
		k8sClient:     k8sClient,
		ksClient:      ksClient,
		groupInformer: groupInformer,
		groupLister:   groupInformer.Lister(),
		groupSynced:   groupInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Group"),
		recorder:      recorder,
	}
	klog.Info("Setting up event handlers")
	groupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueGroup,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueGroup(new)
		},
		DeleteFunc: ctl.enqueueGroup,
	})
	return ctl
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Group controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")

	synced := []cache.InformerSynced{c.groupSynced}

	if ok := cache.WaitForCacheSync(stopCh, synced...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
	return nil
}

func (c *Controller) enqueueGroup(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
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

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) reconcile(key string) error {

	group, err := c.groupLister.Get(key)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("group '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}
	if group.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(group.Finalizers, finalizer) {
			group.ObjectMeta.Finalizers = append(group.ObjectMeta.Finalizers, finalizer)

			if group, err = c.ksClient.IamV1alpha2().Groups().Update(group); err != nil {
				return err
			}
			// Skip reconcile when group is updated.
			return nil
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(group.ObjectMeta.Finalizers, finalizer) {
			if err = c.deleteGroupBindings(group); err != nil {
				klog.Error(err)
				return err
			}

			if err = c.deleteRoleBindings(group); err != nil {
				klog.Error(err)
				return err
			}

			// remove our finalizer from the list and update it.
			group.Finalizers = sliceutil.RemoveString(group.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if group, err = c.ksClient.IamV1alpha2().Groups().Update(group); err != nil {
				return err
			}
		}
		// Our finalizer has finished, so the reconciler can do nothing.
		return nil
	}

	c.recorder.Event(group, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

func (c *Controller) deleteGroupBindings(group *iam1alpha2.Group) error {

	// Groupbindings that created by kubeshpere will be deleted directly.
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{iam1alpha2.GroupReferenceLabel: group.Name}).String(),
	}
	deleteOptions := metav1.NewDeleteOptions(0)

	if err := c.ksClient.IamV1alpha2().GroupBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *Controller) deleteRoleBindings(group *iam1alpha2.Group) error {
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{iam1alpha2.GroupReferenceLabel: group.Name}).String(),
	}
	deleteOptions := metav1.NewDeleteOptions(0)

	if err := c.ksClient.IamV1alpha2().WorkspaceRoleBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if err := c.k8sClient.RbacV1().ClusterRoleBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if result, err := c.k8sClient.CoreV1().Namespaces().List(metav1.ListOptions{}); err != nil {
		klog.Error(err)
		return err
	} else {
		for _, namespace := range result.Items {
			if err = c.k8sClient.RbacV1().RoleBindings(namespace.Name).DeleteCollection(deleteOptions, listOptions); err != nil {
				klog.Error(err)
				return err
			}
		}
	}

	return nil
}
