/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package rbac

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/open-policy-agent/opa/ast"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const defaultRegoFileName = "authz.rego"

const (
	AggregateRoleTemplateFailed = "AggregateRoleTemplateFailed"
	MessageResourceSynced       = "Aggregating roleTemplates successfully"
)

type Helper struct {
	client.Client
}

func NewHelper(c client.Client) *Helper {
	return &Helper{c}
}

func (h *Helper) aggregateRoleTemplateRule(roleTemplates []iamv1beta1.RoleTemplate) ([]rbacv1.PolicyRule, []string, error) {
	rules := []rbacv1.PolicyRule{}
	newTemplateNames := []string{}
	for _, rt := range roleTemplates {
		newTemplateNames = append(newTemplateNames, rt.Name)
		for _, rule := range rt.Spec.Rules {
			if !ruleExists(rules, rule) {
				rules = append(rules, rule)
			}
		}
	}

	return rules, newTemplateNames, nil
}

func (h *Helper) aggregateRoleTemplateRegoPolicy(roleTemplates []iamv1beta1.RoleTemplate) (string, error) {
	mergedPolicy := &ast.Module{
		Rules: make([]*ast.Rule, 0),
	}
	for _, rt := range roleTemplates {
		rawPolicy := rt.Annotations[iamv1beta1.RegoOverrideAnnotation]
		if rawPolicy == "" {
			continue
		}
		module, err := ast.ParseModule(defaultRegoFileName, rawPolicy)
		if err != nil {
			return "", err
		}
		if mergedPolicy.Package == nil {
			mergedPolicy.Package = module.Package
		}
		if module != nil {
			mergedPolicy.Rules = append(mergedPolicy.Rules, module.Rules...)
		}
	}

	if len(mergedPolicy.Rules) == 0 {
		return "", nil
	}

	seenRules := make(map[string]struct{})

	uniqueMergedPolicy := &ast.Module{
		Package: mergedPolicy.Package,
		Rules:   make([]*ast.Rule, 0),
	}

	for _, rule := range mergedPolicy.Rules {
		ruleString := rule.String()
		if _, seen := seenRules[ruleString]; !seen {
			uniqueMergedPolicy.Rules = append(uniqueMergedPolicy.Rules, rule)
			seenRules[ruleString] = struct{}{}
		}
	}
	return uniqueMergedPolicy.String(), nil
}

func (h *Helper) getRoleTemplates(ctx context.Context, owner RuleOwner) ([]iamv1beta1.RoleTemplate, error) {
	aggregationRule := owner.GetAggregationRule()
	logger := logr.FromContextOrDiscard(ctx)

	if aggregationRule.RoleSelector == nil {
		roletemplates := []iamv1beta1.RoleTemplate{}
		for _, templateName := range aggregationRule.TemplateNames {
			roleTemplate := &iamv1beta1.RoleTemplate{}
			if err := h.Get(ctx, types.NamespacedName{Name: templateName}, roleTemplate); err != nil {
				if errors.IsNotFound(err) {
					logger.V(4).Info("aggregation role template not found", "name", templateName, "role", owner.GetObject())
					continue
				}
				return nil, err
			}
			roletemplates = append(roletemplates, *roleTemplate)
		}
		return roletemplates, nil
	}

	selector := aggregationRule.RoleSelector.DeepCopy()
	roleTemplateList := &iamv1beta1.RoleTemplateList{}
	// Ensure the roleTemplate can be aggregated at the specific role scope
	selector.MatchLabels = labels.Merge(selector.MatchLabels, map[string]string{iamv1beta1.ScopeLabel: owner.GetRuleOwnerScope()})
	asSelector, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		logger.Error(err, "failed to parse role selector", "scope", owner.GetRuleOwnerScope(), "name", owner.GetName())
		return nil, err
	}
	if err = h.List(ctx, roleTemplateList, &client.ListOptions{LabelSelector: asSelector}); err != nil {
		return nil, err
	}
	return roleTemplateList.Items, nil
}

