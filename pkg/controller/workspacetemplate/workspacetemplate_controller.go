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

package workspacetemplate

import (
	"bytes"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha2"
	typesv1beta1 "kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	tenantv1alpha1informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/tenant/v1alpha1"
	tenantv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/tenant/v1alpha2"
	typesv1beta1informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/types/v1beta1"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	tenantv1alpha1listers "kubesphere.io/kubesphere/pkg/client/listers/tenant/v1alpha1"
	tenantv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/tenant/v1alpha2"
	typesv1beta1listers "kubesphere.io/kubesphere/pkg/client/listers/types/v1beta1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced = "WorkspaceTemplate synced successfully"
	controllerName        = "workspacetemplate-controller"
)

type controller struct {
	k8sClient                kubernetes.Interface
	ksClient                 kubesphere.Interface
	workspaceTemplateLister  tenantv1alpha2listers.WorkspaceTemplateLister
	workspaceTemplateSynced  cache.InformerSynced
	workspaceRoleLister      iamv1alpha2listers.WorkspaceRoleLister
	workspaceRoleSynced      cache.InformerSynced
	roleBaseLister           iamv1alpha2listers.RoleBaseLister
	roleBaseSynced           cache.InformerSynced
	workspaceLister          tenantv1alpha1listers.WorkspaceLister
	workspaceSynced          cache.InformerSynced
	federatedWorkspaceLister typesv1beta1listers.FederatedWorkspaceLister
	federatedWorkspaceSynced cache.InformerSynced
	multiClusterEnabled      bool
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

func NewController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface,
	workspaceTemplateInformer tenantv1alpha2informers.WorkspaceTemplateInformer,
	workspaceInformer tenantv1alpha1informers.WorkspaceInformer,
	roleBaseInformer iamv1alpha2informers.RoleBaseInformer,
	workspaceRoleInformer iamv1alpha2informers.WorkspaceRoleInformer,
	federatedWorkspaceInformer typesv1beta1informers.FederatedWorkspaceInformer,
	multiClusterEnabled bool) *controller {

	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	ctl := &controller{
		k8sClient:               k8sClient,
		ksClient:                ksClient,
		workspaceTemplateLister: workspaceTemplateInformer.Lister(),
		workspaceTemplateSynced: workspaceTemplateInformer.Informer().HasSynced,
		workspaceLister:         workspaceInformer.Lister(),
		workspaceSynced:         workspaceInformer.Informer().HasSynced,
		workspaceRoleLister:     workspaceRoleInformer.Lister(),
		workspaceRoleSynced:     workspaceRoleInformer.Informer().HasSynced,
		roleBaseLister:          roleBaseInformer.Lister(),
		roleBaseSynced:          roleBaseInformer.Informer().HasSynced,
		multiClusterEnabled:     multiClusterEnabled,
		workqueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "WorkspaceTemplate"),
		recorder:                recorder,
	}

	if multiClusterEnabled {
		ctl.federatedWorkspaceLister = federatedWorkspaceInformer.Lister()
		ctl.federatedWorkspaceSynced = federatedWorkspaceInformer.Informer().HasSynced
	}

	klog.Info("Setting up event handlers")
	workspaceTemplateInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueWorkspaceTemplate,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueWorkspaceTemplate(new)
		},
		DeleteFunc: ctl.enqueueWorkspaceTemplate,
	})
	return ctl
}

func (c *controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting WorkspaceTemplate controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")

	synced := make([]cache.InformerSynced, 0)
	synced = append(synced, c.workspaceTemplateSynced, c.workspaceSynced, c.workspaceRoleSynced, c.roleBaseSynced)
	if c.multiClusterEnabled {
		synced = append(synced, c.federatedWorkspaceSynced)
	}
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

func (c *controller) enqueueWorkspaceTemplate(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *controller) processNextWorkItem() bool {
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

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *controller) reconcile(key string) error {
	workspaceTemplate, err := c.workspaceTemplateLister.Get(key)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("workspace template '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}

	if err = c.initRoles(workspaceTemplate); err != nil {
		klog.Error(err)
		return err
	}

	if err = c.initManagerRoleBinding(workspaceTemplate); err != nil {
		klog.Error(err)
		return err
	}

	if c.multiClusterEnabled {
		if err = c.multiClusterSync(workspaceTemplate); err != nil {
			klog.Error(err)
			return err
		}
	} else {
		if err = c.sync(workspaceTemplate); err != nil {
			klog.Error(err)
			return err
		}
	}

	c.recorder.Event(workspaceTemplate, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *controller) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

func (c *controller) multiClusterSync(workspaceTemplate *tenantv1alpha2.WorkspaceTemplate) error {
	// multi cluster environment, synchronize workspaces with kubefed
	federatedWorkspace, err := c.federatedWorkspaceLister.Get(workspaceTemplate.Name)
	if err != nil {
		// create federatedworkspace if not found
		if errors.IsNotFound(err) {
			return c.createFederatedWorkspace(workspaceTemplate)
		}
		klog.Error(err)
		return err
	}
	// update spec
	if !reflect.DeepEqual(federatedWorkspace.Spec, workspaceTemplate.Spec) {
		federatedWorkspace.Spec = workspaceTemplate.Spec
		if err = c.updateFederatedWorkspace(federatedWorkspace); err != nil {
			klog.Error(err)
			return err
		}
	}

	return nil
}

func (c *controller) createFederatedWorkspace(workspaceTemplate *tenantv1alpha2.WorkspaceTemplate) error {
	federatedWorkspace := &typesv1beta1.FederatedWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: workspaceTemplate.Name,
		},
		Spec: workspaceTemplate.Spec,
	}

	if err := controllerutil.SetControllerReference(workspaceTemplate, federatedWorkspace, scheme.Scheme); err != nil {
		return err
	}

	if _, err := c.ksClient.TypesV1beta1().FederatedWorkspaces().Create(federatedWorkspace); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	return nil
}

func (c *controller) sync(workspaceTemplate *tenantv1alpha2.WorkspaceTemplate) error {
	workspace, err := c.workspaceLister.Get(workspaceTemplate.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.createWorkspace(workspaceTemplate)
		}
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(workspace.Spec, workspaceTemplate.Spec.Template.Spec) ||
		!reflect.DeepEqual(workspace.Labels, workspaceTemplate.Spec.Template.Labels) ||
		!reflect.DeepEqual(workspace.Annotations, workspaceTemplate.Spec.Template.Annotations) {

		workspace = workspace.DeepCopy()
		workspace.Spec = workspaceTemplate.Spec.Template.Spec
		workspace.Labels = workspaceTemplate.Spec.Template.Labels
		workspace.Annotations = workspaceTemplate.Spec.Template.Annotations

		return c.updateWorkspace(workspace)
	}

	return nil
}

func (c *controller) createWorkspace(workspaceTemplate *tenantv1alpha2.WorkspaceTemplate) error {
	workspace := &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        workspaceTemplate.Name,
			Labels:      workspaceTemplate.Spec.Template.Labels,
			Annotations: workspaceTemplate.Spec.Template.Annotations,
		},
		Spec: workspaceTemplate.Spec.Template.Spec,
	}

	err := controllerutil.SetControllerReference(workspaceTemplate, workspace, scheme.Scheme)
	if err != nil {
		return err
	}

	_, err = c.ksClient.TenantV1alpha1().Workspaces().Create(workspace)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	return nil
}

