package rbac

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
)

type RoleRuleOwner struct {
	Role *iamv1beta1.Role
}

func (r RoleRuleOwner) RuleOwnerScopeKey() string {
	return LabelNamespaceScope
}

func (r RoleRuleOwner) GetObject() runtime.Object {
	return r.Role
}

func (r RoleRuleOwner) GetNamespace() string {
	return r.Role.Namespace
}

func (r RoleRuleOwner) GetName() string {
	return r.Role.Name
}

func (r RoleRuleOwner) GetLabels() map[string]string {
	return r.Role.Labels
}

func (r RoleRuleOwner) SetLabels(m map[string]string) {
	r.Role.Labels = m
}

func (r RoleRuleOwner) GetAnnotations() map[string]string {
	return r.Role.Annotations
}

func (r RoleRuleOwner) SetAnnotations(m map[string]string) {
	r.Role.Annotations = m
}

func (r RoleRuleOwner) GetRules() []rbacv1.PolicyRule {
	return r.Role.Rules
}

func (r RoleRuleOwner) SetRules(rules []rbacv1.PolicyRule) {
	r.Role.Rules = rules
}

func (r RoleRuleOwner) GetAggregationRule() *iamv1beta1.AggregationRoleTemplates {
	return r.Role.AggregationRoleTemplates
}

func (r RoleRuleOwner) SetAggregationRule(i *iamv1beta1.AggregationRoleTemplates) {
	r.Role.AggregationRoleTemplates = i
}

func (r RoleRuleOwner) DeepCopyRuleOwner() RuleOwner {
	return RoleRuleOwner{Role: r.Role.DeepCopy()}
}

var _ RuleOwner = RoleRuleOwner{}
