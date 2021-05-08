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

package workspacerolebinding

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
	typesv1beta1 "kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/constants"
	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

const (
	controllerName = "workspacerolebinding-controller"
)

// Reconciler reconciles a WorkspaceRoleBinding object
type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Scheme                  *runtime.Scheme
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
	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
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
		For(&iamv1alpha2.WorkspaceRoleBinding{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=workspacerolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=types.kubefed.io,resources=federatedworkspacerolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("workspacerolebinding", req.NamespacedName)
	rootCtx := context.Background()
	workspaceRoleBinding := &iamv1alpha2.WorkspaceRoleBinding{}
	if err := r.Get(rootCtx, req.NamespacedName, workspaceRoleBinding); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// controlled kubefed-controller-manager
	if workspaceRoleBinding.Labels[constants.KubefedManagedLabel] == "true" {
		return ctrl.Result{}, nil
	}

	if err := r.bindWorkspace(rootCtx, logger, workspaceRoleBinding); err != nil {
		return ctrl.Result{}, err
	}

	if r.MultiClusterEnabled {
		if err := r.multiClusterSync(rootCtx, logger, workspaceRoleBinding); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.Recorder.Event(workspaceRoleBinding, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, logger logr.Logger, workspaceRoleBinding *iamv1alpha2.WorkspaceRoleBinding) error {
	workspaceName := workspaceRoleBinding.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return nil
	}
	workspace := &tenantv1alpha2.WorkspaceTemplate{}
	if err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, workspace); err != nil {
		// skip if workspace not found
		return client.IgnoreNotFound(err)
	}
	// owner reference not match workspace label
	if !metav1.IsControlledBy(workspaceRoleBinding, workspace) {
		workspaceRoleBinding := workspaceRoleBinding.DeepCopy()
		workspaceRoleBinding.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(workspaceRoleBinding.OwnerReferences)
		if err := controllerutil.SetControllerReference(workspace, workspaceRoleBinding, r.Scheme); err != nil {
			logger.Error(err, "set controller reference failed")
			return err
		}
		logger.V(4).Info("update owner reference")
		if err := r.Update(ctx, workspaceRoleBinding); err != nil {
			logger.Error(err, "update owner reference failed")
			return err
		}
	}
	return nil
}

func (r *Reconciler) multiClusterSync(ctx context.Context, logger logr.Logger, workspaceRoleBinding *iamv1alpha2.WorkspaceRoleBinding) error {
	if err := r.ensureNotControlledByKubefed(ctx, logger, workspaceRoleBinding); err != nil {
		return err
	}
	federatedWorkspaceRoleBinding := &typesv1beta1.FederatedWorkspaceRoleBinding{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: workspaceRoleBinding.Name}, federatedWorkspaceRoleBinding); err != nil {
		if errors.IsNotFound(err) {
			if federatedWorkspaceRoleBinding, err := newFederatedWorkspaceRole(workspaceRoleBinding); err != nil {
				logger.Error(err, "generate federated workspace role binding failed")
				return err
			} else {
				if err := r.Create(ctx, federatedWorkspaceRoleBinding); err != nil {
					logger.Error(err, "create federated workspace role binding failed")
					return err
				}
			}
		}
		logger.Error(err, "get federated workspace role binding failed")
		return err
	}

	if !reflect.DeepEqual(federatedWorkspaceRoleBinding.Spec.Template.RoleRef, workspaceRoleBinding.RoleRef) ||
		!reflect.DeepEqual(federatedWorkspaceRoleBinding.Spec.Template.Subjects, workspaceRoleBinding.Subjects) ||
		!reflect.DeepEqual(federatedWorkspaceRoleBinding.Spec.Template.Labels, workspaceRoleBinding.Labels) {

		federatedWorkspaceRoleBinding.Spec.Template.RoleRef = workspaceRoleBinding.RoleRef
		federatedWorkspaceRoleBinding.Spec.Template.Subjects = workspaceRoleBinding.Subjects
		federatedWorkspaceRoleBinding.Spec.Template.Labels = workspaceRoleBinding.Labels

		if err := r.Update(ctx, federatedWorkspaceRoleBinding); err != nil {
			logger.Error(err, "update federated workspace role failed")
			return err
		}
	}

	return nil
}

func newFederatedWorkspaceRole(workspaceRoleBinding *iamv1alpha2.WorkspaceRoleBinding) (*typesv1beta1.FederatedWorkspaceRoleBinding, error) {
	federatedWorkspaceRole := &typesv1beta1.FederatedWorkspaceRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       typesv1beta1.FederatedWorkspaceRoleBindingKind,
			APIVersion: typesv1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: workspaceRoleBinding.Name,
		},
		Spec: typesv1beta1.FederatedWorkspaceRoleBindingSpec{
			Template: typesv1beta1.WorkspaceRoleBindingTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: workspaceRoleBinding.Labels,
				},
				RoleRef:  workspaceRoleBinding.RoleRef,
				Subjects: workspaceRoleBinding.Subjects,
			},
			Placement: typesv1beta1.GenericPlacementFields{
				ClusterSelector: &metav1.LabelSelector{},
			},
		},
	}
	if err := controllerutil.SetControllerReference(workspaceRoleBinding, federatedWorkspaceRole, scheme.Scheme); err != nil {
		return nil, err
	}
	return federatedWorkspaceRole, nil
}

func (r *Reconciler) ensureNotControlledByKubefed(ctx context.Context, logger logr.Logger, workspaceRoleBinding *iamv1alpha2.WorkspaceRoleBinding) error {
	if workspaceRoleBinding.Labels[constants.KubefedManagedLabel] != "false" {
		if workspaceRoleBinding.Labels == nil {
			workspaceRoleBinding.Labels = make(map[string]string)
		}
		workspaceRoleBinding = workspaceRoleBinding.DeepCopy()
		workspaceRoleBinding.Labels[constants.KubefedManagedLabel] = "false"
		logger.V(4).Info("update kubefed managed label")
		if err := r.Update(ctx, workspaceRoleBinding); err != nil {
			logger.Error(err, "update kubefed managed label failed")
			return err
		}
	}
	return nil
}
