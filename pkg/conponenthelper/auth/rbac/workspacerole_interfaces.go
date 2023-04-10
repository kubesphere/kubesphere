package rbac

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
)

type WorkspaceRoleRuleOwner struct {
	WorkspaceRole *iamv1beta1.WorkspaceRole
}

func (w WorkspaceRoleRuleOwner) RuleOwnerScopeKey() string {
	return LabelWorkspaceScope
}

func (w WorkspaceRoleRuleOwner) GetObject() runtime.Object {
	return w.WorkspaceRole
}

func (w WorkspaceRoleRuleOwner) GetNamespace() string {
	return ""
}

func (w WorkspaceRoleRuleOwner) GetName() string {
	return w.WorkspaceRole.Name
}

func (w WorkspaceRoleRuleOwner) GetLabels() map[string]string {
	return w.WorkspaceRole.Labels
}

func (w WorkspaceRoleRuleOwner) SetLabels(m map[string]string) {
	w.WorkspaceRole.Labels = m
}

func (w WorkspaceRoleRuleOwner) GetAnnotations() map[string]string {
	return w.WorkspaceRole.Annotations
}

func (w WorkspaceRoleRuleOwner) SetAnnotations(m map[string]string) {
	w.WorkspaceRole.Annotations = m
}

func (w WorkspaceRoleRuleOwner) GetRules() []rbacv1.PolicyRule {
	return w.WorkspaceRole.Rules
}

func (w WorkspaceRoleRuleOwner) SetRules(rules []rbacv1.PolicyRule) {
	w.WorkspaceRole.Rules = rules
}

func (w WorkspaceRoleRuleOwner) GetAggregationRule() *iamv1beta1.AggregationRoleTemplates {
	return w.WorkspaceRole.AggregationRoleTemplates
}

func (w WorkspaceRoleRuleOwner) SetAggregationRule(i *iamv1beta1.AggregationRoleTemplates) {
	w.WorkspaceRole.AggregationRoleTemplates = i
}

func (w WorkspaceRoleRuleOwner) DeepCopyRuleOwner() RuleOwner {
	return WorkspaceRoleRuleOwner{WorkspaceRole: w.WorkspaceRole.DeepCopy()}
}

var _ RuleOwner = WorkspaceRoleRuleOwner{}
