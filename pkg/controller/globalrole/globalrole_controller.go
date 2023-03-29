package globalrole

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
)

const (
	iamLabelGlobalScope = "scope.iam.kubesphere.io/global"

	controllerName = "globalrole-controller"

	messageResourceSynced = "Aggregating roleTemplates successfully"
)

// GlobalRoleReconciler reconciles a GlobalRole object
type GlobalRoleReconciler struct {
	client.Client
	logger   logr.Logger
	scheme   *runtime.Scheme
	recorder record.EventRecorder
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
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if globalrole.AggregationRoleTemplates != nil {
		newPolicyRules, newTemplateNames, err := r.getAggregationRoleTemplateRule(ctx, iamLabelGlobalScope, globalrole.AggregationRoleTemplates)
		if err != nil {
			r.recorder.Event(globalrole, corev1.EventTypeWarning, controllerutils.FailedSynced, err.Error())
			return ctrl.Result{}, err
		}

		if equality.Semantic.DeepEqual(newTemplateNames, globalrole.AggregationRoleTemplates.TemplateNames) {
			return ctrl.Result{}, nil
		}
		globalrole.Rules = newPolicyRules
		globalrole.AggregationRoleTemplates.TemplateNames = newTemplateNames

		err = r.Update(ctx, globalrole)
		if err != nil {
			r.recorder.Event(globalrole, corev1.EventTypeWarning, controllerutils.FailedSynced, err.Error())
			return ctrl.Result{}, err
		}
		r.recorder.Event(globalrole, corev1.EventTypeNormal, controllerutils.SuccessSynced, messageResourceSynced)
	}

	return ctrl.Result{}, nil
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
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&iamv1beta1.GlobalRole{}).
		Complete(r)
}

// TODO This function could be a generic function for all role types including the field AggregationTemplates
func (r *GlobalRoleReconciler) getAggregationRoleTemplateRule(ctx context.Context, scopeKey string, templates *iamv1beta1.AggregationRoleTemplates) ([]rbacv1.PolicyRule, []string, error) {
	rules := make([]rbacv1.PolicyRule, 0)
	newTemplateNames := make([]string, 0)
	if len(templates.RoleSelectors) == 0 {
		for _, name := range templates.TemplateNames {
			roleTemplate := &iamv1beta1.RoleTemplate{}
			err := r.Get(ctx, types.NamespacedName{Name: name}, roleTemplate)
			if err != nil {
				if errors.IsNotFound(err) {
					klog.Errorf("Get RoleTemplate %s failed: %s", name, err)
					continue
				} else {
					return nil, nil, err
				}
			}

			// Ensure the roleTemplate can be aggregated at the specific role scope
			if _, exist := roleTemplate.Labels[scopeKey]; !exist {
				klog.Errorf("RoleTemplate %s not match scope", roleTemplate.Name)
				continue
			}
			rules = append(rules, roleTemplate.Spec.Rules...)
		}
		newTemplateNames = templates.TemplateNames
	} else {
		for _, selector := range templates.RoleSelectors {
			roleTemplateList := &iamv1beta1.RoleTemplateList{}
			// Ensure the roleTemplate can be aggregated at the specific role scope
			selector.MatchLabels = labels.Merge(selector.MatchLabels, map[string]string{scopeKey: ""})
			asSelector, err := metav1.LabelSelectorAsSelector(&selector)
			if err != nil {
				return nil, nil, err
			}
			if err = r.List(ctx, roleTemplateList, &client.ListOptions{LabelSelector: asSelector}); err != nil {
				return nil, nil, err
			}

			for _, roleTemplate := range roleTemplateList.Items {
				newTemplateNames = append(newTemplateNames, roleTemplate.Name)
				for _, rule := range roleTemplate.Spec.Rules {
					if !ruleExists(rules, rule) {
						rules = append(rules, rule)
					}
				}
			}
		}
	}
	return rules, newTemplateNames, nil
}

func ruleExists(haystack []rbacv1.PolicyRule, needle rbacv1.PolicyRule) bool {
	for _, curr := range haystack {
		if equality.Semantic.DeepEqual(curr, needle) {
			return true
		}
	}
	return false
}
