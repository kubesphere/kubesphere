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

package v1alpha1

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"
	prommodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

const (
	RuleLevelCluster   RuleLevel = "cluster"
	RuleLevelNamespace RuleLevel = "namespace"
)

var (
	ErrThanosRulerNotEnabled     = errors.New("The request operation to custom alerting rule could not be done because thanos ruler is not enabled")
	ErrAlertingRuleNotFound      = errors.New("The alerting rule was not found")
	ErrAlertingRuleAlreadyExists = errors.New("The alerting rule already exists")

	ruleLabelNameMatcher = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
)

type RuleLevel string

type AlertingRule struct {
	Id   string `json:"id,omitempty" description:"rule id is only used by built-in alerting rules"`
	Name string `json:"name,omitempty" description:"rule name should be unique in one namespace for custom alerting rules"`

	Query       string            `json:"query,omitempty" description:"prometheus query expression, grammars of which may be referred to https://prometheus.io/docs/prometheus/latest/querying/basics/"`
	Duration    string            `json:"duration,omitempty" description:"duration an alert transitions from Pending to Firing state, which must match ^([0-9]+)(y|w|d|h|m|s|ms)$"`
	Labels      map[string]string `json:"labels,omitempty" description:"extra labels to attach to the resulting alert sample vectors (the key string has to match [a-zA-Z_][a-zA-Z0-9_]*). eg: a typical label called severity, whose value may be info, warning, error, critical, is usually used to indicate the severity of an alert"`
	Annotations map[string]string `json:"annotations,omitempty" description:"non-identifying key/value pairs. summary, message, description are the commonly used annotation names"`
}

type PostableAlertingRule struct {
	AlertingRule `json:",omitempty"`
}

