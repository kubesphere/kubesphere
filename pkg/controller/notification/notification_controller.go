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
	"encoding/json"
	"fmt"
	"reflect"
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
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/api/cluster/v1alpha1"
	"kubesphere.io/api/notification/v2beta2"
	"kubesphere.io/api/types/v1beta1"
	"kubesphere.io/api/types/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced         = "Synced"
	controllerName        = "notification-controller"
	messageResourceSynced = "Notification synced successfully"
)

type Controller struct {
	client.Client
	ksCache        cache.Cache
	reconciledObjs []client.Object
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
	c.reconciledObjs = append(c.reconciledObjs, &v2beta2.Config{})
	c.reconciledObjs = append(c.reconciledObjs, &v2beta2.Receiver{})
	c.reconciledObjs = append(c.reconciledObjs, &v2beta2.Router{})
	c.reconciledObjs = append(c.reconciledObjs, &v2beta2.Silence{})
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

	// Watch the cluster add and delete operations.
	if informer, err := c.ksCache.GetInformer(context.Background(), &v1alpha1.Cluster{}); err != nil {
		klog.Errorf("get cluster informer error, %v", err)
		return err
	} else {
		informer.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
			AddFunc:    c.enqueue,
			DeleteFunc: c.enqueue,
		})
		c.informerSynced = append(c.informerSynced, informer.HasSynced)
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

	runtimeObj, ok := obj.(client.Object)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("object does not implement the Object interfaces"))
		return nil
	}

	// Only reconcile the secret which created by notification manager.
	if secret, ok := obj.(*corev1.Secret); ok {
		if secret.Namespace != constants.NotificationSecretNamespace ||
			secret.Labels == nil || secret.Labels[constants.NotificationManagedLabel] != "true" {
			klog.V(8).Infof("No need to reconcile secret %s/%s", runtimeObj.GetNamespace(), runtimeObj.GetName())
			return nil
		}
	}

	name := runtimeObj.GetName()

	// The notification controller should update the annotations of secrets managed by itself
	// whenever a cluster is added or deleted. This way, the controller will have a chance to override the secret.
	if _, ok := obj.(*v1alpha1.Cluster); ok {
		err := c.updateSecret()
		if err != nil {
			klog.Errorf("update secret failed, %s", err)
			return err
		}

		return nil
	}

	err := c.Get(context.Background(), client.ObjectKey{Name: runtimeObj.GetName(), Namespace: runtimeObj.GetNamespace()}, runtimeObj)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("obj '%s' in work queue no longer exists", name))
			c.recorder.Event(runtimeObj, corev1.EventTypeNormal, successSynced, messageResourceSynced)
			klog.Infof("Successfully synced %s", name)
			return nil
		}
		klog.Error(err)
		return err
	}

	if err = c.multiClusterSync(context.Background(), runtimeObj); err != nil {
		return err
	}

	c.recorder.Event(runtimeObj, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	klog.Infof("Successfully synced %s", name)
	return nil
}

func (c *Controller) Start(ctx context.Context) error {
	return c.Run(4, ctx.Done())
}

func (c *Controller) multiClusterSync(ctx context.Context, obj client.Object) error {

	if err := c.ensureNotControlledByKubefed(ctx, obj); err != nil {
		klog.Error(err)
		return err
	}

	switch obj := obj.(type) {
	case *v2beta2.Config:
		return c.syncFederatedConfig(obj)
	case *v2beta2.Receiver:
		return c.syncFederatedReceiver(obj)
	case *v2beta2.Router:
		return c.syncFederatedRouter(obj)
	case *v2beta2.Silence:
		return c.syncFederatedSilence(obj)
	case *corev1.Secret:
		return c.syncFederatedSecret(obj)
	default:
		klog.Errorf("unknown type for notification, %v", obj)
		return nil
	}
}

