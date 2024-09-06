/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package roletemplate

import (
	"context"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rbachelper "kubesphere.io/kubesphere/pkg/componenthelper/auth/rbac"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	autoAggregateIndexKey = ".metadata.annotations[iam.kubesphere.io/auto-aggregate]"
	autoAggregationLabel  = "iam.kubesphere.io/auto-aggregate"
	controllerName        = "roletemplate"
	reasonFailedSync      = "FailedInjectRoleTemplate"
	messageResourceSynced = "RoleTemplate injected successfully"
)

var _ kscontroller.Controller = &Reconciler{}

// Reconciler reconciles a RoleTemplate object
type Reconciler struct {
	client.Client
	recorder record.EventRecorder
	logger   klog.Logger
}

func (r *Reconciler) Name() string {
	return controllerName
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.logger = mgr.GetLogger().WithName(controllerName)
	r.Client = mgr.GetClient()

	if err := mgr.GetCache().IndexField(context.Background(), &iamv1beta1.GlobalRole{}, autoAggregateIndexKey, globalRoleIndexByAnnotation); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.Background(), &iamv1beta1.WorkspaceRole{}, autoAggregateIndexKey, workspaceRoleIndexByAnnotation); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.Background(), &iamv1beta1.ClusterRole{}, autoAggregateIndexKey, clusterRoleIndexByAnnotation); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.Background(), &iamv1beta1.Role{}, autoAggregateIndexKey, roleIndexByAnnotation); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&iamv1beta1.RoleTemplate{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	roleTemplate := &iamv1beta1.RoleTemplate{}

	if err := r.Get(ctx, req.NamespacedName, roleTemplate); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.injectRoleTemplateToRuleOwner(ctx, roleTemplate); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) injectRoleTemplateToRuleOwner(ctx context.Context, roleTemplate *iamv1beta1.RoleTemplate) error {
	if roleTemplate.Labels[iamv1beta1.ScopeLabel] == iamv1beta1.ScopeGlobal {
		if err := r.aggregateGlobalRoles(ctx, roleTemplate); err != nil {
			return err
		}
	}
	if roleTemplate.Labels[iamv1beta1.ScopeLabel] == iamv1beta1.ScopeWorkspace {
		if err := r.aggregateWorkspaceRoles(ctx, roleTemplate); err != nil {
			return err
		}
	}
	if roleTemplate.Labels[iamv1beta1.ScopeLabel] == iamv1beta1.ScopeCluster {
		if err := r.aggregateClusterRoles(ctx, roleTemplate); err != nil {
			return err
		}
	}
	if roleTemplate.Labels[iamv1beta1.ScopeLabel] == iamv1beta1.ScopeNamespace {
		if err := r.aggregateRoles(ctx, roleTemplate); err != nil {
			return err
		}
	}
	return nil
}

// aggregateGlobalRoles automatic inject the RoleTemplate`s rules to the role
// (all role types have the field AggregationRoleTemplates) matching the AggregationRoleTemplates filed.
// Note that autoAggregateRoles just aggregate the templates by field ".aggregationRoleTemplates.roleSelectors",
// and if the roleTemplate content is changed, the role including the roleTemplate should not be updated.
func (r *Reconciler) aggregateGlobalRoles(ctx context.Context, roleTemplate *iamv1beta1.RoleTemplate) error {
	list := &iamv1beta1.GlobalRoleList{}
	if err := r.List(ctx, list, client.MatchingFields(fields.Set{autoAggregateIndexKey: "true"})); err != nil {
		return err
	}
	for _, role := range list.Items {
		err := r.aggregate(ctx, rbachelper.GlobalRoleRuleOwner{GlobalRole: &role}, roleTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) aggregateWorkspaceRoles(ctx context.Context, roleTemplate *iamv1beta1.RoleTemplate) error {
	list := &iamv1beta1.WorkspaceRoleList{}
	if err := r.List(ctx, list, client.MatchingFields(fields.Set{autoAggregateIndexKey: "true"})); err != nil {
		return err
	}

	for _, role := range list.Items {
		err := r.aggregate(ctx, rbachelper.WorkspaceRoleRuleOwner{WorkspaceRole: &role}, roleTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) aggregateClusterRoles(ctx context.Context, roleTemplate *iamv1beta1.RoleTemplate) error {
	list := &iamv1beta1.ClusterRoleList{}
	if err := r.List(ctx, list, client.MatchingFields(fields.Set{autoAggregateIndexKey: "true"})); err != nil {
		return err
	}

	for _, role := range list.Items {
		err := r.aggregate(ctx, rbachelper.ClusterRoleRuleOwner{ClusterRole: &role}, roleTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) aggregateRoles(ctx context.Context, roleTemplate *iamv1beta1.RoleTemplate) error {
	list := &iamv1beta1.RoleList{}
	if err := r.List(ctx, list, client.MatchingFields(fields.Set{autoAggregateIndexKey: "true"})); err != nil {
		return err
	}
	for _, role := range list.Items {
		err := r.aggregate(ctx, rbachelper.RoleRuleOwner{Role: &role}, roleTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

// aggregate the role-template rules to the ruleOwner. If the role-template is updated but has already been aggregated by the ruleOwner,
// the ruleOwner cannot update the new role-template rule to the ruleOwner.
func (r *Reconciler) aggregate(ctx context.Context, ruleOwner rbachelper.RuleOwner, roleTemplate *iamv1beta1.RoleTemplate) error {
	aggregation := ruleOwner.GetAggregationRule()
	if aggregation == nil {
		return nil
	}
	hasTemplateName := sliceutil.HasString(aggregation.TemplateNames, roleTemplate.Name)
	if hasTemplateName {
		return nil
	}
	selector, err := metav1.LabelSelectorAsSelector(aggregation.RoleSelector)
	if err != nil {
		r.logger.V(4).Error(err, "failed to pares role selector", "template", ruleOwner.GetName())
		return nil
	}
	if !selector.Matches(labels.Set(roleTemplate.Labels)) {
		return nil
	}
	cover, _ := rbachelper.Covers(ruleOwner.GetRules(), roleTemplate.Spec.Rules)
	if cover && hasTemplateName {
		return nil
	}

	if !cover {
		ruleOwner.SetRules(append(ruleOwner.GetRules(), roleTemplate.Spec.Rules...))
	}

	aggregation.TemplateNames = append(aggregation.TemplateNames, roleTemplate.Name)
	ruleOwner.SetAggregationRule(aggregation)

	if err := r.Update(ctx, ruleOwner.GetObject().(client.Object)); err != nil {
		r.recorder.Event(ruleOwner.GetObject(), corev1.EventTypeWarning, reasonFailedSync, err.Error())
		return err
	}

	r.recorder.Event(ruleOwner.GetObject(), corev1.EventTypeNormal, "Synced", messageResourceSynced)
	return nil
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
