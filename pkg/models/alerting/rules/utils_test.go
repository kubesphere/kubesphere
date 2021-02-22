package rules

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/rules"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
)

func TestGetAlertingRulesStatus(t *testing.T) {
	var tests = []struct {
		description       string
		ruleNamespace     string
		resourceRuleChunk *ResourceRuleChunk
		ruleGroups        []*alerting.RuleGroup
		extLabels         func() map[string]string
		expected          []*v2alpha1.GettableAlertingRule
	}{{
		description:   "get alerting rules status",
		ruleNamespace: "test",
		resourceRuleChunk: &ResourceRuleChunk{
			Level:  v2alpha1.RuleLevelNamespace,
			Custom: true,
			ResourceRulesMap: map[string]*ResourceRuleCollection{
				"custom-alerting-rule-jqbgn": &ResourceRuleCollection{
					GroupSet: map[string]struct{}{"alerting.custom.defaults": struct{}{}},
					NameRules: map[string][]*ResourceRuleItem{
						"ca7f09e76954e67c": []*ResourceRuleItem{{
							ResourceName: "custom-alerting-rule-jqbgn",
							RuleWithGroup: RuleWithGroup{
								Group: "alerting.custom.defaults",
								Id:    "ca7f09e76954e67c",
								Rule: promresourcesv1.Rule{
									Alert: "TestCPUUsageHigh",
									Expr:  intstr.FromString(`namespace:workload_cpu_usage:sum{namespace="test"} > 1`),
									For:   "1m",
									Annotations: map[string]string{
										"alias":       "The alias is here",
										"description": "The description is here",
									},
								},
							},
						}},
					},
				},
			},
		},
		ruleGroups: []*alerting.RuleGroup{{
			Name: "alerting.custom.defaults",
			File: "/etc/thanos/rules/thanos-ruler-thanos-ruler-rulefiles-0/test-custom-alerting-rule-jqbgn.yaml",
			Rules: []*alerting.AlertingRule{{
				Name:     "TestCPUUsageHigh",
				Query:    `namespace:workload_cpu_usage:sum{namespace="test"} > 1`,
				Duration: 60,
				Health:   string(rules.HealthGood),
				State:    stateInactiveString,
				Annotations: map[string]string{
					"alias":       "The alias is here",
					"description": "The description is here",
				},
			}},
		}},
		expected: []*v2alpha1.GettableAlertingRule{{
			AlertingRule: v2alpha1.AlertingRule{
				Id:       "ca7f09e76954e67c",
				Name:     "TestCPUUsageHigh",
				Query:    `namespace:workload_cpu_usage:sum{namespace="test"} > 1`,
				Duration: "1m",
				Annotations: map[string]string{
					"alias":       "The alias is here",
					"description": "The description is here",
				},
			},
			Health: string(rules.HealthGood),
			State:  stateInactiveString,
		}},
	}}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			rules, err := GetAlertingRulesStatus(test.ruleNamespace, test.resourceRuleChunk, test.ruleGroups, test.extLabels)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(rules, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}
