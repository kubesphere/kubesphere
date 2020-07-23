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

package workspacerole

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	tenantv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/tenant/v1alpha2"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	tenantv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/tenant/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced = "WorkspaceRole synced successfully"
	controllerName        = "workspacerole-controller"
)

type Controller struct {
	scheme                          *runtime.Scheme
	k8sClient                       kubernetes.Interface
	ksClient                        kubesphere.Interface
	workspaceRoleInformer           iamv1alpha2informers.WorkspaceRoleInformer
	workspaceRoleLister             iamv1alpha2listers.WorkspaceRoleLister
	workspaceRoleSynced             cache.InformerSynced
	workspaceTemplateInformer       tenantv1alpha2informers.WorkspaceTemplateInformer
	workspaceTemplateLister         tenantv1alpha2listers.WorkspaceTemplateLister
	workspaceTemplateSynced         cache.InformerSynced
	fedWorkspaceRoleCache           cache.Store
	fedWorkspaceRoleCacheController cache.Controller
	multiClusterEnabled             bool
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

func NewController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface, workspaceRoleInformer iamv1alpha2informers.WorkspaceRoleInformer,
	fedWorkspaceRoleCache cache.Store, fedWorkspaceRoleCacheController cache.Controller, workspaceTemplateInformer tenantv1alpha2informers.WorkspaceTemplateInformer, multiClusterEnabled bool) *Controller {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	ctl := &Controller{
		k8sClient:                       k8sClient,
		ksClient:                        ksClient,
		workspaceRoleInformer:           workspaceRoleInformer,
		workspaceRoleLister:             workspaceRoleInformer.Lister(),
		workspaceRoleSynced:             workspaceRoleInformer.Informer().HasSynced,
		fedWorkspaceRoleCache:           fedWorkspaceRoleCache,
		fedWorkspaceRoleCacheController: fedWorkspaceRoleCacheController,
		workspaceTemplateInformer:       workspaceTemplateInformer,
		workspaceTemplateLister:         workspaceTemplateInformer.Lister(),
		workspaceTemplateSynced:         workspaceTemplateInformer.Informer().HasSynced,
		multiClusterEnabled:             multiClusterEnabled,
		workqueue:                       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "WorkspaceRole"),
		recorder:                        recorder,
	}
	klog.Info("Setting up event handlers")
	workspaceRoleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueClusterRole,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueClusterRole(new)
		},
		DeleteFunc: ctl.enqueueClusterRole,
	})
	return ctl
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting WorkspaceRole controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")

	synced := make([]cache.InformerSynced, 0)
	synced = append(synced, c.workspaceRoleSynced, c.workspaceTemplateSynced)
	if c.multiClusterEnabled {
		synced = append(synced, c.fedWorkspaceRoleCacheController.HasSynced)
	}

	if ok := cache.WaitForCacheSync(stopCh); !ok {
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

func (c *Controller) enqueueClusterRole(obj interface{}) {
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

	workspaceRole, err := c.workspaceRoleLister.Get(key)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("workspacerole '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}

	if err = c.bindWorkspace(workspaceRole); err != nil {
		klog.Error(err)
		return err
	}

	if c.multiClusterEnabled {
		if err = c.multiClusterSync(workspaceRole); err != nil {
			klog.Error(err)
			return err
		}
	}

	c.recorder.Event(workspaceRole, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

func (c *Controller) bindWorkspace(workspaceRole *iamv1alpha2.WorkspaceRole) error {

	workspaceName := workspaceRole.Labels[constants.WorkspaceLabelKey]

	if workspaceName == "" {
		return nil
	}

	workspace, err := c.workspaceTemplateLister.Get(workspaceName)

	if err != nil {
		// skip if workspace not found
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	if !metav1.IsControlledBy(workspaceRole, workspace) {
		workspaceRole.OwnerReferences = nil
		if err := controllerutil.SetControllerReference(workspace, workspaceRole, scheme.Scheme); err != nil {
			klog.Error(err)
			return err
		}
		_, err = c.ksClient.IamV1alpha2().WorkspaceRoles().Update(workspaceRole)
		if err != nil {
			klog.Error(err)
			return err
		}
	}
	return nil
}

func (c *Controller) multiClusterSync(workspaceRole *iamv1alpha2.WorkspaceRole) error {

	if err := c.ensureNotControlledByKubefed(workspaceRole); err != nil {
		klog.Error(err)
		return err
	}

	obj, exist, err := c.fedWorkspaceRoleCache.GetByKey(workspaceRole.Name)
	if !exist {
		return c.createFederatedWorkspaceRole(workspaceRole)
	}
	if err != nil {
		klog.Error(err)
		return err
	}

	var federatedWorkspaceRole iamv1alpha2.FederatedRole
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.(*unstructured.Unstructured).Object, &federatedWorkspaceRole); err != nil {
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(federatedWorkspaceRole.Spec.Template.Rules, workspaceRole.Rules) ||
		!reflect.DeepEqual(federatedWorkspaceRole.Spec.Template.Labels, workspaceRole.Labels) ||
		!reflect.DeepEqual(federatedWorkspaceRole.Spec.Template.Annotations, workspaceRole.Annotations) {

		federatedWorkspaceRole.Spec.Template.Rules = workspaceRole.Rules
		federatedWorkspaceRole.Spec.Template.Annotations = workspaceRole.Annotations
		federatedWorkspaceRole.Spec.Template.Labels = workspaceRole.Labels

		return c.updateFederatedGlobalRole(&federatedWorkspaceRole)
	}

	return nil
}

func (c *Controller) createFederatedWorkspaceRole(workspaceRole *iamv1alpha2.WorkspaceRole) error {
	federatedWorkspaceRole := &iamv1alpha2.FederatedRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       iamv1alpha2.FedWorkspaceRoleKind,
			APIVersion: iamv1alpha2.FedWorkspaceRoleResource.Group + "/" + iamv1alpha2.FedWorkspaceRoleResource.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: workspaceRole.Name,
		},
		Spec: iamv1alpha2.FederatedRoleSpec{
			Template: iamv1alpha2.RoleTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      workspaceRole.Labels,
					Annotations: workspaceRole.Annotations,
				},
				Rules: workspaceRole.Rules,
			},
			Placement: iamv1alpha2.Placement{
				ClusterSelector: iamv1alpha2.ClusterSelector{},
			},
		},
	}

	err := controllerutil.SetControllerReference(workspaceRole, federatedWorkspaceRole, scheme.Scheme)
	if err != nil {
		return err
	}

	data, err := json.Marshal(federatedWorkspaceRole)
	if err != nil {
		return err
	}

	cli := c.k8sClient.(*kubernetes.Clientset)
	err = cli.RESTClient().Post().
		AbsPath(fmt.Sprintf("/apis/%s/%s/%s", iamv1alpha2.FedWorkspaceRoleResource.Group,
			iamv1alpha2.FedWorkspaceRoleResource.Version, iamv1alpha2.FedWorkspaceRoleResource.Name)).
		Body(data).
		Do().Error()

	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func (c *Controller) updateFederatedGlobalRole(federatedWorkspaceRole *iamv1alpha2.FederatedRole) error {

	data, err := json.Marshal(federatedWorkspaceRole)
	if err != nil {
		return err
	}

	cli := c.k8sClient.(*kubernetes.Clientset)

	err = cli.RESTClient().Put().
		AbsPath(fmt.Sprintf("/apis/%s/%s/%s/%s", iamv1alpha2.FedWorkspaceRoleResource.Group,
			iamv1alpha2.FedWorkspaceRoleResource.Version, iamv1alpha2.FedWorkspaceRoleResource.Name,
			federatedWorkspaceRole.Name)).
		Body(data).
		Do().Error()
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func (c *Controller) ensureNotControlledByKubefed(workspaceRole *iamv1alpha2.WorkspaceRole) error {
	if workspaceRole.Labels[constants.KubefedManagedLabel] != "false" {
		if workspaceRole.Labels == nil {
			workspaceRole.Labels = make(map[string]string, 0)
		}
		workspaceRole = workspaceRole.DeepCopy()
		workspaceRole.Labels[constants.KubefedManagedLabel] = "false"
		_, err := c.ksClient.IamV1alpha2().WorkspaceRoles().Update(workspaceRole)
		if err != nil {
			klog.Error(err)
		}
	}
	return nil
}
