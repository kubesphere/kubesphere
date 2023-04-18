package clusterrole

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	rbachelper "kubesphere.io/kubesphere/pkg/conponenthelper/auth/rbac"
)

const controllerName = "clusterrole-controller"

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
		For(&iamv1beta1.ClusterRole{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	clusterRole := &iamv1beta1.ClusterRole{}
	err := r.Get(ctx, req.NamespacedName, clusterRole)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if clusterRole.AggregationRoleTemplates != nil {
		err = r.helper.AggregationRole(ctx, rbachelper.ClusterRoleRuleOwner{ClusterRole: clusterRole}, r.Recorder)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
