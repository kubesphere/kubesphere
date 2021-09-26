package polciyTemplates

import (
	"context"
	"kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

type PolicyTemplateManagerInterface interface {
	// GetPolicyTemplate gets the admission policy template with the given name.
	GetPolicyTemplate(ctx context.Context, templateName string) (*v2alpha1.GettableAlertingRule, error)
	// ListPolicyTemplates lists the alerts of the custom alerting rule with the given name.
	ListPolicyTemplates(ctx context.Context) (*v2alpha1.AlertList, error)
}

type PolicyTemplateManager struct {
	ksInformers externalversions.SharedInformerFactory
}

func NewPolicyTemplateManager(ksInformers externalversions.SharedInformerFactory) *PolicyTemplateManager {
	return &PolicyTemplateManager{ksInformers: ksInformers}
}

func (p *PolicyTemplateManager) GetPolicyTemplate(ctx context.Context, templateName string) (*v2alpha1.GettableAlertingRule, error) {
	panic("implement me")
}

func (p *PolicyTemplateManager) ListPolicyTemplates(ctx context.Context) (*v2alpha1.AlertList, error) {
	panic("implement me")
}
