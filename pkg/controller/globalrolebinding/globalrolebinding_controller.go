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

package globalrolebinding

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced           = "GlobalRoleBinding synced successfully"
	controllerName                  = "globalrolebinding-controller"
	federatedClusterRoleBindingKind = "FederatedClusterRoleBinding"
	federatedResourceVersion        = "types.kubefed.io/v1beta1"
	federatedResourceAPIPath        = "/apis/types.kubefed.io/v1beta1/federatedclusterrolebindings"
)

type Controller struct {
	k8sClient kubernetes.Interface
	informer  iamv1alpha2informers.GlobalRoleBindingInformer
	lister    iamv1alpha2listers.GlobalRoleBindingLister
	synced    cache.InformerSynced
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder            record.EventRecorder
	multiClusterEnabled bool
}

func NewController(k8sClient kubernetes.Interface, globalRoleBindingInformer iamv1alpha2informers.GlobalRoleBindingInformer, multiClusterEnabled bool) *Controller {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	ctl := &Controller{
		k8sClient:           k8sClient,
		informer:            globalRoleBindingInformer,
		lister:              globalRoleBindingInformer.Lister(),
		synced:              globalRoleBindingInformer.Informer().HasSynced,
		workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ClusterRoleBinding"),
		recorder:            recorder,
		multiClusterEnabled: multiClusterEnabled,
	}
	klog.Info("Setting up event handlers")
	globalRoleBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueClusterRoleBinding,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueClusterRoleBinding(new)
		},
		DeleteFunc: ctl.enqueueClusterRoleBinding,
	})
	return ctl
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	//init client

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting User controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.synced); !ok {
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

func (c *Controller) enqueueClusterRoleBinding(obj interface{}) {
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

	globalRoleBinding, err := c.lister.Get(key)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("clusterrolebinding '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}

	if globalRoleBinding.RoleRef.Name == iamv1alpha2.PlatformAdmin {
		if err := c.relateToClusterAdmin(globalRoleBinding); err != nil {
			klog.Error(err)
			return err
		}
	}

	c.recorder.Event(globalRoleBinding, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

func (c *Controller) relateToClusterAdmin(globalRoleBinding *iamv1alpha2.GlobalRoleBinding) error {

	if c.multiClusterEnabled {

		federatedClusterRoleBinding := &iamv1alpha2.FederatedClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       federatedClusterRoleBindingKind,
				APIVersion: federatedResourceVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("fed-%s", globalRoleBinding.Name),
			},
			Spec: iamv1alpha2.FederatedClusterRoleBindingSpec{
				Template: iamv1alpha2.Template{
					Subjects: ensureSubjectAPIVersionIsValid(globalRoleBinding.Subjects),
					RoleRef: rbacv1.RoleRef{
						APIGroup: "rbac.authorization.k8s.io",
						Kind:     iamv1alpha2.ResourceKindClusterRole,
						Name:     iamv1alpha2.ClusterAdmin,
					},
				},
				Placement: iamv1alpha2.Placement{
					ClusterSelector: iamv1alpha2.ClusterSelector{},
				},
			},
		}

		err := controllerutil.SetControllerReference(globalRoleBinding, federatedClusterRoleBinding, scheme.Scheme)

		if err != nil {
			return err
		}

		data, err := json.Marshal(federatedClusterRoleBinding)

		if err != nil {
			return err
		}

		cli := c.k8sClient.(*kubernetes.Clientset)

		err = cli.RESTClient().Post().
			AbsPath(federatedResourceAPIPath).
			Body(data).
			Do().Error()

		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		}
	} else {

		clusterRoleBinding := &rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("fed-%s", globalRoleBinding.Name),
			},
			Subjects: ensureSubjectAPIVersionIsValid(globalRoleBinding.Subjects),
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     iamv1alpha2.ResourceKindClusterRole,
				Name:     iamv1alpha2.ClusterAdmin,
			},
		}

		err := controllerutil.SetControllerReference(globalRoleBinding, clusterRoleBinding, scheme.Scheme)

		if err != nil {
			return err
		}

		_, err = c.k8sClient.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)

		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		}
	}

	return nil
}

func ensureSubjectAPIVersionIsValid(subjects []rbacv1.Subject) []rbacv1.Subject {
	validSubjects := make([]rbacv1.Subject, 0)
	for _, subject := range subjects {
		if subject.Kind == iamv1alpha2.ResourceKindUser {
			validSubject := rbacv1.Subject{
				Kind:     iamv1alpha2.ResourceKindUser,
				APIGroup: "rbac.authorization.k8s.io",
				Name:     subject.Name,
			}
			validSubjects = append(validSubjects, validSubject)
		}
	}
	return validSubjects
}
