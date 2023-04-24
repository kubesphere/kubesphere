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

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"

	rbachelper "kubesphere.io/kubesphere/pkg/conponenthelper/auth/rbac"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

const (
	controllerName = "workspacerole-controller"
)

// Reconciler reconciles a WorkspaceRole object
type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
	helper                  *rbachelper.Helper
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Logger.GetSink() == nil {
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

	if r.helper == nil {
		r.helper = rbachelper.NewHelper(r.Client)
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		For(&iamv1beta1.WorkspaceRole{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=workspaceroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=types.kubefed.io,resources=federatedworkspaceroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=workspaces,verbs=get;list;watch;
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues("workspacerole", req.NamespacedName)
	workspaceRole := &iamv1beta1.WorkspaceRole{}
	err := r.Get(ctx, req.NamespacedName, workspaceRole)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.bindWorkspace(ctx, logger, workspaceRole); err != nil {
		return ctrl.Result{}, err
	}

	if workspaceRole.AggregationRoleTemplates != nil {
		err = r.helper.AggregationRole(ctx, rbachelper.WorkspaceRoleRuleOwner{WorkspaceRole: workspaceRole}, r.Recorder)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) bindWorkspace(ctx context.Context, logger logr.Logger, workspaceRole *iamv1beta1.WorkspaceRole) error {
	workspaceName := workspaceRole.Labels[constants.WorkspaceLabelKey]
	if workspaceName == "" {
		return nil
	}
	var workspace tenantv1alpha2.WorkspaceTemplate
	if err := r.Get(ctx, types.NamespacedName{Name: workspaceName}, &workspace); err != nil {
		return client.IgnoreNotFound(err)
	}
	if !metav1.IsControlledBy(workspaceRole, &workspace) {
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
