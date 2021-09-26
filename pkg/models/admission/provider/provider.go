package provider

import (
	"context"
	"kubesphere.io/api/admission/v1alpha1"
)

type Provider interface {
	AddPolicy(ctx context.Context, policy *v1alpha1.Policy) error
	RemovePolicy(ctx context.Context, policy *v1alpha1.Policy) error

	AddRule(ctx context.Context, rule *v1alpha1.Rule) error
	RemoveRule(ctx context.Context, rule *v1alpha1.Rule) error
}
