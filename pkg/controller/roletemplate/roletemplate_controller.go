package roletemplate

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	rbachelper "kubesphere.io/kubesphere/pkg/conponenthelper/auth/rbac"
	controllerutils "kubesphere.io/kubesphere/pkg/controller/utils/controller"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	autoAggregateIndexKey = ".metadata.annotations[iam.kubesphere.io/auto-aggregate]"

	autoAggregationLabel = "iam.kubesphere.io/auto-aggregate"

	controllerName = "roletemplate-controller"

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
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err = r.injectRoleTemplateToRuleOwner(ctx, roletemplate)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RoleTemplateReconciler) injectRoleTemplateToRuleOwner(ctx context.Context, roletemplate *iamv1beta1.RoleTemplate) error {
	if _, exist := roletemplate.Labels[rbachelper.LabelGlobalScope]; exist {
		if err := r.aggregateGlobalRoles(ctx, roletemplate); err != nil {
			return err
		}
	}
	if _, exist := roletemplate.Labels[rbachelper.LabelWorkspaceScope]; exist {
		if err := r.aggregateWorkspaceRoles(ctx, roletemplate); err != nil {
			return err
		}
	}
	if _, exist := roletemplate.Labels[rbachelper.LabelClusterScope]; exist {
		if err := r.aggregateClusterRoles(ctx, roletemplate); err != nil {
			return err
		}
	}
	if _, exist := roletemplate.Labels[rbachelper.LabelNamespaceScope]; exist {
		if err := r.aggregateRoles(ctx, roletemplate); err != nil {
			return err
		}
	}
	return nil
}

// aggregateGlobalRoles automatic inject the RoleTemplate`s rules to the role
// (all role types have the field AggregationRoleTemplates) matching the AggregationRoleTemplates filed.
// Note that autoAggregateRoles just aggregate the templates by field ".aggregationRoleTemplates.roleSelectors",
// and if the roleTemplate content is changed, the role including the roleTemplate should not be updated.
func (r *RoleTemplateReconciler) aggregateGlobalRoles(ctx context.Context, roletemplate *iamv1beta1.RoleTemplate) error {
	list := &iamv1beta1.GlobalRoleList{}
	l := map[string]string{autoAggregateIndexKey: "true"}
	err := r.Cache.List(ctx, list, client.MatchingFields(l))
	if err != nil {
		return err
	}

	for _, role := range list.Items {
		err := r.aggregate(ctx, rbachelper.GlobalRoleRuleOwner{GlobalRole: &role}, roletemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleTemplateReconciler) aggregateWorkspaceRoles(ctx context.Context, roletemplate *iamv1beta1.RoleTemplate) error {
	list := &iamv1beta1.WorkspaceRoleList{}
	l := map[string]string{autoAggregateIndexKey: "true"}
	err := r.Cache.List(ctx, list, client.MatchingFields(l))
	if err != nil {
		return err
	}

	for _, role := range list.Items {
		err := r.aggregate(ctx, rbachelper.WorkspaceRoleRuleOwner{WorkspaceRole: &role}, roletemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleTemplateReconciler) aggregateClusterRoles(ctx context.Context, roletemplate *iamv1beta1.RoleTemplate) error {
	list := &iamv1beta1.ClusterRoleList{}
	l := map[string]string{autoAggregateIndexKey: "true"}
	err := r.Cache.List(ctx, list, client.MatchingFields(l))
	if err != nil {
		return err
	}

	for _, role := range list.Items {
		err := r.aggregate(ctx, rbachelper.ClusterRoleRuleOwner{ClusterRole: &role}, roletemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleTemplateReconciler) aggregateRoles(ctx context.Context, roletemplate *iamv1beta1.RoleTemplate) error {
	list := &iamv1beta1.RoleList{}
	l := map[string]string{autoAggregateIndexKey: "true"}
	err := r.Cache.List(ctx, list, client.MatchingFields(l))
	if err != nil {
		return err
	}

	for _, role := range list.Items {
		err := r.aggregate(ctx, rbachelper.RoleRuleOwner{Role: &role}, roletemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

// aggregate the roletemplate rules to the ruleOwner. If the roletemplate is updated but has already been aggregated by the ruleOwner,
// the ruleOwner cannot update the new roletemplate rule to the ruleOwner.
func (r *RoleTemplateReconciler) aggregate(ctx context.Context, ruleOwner rbachelper.RuleOwner, roletemplate *iamv1beta1.RoleTemplate) error {
	aggregation := ruleOwner.GetAggregationRule()
	hasTemplateName := sliceutil.HasString(aggregation.TemplateNames, roletemplate.Name)
	if aggregation != nil && !hasTemplateName {
		if isContainsLabels(roletemplate.Labels, aggregation.RoleSelector.MatchLabels) {
			cover, _ := rbachelper.Covers(ruleOwner.GetRules(), roletemplate.Spec.Rules)
			if cover && hasTemplateName {
				return nil
			}

			if !cover {
				ruleOwner.SetRules(append(ruleOwner.GetRules(), roletemplate.Spec.Rules...))
			}

			if !hasTemplateName {
				aggregation.TemplateNames = append(aggregation.TemplateNames, roletemplate.Name)
				ruleOwner.SetAggregationRule(aggregation)
			}

			err := r.Client.Update(ctx, ruleOwner.GetObject().(client.Object))
			if err != nil {
				r.recorder.Event(ruleOwner.GetObject(), corev1.EventTypeWarning, reasonFailedSync, err.Error())
				return err
			}
			r.recorder.Event(ruleOwner.GetObject(), corev1.EventTypeNormal, controllerutils.SuccessSynced, messageResourceSynced)
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
	err = r.Cache.IndexField(context.Background(), &iamv1beta1.WorkspaceRole{}, autoAggregateIndexKey, workspaceRoleIndexByAnnotation)
	if err != nil {
		return err
	}
	err = r.Cache.IndexField(context.Background(), &iamv1beta1.ClusterRole{}, autoAggregateIndexKey, clusterRoleIndexByAnnotation)
	if err != nil {
		return err
	}
	err = r.Cache.IndexField(context.Background(), &iamv1beta1.Role{}, autoAggregateIndexKey, roleIndexByAnnotation)
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

func workspaceRoleIndexByAnnotation(obj client.Object) []string {
	role := obj.(*iamv1beta1.WorkspaceRole)
	if val, ok := role.Annotations[autoAggregationLabel]; ok {
		return []string{val}
	}
	return []string{}
}

func clusterRoleIndexByAnnotation(obj client.Object) []string {
	role := obj.(*iamv1beta1.ClusterRole)
	if val, ok := role.Annotations[autoAggregationLabel]; ok {
		return []string{val}
	}
	return []string{}
}

func roleIndexByAnnotation(obj client.Object) []string {
	role := obj.(*iamv1beta1.Role)
	if val, ok := role.Annotations[autoAggregationLabel]; ok {
		return []string{val}
	}
	return []string{}
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
