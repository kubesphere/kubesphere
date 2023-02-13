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

package v2alpha1

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"
	prommodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/template"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

const (
	RuleLevelCluster   RuleLevel = "cluster"
	RuleLevelNamespace RuleLevel = "namespace"

	AnnotationKeyRuleUpdateTime = "rule_update_time"
)

var (
	ErrThanosRulerNotEnabled     = errors.New("The request operation to custom alerting rule could not be done because thanos ruler is not enabled")
	ErrAlertingRuleNotFound      = errors.New("The alerting rule was not found")
	ErrAlertingRuleAlreadyExists = errors.New("The alerting rule already exists")
	ErrAlertingAPIV2NotEnabled   = errors.New("The alerting v2 API is not enabled")

	templateTestData       = template.AlertTemplateData(map[string]string{}, map[string]string{}, "", 0)
	templateTestTextPrefix = "{{$labels := .Labels}}{{$externalLabels := .ExternalLabels}}{{$value := .Value}}"

	ruleNameMatcher = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
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
	} else {
		if !ruleNameMatcher.MatchString(r.Name) {
			errs = append(errs, errors.New("rule name must match regular expression ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"))
		}
	}

	if r.Query == "" {
		errs = append(errs, errors.New("query can not be empty"))
	} else if _, err := parser.ParseExpr(r.Query); err != nil {
		errs = append(errs, errors.Wrapf(err, "query is invalid: %s", r.Query))
	}
	if r.Duration != "" {
		if _, err := prommodel.ParseDuration(r.Duration); err != nil {
			errs = append(errs, errors.Wrapf(err, "duration is invalid: %s", r.Duration))
		}
	}

	parseTest := func(text string) error {
		tmpl := template.NewTemplateExpander(
			context.TODO(),
			templateTestTextPrefix+text,
			"__alert_"+r.Name,
			templateTestData,
			prommodel.Time(timestamp.FromTime(time.Now())),
			nil,
			nil,
			nil,
		)
		return tmpl.ParseTest()
	}

	if len(r.Labels) > 0 {
		for name, v := range r.Labels {
			if !prommodel.LabelName(name).IsValid() || strings.HasPrefix(name, "__") {
				errs = append(errs, errors.Errorf(
					"label name (%s) is not valid. The name must match [a-zA-Z_][a-zA-Z0-9_]* and has not the __ prefix (label names with this prefix are for internal use)", name))
			}
			if !prommodel.LabelValue(v).IsValid() {
				errs = append(errs, errors.Errorf("invalid label value: %s", v))
			}
			if err := parseTest(v); err != nil {
				errs = append(errs, errors.Errorf("invalid label value: %s", v))
			}
		}
	}

	if len(r.Annotations) > 0 {
		for name, v := range r.Annotations {
			if !prommodel.LabelName(name).IsValid() {
				errs = append(errs, errors.Errorf("invalid annotation name: %s", v))
			}
			if err := parseTest(v); err != nil {
				errs = append(errs, errors.Errorf("invalid annotation value: %s", v))
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

	PageNum   int
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
	if q.NameContainFilter != "" && !containsCaseInsensitive(rule.Name, q.NameContainFilter) {
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
		if fv, ok := rule.Labels[k]; !ok || !containsCaseInsensitive(fv, v) {
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
	baseCompare := func(left, right *GettableAlertingRule) bool {
		var leftUpdateTime, rightUpdateTime string
		if len(left.Annotations) > 0 {
			leftUpdateTime = left.Annotations[AnnotationKeyRuleUpdateTime]
		}
		if len(right.Annotations) > 0 {
			rightUpdateTime = right.Annotations[AnnotationKeyRuleUpdateTime]
		}

		if leftUpdateTime != rightUpdateTime {
			return leftUpdateTime > rightUpdateTime
		}

		return AlertingRuleIdCompare(left.Id, right.Id)
	}
	var compare = baseCompare
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
				return baseCompare(left, right)
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
				return baseCompare(left, right)
			}
		case "evaluationTime":
			compare = func(left, right *GettableAlertingRule) bool {
				if left.EvaluationDurationSeconds != right.EvaluationDurationSeconds {
					if reverse {
						return left.EvaluationDurationSeconds > right.EvaluationDurationSeconds
					}
					return left.EvaluationDurationSeconds < right.EvaluationDurationSeconds
				}
				return baseCompare(left, right)
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
		start, stop = (q.PageNum-1)*q.Limit, q.PageNum*q.Limit
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

	PageNum int
	Limit   int
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
		if fv, ok := alert.Labels[k]; !ok || !containsCaseInsensitive(fv, v) {
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
		start, stop = (q.PageNum-1)*q.Limit, q.PageNum*q.Limit
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
	q.PageNum, err = strconv.Atoi(req.QueryParameter("page"))
	if err != nil {
		q.PageNum = 1
	}
	if q.PageNum <= 0 {
		q.PageNum = 1
	}
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
	q.PageNum, err = strconv.Atoi(req.QueryParameter("page"))
	if err != nil {
		q.PageNum = 1
	}
	if q.PageNum <= 0 {
		q.PageNum = 1
	}
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

const (
	ErrBadData       ErrorType = "bad_data"
	ErrDuplicateName ErrorType = "duplicate_name"
	ErrNotFound      ErrorType = "not_found"
	ErrServer        ErrorType = "server_error"

	StatusSuccess Status = "success"
	StatusError   Status = "error"

	ResultCreated Result = "created"
	ResultUpdated Result = "updated"
	ResultDeleted Result = "deleted"
)

type Status string

type ErrorType string

type Result string

type BulkResponse struct {
	Errors bool                `json:"errors" description:"If true, one or more operations in the bulk request don't complete successfully"`
	Items  []*BulkItemResponse `json:"items" description:"It contains the result of each operation in the bulk request"`
}

// MakeBulkResponse tidies the internal items and sets the errors
func (br *BulkResponse) MakeBulkResponse() *BulkResponse {
	var (
		items   []*BulkItemResponse
		itemMap = make(map[string]*BulkItemResponse)
	)
	for i, item := range br.Items {
		if item.Status == StatusError {
			br.Errors = true
		}
		pitem, ok := itemMap[item.RuleName]
		if !ok || (pitem.Status == StatusSuccess || item.Status == StatusError) {
			itemMap[item.RuleName] = br.Items[i]
		}
	}
	for k := range itemMap {
		item := itemMap[k]
		if item.Error != nil {
			item.ErrorStr = item.Error.Error()
		}
		items = append(items, itemMap[k])
	}
	br.Items = items
	return br
}

type BulkItemResponse struct {
	RuleName  string    `json:"ruleName,omitempty"`
	Status    Status    `json:"status,omitempty" description:"It may be success or error"`
	Result    Result    `json:"result,omitempty" description:"It may be created, updated or deleted, and only for successful operations"`
	ErrorType ErrorType `json:"errorType,omitempty" description:"It may be bad_data, duplicate_name, not_found or server_error, and only for failed operations"`
	Error     error     `json:"-"`
	ErrorStr  string    `json:"error,omitempty" description:"It is only returned for failed operations"`
}

func NewBulkItemSuccessResponse(ruleName string, result Result) *BulkItemResponse {
	return &BulkItemResponse{
		RuleName: ruleName,
		Status:   StatusSuccess,
		Result:   result,
	}
}

func NewBulkItemErrorServerResponse(ruleName string, err error) *BulkItemResponse {
	return &BulkItemResponse{
		RuleName:  ruleName,
		Status:    StatusError,
		ErrorType: ErrServer,
		Error:     err,
	}
}

// containsCaseInsensitive reports whether substr is case-insensitive within s.
func containsCaseInsensitive(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr))
}
