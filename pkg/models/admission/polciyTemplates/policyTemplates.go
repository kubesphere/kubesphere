package polciyTemplates

import (
	"context"
	admission "kubesphere.io/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/api/admission/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/admission/policytemplates"
	"strings"
)

type PolicyTemplateManagerInterface interface {
	// GetPolicyTemplate gets the admission policy template with the given name.
	GetPolicyTemplate(ctx context.Context, templateName string) (*v1alpha1.PolicyTemplateDetail, error)
	// ListPolicyTemplates lists the alerts of the custom alerting rule with the given name.
	ListPolicyTemplates(ctx context.Context, query *query.Query) (*v1alpha1.PolicyTemplateList, error)
}

type PolicyTemplateManager struct {
	getter resources.Interface
}

func NewPolicyTemplateManager(ksInformers externalversions.SharedInformerFactory) *PolicyTemplateManager {
	return &PolicyTemplateManager{policytemplates.New(ksInformers)}
}

func (m *PolicyTemplateManager) GetPolicyTemplate(_ context.Context, templateName string) (*v1alpha1.PolicyTemplateDetail, error) {
	obj, err := m.getter.Get("", PolicyTemplateUniqueName(templateName))
	if err != nil {
		return nil, err
	}
	template := obj.(*admission.PolicyTemplate).DeepCopy()
	if template == nil {
		return nil, v1alpha1.ErrPolicyTemplateNotFound
	}
	return PolicyTemplateDetail(template), nil
}

func (m *PolicyTemplateManager) ListPolicyTemplates(_ context.Context, query *query.Query) (*v1alpha1.PolicyTemplateList, error) {
	objs, err := m.getter.List("", query)
	if err != nil {
		return nil, err
	}
	list := &v1alpha1.PolicyTemplateList{
		Total: objs.TotalItems,
	}
	for _, obj := range objs.Items {
		template := obj.(*admission.PolicyTemplate).DeepCopy()
		list.Items = append(list.Items, PolicyTemplate(template))
	}
	return list, err
}

func PolicyTemplateDetail(policy *admission.PolicyTemplate) *v1alpha1.PolicyTemplateDetail {
	detail := &v1alpha1.PolicyTemplateDetail{
		PolicyTemplate: *PolicyTemplate(policy),
	}
	return detail
}

func PolicyTemplate(policy *admission.PolicyTemplate) *v1alpha1.PolicyTemplate {
	targets := policy.Spec.Content.Targets

	var templateTargets []v1alpha1.PolicyTemplateTarget
	for _, target := range targets {
		templateTargets = append(templateTargets, v1alpha1.PolicyTemplateTarget{
			Target:     target.Target,
			Expression: target.Expression,
			Import:     target.Import,
			Provider:   target.Provider,
		})
	}

	validation := policy.Spec.Content.Spec.Parameters.Validation

	return &v1alpha1.PolicyTemplate{
		Name:        policy.Spec.Name,
		Description: policy.Spec.Description,
		Targets:     templateTargets,
		Parameters: v1alpha1.Parameters{
			Validation: &v1alpha1.Validation{
				OpenAPIV3Schema: validation.OpenAPIV3Schema,
				LegacySchema:    validation.LegacySchema,
			},
		},
	}
}

func PolicyTemplateUniqueName(templateName string) string {
	return strings.ToLower(templateName)
}
