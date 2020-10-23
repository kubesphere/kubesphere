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

package groupbinding

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced = "GroupBinding synced successfully"
	controllerName        = "groupbinding-controller"
	finalizer             = "finalizers.kubesphere.io/groupsbindings"
)

type Controller struct {
	scheme               *runtime.Scheme
	k8sClient            kubernetes.Interface
	ksClient             kubesphere.Interface
	groupBindingInformer iamv1alpha2informers.GroupBindingInformer
	groupBindingLister   iamv1alpha2listers.GroupBindingLister
	groupBindingSynced   cache.InformerSynced
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

// NewController creates GroupBinding Controller instance
func NewController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface, groupBindingInformer iamv1alpha2informers.GroupBindingInformer) *Controller {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	ctl := &Controller{
		k8sClient:            k8sClient,
		ksClient:             ksClient,
		groupBindingInformer: groupBindingInformer,
		groupBindingLister:   groupBindingInformer.Lister(),
		groupBindingSynced:   groupBindingInformer.Informer().HasSynced,
		workqueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "GroupBinding"),
		recorder:             recorder,
	}
	klog.Info("Setting up event handlers")
	groupBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueGroupBinding,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueGroupBinding(new)
		},
		DeleteFunc: ctl.enqueueGroupBinding,
	})
	return ctl
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting GroupBinding controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")

	synced := []cache.InformerSynced{c.groupBindingSynced}

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

func (c *Controller) enqueueGroupBinding(obj interface{}) {
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

	groupBinding, err := c.groupBindingLister.Get(key)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("groupbinding '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}
	if groupBinding.ObjectMeta.DeletionTimestamp.IsZero() {

		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(groupBinding.Finalizers, finalizer) {
			groupBinding.ObjectMeta.Finalizers = append(groupBinding.ObjectMeta.Finalizers, finalizer)
			if groupBinding, err = c.ksClient.IamV1alpha2().GroupBindings().Update(groupBinding); err != nil {
				return err
			}
			// Skip reconcile when groupbinding is updated.
			return nil
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(groupBinding.ObjectMeta.Finalizers, finalizer) {
			if err = c.bindUser(groupBinding); err != nil {
				klog.Error(err)
				return err
			}

			// remove our finalizer from the list and update it.
			groupBinding.Finalizers = sliceutil.RemoveString(groupBinding.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if groupBinding, err = c.ksClient.IamV1alpha2().GroupBindings().Update(groupBinding); err != nil {
				return err
			}
		}
		// Our finalizer has finished, so the reconciler can do nothing.
		return nil
	}

	if err = c.bindUser(groupBinding); err != nil {
		klog.Error(err)
		return err
	}

	c.recorder.Event(groupBinding, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

// Udpate user's Group property. So no need to query user's groups when authorizing.
func (c *Controller) bindUser(groupBinding *iamv1alpha2.GroupBinding) error {

	users := make([]string, 0)
	// Ignore the user if the user if being deleted.
	for _, u := range groupBinding.Users {
		if user, err := c.ksClient.IamV1alpha2().Users().Get(u, metav1.GetOptions{}); err == nil && user.ObjectMeta.DeletionTimestamp.IsZero() {
			users = append(users, u)
		}
	}

	// Nothing to do
	if len(users) == 0 {
		return nil
	}

	// Get all GroupBindings and check whether user exists in the Group.
	listOptions := metav1.ListOptions{}
	groupBindingList, err := c.ksClient.IamV1alpha2().GroupBindings().List(listOptions)
	if err != nil {
		klog.Error(err)
		return err
	}

	userGroups := make(map[string][]string)
	for _, item := range groupBindingList.Items {
		if item.ObjectMeta.DeletionTimestamp.IsZero() {
			for _, u := range users {
				if sliceutil.HasString(item.Users, u) {
					if userGroups[u] == nil {
						userGroups[u] = make([]string, 0)
					}
					userGroups[u] = append(userGroups[u], item.GroupRef.Name)
				}
			}
		}
	}
	for k, v := range userGroups {
		if err := c.patchUser(k, v); err != nil {
			if errors.IsNotFound(err) {
				klog.Infof("user %s doesn't exist any more", k)
				return nil
			}
			klog.Error(err)
			return err
		}
	}
	return nil
}

func (c *Controller) patchUser(userName string, groups []string) error {
	if user, err := c.ksClient.IamV1alpha2().Users().Get(userName, metav1.GetOptions{}); err == nil && user.ObjectMeta.DeletionTimestamp.IsZero() {
		newUser := user.DeepCopy()
		newUser.Spec.Groups = groups
		patch := client.MergeFrom(user)
		patchData, _ := patch.Data(newUser)
		if _, err := c.ksClient.IamV1alpha2().Users().
			Patch(userName, patch.Type(), patchData); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}
