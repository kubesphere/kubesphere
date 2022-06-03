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
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/api/application/v1alpha1"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
	typesv1beta1 "kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/constants"
	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	controllerName                  = "workspacetemplate-controller"
	workspaceTemplateFinalizer      = "finalizers.workspacetemplate.kubesphere.io"
	orphanFinalizer                 = "orphan.finalizers.kubesphere.io"
	orphanDeleteOptionAnnotationKey = "kubefed.io/deleteoption"
	orphanDeleteOptionAnnotation    = "{\"propagationPolicy\":\"Orphan\"}"
)

// Reconciler reconciles a WorkspaceRoleBinding object
type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
	MultiClusterEnabled     bool
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Logger == nil {
		r.Logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	}
	if r.Recorder == nil {
		r.Recorder = mgr.GetEventRecorderFor(controllerName)
	}
	if r.MaxConcurrentReconciles <= 0 {
		r.MaxConcurrentReconciles = 1
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		For(&tenantv1alpha2.WorkspaceTemplate{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=workspacerolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=types.kubefed.io,resources=federatedworkspacerolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("workspacetemplate", req.NamespacedName)
	workspaceTemplate := &tenantv1alpha2.WorkspaceTemplate{}
	if err := r.Get(ctx, req.NamespacedName, workspaceTemplate); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if workspaceTemplate.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(workspaceTemplate.ObjectMeta.Finalizers, workspaceTemplateFinalizer) {
			workspaceTemplate.ObjectMeta.Finalizers = append(workspaceTemplate.ObjectMeta.Finalizers, workspaceTemplateFinalizer)
			if err := r.Update(ctx, workspaceTemplate); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(workspaceTemplate.ObjectMeta.Finalizers, workspaceTemplateFinalizer) ||
			sliceutil.HasString(workspaceTemplate.ObjectMeta.Finalizers, orphanFinalizer) {
			if err := r.deleteOpenPitrixResourcesInWorkspace(ctx, workspaceTemplate.Name); err != nil {
				logger.Error(err, "failed to delete related openpitrix resource")
				return ctrl.Result{}, err
			}

			if err := r.deleteWorkspace(ctx, workspaceTemplate); err != nil {
				if errors.IsNotFound(err) {
					logger.V(4).Info("related workspace not found")
				} else {
					logger.Error(err, "failed to delete related workspace")
					return ctrl.Result{}, nil
				}
			}

			// remove our finalizer from the list and update it.
			workspaceTemplate.ObjectMeta.Finalizers = sliceutil.RemoveString(workspaceTemplate.ObjectMeta.Finalizers, func(item string) bool {
				return item == workspaceTemplateFinalizer || item == orphanFinalizer
			})

			logger.V(4).Info("update workspace template")
			if err := r.Update(ctx, workspaceTemplate); err != nil {
				logger.Error(err, "update workspace template failed")
				return ctrl.Result{}, err
			}
		}
		// Our finalizer has finished, so the reconciler can do nothing.
		return ctrl.Result{}, nil
	}

	if r.MultiClusterEnabled {
		if err := r.multiClusterSync(ctx, logger, workspaceTemplate); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.singleClusterSync(ctx, logger, workspaceTemplate); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err := r.initWorkspaceRoles(ctx, logger, workspaceTemplate); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.initManagerRoleBinding(ctx, logger, workspaceTemplate); err != nil {
		return ctrl.Result{}, err
	}
	r.Recorder.Event(workspaceTemplate, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) singleClusterSync(ctx context.Context, logger logr.Logger, workspaceTemplate *tenantv1alpha2.WorkspaceTemplate) error {
	workspace := &tenantv1alpha1.Workspace{}
	if err := r.Get(ctx, types.NamespacedName{Name: workspaceTemplate.Name}, workspace); err != nil {
		if errors.IsNotFound(err) {
			if workspace, err := newWorkspace(workspaceTemplate); err != nil {
				logger.Error(err, "generate workspace failed")
				return err
			} else {
				if err := r.Create(ctx, workspace); err != nil {
					logger.Error(err, "create workspace failed")
					return err
				}
				return nil
			}
		}
		logger.Error(err, "get workspace failed")
		return err
	}
	if !reflect.DeepEqual(workspace.Spec, workspaceTemplate.Spec.Template.Spec) ||
		!reflect.DeepEqual(workspace.Labels, workspaceTemplate.Spec.Template.Labels) {

		workspace = workspace.DeepCopy()
		workspace.Spec = workspaceTemplate.Spec.Template.Spec
		workspace.Labels = workspaceTemplate.Spec.Template.Labels

		if err := r.Update(ctx, workspace); err != nil {
			logger.Error(err, "update workspace failed")
			return err
		}
	}
	return nil
}

func (r *Reconciler) multiClusterSync(ctx context.Context, logger logr.Logger, workspaceTemplate *tenantv1alpha2.WorkspaceTemplate) error {
	if err := r.ensureNotControlledByKubefed(ctx, logger, workspaceTemplate); err != nil {
		return err
	}
	federatedWorkspace := &typesv1beta1.FederatedWorkspace{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: workspaceTemplate.Name}, federatedWorkspace); err != nil {
		if errors.IsNotFound(err) {
			if federatedWorkspace, err := newFederatedWorkspace(workspaceTemplate); err != nil {
				logger.Error(err, "generate federated workspace failed")
				return err
			} else {
				if err := r.Create(ctx, federatedWorkspace); err != nil {
					logger.Error(err, "create federated workspace failed")
					return err
				}
				return nil
			}
		}
		logger.Error(err, "get federated workspace failed")
		return err
	}

	if !reflect.DeepEqual(federatedWorkspace.Spec, workspaceTemplate.Spec) ||
		!reflect.DeepEqual(federatedWorkspace.Labels, workspaceTemplate.Labels) {

		federatedWorkspace.Spec = workspaceTemplate.Spec
		federatedWorkspace.Labels = workspaceTemplate.Labels

		if err := r.Update(ctx, federatedWorkspace); err != nil {
			logger.Error(err, "update federated workspace failed")
			return err
		}
	}

	return nil
}

func newFederatedWorkspace(template *tenantv1alpha2.WorkspaceTemplate) (*typesv1beta1.FederatedWorkspace, error) {
	federatedWorkspace := &typesv1beta1.FederatedWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   template.Name,
			Labels: template.Labels,
		},
		Spec: template.Spec,
	}
	return federatedWorkspace, nil
}

func newWorkspace(template *tenantv1alpha2.WorkspaceTemplate) (*tenantv1alpha1.Workspace, error) {
	workspace := &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   template.Name,
			Labels: template.Spec.Template.Labels,
		},
		Spec: template.Spec.Template.Spec,
	}

	return workspace, nil
}

