/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package rbac

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
)

type GlobalRoleRuleOwner struct {
	GlobalRole *iamv1beta1.GlobalRole
}

func (g GlobalRoleRuleOwner) GetRuleOwnerScope() string {
	return iamv1beta1.ScopeGlobal
}

func (g GlobalRoleRuleOwner) GetObject() runtime.Object {
	return g.GlobalRole
}

func (g GlobalRoleRuleOwner) GetNamespace() string {
	return ""
}

func (g GlobalRoleRuleOwner) GetName() string {
	return g.GlobalRole.Name
}

func (g GlobalRoleRuleOwner) GetLabels() map[string]string {
	return g.GlobalRole.Labels
}

func (g GlobalRoleRuleOwner) SetLabels(m map[string]string) {
	g.GlobalRole.Labels = m
}

func (g GlobalRoleRuleOwner) GetAnnotations() map[string]string {
	return g.GlobalRole.Annotations
}

func (g GlobalRoleRuleOwner) SetAnnotations(m map[string]string) {
	g.GlobalRole.Annotations = m
}

func (g GlobalRoleRuleOwner) GetRules() []rbacv1.PolicyRule {
	return g.GlobalRole.Rules
}

func (g GlobalRoleRuleOwner) SetRules(rules []rbacv1.PolicyRule) {
	g.GlobalRole.Rules = rules
}

func (g GlobalRoleRuleOwner) GetAggregationRule() *iamv1beta1.AggregationRoleTemplates {
	return g.GlobalRole.AggregationRoleTemplates
}

func (g GlobalRoleRuleOwner) SetAggregationRule(aggregationRoleTemplates *iamv1beta1.AggregationRoleTemplates) {
	g.GlobalRole.AggregationRoleTemplates = aggregationRoleTemplates
}

func (g GlobalRoleRuleOwner) DeepCopyRuleOwner() RuleOwner {
	return GlobalRoleRuleOwner{GlobalRole: g.GlobalRole.DeepCopy()}
}

func (g GlobalRoleRuleOwner) GetRegoPolicy() string {
	if g.GlobalRole.ObjectMeta.Annotations != nil {
		return g.GlobalRole.ObjectMeta.Annotations[iamv1beta1.RegoOverrideAnnotation]
	}
	return ""
}

func (g GlobalRoleRuleOwner) SetRegoPolicy(rego string) {
	if g.GlobalRole.ObjectMeta.Annotations == nil {
		g.GlobalRole.ObjectMeta.Annotations = make(map[string]string)
	}
	g.GlobalRole.ObjectMeta.Annotations[iamv1beta1.RegoOverrideAnnotation] = rego
}

var _ RuleOwner = GlobalRoleRuleOwner{}
