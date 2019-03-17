package models

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiali/kiali/kubernetes"
)

type IstioRuleList struct {
	Namespace Namespace   `json:"namespace"`
	Rules     []IstioRule `json:"rules"`
}

// IstioRules istioRules
//
// This type type is used for returning an array of IstioRules
//
// swagger:model istioRules
// An array of istioRule
// swagger:allOf
type IstioRules []IstioRule

// IstioRule istioRule
//
// This type type is used for returning a IstioRule
//
// swagger:model istioRule
type IstioRule struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Match   interface{} `json:"match"`
		Actions interface{} `json:"actions"`
	} `json:"spec"`
}

// IstioAdapters istioAdapters
//
// This type type is used for returning an array of IstioAdapters
//
// swagger:model istioAdapters
// An array of istioAdapter
// swagger:allOf
type IstioAdapters []IstioAdapter

// IstioAdapter istioAdapter
//
// This type type is used for returning a IstioAdapter
//
// swagger:model istioAdapter
type IstioAdapter struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     interface{}        `json:"spec"`
	Adapter  string             `json:"adapter"`
	// We need to bring the plural to use it from the UI to build the API
	Adapters string `json:"adapters"`
}

// IstioTemplates istioTemplates
//
// This type type is used for returning an array of IstioTemplates
//
// swagger:model istioTemplates
// An array of istioTemplates
// swagger:allOf
type IstioTemplates []IstioTemplate

// IstioTemplate istioTemplate
//
// This type type is used for returning a IstioTemplate
//
// swagger:model istioTemplate
type IstioTemplate struct {
	Metadata meta_v1.ObjectMeta `json:"metadata"`
	Spec     interface{}        `json:"spec"`
	Template string             `json:"template"`
	// We need to bring the plural to use it from the UI to build the API
	Templates string `json:"templates"`
}

func CastIstioRulesCollection(rules []kubernetes.IstioObject) IstioRules {
	istioRules := make([]IstioRule, len(rules))
	for i, rule := range rules {
		istioRules[i] = CastIstioRule(rule)
	}
	return istioRules
}

func CastIstioRule(rule kubernetes.IstioObject) IstioRule {
	istioRule := IstioRule{}
	istioRule.Metadata = rule.GetObjectMeta()
	istioRule.Spec.Match = rule.GetSpec()["match"]
	istioRule.Spec.Actions = rule.GetSpec()["actions"]
	return istioRule
}

func CastIstioAdaptersCollection(adapters []kubernetes.IstioObject) IstioAdapters {
	istioAdapters := make([]IstioAdapter, len(adapters))
	for i, adapter := range adapters {
		istioAdapters[i] = CastIstioAdapter(adapter)
	}
	return istioAdapters
}

func CastIstioAdapter(adapter kubernetes.IstioObject) IstioAdapter {
	istioAdapter := IstioAdapter{}
	istioAdapter.Metadata = adapter.GetObjectMeta()
	istioAdapter.Spec = adapter.GetSpec()
	istioAdapter.Adapter = adapter.GetObjectMeta().Labels["adapter"]
	istioAdapter.Adapters = adapter.GetObjectMeta().Labels["adapters"]
	return istioAdapter
}

func CastIstioTemplatesCollection(templates []kubernetes.IstioObject) IstioTemplates {
	istioTemplates := make([]IstioTemplate, len(templates))
	for i, template := range templates {
		istioTemplates[i] = CastIstioTemplate(template)
	}
	return istioTemplates
}

func CastIstioTemplate(template kubernetes.IstioObject) IstioTemplate {
	istioTemplate := IstioTemplate{}
	istioTemplate.Metadata = template.GetObjectMeta()
	istioTemplate.Spec = template.GetSpec()
	istioTemplate.Template = template.GetObjectMeta().Labels["template"]
	istioTemplate.Templates = template.GetObjectMeta().Labels["templates"]
	return istioTemplate
}