func (r *Reconciler) deleteWorkspace(ctx context.Context, template *tenantv1alpha2.WorkspaceTemplate) error {
	if r.MultiClusterEnabled {
		federatedWorkspace := &typesv1beta1.FederatedWorkspace{}
		if err := r.Client.Get(ctx, types.NamespacedName{Name: template.Name}, federatedWorkspace); err != nil {
			return err
		}

		// Workspace will be deleted with Orphan Option when it has a orphan finalizer.
		// Reousrces that owned by the Workspace will not be deleted.
		if sliceutil.HasString(template.ObjectMeta.Finalizers, orphanFinalizer) {
			if federatedWorkspace.Annotations == nil {
				federatedWorkspace.Annotations = make(map[string]string, 1)
			}
			federatedWorkspace.Annotations[orphanDeleteOptionAnnotationKey] = orphanDeleteOptionAnnotation
			if err := r.Update(ctx, federatedWorkspace); err != nil {
				return err
			}
		} else {
			// Usually namespace will bind the lifecycle of workspace with ownerReference,
			// in multi-cluster environment workspace will not be created in host cluster
			// if the cluster is not be granted or kubefed-controller-manager is unavailable,
			// this will cause the federated namespace left an orphan object in host cluster.
			// After workspaceTemplate deleted we need to deleted orphan namespace in host cluster directly.
			if err := r.deleteNamespacesInWorkspace(ctx, template); err != nil {
				return err
			}
		}

		if err := r.Delete(ctx, federatedWorkspace); err != nil {
			return err
		}
	}

	opt := &client.DeleteOptions{}
	// Dependents won't be deleted when it's has a orphanFinalizer
	if sliceutil.HasString(template.ObjectMeta.Finalizers, orphanFinalizer) {
		orphan := metav1.DeletePropagationOrphan
		opt = &client.DeleteOptions{PropagationPolicy: &orphan}
	}

	ws := &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: template.Name,
		},
	}
	if err := r.Delete(ctx, ws, opt); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) ensureNotControlledByKubefed(ctx context.Context, logger logr.Logger, workspaceTemplate *tenantv1alpha2.WorkspaceTemplate) error {
	if workspaceTemplate.Labels[constants.KubefedManagedLabel] != "false" {
		if workspaceTemplate.Labels == nil {
			workspaceTemplate.Labels = make(map[string]string)
		}
		workspaceTemplate.Labels[constants.KubefedManagedLabel] = "false"
		logger.V(4).Info("update kubefed managed label")
		if err := r.Update(ctx, workspaceTemplate); err != nil {
			logger.Error(err, "update kubefed managed label failed")
			return err
		}
	}
	return nil
}

