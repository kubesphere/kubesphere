package rules

import (
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"kubesphere.io/kubesphere/pkg/api/customalerting/v1alpha1"
)

type ResourceRules struct {
	GroupSet  map[string]struct{}
	IdRules   map[string]*ResourceRule
	NameRules map[string][]*ResourceRule
}

type ResourceRule struct {
	ResourceName string
	Group        string
	Id           string
	Rule         *promresourcesv1.Rule
}

type ResourceRuleSole struct {
	Level  v1alpha1.RuleLevel
	Custom bool
	ResourceRule
}

type ResourceRuleChunk struct {
	Level            v1alpha1.RuleLevel
	Custom           bool
	ResourceRulesMap map[string]*ResourceRules
}