func (h *Helper) AggregationRole(ctx context.Context, ruleOwner RuleOwner, recorder record.EventRecorder) error {
	var needUpdate bool
	if ruleOwner.GetAggregationRule() == nil {
		return nil
	}
	templates, err := h.getRoleTemplates(ctx, ruleOwner)
	if err != nil {
		recorder.Event(ruleOwner.GetObject(), corev1.EventTypeWarning, AggregateRoleTemplateFailed, err.Error())
		return err
	}
	newPolicyRules, newTemplateNames, err := h.aggregateRoleTemplateRule(templates)
	if err != nil {
		recorder.Event(ruleOwner.GetObject(), corev1.EventTypeWarning, AggregateRoleTemplateFailed, err.Error())
		return err
	}

	cover, uncovered := Covers(ruleOwner.GetRules(), newPolicyRules)

	aggregationRule := ruleOwner.GetAggregationRule()
	templateNamesEqual := false
	if aggregationRule != nil {
		templateNamesEqual = sliceutil.Equal(aggregationRule.TemplateNames, newTemplateNames)
	}

	if !cover {
		needUpdate = true
		newRule := append(ruleOwner.GetRules(), uncovered...)
		squashedRules := SquashRules(len(newRule), newRule)
		ruleOwner.SetRules(squashedRules)
	}

	if !templateNamesEqual {
		needUpdate = true
		aggregationRule.TemplateNames = newTemplateNames
		ruleOwner.SetAggregationRule(aggregationRule)
	}

	newRegoPolicy, err := h.aggregateRoleTemplateRegoPolicy(templates)
	if err != nil {
		recorder.Event(ruleOwner.GetObject(), corev1.EventTypeWarning, AggregateRoleTemplateFailed, err.Error())
		return err
	}

	policyCover, err := regoPolicyCover(ruleOwner.GetRegoPolicy(), newRegoPolicy)
	if err != nil {
		recorder.Event(ruleOwner.GetObject(), corev1.EventTypeWarning, AggregateRoleTemplateFailed, err.Error())
		return err
	}
	if !policyCover {
		needUpdate = true
		ruleOwner.SetRegoPolicy(newRegoPolicy)
	}
	if needUpdate {
		if err = h.Update(ctx, ruleOwner.GetObject().(client.Object)); err != nil {
			recorder.Event(ruleOwner.GetObject(), corev1.EventTypeWarning, AggregateRoleTemplateFailed, err.Error())
			return err
		}
		recorder.Event(ruleOwner.GetObject(), corev1.EventTypeNormal, "Synced", MessageResourceSynced)
	}

	return nil
}

func ruleExists(haystack []rbacv1.PolicyRule, needle rbacv1.PolicyRule) bool {
	covers, _ := Covers(haystack, []rbacv1.PolicyRule{needle})
	return covers
}

func regoPolicyCover(owner, servant string) (bool, error) {
	if servant == "" {
		return true, nil
	}

	if owner == "" && servant != "" {
		return false, nil
	}

	ownerModule, err := ast.ParseModule(defaultRegoFileName, owner)
	if err != nil {
		return false, err
	}

	servantModule, err := ast.ParseModule(defaultRegoFileName, servant)
	if err != nil {
		return false, err
	}
	cover := ownerModule.Compare(servantModule) >= 0

	return cover, nil
}

func SquashRules(deep int, rules []rbacv1.PolicyRule) []rbacv1.PolicyRule {
	var resultRules []rbacv1.PolicyRule
	for _, rule := range rules {
		merged := false
		if cover, _ := Covers(resultRules, []rbacv1.PolicyRule{rule}); cover {
			continue
		}
		for i, rRule := range resultRules {
			if (containRules(rRule.APIGroups, rule.APIGroups) && equalRules(rRule.Resources, rule.Resources)) ||
				(containRules(rRule.APIGroups, rule.APIGroups) && equalRules(rRule.Verbs, rule.Verbs)) {
				merged = true
				resultRules[i] = mergeRules(rRule, rule)
				break
			}
		}

		if !merged {
			resultRules = append(resultRules, rule)
		}
	}

	if len(resultRules) == deep {
		return resultRules
	}
	return SquashRules(len(resultRules), resultRules)
}

func mergeRules(base, rule rbacv1.PolicyRule) rbacv1.PolicyRule {
	if !sliceutil.HasString(base.APIGroups, "*") {
		base.APIGroups = merge(base.APIGroups, rule.APIGroups)
	}
	if !sliceutil.HasString(base.Resources, "*") {
		base.Resources = merge(base.Resources, rule.Resources)
	}
	if !sliceutil.HasString(base.Verbs, "*") {
		base.Verbs = merge(base.Verbs, rule.Verbs)
	}
	return base
}

func merge(base, rule []string) []string {
	for _, r := range rule {
		if !sliceutil.HasString(base, r) {
			base = append(base, r)
		}
	}
	return base
}

func containRules(base, rule []string) bool {
	if sliceutil.HasString(base, "*") {
		return true
	}

	for _, b := range base {
		if !sliceutil.HasString(rule, b) {
			return false
		}
	}
	return true
}

func equalRules(base, rule []string) bool {
	if len(base) != len(rule) {
		return false
	}

	baseMap := make(map[string]int)
	for _, item := range base {
		baseMap[item]++
	}

	for _, item := range rule {
		count, exists := baseMap[item]
		if !exists || count == 0 {
			return false
		}
		baseMap[item]--
	}

	return true
}