func (c *controller) updateWorkspace(workspace *tenantv1alpha1.Workspace) error {
	_, err := c.ksClient.TenantV1alpha1().Workspaces().Update(workspace)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *controller) initRoles(workspace *tenantv1alpha2.WorkspaceTemplate) error {
	roleBases, err := c.roleBaseLister.List(labels.Everything())
	if err != nil {
		klog.Error(err)
		return err
	}
	for _, roleBase := range roleBases {
		var role iamv1alpha2.WorkspaceRole
		if err = yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(roleBase.Role.Raw), 1024).Decode(&role); err == nil && role.Kind == iamv1alpha2.ResourceKindWorkspaceRole {
			roleName := fmt.Sprintf("%s-%s", workspace.Name, role.Name)
			if role.Labels == nil {
				role.Labels = make(map[string]string, 0)
			}
			// make sure workspace label always exist
			role.Labels[tenantv1alpha1.WorkspaceLabel] = workspace.Name
			role.Name = roleName
			old, err := c.workspaceRoleLister.Get(roleName)
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = c.ksClient.IamV1alpha2().WorkspaceRoles().Create(&role)
					if err != nil {
						klog.Error(err)
						return err
					}
					continue
				}
			}
			if !reflect.DeepEqual(role.Labels, old.Labels) ||
				!reflect.DeepEqual(role.Annotations, old.Annotations) ||
				!reflect.DeepEqual(role.Rules, old.Rules) {
				updated := old.DeepCopy()
				updated.Labels = role.Labels
				updated.Annotations = role.Annotations
				updated.Rules = role.Rules
				_, err = c.ksClient.IamV1alpha2().WorkspaceRoles().Update(updated)
				if err != nil {
					klog.Error(err)
					return err
				}
			}
		}
	}
	return nil
}

func (c *controller) resetWorkspaceOwner(workspace *tenantv1alpha2.WorkspaceTemplate) error {
	workspace = workspace.DeepCopy()
	workspace.Spec.Template.Spec.Manager = ""
	_, err := c.ksClient.TenantV1alpha2().WorkspaceTemplates().Update(workspace)
	klog.V(4).Infof("update workspace after manager has been deleted")
	return err
}

func (c *controller) initManagerRoleBinding(workspace *tenantv1alpha2.WorkspaceTemplate) error {
	manager := workspace.Spec.Template.Spec.Manager
	if manager == "" {
		return nil
	}

	user, err := c.ksClient.IamV1alpha2().Users().Get(manager, metav1.GetOptions{})
	if err != nil {
		// skip if user has been deleted
		if errors.IsNotFound(err) {
			return c.resetWorkspaceOwner(workspace)
		}
		klog.Error(err)
		return err
	}

	// skip if user has been deleted
	if !user.DeletionTimestamp.IsZero() {
		return c.resetWorkspaceOwner(workspace)
	}

	workspaceAdminRoleName := fmt.Sprintf(iamv1alpha2.WorkspaceAdminFormat, workspace.Name)
	managerRoleBinding := &iamv1alpha2.WorkspaceRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", manager, workspaceAdminRoleName),
			Labels: map[string]string{
				tenantv1alpha1.WorkspaceLabel:  workspace.Name,
				iamv1alpha2.UserReferenceLabel: manager,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:     iamv1alpha2.ResourceKindWorkspaceRole,
			Name:     workspaceAdminRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Name:     manager,
				Kind:     iamv1alpha2.ResourceKindUser,
				APIGroup: rbacv1.GroupName,
			},
		},
	}
	_, err = c.ksClient.IamV1alpha2().WorkspaceRoleBindings().Create(managerRoleBinding)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	return nil
}

func (c *controller) updateFederatedWorkspace(workspace *typesv1beta1.FederatedWorkspace) error {
	_, err := c.ksClient.TypesV1beta1().FederatedWorkspaces().Update(workspace)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
