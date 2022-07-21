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
	FieldState = "state"

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
	State          string       `json:"state,omitempty" description:"state of a rule based on its alerts, one of firing, pending, inactive"`
	EvaluationTime *float64     `json:"evaluationTime,omitempty" description:"taken seconds for evaluation"`
	LastEvaluation *time.Time   `json:"lastEvaluation,omitempty" description:"time for last evaluation"`
	RuleStatuses   []RuleStatus `json:"ruleStatuses,omitempty" description:"rule statuses"`
}

type RuleStatus struct {
	State                     string     `json:"state,omitempty" description:"state of a rule based on its alerts, one of firing, pending, inactive"`
	Health                    string     `json:"health,omitempty" description:"health state of a rule based on the last execution, one of ok, err, unknown"`
	LastError                 string     `json:"lastError,omitempty" description:"error for the last execution"`
	EvaluationDurationSeconds float64    `json:"evaluationTime,omitempty" description:"taken seconds for evaluation of query expression"`
	LastEvaluation            *time.Time `json:"lastEvaluation,omitempty" description:"time for last evaluation of query expression"`

	Alerts []*Alert `json:"alerts,omitempty" description:"alerts"`
}

type Alert struct {
	ActiveAt    *time.Time        `json:"activeAt,omitempty" description:"time when alert is active"`
	Annotations map[string]string `json:"annotations,omitempty" description:"annotations"`
	Labels      map[string]string `json:"labels,omitempty" description:"labels"`
	State       string            `json:"state,omitempty" description:"state"`
	Value       string            `json:"value,omitempty" description:"the value at the last evaluation of the query expression"`
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
