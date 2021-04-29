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
	controllerName = "workspacerole-controller"
)

// Reconciler reconciles a WorkspaceRole object
type Reconciler struct {
	client.Client
	MultiClusterEnabled     bool
	Logger                  logr.Logger
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
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
		For(&iamv1alpha2.WorkspaceRole{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=workspaceroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=types.kubefed.io,resources=federatedworkspaceroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("workspacerole", req.NamespacedName)
	rootCtx := context.Background()
	workspaceRole := &iamv1alpha2.WorkspaceRole{}
	err := r.Get(rootCtx, req.NamespacedName, workspaceRole)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// controlled kubefed-controller-manager
	if workspaceRole.Labels[constants.KubefedManagedLabel] == "true" {
		return ctrl.Result{}, nil
	}

	if err := r.bindWorkspace(rootCtx, logger, workspaceRole); err != nil {
		return ctrl.Result{}, err
	}

	if r.MultiClusterEnabled {
		if err = r.multiClusterSync(rootCtx, logger, workspaceRole); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.Recorder.Event(workspaceRole, corev1.EventTypeNormal, controllerutils.SuccessSynced, controllerutils.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, logger logr.Logger, workspaceRole *iamv1alpha2.WorkspaceRole) error {
	workspaceName := workspaceRole.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return nil
	}
	var workspace tenantv1alpha2.WorkspaceTemplate
	if err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, &workspace); err != nil {
		return client.IgnoreNotFound(err)
	}
	if !metav1.IsControlledBy(workspaceRole, &workspace) {
		workspaceRole = workspaceRole.DeepCopy()
		workspaceRole.OwnerReferences = k8sutil.RemoveWorkspaceOwnerReference(workspaceRole.OwnerReferences)
		if err := controllerutil.SetControllerReference(&workspace, workspaceRole, r.Scheme); err != nil {
			logger.Error(err, "set controller reference failed")
			return err
		}
		if err := r.Update(ctx, workspaceRole); err != nil {
			logger.Error(err, "update workspace role failed")
			return err
		}
	}
	return nil
}

func (r *Reconciler) multiClusterSync(ctx context.Context, logger logr.Logger, workspaceRole *iamv1alpha2.WorkspaceRole) error {
	if err := r.ensureNotControlledByKubefed(ctx, logger, workspaceRole); err != nil {
		return err
	}
	federatedWorkspaceRole := &typesv1beta1.FederatedWorkspaceRole{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: workspaceRole.Name}, federatedWorkspaceRole); err != nil {
		if errors.IsNotFound(err) {
			if federatedWorkspaceRole, err := newFederatedWorkspaceRole(workspaceRole); err != nil {
				logger.Error(err, "create federated workspace role failed")
				return err
			} else {
				if err := r.Create(ctx, federatedWorkspaceRole); err != nil {
					logger.Error(err, "create federated workspace role failed")
					return err
				}
			}
		}
		logger.Error(err, "get federated workspace role failed")
		return err
	}

	if !reflect.DeepEqual(federatedWorkspaceRole.Spec.Template.Rules, workspaceRole.Rules) ||
		!reflect.DeepEqual(federatedWorkspaceRole.Spec.Template.Labels, workspaceRole.Labels) {

		federatedWorkspaceRole.Spec.Template.Rules = workspaceRole.Rules
		federatedWorkspaceRole.Spec.Template.Labels = workspaceRole.Labels

		if err := r.Update(ctx, federatedWorkspaceRole); err != nil {
			logger.Error(err, "update federated workspace role failed")
			return err
		}
	}

	return nil
}

func newFederatedWorkspaceRole(workspaceRole *iamv1alpha2.WorkspaceRole) (*typesv1beta1.FederatedWorkspaceRole, error) {
	federatedWorkspaceRole := &typesv1beta1.FederatedWorkspaceRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       typesv1beta1.FederatedWorkspaceRoleKind,
			APIVersion: typesv1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: workspaceRole.Name,
		},
		Spec: typesv1beta1.FederatedWorkspaceRoleSpec{
			Template: typesv1beta1.WorkspaceRoleTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels: workspaceRole.Labels,
				},
				Rules: workspaceRole.Rules,
			},
			Placement: typesv1beta1.GenericPlacementFields{
				ClusterSelector: &metav1.LabelSelector{},
			},
		},
	}

	if err := controllerutil.SetControllerReference(workspaceRole, federatedWorkspaceRole, scheme.Scheme); err != nil {
		return nil, err
	}

	return federatedWorkspaceRole, nil
}

func (r *Reconciler) ensureNotControlledByKubefed(ctx context.Context, logger logr.Logger, workspaceRole *iamv1alpha2.WorkspaceRole) error {
	if workspaceRole.Labels[constants.KubefedManagedLabel] != "false" {
		if workspaceRole.Labels == nil {
			workspaceRole.Labels = make(map[string]string)
		}
		workspaceRole = workspaceRole.DeepCopy()
		workspaceRole.Labels[constants.KubefedManagedLabel] = "false"
		if err := r.Update(ctx, workspaceRole); err != nil {
			logger.Error(err, "update kubefed managed label failed")
			return err
		}
	}
	return nil
}
