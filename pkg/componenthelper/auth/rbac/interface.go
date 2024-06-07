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

type RuleOwner interface {
	GetObject() runtime.Object
	GetNamespace() string
	GetName() string
	GetLabels() map[string]string
	SetLabels(map[string]string)
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
	GetRules() []rbacv1.PolicyRule
	SetRules([]rbacv1.PolicyRule)
	GetRegoPolicy() string
	SetRegoPolicy(string)
	GetAggregationRule() *iamv1beta1.AggregationRoleTemplates
	SetAggregationRule(*iamv1beta1.AggregationRoleTemplates)
	DeepCopyRuleOwner() RuleOwner
	GetRuleOwnerScope() string
}