func (r *Reconciler) initWorkspaceRoles(ctx context.Context, logger logr.Logger, workspace *tenantv1alpha2.WorkspaceTemplate) error {
	var templates iamv1alpha2.RoleBaseList
	if err := r.List(ctx, &templates); err != nil {
		logger.Error(err, "list role base failed")
		return err
	}
	for _, template := range templates.Items {
		var expected iamv1alpha2.WorkspaceRole
		if err := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(template.Role.Raw), 1024).Decode(&expected); err == nil && expected.Kind == iamv1alpha2.ResourceKindWorkspaceRole {
			expected.Name = fmt.Sprintf("%s-%s", workspace.Name, expected.Name)
			if expected.Labels == nil {
				expected.Labels = make(map[string]string)
			}
			expected.Labels[tenantv1alpha1.WorkspaceLabel] = workspace.Name
			workspaceRole := &iamv1alpha2.WorkspaceRole{}
			if err := r.Get(ctx, types.NamespacedName{Name: expected.Name}, workspaceRole); err != nil {
				if errors.IsNotFound(err) {
					logger.V(4).Info("create workspace role", "workspacerole", expected.Name)
					if err := r.Create(ctx, &expected); err != nil {
						logger.Error(err, "create workspace role failed")
						return err
					}
					continue
				} else {
					logger.Error(err, "get workspace role failed")
					return err
				}
			}
			if !reflect.DeepEqual(expected.Labels, workspaceRole.Labels) ||
				!reflect.DeepEqual(expected.Annotations, workspaceRole.Annotations) ||
				!reflect.DeepEqual(expected.Rules, workspaceRole.Rules) {
				workspaceRole.Labels = expected.Labels
				workspaceRole.Annotations = expected.Annotations
				workspaceRole.Rules = expected.Rules
				logger.V(4).Info("update workspace role", "workspacerole", workspaceRole.Name)
				if err := r.Update(ctx, workspaceRole); err != nil {
					logger.Error(err, "update workspace role failed")
					return err
				}
			}
		} else if err != nil {
			logger.Error(fmt.Errorf("invalid role base found"), "init workspace roles failed", "name", template.Name)
		}
	}
	return nil
}

