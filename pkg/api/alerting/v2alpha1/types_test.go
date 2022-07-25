package v2alpha1

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/prometheus/prometheus/rules"
	"github.com/stretchr/testify/assert"
)

var (
	alertingRule = &AlertingRule{
		Id:       "ca7f09e76954e67c",
		Name:     "test-cpu-usage-high",
		Query:    `namespace:workload_cpu_usage:sum{namespace="test"} > 1`,
		Duration: "1m",
		Annotations: map[string]string{
			"alias":       "The alias is here",
			"description": "The description is here",
		},
		Labels: map[string]string{
			"flags":   "The flags is here",
			"cpu_num": "The cpu_num is here",
		},
	}

	postableAlertingRule = &PostableAlertingRule{
		AlertingRule: *alertingRule,
	}

	gettableAlertingRule = &GettableAlertingRule{
		AlertingRule: *alertingRule,
		Health:       string(rules.HealthGood),
		State:        rules.StateInactive.String(),
	}

	gettableAlertingRules = []*GettableAlertingRule{
		{
			AlertingRule:              *alertingRule,
			Health:                    string(rules.HealthGood),
			State:                     rules.StateInactive.String(),
			EvaluationDurationSeconds: 1,
			LastEvaluation:            &time.Time{},
		},
		{
			AlertingRule: AlertingRule{
				Id:       "ca7f09e76954e688",
				Name:     "test-cpu-usage-high-2",
				Query:    `namespace:workload_cpu_usage:sum{namespace="test"} > 1`,
				Duration: "1m",
				Annotations: map[string]string{
					"alias":       "The alias is here",
					"description": "The description is here",
				},
				Labels: map[string]string{
					"flags":   "The flags is here",
					"cpu_num": "The cpu_num is here",
				},
			},
			State:                     rules.StateInactive.String(),
			Health:                    string(rules.HealthGood),
			EvaluationDurationSeconds: 0,
			LastEvaluation:            &time.Time{},
		},
	}

	alertingRuleQueryParams = &AlertingRuleQueryParams{
		NameContainFilter: "test-cpu-usage-high",
		State:             rules.StateInactive.String(),
		Health:            string(rules.HealthGood),
		LabelEqualFilters: map[string]string{
			"flags":   "The flags is here",
			"cpu_num": "The cpu_num is here",
		},
		LabelContainFilters: map[string]string{
			"alias":       "The alias is here",
			"description": "The description is here",
		},
		PageNum:   1,
		Limit:     10,
		SortField: "name",
		SortType:  "desc",
	}
)

func TestPostableAlertingRule_Validate(t *testing.T) {
	// validate AlertingRule name field
	// Name is empty
	postableAlertingRule.Name = ""
	err := postableAlertingRule.Validate()
	assert.Equal(t, "name can not be empty", err.Error())

	// Name do not match regx
	postableAlertingRule.Name = "TestCPUUsageHigh"
	err = postableAlertingRule.Validate()
	assert.Equal(t, "rule name must match regular expression ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", err.Error())

	// validate AlertingRule Query field
	postableAlertingRule.Name = "test-cpu-usage-high"
	// Query is empty
	postableAlertingRule.Query = ""
	err = postableAlertingRule.Validate()
	assert.Equal(t, "query can not be empty", err.Error())

	postableAlertingRule.Query = `namespace:workload_cpu_usage:sum{namespace="test"} > 1`
	// test no error

	err = postableAlertingRule.Validate()
	assert.Empty(t, err)

}

func TestAlertQueryParams_Filter(t *testing.T) {
	// empty
	gettableAlertingRules := []*GettableAlertingRule{}
	queryParams := &AlertingRuleQueryParams{}
	ret := queryParams.Filter(gettableAlertingRules)
	assert.Empty(t, ret)

	queryParams.NameContainFilter = "test-cpu-usage-high"
	gettableAlertingRules = []*GettableAlertingRule{
		gettableAlertingRule,
	}
	ret = queryParams.Filter(gettableAlertingRules)
	assert.NotEmpty(t, ret)
}

func TestAlertQueryParams_match(t *testing.T) {
	// empty
	rule := &GettableAlertingRule{}
	queryParams := &AlertingRuleQueryParams{}
	ret := queryParams.matches(rule)
	assert.True(t, ret)

	// Not empty
	// test false case
	queryParams.NameContainFilter = "test-cpu"
	ret = queryParams.matches(rule)
	assert.False(t, ret)

	//test true case
	rule = gettableAlertingRule
	queryParams.NameContainFilter = "test-cpu-usage-high"
	ret = queryParams.matches(rule)
	assert.True(t, ret)
}

func TestAlertingRuleIdCompare(t *testing.T) {
	leftId := "test-id-1"
	rightId := "test-id-2"

	ret := AlertingRuleIdCompare(leftId, rightId)
	assert.True(t, ret)
}

func TestAlertQueryParams_Sort(t *testing.T) {
	queryParams := alertingRuleQueryParams
	rules := []*GettableAlertingRule{
		gettableAlertingRule,
		gettableAlertingRule,
	}

	// sort by name
	queryParams.Sort(rules)

	// sort by lastEvaluation
	queryParams.SortField = "lastEvaluation"
	queryParams.Sort(rules)

	// sort by evaluationTime
	queryParams.SortField = "evaluationTime"
	queryParams.Sort(rules)
}

func TestAlertQueryParams_Sub(t *testing.T) {
	rules := gettableAlertingRules
	queryParams := alertingRuleQueryParams
	queryParams.Sub(rules)
}

