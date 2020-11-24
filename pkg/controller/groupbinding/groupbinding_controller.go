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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
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
	"kubesphere.io/kubesphere/pkg/controller/utils/controller"
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
	controller.BaseController
	scheme               *runtime.Scheme
	k8sClient            kubernetes.Interface
	ksClient             kubesphere.Interface
	groupBindingInformer iamv1alpha2informers.GroupBindingInformer
	groupBindingLister   iamv1alpha2listers.GroupBindingLister
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
		BaseController: controller.BaseController{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "GroupBinding"),
			Synced:    []cache.InformerSynced{groupBindingInformer.Informer().HasSynced},
			Name:      controllerName,
		},
		k8sClient:            k8sClient,
		ksClient:             ksClient,
		groupBindingInformer: groupBindingInformer,
		groupBindingLister:   groupBindingInformer.Lister(),
		recorder:             recorder,
	}
	ctl.Handler = ctl.reconcile
	klog.Info("Setting up event handlers")
	groupBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.Enqueue,
		UpdateFunc: func(old, new interface{}) {
			ctl.Enqueue(new)
		},
		DeleteFunc: ctl.Enqueue,
	})
	return ctl
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