func (r *Reconciler) initManagerRoleBinding(ctx context.Context, logger logr.Logger, workspace *tenantv1alpha2.WorkspaceTemplate) error {
	manager := workspace.Spec.Template.Spec.Manager
	if manager == "" {
		return nil
	}

	var user iamv1alpha2.User
	if err := r.Get(ctx, types.NamespacedName{Name: manager}, &user); err != nil {
		return client.IgnoreNotFound(err)
	}

	// skip if user has been deleted
	if !user.DeletionTimestamp.IsZero() {
		return nil
	}

	workspaceAdminRoleName := fmt.Sprintf("%s-admin", workspace.Name)
	managerRoleBinding := &iamv1alpha2.WorkspaceRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: workspaceAdminRoleName,
		},
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, managerRoleBinding, workspaceRoleBindingChanger(managerRoleBinding, workspace.Name, manager, workspaceAdminRoleName)); err != nil {
		logger.Error(err, "create workspace manager role binding failed")
		return err
	}

	return nil
}
func (r *Reconciler) deleteOpenPitrixResourcesInWorkspace(ctx context.Context, ws string) error {
	if len(ws) == 0 {
		return nil
	}

	var err error
	// helm release, apps and appVersion only exist in host cluster. Delete these resource in workspace template controller
	if err = r.deleteHelmReleases(ctx, ws); err != nil {
		return err
	}

	if err = r.deleteHelmApps(ctx, ws); err != nil {
		return err
	}

	if err = r.deleteHelmRepos(ctx, ws); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) deleteHelmApps(ctx context.Context, ws string) error {
	if len(ws) == 0 {
		return nil
	}

	apps := v1alpha1.HelmApplicationList{}
	err := r.List(ctx, &apps, &client.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{
		constants.WorkspaceLabelKey: ws}),
	})
	if err != nil {
		return err
	}
	for _, app := range apps.Items {
		if app.Annotations == nil {
			app.Annotations = map[string]string{}
		}
		if _, exists := app.Annotations[constants.DanglingAppCleanupKey]; !exists {
			// Mark the app, the cleanup is in the application controller.
			appCopy := app.DeepCopy()
			appCopy.Annotations[constants.DanglingAppCleanupKey] = constants.CleanupDanglingAppOngoing
			appPatch := client.MergeFrom(&app)
			err = r.Patch(ctx, appCopy, appPatch)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete all helm releases in the workspace ws
func (r *Reconciler) deleteHelmReleases(ctx context.Context, ws string) error {
	if len(ws) == 0 {
		return nil
	}
	rls := &v1alpha1.HelmRelease{}
	err := r.DeleteAllOf(ctx, rls, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{
			constants.WorkspaceLabelKey: ws,
		}),
		}})
	return err
}

func (r *Reconciler) deleteHelmRepos(ctx context.Context, ws string) error {
	if len(ws) == 0 {
		return nil
	}
	rls := &v1alpha1.HelmRepo{}
	err := r.DeleteAllOf(ctx, rls, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{
			constants.WorkspaceLabelKey: ws,
		}),
		}})

	return err
}

// deleteNamespacesInWorkspace Deletes the namespace associated with the workspace, which match the workspace label selector
func (r *Reconciler) deleteNamespacesInWorkspace(ctx context.Context, template *tenantv1alpha2.WorkspaceTemplate) error {
	namespaceList := &corev1.NamespaceList{}
	err := r.Client.List(ctx, namespaceList, client.MatchingLabels{tenantv1alpha1.WorkspaceLabel: template.Name})
	if err != nil {
		return err
	}

	for _, namespace := range namespaceList.Items {
		err = r.Client.Delete(ctx, &namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

func workspaceRoleBindingChanger(workspaceRoleBinding *iamv1alpha2.WorkspaceRoleBinding, workspace, username, workspaceRoleName string) controllerutil.MutateFn {
	return func() error {
		workspaceRoleBinding.Labels = map[string]string{
			tenantv1alpha1.WorkspaceLabel:  workspace,
			iamv1alpha2.UserReferenceLabel: username,
		}

		workspaceRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:     iamv1alpha2.ResourceKindWorkspaceRole,
			Name:     workspaceRoleName,
		}

		workspaceRoleBinding.Subjects = []rbacv1.Subject{
			{
				Name:     username,
				Kind:     rbacv1.UserKind,
				APIGroup: rbacv1.GroupName,
			},
		}
		return nil
	}
}