var (
	alertQueryParams = &AlertQueryParams{
		State: rules.StateFiring.String(),
		LabelEqualFilters: map[string]string{
			"alias":       "The alias is here",
			"description": "The description is here",
		},
		LabelContainFilters: map[string]string{
			"alias":       "The alias is here",
			"description": "The description is here",
		},
		PageNum: 1,
		Limit:   10,
	}

	alert = &Alert{
		RuleId:   "ca7f09e76954e67c",
		RuleName: "test-cpu-usage-high",
		State:    rules.StateFiring.String(),
		Value:    `namespace:workload_cpu_usage:sum{namespace="test"} > 1`,
		Annotations: map[string]string{
			"alias":       "The alias is here",
			"description": "The description is here",
		},
		Labels: map[string]string{
			"alias":       "The alias is here",
			"description": "The description is here",
		},
		ActiveAt: &time.Time{},
	}

	alerts = []*Alert{
		alert,
		{
			RuleId:   "ca7f09e76954e699",
			RuleName: "test-cpu-usage-high-2",
			State:    rules.StateFiring.String(),
			Value:    `namespace:workload_cpu_usage:sum{namespace="test"} > 1`,
			Annotations: map[string]string{
				"alias":       "The alias is here",
				"description": "The description is here",
			},
			Labels: map[string]string{
				"alias":       "The alias is here",
				"description": "The description is here",
			},
			ActiveAt: &time.Time{},
		},
	}
)

func TestAlertingRuleQueryParams_Filter(t *testing.T) {
	// empty
	// Alert Array is empty
	loacalAlerts := []*Alert{}
	queryParam := &AlertQueryParams{}
	ret := queryParam.Filter(loacalAlerts)
	assert.Empty(t, ret)

	// Alert Array has nil
	loacalAlerts = []*Alert{nil}
	ret = queryParam.Filter(loacalAlerts)
	assert.Empty(t, ret)

	// all pass
	loacalAlerts = alerts
	ret = queryParam.Filter(loacalAlerts)
	assert.NotEmpty(t, ret)
}

func TestAlertingRuleQueryParams_matches(t *testing.T) {
	loacalAlert := alert
	queryParam := alertQueryParams
	queryParam.LabelEqualFilters = map[string]string{
		"flags":   "The flags is here",
		"cpu_num": "The cpu_num is here",
	}
	ret := queryParam.matches(loacalAlert)
	assert.False(t, ret)

	queryParam.LabelEqualFilters = map[string]string{
		"alias":       "The alias is here",
		"description": "The description is here",
	}
	ret = queryParam.matches(loacalAlert)
	assert.True(t, ret)
}

func TestAlertingRuleQueryParams_Sort(t *testing.T) {
	loacalAlerts := alerts
	queryParam := alertQueryParams
	queryParam.Sort(loacalAlerts)
}

func TestAlertingRuleQueryParams_Sub(t *testing.T) {
	loacalAlerts := alerts
	queryParam := alertQueryParams
	queryParam.Sub(loacalAlerts)
}

var (
	succBulkItemResponse = BulkItemResponse{
		RuleName:  "test-cpu-usage-high",
		Status:    StatusSuccess,
		Result:    ResultCreated,
		ErrorType: "",
		Error:     nil,
		ErrorStr:  "",
	}

	errBulkItemResponse = BulkItemResponse{
		RuleName:  "test-mem-usage-high",
		Status:    StatusError,
		Result:    ResultUpdated,
		ErrorType: ErrBadData,
		Error:     errors.New(string(ErrBadData)),
		ErrorStr:  string(ErrBadData),
	}

	bulkResponse = BulkResponse{
		Items: []*BulkItemResponse{
			&succBulkItemResponse,
			&errBulkItemResponse,
		},
	}
)

func TestParseAlertingRuleQueryParams(t *testing.T) {
	queryParam := "name=test-cpu&state=firing&health=ok&sort_field=lastEvaluation&sort_type=desc&label_filters=name~test"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/alerting.kubesphere.io/v2alpha1/rules?%s", queryParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := AlertingRuleQueryParams{
		NameContainFilter:   "test-cpu",
		State:               "firing",
		Health:              "ok",
		LabelEqualFilters:   make(map[string]string),
		LabelContainFilters: map[string]string{"name": "test"},

		PageNum:   1,
		Limit:     10,
		SortField: "lastEvaluation",
		SortType:  "desc",
	}

	request := restful.NewRequest(req)
	actual, err := ParseAlertingRuleQueryParams(request)
	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}

func TestParseAlertQueryParams(t *testing.T) {
	queryParam := "state=firing&label_filters=name~test"
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/alerting.kubesphere.io/v2alpha1/alerts?%s", queryParam), nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := AlertQueryParams{
		State:               "firing",
		LabelEqualFilters:   make(map[string]string),
		LabelContainFilters: map[string]string{"name": "test"},

		PageNum: 1,
		Limit:   10,
	}

	request := restful.NewRequest(req)
	actual, err := ParseAlertQueryParams(request)
	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}

func TestBulkResponse_MakeBulkResponse(t *testing.T) {
	br := bulkResponse.MakeBulkResponse()
	assert.True(t, br.Errors)
}

func TestBulkResponse_NewBulkItemSuccessResponse(t *testing.T) {
	ruleName := "test-cpu-usage-high"
	result := ResultCreated
	ret := NewBulkItemSuccessResponse(ruleName, result)
	assert.Equal(t, ResultCreated, ret.Result)
}

func TestBulkResponse_NewBulkItemErrorServerResponse(t *testing.T) {
	ruleName := "test-cpu-usage-high"
	err := errors.New(string(ErrBadData))
	ret := NewBulkItemErrorServerResponse(ruleName, err)
	assert.Equal(t, errors.New(string(ErrBadData)), ret.Error)
}

func TestContainsCaseInsensitive(t *testing.T) {
	left := "left"
	right := "RIGHT"
	ret := containsCaseInsensitive(left, right)
	assert.False(t, ret)
}