func (c *Controller) syncFederatedConfig(obj *v2beta2.Config) error {

	fedObj := &v1beta2.FederatedNotificationConfig{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta2.FederatedNotificationConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta2.FederatedNotificationConfigKind,
					APIVersion: v1beta2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta2.FederatedNotificationConfigSpec{
					Template: v1beta2.NotificationConfigTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta2.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err = controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				klog.Errorf("FederatedNotificationConfig '%s' SetControllerReference failed, %s", obj.Name, err)
				return err
			}

			if err = c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create FederatedNotificationConfig '%s' failed, %s", obj.Name, err)
				return err
			}

			return nil
		}
		klog.Errorf("get FederatedNotificationConfig '%s' failed, %s", obj.Name, err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update FederatedNotificationConfig '%s' failed, %s", obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedReceiver(obj *v2beta2.Receiver) error {

	fedObj := &v1beta2.FederatedNotificationReceiver{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta2.FederatedNotificationReceiver{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta2.FederatedNotificationReceiverKind,
					APIVersion: v1beta2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta2.FederatedNotificationReceiverSpec{
					Template: v1beta2.NotificationReceiverTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta2.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err = controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				klog.Errorf("FederatedNotificationReceiver '%s' SetControllerReference failed, %s", obj.Name, err)
				return err
			}

			if err = c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create FederatedNotificationReceiver '%s' failed, %s", obj.Name, err)
				return err
			}

			return nil
		}
		klog.Errorf("get FederatedNotificationReceiver '%s' failed, %s", obj.Name, err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update FederatedNotificationReceiver '%s' failed, %s", obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedRouter(obj *v2beta2.Router) error {

	fedObj := &v1beta2.FederatedNotificationRouter{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta2.FederatedNotificationRouter{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta2.FederatedNotificationReceiverKind,
					APIVersion: v1beta2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta2.FederatedNotificationRouterSpec{
					Template: v1beta2.NotificationRouterTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta2.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err = controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				klog.Errorf("FederatedNotificationRouter '%s' SetControllerReference failed, %s", obj.Name, err)
				return err
			}

			if err = c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create FederatedNotificationRouter '%s' failed, %s", obj.Name, err)
				return err
			}

			return nil
		}
		klog.Errorf("get FederatedNotificationRouter '%s' failed, %s", obj.Name, err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update FederatedNotificationRouter '%s' failed, %s", obj.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) syncFederatedSilence(obj *v2beta2.Silence) error {

	fedObj := &v1beta2.FederatedNotificationSilence{}
	err := c.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
	if err != nil {
		if errors.IsNotFound(err) {
			fedObj = &v1beta2.FederatedNotificationSilence{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta2.FederatedNotificationReceiverKind,
					APIVersion: v1beta2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
				},
				Spec: v1beta2.FederatedNotificationSilenceSpec{
					Template: v1beta2.NotificationSilenceTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Labels: obj.Labels,
						},
						Spec: obj.Spec,
					},
					Placement: v1beta2.GenericPlacementFields{
						ClusterSelector: &metav1.LabelSelector{},
					},
				},
			}

			err = controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				klog.Errorf("FederatedNotificationSilence '%s' SetControllerReference failed, %s", obj.Name, err)
				return err
			}

			if err = c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create FederatedNotificationSilence '%s' failed, %s", obj.Name, err)
				return err
			}

			return nil
		}
		klog.Errorf("get FederatedNotificationSilence '%s' failed, %s", obj.Name, err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Labels, obj.Labels) || !reflect.DeepEqual(fedObj.Spec.Template.Spec, obj.Spec) {

		fedObj.Spec.Template.Spec = obj.Spec
		fedObj.Spec.Template.Labels = obj.Labels

		if err := c.Update(context.Background(), fedObj); err != nil {
			klog.Errorf("update FederatedNotificationSilence '%s' failed, %s", obj.Name, err)
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

			err = c.updateOverrides(obj, fedObj)
			if err != nil {
				klog.Errorf("update FederatedSecret '%s' overrides failed, %s", obj.Name, err)
				return err
			}

			err = controllerutil.SetControllerReference(obj, fedObj, scheme.Scheme)
			if err != nil {
				klog.Errorf("FederatedSecret '%s' SetControllerReference failed, %s", obj.Name, err)
				return err
			}

			if err = c.Create(context.Background(), fedObj); err != nil {
				klog.Errorf("create FederatedSecret '%s' failed, %s", obj.Name, err)
				return err
			}

			return nil
		}
		klog.Errorf("get FederatedSecret '%s' failed, %s", obj.Name, err)
		return err
	}

	if !reflect.DeepEqual(fedObj.Spec.Template.Data, obj.Data) ||
		!reflect.DeepEqual(fedObj.Spec.Template.StringData, obj.StringData) ||
		!reflect.DeepEqual(fedObj.Spec.Template.Type, obj.Type) {

		fedObj.Spec.Template.Data = obj.Data
		fedObj.Spec.Template.StringData = obj.StringData
		fedObj.Spec.Template.Type = obj.Type
	}

	err = c.updateOverrides(obj, fedObj)
	if err != nil {
		klog.Errorf("update FederatedSecret '%s' overrides failed, %s", obj.Name, err)
		return err
	}

	if err := c.Update(context.Background(), fedObj); err != nil {
		klog.Errorf("update FederatedSecret '%s' failed, %s", obj.Name, err)
		return err
	}

	return nil
}

func (c *Controller) updateOverrides(obj *corev1.Secret, fedSecret *v1beta1.FederatedSecret) error {
	clusterList := &v1alpha1.ClusterList{}
	err := c.ksCache.List(context.Background(), clusterList)
	if err != nil {
		return err
	}

	bs, err := json.Marshal(obj.Labels)
	if err != nil {
		return err
	}

	fedSecret.Spec.Overrides = fedSecret.Spec.Overrides[:0]
	for _, cluster := range clusterList.Items {
		fedSecret.Spec.Overrides = append(fedSecret.Spec.Overrides, v1beta1.GenericOverrideItem{
			ClusterName: cluster.Name,
			ClusterOverrides: []v1beta1.ClusterOverride{
				{
					Path: "/metadata/labels",
					Value: runtime.RawExtension{
						Raw: bs,
					},
				},
			},
		})
	}

	return nil
}

// Update the annotations of secrets managed by the notification controller to trigger a reconcile.
func (c *Controller) updateSecret() error {

	secretList := &corev1.SecretList{}
	err := c.ksCache.List(context.Background(), secretList,
		client.InNamespace(constants.NotificationSecretNamespace),
		client.MatchingLabels{
			constants.NotificationManagedLabel: "true",
		})
	if err != nil {
		return err
	}

	for _, secret := range secretList.Items {
		if secret.Annotations == nil {
			secret.Annotations = make(map[string]string)
		}

		secret.Annotations["reloadtimestamp"] = time.Now().String()
		if err := c.Update(context.Background(), &secret); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) ensureNotControlledByKubefed(ctx context.Context, obj client.Object) error {

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string, 0)
	}

	if labels[constants.KubefedManagedLabel] != "false" {
		labels[constants.KubefedManagedLabel] = "false"
		obj.SetLabels(labels)
		err := c.Update(ctx, obj)
		if err != nil {
			klog.Error(err)
		}
	}
	return nil
}
