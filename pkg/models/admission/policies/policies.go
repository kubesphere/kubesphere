package policies

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	admission "kubesphere.io/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/admission/polciyTemplates"
	"kubesphere.io/kubesphere/pkg/models/admission/provider"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/admission/policies"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/admission/policytemplates"
	"strings"
)

type PolicyManagerInterface interface {
	// GetPolicy gets the admission policy with the given name.
	GetPolicy(ctx context.Context, policyName string) (*v1alpha1.PolicyDetail, error)
	// ListPolicies lists the admission policies with the given name.
	ListPolicies(ctx context.Context, query *query.Query) (*v1alpha1.PolicyList, error)
	// CreatePolicy creates an admission policy.
	CreatePolicy(ctx context.Context, policy *v1alpha1.PostPolicy) error
	// UpdatePolicy updates the admission policy with the given name.
	UpdatePolicy(ctx context.Context, policyName string, policy *v1alpha1.PostPolicy) error
	// DeletePolicy deletes the admission policy with the given name.
	DeletePolicy(ctx context.Context, policyName string) error
}

type PolicyManager struct {
	ksClient       kubesphere.Interface
	getter         resources.Interface
	templateGetter resources.Interface
	providers      map[string]provider.Provider
}

func NewPolicyManager(ksClient kubesphere.Interface, ksInformers externalversions.SharedInformerFactory, providers map[string]provider.Provider) *PolicyManager {
	return &PolicyManager{ksClient: ksClient, getter: policies.New(ksInformers), templateGetter: policytemplates.New(ksInformers), providers: providers}
}

func (m *PolicyManager) GetPolicy(_ context.Context, policyName string) (*v1alpha1.PolicyDetail, error) {
	obj, err := m.getter.Get("", PolicyUniqueName(policyName))
	if err != nil {
		return nil, err
	}
	policy := obj.(*admission.Policy).DeepCopy()
	if policy == nil {
		return nil, v1alpha1.ErrPolicyNotFound
	}
	return PolicyDetail(policy), nil
}

func (m *PolicyManager) ListPolicies(_ context.Context, query *query.Query) (*v1alpha1.PolicyList, error) {
	objs, err := m.getter.List("", query)
	if err != nil {
		return nil, err
	}
	list := &v1alpha1.PolicyList{
		Total: objs.TotalItems,
	}
	for _, obj := range objs.Items {
		policy := obj.(*admission.Policy).DeepCopy()
		list.Items = append(list.Items, Policy(policy))
	}
	return list, err
}

