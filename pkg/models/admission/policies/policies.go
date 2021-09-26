package policies

import (
	"context"
	"kubesphere.io/kubesphere/pkg/api/admission/v1alpha1"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/admission/provider"
)

type PolicyManagerInterface interface {
	// GetPolicy gets the admission policy with the given name.
	GetPolicy(ctx context.Context, namespace, policyName string) (*v1alpha1.PolicyDetail, error)
	// ListPolicies lists the admission policies with the given name.
	ListPolicies(ctx context.Context, namespace string) (*v1alpha1.PolicyList, error)
	// CreatePolicy creates an admission policy.
	CreatePolicy(ctx context.Context, namespace string, policy *v1alpha1.PostPolicy) error
	// UpdatePolicy updates the admission policy with the given name.
	UpdatePolicy(ctx context.Context, namespace, policyName string, policy *v1alpha1.PostPolicy) error
	// DeletePolicy deletes the admission policy with the given name.
	DeletePolicy(ctx context.Context, namespace, policyName string) error
}

type PolicyManager struct {
	ksClient    kubesphere.Interface
	ksInformers externalversions.SharedInformerFactory
	Providers   map[string]provider.Provider
}

func NewPolicyManager(ksClient kubesphere.Interface, ksInformers externalversions.SharedInformerFactory, providers map[string]provider.Provider) *PolicyManager {
	return &PolicyManager{ksClient: ksClient, ksInformers: ksInformers, Providers: providers}
}

func (p PolicyManager) GetPolicy(ctx context.Context, namespace, policyName string) (*v1alpha1.PolicyDetail, error) {
	panic("implement me")
}

func (p PolicyManager) ListPolicies(ctx context.Context, namespace string) (*v1alpha1.PolicyList, error) {
	panic("implement me")
}

func (p PolicyManager) CreatePolicy(ctx context.Context, namespace string, policy *v1alpha1.PostPolicy) error {
	panic("implement me")
}

func (p PolicyManager) UpdatePolicy(ctx context.Context, namespace, policyName string, policy *v1alpha1.PostPolicy) error {
	panic("implement me")
}

func (p PolicyManager) DeletePolicy(ctx context.Context, namespace, policyName string) error {
	panic("implement me")
}
