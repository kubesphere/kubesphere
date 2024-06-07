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

type ClusterRoleRuleOwner struct {
	ClusterRole *iamv1beta1.ClusterRole
}

func (c ClusterRoleRuleOwner) GetRuleOwnerScope() string {
	return iamv1beta1.ScopeCluster
}

func (c ClusterRoleRuleOwner) GetObject() runtime.Object {
	return c.ClusterRole
}

func (c ClusterRoleRuleOwner) GetNamespace() string {
	return c.ClusterRole.Namespace
}

func (c ClusterRoleRuleOwner) GetName() string {
	return c.ClusterRole.Name
}

func (c ClusterRoleRuleOwner) GetLabels() map[string]string {
	return c.ClusterRole.Labels
}

func (c ClusterRoleRuleOwner) SetLabels(label map[string]string) {
	c.ClusterRole.Labels = label
}

func (c ClusterRoleRuleOwner) GetAnnotations() map[string]string {
	return c.ClusterRole.Annotations
}

func (c ClusterRoleRuleOwner) SetAnnotations(annotation map[string]string) {
	c.ClusterRole.Annotations = annotation
}

func (c ClusterRoleRuleOwner) GetRules() []rbacv1.PolicyRule {
	return c.ClusterRole.Rules
}

func (c ClusterRoleRuleOwner) SetRules(rules []rbacv1.PolicyRule) {
	c.ClusterRole.Rules = rules
}

func (c ClusterRoleRuleOwner) GetAggregationRule() *iamv1beta1.AggregationRoleTemplates {
	return c.ClusterRole.AggregationRoleTemplates
}

func (c ClusterRoleRuleOwner) SetAggregationRule(aggregationRoleTemplates *iamv1beta1.AggregationRoleTemplates) {
	c.ClusterRole.AggregationRoleTemplates = aggregationRoleTemplates
}

func (c ClusterRoleRuleOwner) GetRegoPolicy() string {
	if c.ClusterRole.ObjectMeta.Annotations != nil {
		return c.ClusterRole.ObjectMeta.Annotations[iamv1beta1.RegoOverrideAnnotation]
	}
	return ""
}

func (c ClusterRoleRuleOwner) SetRegoPolicy(rego string) {
	if c.ClusterRole.ObjectMeta.Annotations == nil {
		c.ClusterRole.ObjectMeta.Annotations = make(map[string]string)
	}
	c.ClusterRole.ObjectMeta.Annotations[iamv1beta1.RegoOverrideAnnotation] = rego
}

func (c ClusterRoleRuleOwner) DeepCopyRuleOwner() RuleOwner {
	return ClusterRoleRuleOwner{ClusterRole: c.ClusterRole.DeepCopy()}
}

var _ RuleOwner = ClusterRoleRuleOwner{}
