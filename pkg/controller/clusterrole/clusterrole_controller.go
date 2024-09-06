/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package clusterrole

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbachelper "kubesphere.io/kubesphere/pkg/componenthelper/auth/rbac"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
	rbacutils "kubesphere.io/kubesphere/pkg/utils/rbac"
)

const (
	controllerName = "clusterrole"
	roleRef        = "iam.kubesphere.io/clusterrole-ref"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

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
	r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.Client = mgr.GetClient()
	r.helper = rbachelper.NewHelper(mgr.GetClient())

	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&iamv1beta1.ClusterRole{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger.WithValues("ClusterRole", req.String())
	clusterRole := &iamv1beta1.ClusterRole{}
	if err := r.Get(ctx, req.NamespacedName, clusterRole); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if clusterRole.AggregationRoleTemplates != nil {
		if err := r.helper.AggregationRole(ctx, rbachelper.ClusterRoleRuleOwner{ClusterRole: clusterRole}, r.recorder); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.syncToKubernetes(ctx, clusterRole); err != nil {
		log.Error(err, "sync cluster role failed")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

func (r *Reconciler) syncToKubernetes(ctx context.Context, clusterRole *iamv1beta1.ClusterRole) error {
	k8sClusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: rbacutils.RelatedK8sResourceName(clusterRole.Name)},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, k8sClusterRole, func() error {
		if k8sClusterRole.Labels == nil {
			k8sClusterRole.Labels = make(map[string]string)
		}
		k8sClusterRole.Labels[roleRef] = clusterRole.Name
		k8sClusterRole.Rules = clusterRole.Rules
		if err := controllerutil.SetOwnerReference(clusterRole, k8sClusterRole, r.Scheme()); err != nil {
			return fmt.Errorf("failed to set owner reference: %s", err)
		}
		return nil
	})

	if err != nil {
		r.logger.Error(err, "sync cluster role failed", "cluster role", clusterRole.Name)
	}

	r.logger.V(4).Info("sync cluster role to K8s", "cluster role", clusterRole.Name, "op", op)
	return nil
}
