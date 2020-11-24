package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/rules"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kubesphere.io/kubesphere/pkg/api/customalerting/v1alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/customalerting"
)

func TestMixAlertingRules(t *testing.T) {
	var tests = []struct {
		description       string
		ruleNamespace     string
		resourceRuleChunk *ResourceRuleChunk
		ruleGroups        []*customalerting.RuleGroup
		extLabels         func() map[string]string
		expected          []*v1alpha1.GettableAlertingRule
	}{{
		description:   "mix custom rules",
		ruleNamespace: "test",
		resourceRuleChunk: &ResourceRuleChunk{
			Level:  v1alpha1.RuleLevelNamespace,
			Custom: true,
			ResourceRulesMap: map[string]*ResourceRules{
				"custom-alerting-rule-jqbgn": &ResourceRules{
					GroupSet: map[string]struct{}{"alerting.custom.defaults": struct{}{}},
					NameRules: map[string][]*ResourceRule{
						"f89836879157ca88": []*ResourceRule{{
							ResourceName: "custom-alerting-rule-jqbgn",
							Group:        "alerting.custom.defaults",
							Id:           "f89836879157ca88",
							Rule: &promresourcesv1.Rule{
								Alert: "TestCPUUsageHigh",
								Expr:  intstr.FromString(`namespace:workload_cpu_usage:sum{namespace="test"} > 1`),
								For:   "1m",
								Labels: map[string]string{
									LabelKeyInternalRuleAlias:       "The alias is here",
									LabelKeyInternalRuleDescription: "The description is here",
								},
							},
						}},
					},
				},
			},
		},
		ruleGroups: []*customalerting.RuleGroup{{
			Name: "alerting.custom.defaults",
			File: "/etc/thanos/rules/thanos-ruler-thanos-ruler-rulefiles-0/test-custom-alerting-rule-jqbgn.yaml",
			Rules: []*customalerting.AlertingRule{{
				Name:     "TestCPUUsageHigh",
				Query:    `namespace:workload_cpu_usage:sum{namespace="test"} > 1`,
				Duration: 60,
				Health:   string(rules.HealthGood),
				State:    stateInactiveString,
				Labels: map[string]string{
					LabelKeyInternalRuleAlias:       "The alias is here",
					LabelKeyInternalRuleDescription: "The description is here",
				},
			}},
		}},
		expected: []*v1alpha1.GettableAlertingRule{{
			AlertingRuleQualifier: v1alpha1.AlertingRuleQualifier{
				Id:     "f89836879157ca88",
				Name:   "TestCPUUsageHigh",
				Level:  v1alpha1.RuleLevelNamespace,
				Custom: true,
			},
			AlertingRuleProps: v1alpha1.AlertingRuleProps{
				Query:    `namespace:workload_cpu_usage:sum{namespace="test"} > 1`,
				Duration: "1m",
				Labels:   map[string]string{},
			},
			Alias:       "The alias is here",
			Description: "The description is here",
			Health:      string(rules.HealthGood),
			State:       stateInactiveString,
		}},
	}}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			rules, err := MixAlertingRules(test.ruleNamespace, test.resourceRuleChunk, test.ruleGroups, test.extLabels)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(rules, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}
