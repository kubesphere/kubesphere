/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package role

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	rbachelper "kubesphere.io/kubesphere/pkg/componenthelper/auth/rbac"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	rbacutils "kubesphere.io/kubesphere/pkg/utils/rbac"
)

const (
	controllerName = "role"
	roleRef        = "iam.kubesphere.io/role-ref"
)

var _ kscontroller.Controller = &Reconciler{}

type Reconciler struct {
	client.Client
	logger   logr.Logger
	recorder record.EventRecorder
	helper   *rbachelper.Helper
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.helper = rbachelper.NewHelper(r.Client)
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&iamv1beta1.Role{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger.WithValues("Role", req.String())
	role := &iamv1beta1.Role{}
	if err := r.Get(ctx, req.NamespacedName, role); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if role.AggregationRoleTemplates != nil {
		if err := r.helper.AggregationRole(ctx, rbachelper.RoleRuleOwner{Role: role}, r.recorder); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.syncToKubernetes(ctx, role); err != nil {
		log.Error(err, "sync role failed")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) syncToKubernetes(ctx context.Context, role *iamv1beta1.Role) error {
	k8sRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Namespace: role.Namespace, Name: rbacutils.RelatedK8sResourceName(role.Name)},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, k8sRole, func() error {
		if k8sRole.Labels == nil {
			k8sRole.Labels = make(map[string]string)
		}
		k8sRole.Labels[roleRef] = role.Name
		k8sRole.Rules = role.Rules
		if err := controllerutil.SetOwnerReference(role, k8sRole, r.Scheme()); err != nil {
			return fmt.Errorf("failed to set owner reference: %s", err)
		}
		return nil
	})

	if err != nil {
		r.logger.Error(err, "sync role failed", "namespace", role.Namespace, "role", role.Name)
		return err
	}

	r.logger.V(4).Info("sync role to K8s", "namespace", role.Namespace, "role", role.Name, "op", op)
	return nil
}
