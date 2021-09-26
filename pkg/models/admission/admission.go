package admission

import (
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/admission/polciyTemplates"
	"kubesphere.io/kubesphere/pkg/models/admission/policies"
	"kubesphere.io/kubesphere/pkg/models/admission/provider"
	"kubesphere.io/kubesphere/pkg/models/admission/rules"
)

type Operator interface {
	polciyTemplates.PolicyTemplateManagerInterface
	policies.PolicyManagerInterface
	rules.RuleManagerInterface
}

type admissionOperator struct {
	*polciyTemplates.PolicyTemplateManager
	*policies.PolicyManager
	*rules.RuleManager
}

func NewOperator(ksClient kubesphere.Interface,
	ksInformers externalversions.SharedInformerFactory,
	providers map[string]provider.Provider) *admissionOperator {
	return &admissionOperator{
		PolicyTemplateManager: polciyTemplates.NewPolicyTemplateManager(
			ksInformers,
		),
		PolicyManager: policies.NewPolicyManager(
			ksClient, ksInformers, providers,
		),
		RuleManager: rules.NewRuleManager(
			ksClient, ksInformers, providers,
		),
	}
}
