package roletemplate

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	autoAggregateIndexKey = ".metadata.annotations[iam.kubesphere.io/auto-aggregate]"

	autoAggregationLabel = "iam.kubesphere.io/auto-aggregate"

	iamLabelGlobalScope = "scope.iam.kubesphere.io/global"
	controllerName      = "roletemplate-controller"

	reasonFailedSync      = "FailedInjectRoleTemplate"
	messageResourceSynced = "RoleTemplate injected successfully"
)

// RoleTemplateReconciler reconciles a GlobalRole object
type RoleTemplateReconciler struct {
	client.Client
	cache.Cache
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
func (r *RoleTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	roletemplate := &iamv1beta1.RoleTemplate{}
	err := r.Client.Get(ctx, req.NamespacedName, roletemplate)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	err = r.autoAggregateRoles(ctx, roletemplate)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// autoAggregateRoles watch the RoleTemplate and automatic inject the RoleTemplate`s rules to the role
// (all role types have the field AggregationRoleTemplates) matching the AggregationRoleTemplates filed.
// Note that autoAggregateRoles just aggregate the templates by .aggregationRoleTemplates.roleSelectors,
// and if the roleTemplate content is changed, the role including the roleTemplate will not be updated.
// TODO support auto-aggregate for ClusterRole, Role, WorkspaceRole
func (r *RoleTemplateReconciler) autoAggregateRoles(ctx context.Context, roletemplate *iamv1beta1.RoleTemplate) error {
	if _, exist := roletemplate.Labels[iamLabelGlobalScope]; exist {
		// list the role having autoAggregate key
		list := &iamv1beta1.GlobalRoleList{}
		l := map[string]string{autoAggregateIndexKey: "true"}
		err := r.Cache.List(ctx, list, client.MatchingFields(l))
		if err != nil {
			return err
		}

		for _, role := range list.Items {
			aggregation := role.AggregationRoleTemplates
			if aggregation != nil &&
				!sliceutil.HasString(aggregation.TemplateNames, roletemplate.Name) {
				for _, selector := range aggregation.RoleSelectors {
					if isContainsLabels(roletemplate.Labels, selector.MatchLabels) {
						err = r.aggregateGlobalRole(ctx, roletemplate, &role)
						if err != nil {
							r.recorder.Event(&role, corev1.EventTypeWarning, reasonFailedSync, err.Error())
							return err
						}
						r.recorder.Event(&role, corev1.EventTypeNormal, controllerutils.SuccessSynced, messageResourceSynced)
						break
					}
				}
			}
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	if r.Cache == nil {
		r.Cache = mgr.GetCache()
	}

	if r.recorder == nil {
		r.recorder = mgr.GetEventRecorderFor(controllerName)
	}

	err := r.Cache.IndexField(context.Background(), &iamv1beta1.GlobalRole{}, autoAggregateIndexKey, globalRoleIndexByAnnotation)
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&iamv1beta1.RoleTemplate{}).
		Complete(r)
}

func globalRoleIndexByAnnotation(obj client.Object) []string {
	role := obj.(*iamv1beta1.GlobalRole)
	if val, ok := role.Annotations[autoAggregationLabel]; ok {
		return []string{val}
	}
	return []string{}
}

func (r *RoleTemplateReconciler) aggregateGlobalRole(ctx context.Context, template *iamv1beta1.RoleTemplate, globalrole *iamv1beta1.GlobalRole) error {
	for _, rule := range template.Spec.Rules {
		if !ruleExists(globalrole.Rules, rule) {
			globalrole.Rules = append(globalrole.Rules, rule)
		}
	}
	// Update templateNames for adding the new template`s name
	globalrole.AggregationRoleTemplates.TemplateNames = append(globalrole.AggregationRoleTemplates.TemplateNames, template.Name)

	return r.Client.Update(ctx, globalrole)
}

func ruleExists(haystack []rbacv1.PolicyRule, needle rbacv1.PolicyRule) bool {
	for _, curr := range haystack {
		if equality.Semantic.DeepEqual(curr, needle) {
			return true
		}
	}
	return false
}

func isContainsLabels(haystack, needle map[string]string) bool {
	var count int
	for key, val := range needle {
		if haystack[key] == val {
			count += 1
		}
	}
	return count == len(needle)
}
