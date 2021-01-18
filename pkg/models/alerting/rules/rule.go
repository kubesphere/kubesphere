package rules

import (
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
)

type ResourceRuleCollection struct {
	GroupSet  map[string]struct{}
	IdRules   map[string]*ResourceRuleItem
	NameRules map[string][]*ResourceRuleItem
}

type ResourceRuleItem struct {
	ResourceName string
	Group        string
	Id           string
	Rule         *promresourcesv1.Rule
}

type ResourceRule struct {
	Level  v2alpha1.RuleLevel
	Custom bool
	ResourceRuleItem
}

type ResourceRuleChunk struct {
	Level            v2alpha1.RuleLevel
	Custom           bool
	ResourceRulesMap map[string]*ResourceRuleCollection
}
