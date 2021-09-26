package rules

import (
	"context"
	admission "kubesphere.io/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/api/admission/v1alpha1"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/admission/provider"
)

type RuleManagerInterface interface {
	// GetRule gets the admission rule for the policy.
	GetRule(ctx context.Context, namespace, policyName, ruleName string) (*v1alpha1.RuleDetail, error)
	// ListRules lists the admission rules from the given policy.
	ListRules(ctx context.Context, namespace, policyName string) (*v1alpha1.RuleList, error)
	// CreateRule creates an admission rule for the policy.
	CreateRule(ctx context.Context, namespace string, policyName, rule *v1alpha1.PostRule) error
	// UpdateRule updates the admission rule for the policy with the given name.
	UpdateRule(ctx context.Context, namespace, policyName, ruleName string, rule *v1alpha1.PostRule) error
	// DeleteRule deletes the admission rule for the policy with the given name.
	DeleteRule(ctx context.Context, namespace, policyName, ruleName string) error
}

type RuleManager struct {
	ksClient    kubesphere.Interface
	ksInformers externalversions.SharedInformerFactory
	Providers   map[string]provider.Provider
}

func NewRuleManager(ksClient kubesphere.Interface, ksInformers externalversions.SharedInformerFactory, providers map[string]provider.Provider) *RuleManager {
	return &RuleManager{ksClient: ksClient, ksInformers: ksInformers, Providers: providers}
}

func (r RuleManager) GetRule(ctx context.Context, namespace, policyName, ruleName string) (*v1alpha1.RuleDetail, error) {
	panic("implement me")
}

func (r RuleManager) ListRules(ctx context.Context, namespace, policyName string) (*v1alpha1.RuleList, error) {
	panic("implement me")
}

func (r RuleManager) CreateRule(ctx context.Context, namespace string, policyName, rule *v1alpha1.PostRule) error {
	panic("implement me")
}

func (r RuleManager) UpdateRule(ctx context.Context, namespace, policyName, ruleName string, rule *v1alpha1.PostRule) error {
	panic("implement me")
}

func (r RuleManager) DeleteRule(ctx context.Context, namespace, policyName, ruleName string) error {
	panic("implement me")
}

func Rule(rule *v1alpha1.PostRule) *admission.Rule {
	return nil
}
