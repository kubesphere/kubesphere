package rules

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	admission "kubesphere.io/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/admission/provider"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/admission/rules"
	"strings"
)

const (
	Kind = "rule.kubesphere.io"
)

type RuleManagerInterface interface {
	// GetRule gets the admission rule for the policy.
	GetRule(ctx context.Context, policyName, ruleName string) (*v1alpha1.RuleDetail, error)
	// ListRules lists the admission rules from the given policy.
	ListRules(ctx context.Context, policyName string, query *query.Query) (*v1alpha1.RuleList, error)
	// CreateRule creates an admission rule for the policy.
	CreateRule(ctx context.Context, policyName string, rule *v1alpha1.PostRule) error
	// UpdateRule updates the admission rule for the policy with the given name.
	UpdateRule(ctx context.Context, policyName, ruleName string, rule *v1alpha1.PostRule) error
	// DeleteRule deletes the admission rule for the policy with the given name.
	DeleteRule(ctx context.Context, policyName, ruleName string) error
}

type RuleManager struct {
	ksClient  kubesphere.Interface
	getter    resources.Interface
	providers map[string]provider.Provider
}

func NewRuleManager(ksClient kubesphere.Interface, ksInformers externalversions.SharedInformerFactory, providers map[string]provider.Provider) *RuleManager {
	return &RuleManager{ksClient: ksClient, getter: rules.New(ksInformers), providers: providers}
}

func (m *RuleManager) GetRule(_ context.Context, policyName, ruleName string) (*v1alpha1.RuleDetail, error) {
	obj, err := m.getter.Get("", RuleUniqueName(policyName, ruleName))
	if err != nil {
		return nil, err
	}
	rule := obj.(*admission.Rule).DeepCopy()
	if rule == nil {
		return nil, v1alpha1.ErrRuleNotFound
	}
	return RuleDetail(rule), nil
}

func (m *RuleManager) ListRules(_ context.Context, policyName string, q *query.Query) (*v1alpha1.RuleList, error) {
	q.Filters[admission.ResourcesSingularPolicy] = query.Value(policyName)
	objs, err := m.getter.List("", q)
	if err != nil {
		return nil, err
	}
	list := &v1alpha1.RuleList{
		Total: objs.TotalItems,
	}
	for _, obj := range objs.Items {
		rule := obj.(*admission.Rule).DeepCopy()
		list.Items = append(list.Items, Rule(rule))
	}
	return list, err
}

func (m *RuleManager) CreateRule(ctx context.Context, policyName string, rule *v1alpha1.PostRule) error {
	obj, err := m.getter.Get("", RuleUniqueName(policyName, rule.Name))
	if err != nil {
		return err
	}
	nowRule := obj.(*admission.Rule)
	if nowRule != nil {
		return v1alpha1.ErrRuleAlreadyExists
	}
	p, ok := m.providers[rule.Provider]
	if !ok {
		return v1alpha1.ErrProviderNotFound
	}
	newRule := NewRule(policyName, rule, admission.RuleInactive)
	err = p.AddRule(ctx, newRule)
	if err != nil {
		return err
	}
	_, err = m.ksClient.AdmissionV1alpha1().Rules().Create(ctx, newRule, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *RuleManager) UpdateRule(ctx context.Context, policyName, ruleName string, rule *v1alpha1.PostRule) error {
	obj, err := m.getter.Get("", RuleUniqueName(policyName, ruleName))
	if err != nil {
		return err
	}
	nowRule := obj.(*admission.Rule)
	if nowRule == nil {
		return v1alpha1.ErrRuleNotFound
	}
	p, ok := m.providers[nowRule.Spec.Provider]
	if !ok {
		return v1alpha1.ErrProviderNotFound
	}
	newRule := NewRule(policyName, rule, nowRule.Status.State)
	err = p.RemoveRule(ctx, nowRule)
	if err != nil {
		return err
	}
	err = p.AddRule(ctx, newRule)
	if err != nil {
		return err
	}
	_, err = m.ksClient.AdmissionV1alpha1().Rules().Update(ctx, newRule, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *RuleManager) DeleteRule(ctx context.Context, policyName, ruleName string) error {
	obj, err := m.getter.Get("", RuleUniqueName(policyName, ruleName))
	if err != nil {
		return err
	}
	rule := obj.(*admission.Rule)
	if rule == nil {
		return v1alpha1.ErrRuleNotFound
	}
	p, ok := m.providers[rule.Spec.Provider]
	if !ok {
		return v1alpha1.ErrProviderNotFound
	}
	err = p.RemoveRule(ctx, rule)
	if err != nil {
		return err
	}
	err = m.ksClient.AdmissionV1alpha1().Policies().Delete(ctx, RuleUniqueName(policyName, ruleName), metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func RuleDetail(rule *admission.Rule) *v1alpha1.RuleDetail {
	detail := &v1alpha1.RuleDetail{
		Rule: *Rule(rule),
	}
	return detail
}

func Rule(rule *admission.Rule) *v1alpha1.Rule {
	obj := make(map[string]interface{})
	err := json.Unmarshal(rule.Spec.Parameters.Raw, &obj)
	if err != nil {
		return nil
	}
	match := v1alpha1.Match{
		Namespaces:         rule.Spec.Match.Namespaces,
		ExcludedNamespaces: rule.Spec.Match.ExcludedNamespaces,
	}
	return &v1alpha1.Rule{
		Name:        rule.Spec.Name,
		Policy:      rule.Spec.Policy,
		Provider:    rule.Spec.Provider,
		Description: rule.Spec.Description,
		Match:       match,
		Parameters:  obj,
	}
}

func NewRule(policyName string, postRule *v1alpha1.PostRule, state admission.RuleState) *admission.Rule {
	postRule.Policy = policyName
	parameters, err := json.Marshal(postRule.Parameters)
	if err != nil {
		klog.Error(err)
		parameters = []byte{}
	}
	if state == "" {
		state = admission.RuleInactive
	}
	rule := &admission.Rule{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: RuleUniqueName(policyName, postRule.Name),
			Labels: map[string]string{
				admission.AdmissionPolicyLabel:   postRule.Policy,
				admission.AdmissionProviderLabel: postRule.Provider,
			},
		},
		Spec: admission.RuleSpec{
			Name:        postRule.Name,
			Policy:      postRule.Policy,
			Provider:    postRule.Provider,
			Description: postRule.Description,
			Match: admission.Match{
				Namespaces:         postRule.Match.Namespaces,
				ExcludedNamespaces: postRule.Match.ExcludedNamespaces,
			},
			Parameters: runtime.RawExtension{
				Raw: parameters,
			},
		},
		Status: admission.RuleStatus{State: state},
	}
	return rule
}

func RuleUniqueName(policyName, ruleName string) string {
	return strings.ToLower(policyName + "_" + ruleName)
}
