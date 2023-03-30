package iam

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
)

const (
	LabelGlobalScope    = "scope.iam.kubesphere.io/global"
	LabelWorkspaceScope = "scope.iam.kubesphere.io/workspace"
	LabelClusterScope   = "scope.iam.kubesphere.io/cluster"
	LabelNamespaceScope = "scope.iam.kubesphere.io/namespace"

	AggregateRoleTemplateFailed = "AggregateRoleTemplateFailed"
	MessageResourceSynced       = "Aggregating roleTemplates successfully"
)

type Helper struct {
	client.Client
}

func NewHelper(c client.Client) *Helper {
	return &Helper{c}
}

func (h *Helper) GetAggregationRoleTemplateRule(ctx context.Context, scopeKey string, templates *iamv1beta1.AggregationRoleTemplates) ([]rbacv1.PolicyRule, []string, error) {
	rules := make([]rbacv1.PolicyRule, 0)
	newTemplateNames := make([]string, 0)
	if len(templates.RoleSelectors) == 0 {
		for _, name := range templates.TemplateNames {
			roleTemplate := &iamv1beta1.RoleTemplate{}
			err := h.Get(ctx, types.NamespacedName{Name: name}, roleTemplate)
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
			if err = h.List(ctx, roleTemplateList, &client.ListOptions{LabelSelector: asSelector}); err != nil {
				return nil, nil, err
			}

			for _, roleTemplate := range roleTemplateList.Items {
				newTemplateNames = append(newTemplateNames, roleTemplate.Name)
				for _, rule := range roleTemplate.Spec.Rules {
					if !RuleExists(rules, rule) {
						rules = append(rules, rule)
					}
				}
			}
		}
	}
	return rules, newTemplateNames, nil
}

func RuleExists(haystack []rbacv1.PolicyRule, needle rbacv1.PolicyRule) bool {
	for _, curr := range haystack {
		if equality.Semantic.DeepEqual(curr, needle) {
			return true
		}
	}
	return false
}
