/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2beta1

import (
	"strings"
	"time"

	alertingv2beta1 "kubesphere.io/api/alerting/v2beta1"
)

const (
	// for rulegroup/alert
	FieldState   = "state"
	FieldBuiltin = "builtin"

	// for rulegroup
	FieldRuleGroupEvaluationTime = "evaluationTime"
	FieldRuleGroupLastEvaluation = "lastEvalution"

	// for alert
	FieldAlertLabelFilters = "label_filters"
	FieldAlertActiveAt     = "activeAt"
)

var SortableFields = []string{
	FieldRuleGroupEvaluationTime,
	FieldRuleGroupLastEvaluation,
	FieldAlertActiveAt,
}

var ComparableFields = []string{
	FieldState,
	FieldAlertLabelFilters,
}

type RuleGroup struct {
	alertingv2beta1.RuleGroup `json:",inline"`
	Status                    RuleGroupStatus `json:"status,omitempty"`
}

type ClusterRuleGroup struct {
	alertingv2beta1.ClusterRuleGroup `json:",inline"`
	Status                           RuleGroupStatus `json:"status,omitempty"`
}

type GlobalRuleGroup struct {
	alertingv2beta1.GlobalRuleGroup `json:",inline"`
	Status                          RuleGroupStatus `json:"status,omitempty"`
}

type RuleGroupStatus struct {
	State          string       `json:"state,omitempty" description:"state of a rulegroup, one of firing, pending or inactive depending on its rules"`
	EvaluationTime *float64     `json:"evaluationTime,omitempty" description:"time spent on rule group evaluation in seconds"`
	LastEvaluation *time.Time   `json:"lastEvaluation,omitempty" description:"time of last evaluation"`
	RulesStatus    []RuleStatus `json:"rulesStatus,omitempty" description:"status of rules in one RuleGroup"`
	RulesStats     RulesStats   `json:"rulesStats,omitempty" description:"statistics of rules in one RuleGroup"`
}

type RulesStats struct {
	Inactive int `json:"inactive" description:"count of rules in the inactive state"`
	Pending  int `json:"pending" description:"count of rules in the pending state"`
	Firing   int `json:"firing" description:"count of rules in the firing state"`
	Disabled int `json:"disabled" description:"count of disabled rules"`
}

type RuleStatus struct {
	Expr           string     `json:"expr,omitempty" description:"expression evaluated, for global rules only"`
	State          string     `json:"state,omitempty" description:"state of a rule, one of firing, pending or inactive depending on its alerts"`
	Health         string     `json:"health,omitempty" description:"health state of a rule, one of ok, err, unknown depending on the last execution result"`
	LastError      string     `json:"lastError,omitempty" description:"error of the last evaluation"`
	EvaluationTime *float64   `json:"evaluationTime,omitempty" description:"time spent on the expression evaluation in seconds"`
	LastEvaluation *time.Time `json:"lastEvaluation,omitempty" description:"time of last evaluation"`
	ActiveAt       *time.Time `json:"activeAt,omitempty" description:"time when this rule became active"`

	Alerts []*Alert `json:"alerts,omitempty" description:"alerts"`
}

type Alert struct {
	ActiveAt    *time.Time        `json:"activeAt,omitempty" description:"time when this alert became active"`
	Annotations map[string]string `json:"annotations,omitempty" description:"annotations"`
	Labels      map[string]string `json:"labels,omitempty" description:"labels"`
	State       string            `json:"state,omitempty" description:"state"`
	Value       string            `json:"value,omitempty" description:"the value from the last expression evaluation"`
}

type LabelFilterOperator string

const (
	LabelFilterOperatorEqual   = "="
	LabelFilterOperatorContain = "~"
)

type LabelFilter struct {
	LabelName  string
	LabelValue string
	Operator   LabelFilterOperator
}

func (f *LabelFilter) Matches(labels map[string]string) bool {
	v, ok := labels[f.LabelName]
	if ok {
		switch f.Operator {
		case LabelFilterOperatorEqual:
			return v == f.LabelValue
		case LabelFilterOperatorContain:
			return strings.Contains(v, f.LabelValue)
		}
	}
	return false
}

type LabelFilters []LabelFilter

func (fs LabelFilters) Matches(labels map[string]string) bool {
	for _, f := range fs {
		if !f.Matches(labels) {
			return false
		}
	}
	return true
}

func ParseLabelFilters(filters string) LabelFilters {
	var fs LabelFilters
	for _, filter := range strings.Split(filters, ",") {
		if i := strings.Index(filter, LabelFilterOperatorEqual); i > 0 {
			fs = append(fs, LabelFilter{
				Operator:   LabelFilterOperatorEqual,
				LabelName:  filter[:i],
				LabelValue: filter[i+1:],
			})
		} else if i := strings.Index(filter, LabelFilterOperatorContain); i > 0 {
			fs = append(fs, LabelFilter{
				Operator:   LabelFilterOperatorContain,
				LabelName:  filter[:i],
				LabelValue: filter[i+1:],
			})
		}
	}
	return fs
}
