package globalrole

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/controller/utils/iam"
)

const (
	controllerName = "globalrole-controller"
)

// GlobalRoleReconciler reconciles a GlobalRole object
type GlobalRoleReconciler struct {
	client.Client
	logger   logr.Logger
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	helper   *iam.Helper
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the GlobalRole object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *GlobalRoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	globalrole := &iamv1beta1.GlobalRole{}
	err := r.Get(ctx, req.NamespacedName, globalrole)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if globalrole.AggregationRoleTemplates != nil {
		err = r.aggregateRoleTemplate(ctx, globalrole)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *GlobalRoleReconciler) aggregateRoleTemplate(ctx context.Context, globalrole *iamv1beta1.GlobalRole) error {
	newPolicyRules, newTemplateNames, err := r.helper.GetAggregationRoleTemplateRule(ctx, iam.LabelGlobalScope, globalrole.AggregationRoleTemplates)
	if err != nil {
		r.recorder.Event(globalrole, corev1.EventTypeWarning, iam.AggregateRoleTemplateFailed, err.Error())
		return err
	}
	if equality.Semantic.DeepEqual(globalrole.Rules, newPolicyRules) &&
		equality.Semantic.DeepEqual(globalrole.AggregationRoleTemplates.TemplateNames, newTemplateNames) {
		return nil
	}
	globalrole.Rules = newPolicyRules
	globalrole.AggregationRoleTemplates.TemplateNames = newTemplateNames

	err = r.Update(ctx, globalrole)
	if err != nil {
		r.recorder.Event(globalrole, corev1.EventTypeWarning, iam.AggregateRoleTemplateFailed, err.Error())
		return err
	}
	r.recorder.Event(globalrole, corev1.EventTypeNormal, controllerutils.SuccessSynced, iam.MessageResourceSynced)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GlobalRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.logger.GetSink() == nil {
		r.logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	}
	if r.scheme == nil {
		r.scheme = mgr.GetScheme()
	}
	if r.recorder == nil {
		r.recorder = mgr.GetEventRecorderFor(controllerName)
	}

	if r.helper == nil {
		r.helper = iam.NewHelper(r.Client)
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&iamv1beta1.GlobalRole{}).
		Complete(r)
}
