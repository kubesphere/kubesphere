/*
Copyright 2019 The KubeSphere Authors.

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

package notification

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/notification/v2alpha1"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced  = "Synced"
	controllerName = "notification-controller"
)

type Controller struct {
	client.Client
	ksCache        cache.Cache
	reconciledObjs []runtime.Object
	informerSynced []toolscache.InformerSynced
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

func NewController(k8sClient kubernetes.Interface, ksClient client.Client, ksCache cache.Cache) (*Controller, error) {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	ctl := &Controller{
		Client:    ksClient,
		ksCache:   ksCache,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Notification"),
		recorder:  recorder,
	}
	klog.Info("Setting up event handlers")

	if err := ctl.setEventHandlers(); err != nil {
		return nil, err
	}

	return ctl, nil
}

func (c *Controller) setEventHandlers() error {

	if c.reconciledObjs != nil && len(c.reconciledObjs) > 0 {
		c.reconciledObjs = c.reconciledObjs[:0]
	}
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.DingTalkConfig{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.DingTalkReceiver{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.EmailConfig{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.EmailReceiver{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.SlackConfig{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.SlackReceiver{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.WebhookConfig{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.WebhookReceiver{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.WechatConfig{})
	c.reconciledObjs = append(c.reconciledObjs, &v2alpha1.WechatReceiver{})
	c.reconciledObjs = append(c.reconciledObjs, &corev1.Secret{})

	if c.informerSynced != nil && len(c.informerSynced) > 0 {
		c.informerSynced = c.informerSynced[:0]
	}

	for _, obj := range c.reconciledObjs {
		if informer, err := c.ksCache.GetInformer(context.Background(), obj); err != nil {
			klog.Errorf("get %s informer error, %v", obj.GetObjectKind().GroupVersionKind().String(), err)
			return err
		} else {
			informer.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
				AddFunc: c.enqueue,
				UpdateFunc: func(old, new interface{}) {
					c.enqueue(new)
				},
				DeleteFunc: c.enqueue,
			})
			c.informerSynced = append(c.informerSynced, informer.HasSynced)
		}
	}

	return nil
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Notification controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")

	if ok := toolscache.WaitForCacheSync(stopCh, c.informerSynced...); !ok {
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

func (c *Controller) enqueue(obj interface{}) {
	c.workqueue.Add(obj)
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

		// Run the reconcile, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.reconcile(obj); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(obj)
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
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
func (c *Controller) reconcile(obj interface{}) error {

	runtimeObj, ok := obj.(runtime.Object)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("object does not implement the Object interfaces"))
		return nil
	}
	runtimeObj = runtimeObj.DeepCopyObject()

	accessor, err := meta.Accessor(runtimeObj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("object does not implement the Object interfaces"))
		return nil
	}

	// Only reconcile the secret which created by notification manager.
	if secret, ok := obj.(*corev1.Secret); ok {
		if secret.Namespace != constants.NotificationSecretNamespace {
			klog.V(8).Infof("No need to reconcile secret %s/%s", accessor.GetNamespace(), accessor.GetName())
			return nil
		}

		if err := c.ensureNotificationNamespaceExist(); err != nil {
			return err
		}
	}

	name := accessor.GetName()
	kind := runtimeObj.GetObjectKind().GroupVersionKind().String()
	err = c.Get(context.Background(), client.ObjectKey{Name: accessor.GetName(), Namespace: accessor.GetNamespace()}, runtimeObj)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("obj '%s, %s' in work queue no longer exists", kind, name))
			c.recorder.Event(runtimeObj, corev1.EventTypeNormal, successSynced, fmt.Sprintf("%s synced successfully", kind))
			klog.Infof("Successfully synced %s:%s", kind, name)
			return nil
		}
		klog.Error(err)
		return err
	}

	if err = c.multiClusterSync(context.Background(), runtimeObj); err != nil {
		return err
	}

	c.recorder.Event(runtimeObj, corev1.EventTypeNormal, successSynced, fmt.Sprintf("%s synced successfully", kind))
	klog.Infof("Successfully synced %s:%s", kind, name)
	return nil
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

func (c *Controller) multiClusterSync(ctx context.Context, obj runtime.Object) error {

	if err := c.ensureNotControlledByKubefed(ctx, obj); err != nil {
		klog.Error(err)
		return err
	}

	switch obj.(type) {
	case *v2alpha1.DingTalkConfig:
		return c.syncFederatedDingTalkConfig(obj.(*v2alpha1.DingTalkConfig))
	case *v2alpha1.DingTalkReceiver:
		return c.syncFederatedDingTalkReceiver(obj.(*v2alpha1.DingTalkReceiver))
	case *v2alpha1.EmailConfig:
		return c.syncFederatedEmailConfig(obj.(*v2alpha1.EmailConfig))
	case *v2alpha1.EmailReceiver:
		return c.syncFederatedEmailReceiver(obj.(*v2alpha1.EmailReceiver))
	case *v2alpha1.SlackConfig:
		return c.syncFederatedSlackConfig(obj.(*v2alpha1.SlackConfig))
	case *v2alpha1.SlackReceiver:
		return c.syncFederatedSlackReceiver(obj.(*v2alpha1.SlackReceiver))
	case *v2alpha1.WebhookConfig:
		return c.syncFederatedWebhookConfig(obj.(*v2alpha1.WebhookConfig))
	case *v2alpha1.WebhookReceiver:
		return c.syncFederatedWebhookReceiver(obj.(*v2alpha1.WebhookReceiver))
	case *v2alpha1.WechatConfig:
		return c.syncFederatedWechatConfig(obj.(*v2alpha1.WechatConfig))
	case *v2alpha1.WechatReceiver:
		return c.syncFederatedWechatReceiver(obj.(*v2alpha1.WechatReceiver))
	case *corev1.Secret:
		return c.syncFederatedSecret(obj.(*corev1.Secret))
	default:
		klog.Errorf("unknown type for notification, %v", obj)
		return nil
	}
}

func (c *Controller) syncFederatedDingTalkConfig(obj *v2alpha1.DingTalkConfig) error {

	fedObj := &v1beta1.FederatedDingTalkConfig{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedDingTalkConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedDingTalkConfigKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedDingTalkConfigSpec{
					Template: v1beta1.DingTalkConfigTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedDingTalkReceiver(obj *v2alpha1.DingTalkReceiver) error {

	fedObj := &v1beta1.FederatedDingTalkReceiver{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedDingTalkReceiver{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedDingTalkReceiverKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedDingTalkReceiverSpec{
					Template: v1beta1.DingTalkReceiverTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedEmailConfig(obj *v2alpha1.EmailConfig) error {

	fedObj := &v1beta1.FederatedEmailConfig{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedEmailConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedEmailConfigKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedEmailConfigSpec{
					Template: v1beta1.EmailConfigTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedEmailReceiver(obj *v2alpha1.EmailReceiver) error {

	fedObj := &v1beta1.FederatedEmailReceiver{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedEmailReceiver{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedEmailReceiverKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedEmailReceiverSpec{
					Template: v1beta1.EmailReceiverTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedSlackConfig(obj *v2alpha1.SlackConfig) error {

	fedObj := &v1beta1.FederatedSlackConfig{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedSlackConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedSlackConfigKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedSlackConfigSpec{
					Template: v1beta1.SlackConfigTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedSlackReceiver(obj *v2alpha1.SlackReceiver) error {

	fedObj := &v1beta1.FederatedSlackReceiver{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedSlackReceiver{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedSlackReceiverKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedSlackReceiverSpec{
					Template: v1beta1.SlackReceiverTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedWebhookConfig(obj *v2alpha1.WebhookConfig) error {

	fedObj := &v1beta1.FederatedWebhookConfig{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedWebhookConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedWebhookConfigKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedWebhookConfigSpec{
					Template: v1beta1.WebhookConfigTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedWebhookReceiver(obj *v2alpha1.WebhookReceiver) error {

	fedObj := &v1beta1.FederatedWebhookReceiver{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedWebhookReceiver{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedWebhookReceiverKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedWebhookReceiverSpec{
					Template: v1beta1.WebhookReceiverTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedWechatConfig(obj *v2alpha1.WechatConfig) error {

	fedObj := &v1beta1.FederatedWechatConfig{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedWechatConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedWechatConfigKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedWechatConfigSpec{
					Template: v1beta1.WechatConfigTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedWechatReceiver(obj *v2alpha1.WechatReceiver) error {

	fedObj := &v1beta1.FederatedWechatReceiver{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedWechatReceiver{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedWechatReceiverKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta1.FederatedWechatReceiverSpec{
					Template: v1beta1.WechatReceiverTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedSecret(obj *corev1.Secret) error {

	fedObj := &v1beta1.FederatedSecret{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name, Namespace: obj.Namespace}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta1.FederatedSecret{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedSecretKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      obj.Name,
					Namespace: obj.Namespace,
				},
				Spec: v1beta1.FederatedSecretSpec{
					Template: v1beta1.SecretTemplate{
						Data:       obj.Data,
						StringData: obj.StringData,
						Type:       obj.Type,
					},
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err := controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				return err
			}
			if err := c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
				return err
			}

			return nil
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Data, obj.Data) ||
		!reflect.DeepEqual(fedObj.Spec.Template.StringData, obj.StringData) ||
		!reflect.DeepEqual(fedObj.Spec.Template.Type, obj.Type) {

		fedObj.Spec.Template.Data = obj.Data
		fedObj.Spec.Template.StringData = obj.StringData
		fedObj.Spec.Template.Type = obj.Type

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update '%s:%s' failed, %s", fedObj.GetObjectKind().GroupVersionKind().String(), obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) ensureNotificationNamespaceExist() error {

	ns := corev1.Namespace{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: constants.NotificationSecretNamespace}, &ns); err != nil {
		return err
	}

	fedNs := v1beta1.FederatedNamespace{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: constants.NotificationSecretNamespace, Namespace: constants.NotificationSecretNamespace}, &fedNs); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}

		if errors.IsNotFound(err) {
			fedNs = v1beta1.FederatedNamespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.FederatedNamespaceKind,
					APIVersion: v1beta1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.NotificationSecretNamespace,
					Namespace: constants.NotificationSecretNamespace,
				},
				Spec: v1beta1.FederatedNamespaceSpec{
					Placement: v1beta1.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			if err := controllerutil.SetControllerReference(&ns, &fedNs, scheme.Scheme); err != nil {
				return err
			}

			return c.Create(context.Background(), &fedNs)
		}

		return err
	}

	return nil

}

func (c *Controller) ensureNotControlledByKubefed(ctx context.Context, obj runtime.Object) error {

	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.Error(err)
		return nil
	}

	labels := accessor.GetLabels()
	if labels == nil {
		labels = make(map[string]string, 0)
	}

	if labels[constants.KubefedManagedLabel] != "false" {
		labels[constants.KubefedManagedLabel] = "false"
		accessor.SetLabels(labels)
		err := c.Update(ctx, accessor.(runtime.Object))
		if err != nil {
			klog.Error(err)
		}
	}
	return nil
}