func (r *PostableAlertingRule) Validate() error {
	errs := []error{}

	if r.Name == "" {
		errs = append(errs, errors.New("name can not be empty"))
	}
	if _, err := parser.ParseExpr(r.Query); err != nil {
		errs = append(errs, errors.Wrapf(err, "query is invalid: %s", r.Query))
	}
	if r.Duration != "" {
		if _, err := prommodel.ParseDuration(r.Duration); err != nil {
			errs = append(errs, errors.Wrapf(err, "duration is invalid: %s", r.Duration))
		}
	}

	if len(r.Labels) > 0 {
		for name, _ := range r.Labels {
			if !ruleLabelNameMatcher.MatchString(name) || strings.HasPrefix(name, "__") {
				errs = append(errs, errors.Errorf(
					"label name (%s) is not valid. The name must match [a-zA-Z_][a-zA-Z0-9_]* and has not the __ prefix (label names with this prefix are for internal use)", name))
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}

type GettableAlertingRule struct {
	AlertingRule `json:",omitempty"`

	State                     string     `json:"state,omitempty" description:"state of a rule based on its alerts, one of firing, pending, inactive"`
	Health                    string     `json:"health,omitempty" description:"health state of a rule based on the last execution, one of ok, err, unknown"`
	LastError                 string     `json:"lastError,omitempty" description:"error for the last execution"`
	EvaluationDurationSeconds float64    `json:"evaluationTime,omitempty" description:"taken seconds for evaluation of query expression"`
	LastEvaluation            *time.Time `json:"lastEvaluation,omitempty" description:"time for last evaluation of query expression"`

	Alerts []*Alert `json:"alerts,omitempty" description:"alerts"`
}

type GettableAlertingRuleList struct {
	Items []*GettableAlertingRule `json:"items"`
	Total int                     `json:"total"`
}

type Alert struct {
	ActiveAt    *time.Time        `json:"activeAt,omitempty" description:"time when alert is active"`
	Annotations map[string]string `json:"annotations,omitempty" description:"annotations"`
	Labels      map[string]string `json:"labels,omitempty" description:"labels"`
	State       string            `json:"state,omitempty" description:"state"`
	Value       string            `json:"value,omitempty" description:"the value at the last evaluation of the query expression"`

	RuleId   string `json:"ruleId,omitempty" description:"rule id triggering the alert"`
	RuleName string `json:"ruleName,omitempty" description:"rule name triggering the alert"`
}

type AlertList struct {
	Items []*Alert `json:"items"`
	Total int      `json:"total"`
}

type AlertingRuleQueryParams struct {
	NameContainFilter   string
	State               string
	Health              string
	LabelEqualFilters   map[string]string
	LabelContainFilters map[string]string

	Offset    int
	Limit     int
	SortField string
	SortType  string
}

func (q *AlertingRuleQueryParams) Filter(rules []*GettableAlertingRule) []*GettableAlertingRule {
	var ret []*GettableAlertingRule
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if q == nil || q.matches(rule) {
			ret = append(ret, rule)
		}
	}
	return ret
}

func (q *AlertingRuleQueryParams) matches(rule *GettableAlertingRule) bool {
	if q.NameContainFilter != "" && !strings.Contains(rule.Name, q.NameContainFilter) {
		return false
	}
	if q.State != "" && q.State != rule.State {
		return false
	}
	if q.Health != "" && q.Health != rule.Health {
		return false
	}
	if len(rule.Labels) == 0 {
		return len(q.LabelEqualFilters) == 0 && len(q.LabelContainFilters) == 0
	}
	for k, v := range q.LabelEqualFilters {
		if fv, ok := rule.Labels[k]; !ok || fv != v {
			return false
		}
	}
	for k, v := range q.LabelContainFilters {
		if fv, ok := rule.Labels[k]; !ok || !strings.Contains(fv, v) {
			return false
		}
	}
	return true
}

// AlertingRuleIdCompare defines the default order for the alerting rules.
// For the alerting rule list, it guarantees a stable sort. For the custom alerting rules with possible same names
// and the builtin alerting rules with possible same ids, it guarantees the stability of get operations.
func AlertingRuleIdCompare(leftId, rightId string) bool {
	// default to ascending order of id
	return leftId <= rightId
}

func (q *AlertingRuleQueryParams) Sort(rules []*GettableAlertingRule) {
	idCompare := func(left, right *GettableAlertingRule) bool {
		return AlertingRuleIdCompare(left.Id, right.Id)
	}
	var compare = idCompare
	if q != nil {
		reverse := q.SortType == "desc"
		switch q.SortField {
		case "name":
			compare = func(left, right *GettableAlertingRule) bool {
				if c := strings.Compare(left.Name, right.Name); c != 0 {
					if reverse {
						return c > 0
					}
					return c < 0
				}
				return idCompare(left, right)
			}
		case "lastEvaluation":
			compare = func(left, right *GettableAlertingRule) bool {
				if left.LastEvaluation == nil {
					if right.LastEvaluation != nil {
						return false
					}
				} else {
					if right.LastEvaluation == nil {
						return true
					} else if left.LastEvaluation.Equal(*right.LastEvaluation) {
						if reverse {
							return left.LastEvaluation.After(*right.LastEvaluation)
						}
						return left.LastEvaluation.Before(*right.LastEvaluation)
					}
				}
				return idCompare(left, right)
			}
		case "evaluationTime":
			compare = func(left, right *GettableAlertingRule) bool {
				if left.EvaluationDurationSeconds != right.EvaluationDurationSeconds {
					if reverse {
						return left.EvaluationDurationSeconds > right.EvaluationDurationSeconds
					}
					return left.EvaluationDurationSeconds < right.EvaluationDurationSeconds
				}
				return idCompare(left, right)
			}
		}
	}
	sort.Slice(rules, func(i, j int) bool {
		return compare(rules[i], rules[j])
	})
}

func (q *AlertingRuleQueryParams) Sub(rules []*GettableAlertingRule) []*GettableAlertingRule {
	start, stop := 0, 10
	if q != nil {
		start, stop = q.Offset, q.Offset+q.Limit
	}
	total := len(rules)
	if start < total {
		if stop > total {
			stop = total
		}
		return rules[start:stop]
	}
	return nil
}

type AlertQueryParams struct {
	State               string
	LabelEqualFilters   map[string]string
	LabelContainFilters map[string]string

	Offset int
	Limit  int
}

func (q *AlertQueryParams) Filter(alerts []*Alert) []*Alert {
	var ret []*Alert
	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		if q == nil || q.matches(alert) {
			ret = append(ret, alert)
		}
	}
	return ret
}

func (q *AlertQueryParams) matches(alert *Alert) bool {
	if q.State != "" && q.State != alert.State {
		return false
	}
	if len(alert.Labels) == 0 {
		return len(q.LabelEqualFilters) == 0 && len(q.LabelContainFilters) == 0
	}
	for k, v := range q.LabelEqualFilters {
		if fv, ok := alert.Labels[k]; !ok || fv != v {
			return false
		}
	}
	for k, v := range q.LabelContainFilters {
		if fv, ok := alert.Labels[k]; !ok || !strings.Contains(fv, v) {
			return false
		}
	}
	return true
}

func (q *AlertQueryParams) Sort(alerts []*Alert) {
	compare := func(left, right *Alert) bool {
		if left.ActiveAt == nil {
			if right.ActiveAt != nil {
				return false
			}
		} else {
			if right.ActiveAt == nil {
				return true
			} else if !left.ActiveAt.Equal(*right.ActiveAt) {
				return left.ActiveAt.After(*right.ActiveAt)
			}
		}
		return prommodel.LabelsToSignature(left.Labels) <= prommodel.LabelsToSignature(right.Labels)
	}
	sort.Slice(alerts, func(i, j int) bool {
		return compare(alerts[i], alerts[j])
	})
}

func (q *AlertQueryParams) Sub(alerts []*Alert) []*Alert {
	start, stop := 0, 10
	if q != nil {
		start, stop = q.Offset, q.Offset+q.Limit
	}
	total := len(alerts)
	if start < total {
		if stop > total {
			stop = total
		}
		return alerts[start:stop]
	}
	return nil
}

func ParseAlertingRuleQueryParams(req *restful.Request) (*AlertingRuleQueryParams, error) {
	var (
		q   = &AlertingRuleQueryParams{}
		err error
	)

	q.NameContainFilter = req.QueryParameter("name")
	q.State = req.QueryParameter("state")
	q.Health = req.QueryParameter("health")
	q.Offset, _ = strconv.Atoi(req.QueryParameter("offset"))
	q.Limit, err = strconv.Atoi(req.QueryParameter("limit"))
	if err != nil {
		q.Limit = 10
		err = nil
	}
	q.LabelEqualFilters, q.LabelContainFilters = parseLabelFilters(req)
	q.SortField = req.QueryParameter("sort_field")
	q.SortType = req.QueryParameter("sort_type")
	return q, err
}

func ParseAlertQueryParams(req *restful.Request) (*AlertQueryParams, error) {
	var (
		q   = &AlertQueryParams{}
		err error
	)

	q.State = req.QueryParameter("state")
	q.Offset, _ = strconv.Atoi(req.QueryParameter("offset"))
	q.Limit, err = strconv.Atoi(req.QueryParameter("limit"))
	if err != nil {
		q.Limit = 10
		err = nil
	}
	q.LabelEqualFilters, q.LabelContainFilters = parseLabelFilters(req)
	return q, err
}

func parseLabelFilters(req *restful.Request) (map[string]string, map[string]string) {
	var (
		labelEqualFilters   = make(map[string]string)
		labelContainFilters = make(map[string]string)
		labelFiltersString  = req.QueryParameter("label_filters")
	)
	for _, filter := range strings.Split(labelFiltersString, ",") {
		if i := strings.Index(filter, "="); i > 0 && len(filter) > i+1 {
			labelEqualFilters[filter[:i]] = filter[i+1:]
		} else if i := strings.Index(filter, "~"); i > 0 && len(filter) > i+1 {
			labelContainFilters[filter[:i]] = filter[i+1:]
		}
	}
	return labelEqualFilters, labelContainFilters
}
