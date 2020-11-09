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
	successSynced         = "Synced"
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
	workqueue            workqueue.RateLimitingInterface
	recorder             record.EventRecorder
}

// NewController creates GroupBinding Controller instance
func NewController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface, groupBindingInformer iamv1alpha2informers.GroupBindingInformer) *Controller {
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

	klog.Info("Starting GroupBinding controller")
	klog.Info("Waiting for informer caches to sync")

	synced := []cache.InformerSynced{c.groupBindingSynced}

	if ok := cache.WaitForCacheSync(stopCh, synced...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
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

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.reconcile(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
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

// reconcile handles GroupBinding informer events, it updates user's Groups property with the current GroupBinding.
func (c *Controller) reconcile(key string) error {

	groupBinding, err := c.groupBindingLister.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("groupbinding '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}
	if groupBinding.ObjectMeta.DeletionTimestamp.IsZero() {
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
			if err = c.unbindUser(groupBinding); err != nil {
				klog.Error(err)
				return err
			}

			groupBinding.Finalizers = sliceutil.RemoveString(groupBinding.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if groupBinding, err = c.ksClient.IamV1alpha2().GroupBindings().Update(groupBinding); err != nil {
				return err
			}
		}
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
	return c.Run(2, stopCh)
}

func (c *Controller) unbindUser(groupBinding *iamv1alpha2.GroupBinding) error {
	return c.updateUserGroups(groupBinding, func(groups []string, group string) (bool, []string) {
		// remove a group from the groups
		if sliceutil.HasString(groups, group) {
			groups := sliceutil.RemoveString(groups, func(item string) bool {
				return item == group
			})
			return true, groups
		}
		return false, groups
	})
}

func (c *Controller) bindUser(groupBinding *iamv1alpha2.GroupBinding) error {
	return c.updateUserGroups(groupBinding, func(groups []string, group string) (bool, []string) {
		// add group to the groups
		if !sliceutil.HasString(groups, group) {
			groups := append(groups, group)
			return true, groups
		}
		return false, groups
	})
}

// Udpate user's Group property. So no need to query user's groups when authorizing.
func (c *Controller) updateUserGroups(groupBinding *iamv1alpha2.GroupBinding, operator func(groups []string, group string) (bool, []string)) error {

	for _, u := range groupBinding.Users {
		// Ignore the user if the user if being deleted.
		if user, err := c.ksClient.IamV1alpha2().Users().Get(u, metav1.GetOptions{}); err == nil && user.ObjectMeta.DeletionTimestamp.IsZero() {

			if errors.IsNotFound(err) {
				klog.Infof("user %s doesn't exist any more", u)
				continue
			}

			if changed, groups := operator(user.Spec.Groups, groupBinding.GroupRef.Name); changed {

				if err := c.patchUser(user, groups); err != nil {
					if errors.IsNotFound(err) {
						klog.Infof("user %s doesn't exist any more", u)
						continue
					}
					klog.Error(err)
					return err
				}
			}
		}
	}
	return nil
}

func (c *Controller) patchUser(user *iamv1alpha2.User, groups []string) error {
	newUser := user.DeepCopy()
	newUser.Spec.Groups = groups
	patch := client.MergeFrom(user)
	patchData, _ := patch.Data(newUser)
	if _, err := c.ksClient.IamV1alpha2().Users().
		Patch(user.Name, patch.Type(), patchData); err != nil {
		return err
	}
	return nil
}