func (m *PolicyManager) CreatePolicy(ctx context.Context, policy *v1alpha1.PostPolicy) error {
	obj, err := m.getter.Get("", PolicyUniqueName(policy.Name))
	if err != nil {
		return err
	}
	nowPolicy := obj.(*admission.Policy)
	if nowPolicy != nil {
		return v1alpha1.ErrPolicyAlreadyExists
	}
	p, ok := m.providers[policy.Provider]
	if !ok {
		return v1alpha1.ErrProviderNotFound
	}

	var newPolicy *admission.Policy
	if policy.PolicyTemplate != "" && len(policy.Targets) == 0 {
		obj, err := m.templateGetter.Get("", polciyTemplates.PolicyTemplateUniqueName(policy.PolicyTemplate))
		if err != nil {
			return err
		}
		template := obj.(*admission.PolicyTemplate).DeepCopy()
		if template == nil {
			return v1alpha1.ErrPolicyTemplateNotFound
		}
		newPolicy, err = NewPolicyFromTemplate(template, policy.Name, policy.Provider, policy.Description, admission.PolicyInactive)
		if err != nil {
			return err
		}
	} else {
		newPolicy = NewPolicy(policy, admission.PolicyInactive)
	}

	err = p.AddPolicy(ctx, newPolicy)
	if err != nil {
		return err
	}
	_, err = m.ksClient.AdmissionV1alpha1().Policies().Create(ctx, newPolicy, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *PolicyManager) UpdatePolicy(ctx context.Context, policyName string, policy *v1alpha1.PostPolicy) error {
	obj, err := m.getter.Get("", PolicyUniqueName(policyName))
	if err != nil {
		return err
	}
	nowPolicy := obj.(*admission.Policy)
	if nowPolicy == nil {
		return v1alpha1.ErrPolicyNotFound
	}
	p, ok := m.providers[nowPolicy.Spec.Provider]
	if !ok {
		return v1alpha1.ErrProviderNotFound
	}
	newPolicy := NewPolicy(policy, nowPolicy.Status.State)
	err = p.RemovePolicy(ctx, nowPolicy)
	if err != nil {
		return err
	}
	err = p.AddPolicy(ctx, newPolicy)
	if err != nil {
		return err
	}
	_, err = m.ksClient.AdmissionV1alpha1().Policies().Update(ctx, newPolicy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *PolicyManager) DeletePolicy(ctx context.Context, policyName string) error {
	obj, err := m.getter.Get("", PolicyUniqueName(policyName))
	if err != nil {
		return err
	}
	policy := obj.(*admission.Policy)
	if policy == nil {
		return v1alpha1.ErrPolicyNotFound
	}
	p, ok := m.providers[policy.Spec.Provider]
	if !ok {
		return v1alpha1.ErrProviderNotFound
	}
	err = p.RemovePolicy(ctx, policy)
	if err != nil {
		return err
	}
	err = m.ksClient.AdmissionV1alpha1().Policies().Delete(ctx, PolicyUniqueName(policyName), metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func PolicyDetail(policy *admission.Policy) *v1alpha1.PolicyDetail {
	detail := &v1alpha1.PolicyDetail{
		Policy: *Policy(policy),
	}
	return detail
}

func Policy(policy *admission.Policy) *v1alpha1.Policy {
	targets := policy.Spec.Content.Targets

	var policyTargets []v1alpha1.PolicyTarget
	for _, target := range targets {
		policyTargets = append(policyTargets, v1alpha1.PolicyTarget{
			Target:     target.Target,
			Expression: target.Expression,
			Import:     target.Import,
		})
	}

	validation := policy.Spec.Content.Spec.Parameters.Validation

	return &v1alpha1.Policy{
		Name:           policy.Spec.Name,
		PolicyTemplate: policy.Spec.PolicyTemplate,
		Provider:       policy.Spec.Provider,
		Description:    policy.Spec.Description,
		Targets:        policyTargets,
		Parameters: v1alpha1.Parameters{
			Validation: &v1alpha1.Validation{
				OpenAPIV3Schema: validation.OpenAPIV3Schema,
				LegacySchema:    validation.LegacySchema,
			},
		},
	}
}

func NewPolicy(postPolicy *v1alpha1.PostPolicy, state admission.PolicyState) *admission.Policy {
	if state == "" {
		state = admission.PolicyInactive
	}

	params := postPolicy.Parameters
	targets := postPolicy.Targets

	var policyTargets []admission.PolicyContentTarget
	for _, target := range targets {
		policyTargets = append(policyTargets, admission.PolicyContentTarget{
			Target:     target.Target,
			Expression: target.Expression,
			Import:     target.Import,
		})
	}

	policy := &admission.Policy{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: PolicyUniqueName(postPolicy.Name),
		},
		Spec: admission.PolicySpec{
			Name:           postPolicy.Name,
			PolicyTemplate: postPolicy.PolicyTemplate,
			Description:    postPolicy.Description,
			Provider:       postPolicy.Provider,
			Content: admission.PolicyContent{
				Spec: admission.PolicyContentSpec{
					Names: admission.Names{
						Name: postPolicy.Name,
					},
					Parameters: admission.Parameters{
						Validation: &admission.Validation{
							OpenAPIV3Schema: params.Validation.OpenAPIV3Schema,
							LegacySchema:    params.Validation.LegacySchema,
						},
					},
				},
				Targets: policyTargets,
			},
		},
		Status: admission.PolicyStatus{State: state},
	}
	return policy
}

func NewPolicyFromTemplate(template *admission.PolicyTemplate, name string, provider string, desc string, state admission.PolicyState) (*admission.Policy, error) {
	if state == "" {
		state = admission.PolicyInactive
	}
	if name == "" {
		name = template.Name
	}
	if desc == "" {
		desc = template.Spec.Description
	}
	templateContent := template.Spec.Content
	targets := templateContent.Targets
	var policyTargets []admission.PolicyContentTarget
	for _, target := range targets {
		if target.Provider != provider {
			continue
		}
		policyTargets = append(policyTargets, admission.PolicyContentTarget{
			Target:     target.Target,
			Expression: target.Expression,
			Import:     target.Import,
		})
	}

	if len(policyTargets) == 0 {
		return nil, v1alpha1.ErrTemplateOfProviderNotSupport
	}

	policy := &admission.Policy{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: PolicyUniqueName(name),
			Labels: map[string]string{
				admission.AdmissionPolicyTemplateLabel: template.Name,
			},
		},
		Spec: admission.PolicySpec{
			Name:           name,
			PolicyTemplate: template.Spec.Name,
			Description:    desc,
			Provider:       provider,
			Content: admission.PolicyContent{
				Spec: admission.PolicyContentSpec{
					Names:      templateContent.Spec.Names,
					Parameters: templateContent.Spec.Parameters,
				},
				Targets: policyTargets,
			},
		},
		Status: admission.PolicyStatus{State: state},
	}
	return policy, nil
}

func PolicyUniqueName(policyName string) string {
	return strings.ToLower(policyName)
}
