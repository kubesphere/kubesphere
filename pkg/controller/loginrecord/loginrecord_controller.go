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

package loginrecord

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	"sort"
	"time"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced = "LoginRecord synced successfully"
	controllerName        = "loginrecord-controller"
)

type loginRecordController struct {
	controller.BaseController
	k8sClient                   kubernetes.Interface
	ksClient                    kubesphere.Interface
	loginRecordLister           iamv1alpha2listers.LoginRecordLister
	loginRecordSynced           cache.InformerSynced
	userLister                  iamv1alpha2listers.UserLister
	userSynced                  cache.InformerSynced
	loginHistoryRetentionPeriod time.Duration
	loginHistoryMaximumEntries  int
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewLoginRecordController(k8sClient kubernetes.Interface,
	ksClient kubesphere.Interface,
	loginRecordInformer iamv1alpha2informers.LoginRecordInformer,
	userInformer iamv1alpha2informers.UserInformer,
	loginHistoryRetentionPeriod time.Duration,
	loginHistoryMaximumEntries int) *loginRecordController {

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	ctl := &loginRecordController{
		BaseController: controller.BaseController{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "LoginRecords"),
			Synced:    []cache.InformerSynced{loginRecordInformer.Informer().HasSynced, userInformer.Informer().HasSynced},
			Name:      controllerName,
		},
		k8sClient:                   k8sClient,
		ksClient:                    ksClient,
		loginRecordLister:           loginRecordInformer.Lister(),
		userLister:                  userInformer.Lister(),
		loginHistoryRetentionPeriod: loginHistoryRetentionPeriod,
		loginHistoryMaximumEntries:  loginHistoryMaximumEntries,
		recorder:                    recorder,
	}
	ctl.Handler = ctl.reconcile
	klog.Info("Setting up event handlers")
	loginRecordInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.Enqueue,
		UpdateFunc: func(old, new interface{}) {
			ctl.Enqueue(new)
		},
		DeleteFunc: ctl.Enqueue,
	})
	return ctl
}

func (c *loginRecordController) Start(stopCh <-chan struct{}) error {
	return c.Run(5, stopCh)
}

func (c *loginRecordController) reconcile(key string) error {
	loginRecord, err := c.loginRecordLister.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("login record '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}

	if !loginRecord.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		// Our finalizer has finished, so the reconciler can do nothing.
		return nil
	}

	user, err := c.userForLoginRecord(loginRecord)
	if err != nil {
		// delete orphan object
		if errors.IsNotFound(err) {
			return c.ksClient.IamV1alpha2().LoginRecords().Delete(context.TODO(), loginRecord.Name, metav1.DeleteOptions{})
		}
		return err
	}

	if err = c.updateUserLastLoginTime(user, loginRecord); err != nil {
		return err
	}

	if err = c.shrinkEntriesFor(user); err != nil {
		return err
	}

	now := time.Now()
	// login record beyonds retention period
	if loginRecord.CreationTimestamp.Add(c.loginHistoryRetentionPeriod).Before(now) {
		if err = c.ksClient.IamV1alpha2().LoginRecords().Delete(context.Background(), loginRecord.Name, *metav1.NewDeleteOptions(0)); err != nil {
			klog.Error(err)
			return err
		}
	} else { // put item back into the queue
		c.Workqueue.AddAfter(key, loginRecord.CreationTimestamp.Add(c.loginHistoryRetentionPeriod).Sub(now))
	}
	c.recorder.Event(loginRecord, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

// updateUserLastLoginTime accepts a login object and set user lastLoginTime field
func (c *loginRecordController) updateUserLastLoginTime(user *iamv1alpha2.User, loginRecord *iamv1alpha2.LoginRecord) error {
	// update lastLoginTime
	if user.DeletionTimestamp.IsZero() &&
		(user.Status.LastLoginTime == nil || user.Status.LastLoginTime.Before(&loginRecord.CreationTimestamp)) {
		user.Status.LastLoginTime = &loginRecord.CreationTimestamp
		_, err := c.ksClient.IamV1alpha2().Users().UpdateStatus(context.Background(), user, metav1.UpdateOptions{})
		return err
	}
	return nil
}

// shrinkEntriesFor will delete old entries out of limit
func (c *loginRecordController) shrinkEntriesFor(user *iamv1alpha2.User) error {
	loginRecords, err := c.loginRecordLister.List(labels.SelectorFromSet(labels.Set{iamv1alpha2.UserReferenceLabel: user.Name}))
	if err != nil {
		return err
	}
	if len(loginRecords) <= c.loginHistoryMaximumEntries {
		return nil
	}
	sort.Slice(loginRecords, func(i, j int) bool {
		return loginRecords[j].CreationTimestamp.After(loginRecords[i].CreationTimestamp.Time)
	})
	oldEntries := loginRecords[:len(loginRecords)-c.loginHistoryMaximumEntries]
	for _, r := range oldEntries {
		err = c.ksClient.IamV1alpha2().LoginRecords().Delete(context.TODO(), r.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *loginRecordController) userForLoginRecord(loginRecord *iamv1alpha2.LoginRecord) (*iamv1alpha2.User, error) {
	username, ok := loginRecord.Labels[iamv1alpha2.UserReferenceLabel]
	if !ok || len(username) == 0 {
		klog.V(4).Info("login doesn't belong to any user")
		return nil, errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularUser), username)
	}
	return c.userLister.Get(username)
}
