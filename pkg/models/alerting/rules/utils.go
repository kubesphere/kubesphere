package rules

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus-community/prom-label-proxy/injectproxy"
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prommodel "github.com/prometheus/common/model"
	promlabels "github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/rules"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
)

const (
	ErrGenRuleId = "error generating rule id"

	LabelKeyInternalRuleGroup    = "__rule_group__"
	LabelKeyInternalRuleName     = "__rule_name__"
	LabelKeyInternalRuleQuery    = "__rule_query__"
	LabelKeyInternalRuleDuration = "__rule_duration__"

	LabelKeyThanosRulerReplica = "thanos_ruler_replica"
	LabelKeyPrometheusReplica  = "prometheus_replica"

	LabelKeyAlertType   = "alerttype"
	LabelValueAlertType = "metric"
)

func FormatExpr(expr string) (string, error) {
	parsedExpr, err := parser.ParseExpr(expr)
	if err == nil {
		return parsedExpr.String(), nil
	}
	return "", errors.Wrapf(err, "failed to parse expr: %s", expr)
}

// InjectExprNamespaceLabel injects an label, whose key is "namespace" and whose value is the given namespace,
// into the prometheus query expression, which will limit the query scope.
func InjectExprNamespaceLabel(expr, namespace string) (string, error) {
	parsedExpr, err := parser.ParseExpr(expr)
	if err != nil {
		return "", err
	}
	if err = injectproxy.NewEnforcer(&promlabels.Matcher{
		Type:  promlabels.MatchEqual,
		Name:  "namespace",
		Value: namespace,
	}).EnforceNode(parsedExpr); err == nil {
		return parsedExpr.String(), nil
	}
	return "", err
}

func FormatDuration(for_ string) (string, error) {
	var duration prommodel.Duration
	var err error
	if for_ != "" {
		duration, err = prommodel.ParseDuration(for_)
		if err != nil {
			return "", errors.Wrapf(err, "failed to parse Duration string(\"%s\") to time.Duration", for_)
		}
	}
	return duration.String(), nil
}

func parseDurationSeconds(durationSeconds float64) string {
	return prommodel.Duration(int64(durationSeconds * float64(time.Second))).String()
}

func GenResourceRuleIdIgnoreFormat(group string, rule *promresourcesv1.Rule) string {
	query, err := FormatExpr(rule.Expr.String())
	if err != nil {
		klog.Warning(errors.Wrapf(err, "invalid alerting rule(%s)", rule.Alert))
		query = rule.Expr.String()
	}
	duration, err := FormatDuration(rule.For)
	if err != nil {
		klog.Warning(errors.Wrapf(err, "invalid alerting rule(%s)", rule.Alert))
		duration = rule.For
	}

	lbls := make(map[string]string)
	for k, v := range rule.Labels {
		lbls[k] = v
	}
	lbls[LabelKeyInternalRuleGroup] = group
	lbls[LabelKeyInternalRuleName] = rule.Alert
	lbls[LabelKeyInternalRuleQuery] = query
	lbls[LabelKeyInternalRuleDuration] = duration

	return prommodel.Fingerprint(prommodel.LabelsToSignature(lbls)).String()
}

func GenEndpointRuleId(group string, epRule *alerting.AlertingRule,
	externalLabels func() map[string]string) (string, error) {
	query, err := FormatExpr(epRule.Query)
	if err != nil {
		return "", err
	}
	duration := parseDurationSeconds(epRule.Duration)

	var extLabels map[string]string
	if externalLabels != nil {
		extLabels = externalLabels()
	}
	labelsMap := make(map[string]string)
	for key, value := range epRule.Labels {
		if key == LabelKeyPrometheusReplica || key == LabelKeyThanosRulerReplica {
			continue
		}
		if extLabels == nil {
			labelsMap[key] = value
			continue
		}
		if v, ok := extLabels[key]; !(ok && value == v) {
			labelsMap[key] = value
		}
	}

	lbls := make(map[string]string)
	for k, v := range labelsMap {
		lbls[k] = v
	}
	lbls[LabelKeyInternalRuleGroup] = group
	lbls[LabelKeyInternalRuleName] = epRule.Name
	lbls[LabelKeyInternalRuleQuery] = query
	lbls[LabelKeyInternalRuleDuration] = duration

	return prommodel.Fingerprint(prommodel.LabelsToSignature(lbls)).String(), nil
}

// GetAlertingRulesStatus mix rules from prometheusrule custom resources and rules from endpoints.
// Use rules from prometheusrule custom resources as the main reference.
func GetAlertingRulesStatus(ruleNamespace string, ruleChunk *ResourceRuleChunk, epRuleGroups []*alerting.RuleGroup,
	extLabels func() map[string]string) ([]*v2alpha1.GettableAlertingRule, error) {

	var (
		idEpRules = make(map[string]*alerting.AlertingRule)
		nameIds   = make(map[string][]string)
		ret       []*v2alpha1.GettableAlertingRule
	)
	for _, group := range epRuleGroups {
		fileShort := strings.TrimSuffix(filepath.Base(group.File), filepath.Ext(group.File))
		if !strings.HasPrefix(fileShort, ruleNamespace+"-") {
			continue
		}
		resourceRules, ok := ruleChunk.ResourceRulesMap[strings.TrimPrefix(fileShort, ruleNamespace+"-")]
		if !ok {
			continue
		}
		if _, ok := resourceRules.GroupSet[group.Name]; !ok {
			continue
		}

		for i, epRule := range group.Rules {
			if eid, err := GenEndpointRuleId(group.Name, epRule, extLabels); err != nil {
				return nil, errors.Wrap(err, ErrGenRuleId)
			} else {
				idEpRules[eid] = group.Rules[i]
				nameIds[epRule.Name] = append(nameIds[epRule.Name], eid)
			}
		}
	}
	if ruleChunk.Custom {
		// guarantee the names of the custom alerting rules not to be repeated
		var m = make(map[string][]*ResourceRuleItem)
		for _, resourceRules := range ruleChunk.ResourceRulesMap {
			for name, rrArr := range resourceRules.NameRules {
				m[name] = append(m[name], rrArr...)
			}
		}
		for _, rrArr := range m {
			if l := len(rrArr); l > 0 {
				if l > 1 {
					sort.Slice(rrArr, func(i, j int) bool {
						return v2alpha1.AlertingRuleIdCompare(rrArr[i].Id, rrArr[j].Id)
					})
				}
				resRule := rrArr[0]
				epRule := idEpRules[resRule.Id]
				if r := getAlertingRuleStatus(resRule, epRule, ruleChunk.Custom, ruleChunk.Level); r != nil {
					ret = append(ret, r)
				}
			}
		}
	} else {
		// guarantee the ids of the builtin alerting rules not to be repeated
		var m = make(map[string]*v2alpha1.GettableAlertingRule)
		for _, resourceRules := range ruleChunk.ResourceRulesMap {
			for id, rule := range resourceRules.IdRules {
				if r := getAlertingRuleStatus(rule, idEpRules[id], ruleChunk.Custom, ruleChunk.Level); r != nil {
					m[id] = r
				}
			}
		}
		for _, r := range m {
			ret = append(ret, r)
		}
	}

	return ret, nil
}

func GetAlertingRuleStatus(ruleNamespace string, rule *ResourceRule, epRuleGroups []*alerting.RuleGroup,
	extLabels func() map[string]string) (*v2alpha1.GettableAlertingRule, error) {

	if rule == nil || rule.Alert == "" {
		return nil, nil
	}

	var epRules = make(map[string]*alerting.AlertingRule)
	for _, group := range epRuleGroups {
		fileShort := strings.TrimSuffix(filepath.Base(group.File), filepath.Ext(group.File))
		if !strings.HasPrefix(fileShort, ruleNamespace+"-") {
			continue
		}
		if strings.TrimPrefix(fileShort, ruleNamespace+"-") != rule.ResourceName {
			continue
		}

		for _, epRule := range group.Rules {
			if eid, err := GenEndpointRuleId(group.Name, epRule, extLabels); err != nil {
				return nil, errors.Wrap(err, ErrGenRuleId)
			} else {
				if rule.Rule.Alert == epRule.Name {
					epRules[eid] = epRule
				}
			}
		}
	}
	var epRule *alerting.AlertingRule
	if rule.Custom {
		// guarantees the stability of the get operations.
		var ids []string
		for k, _ := range epRules {
			ids = append(ids, k)
		}
		if l := len(ids); l > 0 {
			if l > 1 {
				sort.Slice(ids, func(i, j int) bool {
					return v2alpha1.AlertingRuleIdCompare(ids[i], ids[j])
				})
			}
			epRule = epRules[ids[0]]
		}
	} else {
		epRule = epRules[rule.Id]
	}

	return getAlertingRuleStatus(&rule.ResourceRuleItem, epRule, rule.Custom, rule.Level), nil
}

func getAlertingRuleStatus(resRule *ResourceRuleItem, epRule *alerting.AlertingRule,
	custom bool, level v2alpha1.RuleLevel) *v2alpha1.GettableAlertingRule {

	if resRule == nil || resRule.Alert == "" {
		return nil
	}

	rule := v2alpha1.GettableAlertingRule{
		AlertingRule: v2alpha1.AlertingRule{
			Id:          resRule.Id,
			Name:        resRule.Rule.Alert,
			Query:       resRule.Rule.Expr.String(),
			Duration:    resRule.Rule.For,
			Labels:      resRule.Rule.Labels,
			Annotations: resRule.Rule.Annotations,
		},
		State:  stateInactiveString,
		Health: string(rules.HealthUnknown),
	}

	if epRule != nil {
		// The state information and alerts associated with the rule are from the rule from the endpoint.
		if epRule.Health != "" {
			rule.Health = epRule.Health
		}
		rule.LastError = epRule.LastError
		rule.LastEvaluation = epRule.LastEvaluation
		rule.EvaluationDurationSeconds = epRule.EvaluationTime

		rState := strings.ToLower(epRule.State)
		cliRuleStateEmpty := rState == ""
		if !cliRuleStateEmpty {
			rule.State = rState
		}
		for _, a := range epRule.Alerts {
			aState := strings.ToLower(a.State)
			if cliRuleStateEmpty {
				// for the rules gotten from prometheus or thanos ruler with a lower version, they may not contain
				// the state property, so compute the rule state by states of its alerts
				if alertState(rState) < alertState(aState) {
					rule.State = aState
				}
			}
			rule.Alerts = append(rule.Alerts, &v2alpha1.Alert{
				ActiveAt:    a.ActiveAt,
				Labels:      a.Labels,
				Annotations: a.Annotations,
				State:       aState,
				Value:       a.Value,

				RuleId:   rule.Id,
				RuleName: rule.Name,
			})
		}
	}
	return &rule
}

func ParseAlertingRules(epRuleGroups []*alerting.RuleGroup, custom bool, level v2alpha1.RuleLevel,
	filterFunc func(group, ruleId string, rule *alerting.AlertingRule) bool) ([]*v2alpha1.GettableAlertingRule, error) {

	var ret []*v2alpha1.GettableAlertingRule
	for _, g := range epRuleGroups {
		for _, r := range g.Rules {
			id, err := GenEndpointRuleId(g.Name, r, nil)
			if err != nil {
				return nil, err
			}
			if filterFunc(g.Name, id, r) {
				rule := &v2alpha1.GettableAlertingRule{
					AlertingRule: v2alpha1.AlertingRule{
						Id:          id,
						Name:        r.Name,
						Query:       r.Query,
						Duration:    parseDurationSeconds(r.Duration),
						Labels:      r.Labels,
						Annotations: r.Annotations,
					},
					State:                     r.State,
					Health:                    string(r.Health),
					LastError:                 r.LastError,
					LastEvaluation:            r.LastEvaluation,
					EvaluationDurationSeconds: r.EvaluationTime,
				}
				if rule.Health != "" {
					rule.Health = string(rules.HealthUnknown)
				}
				ruleStateEmpty := rule.State == ""
				rule.State = stateInactiveString
				for _, a := range r.Alerts {
					aState := strings.ToLower(a.State)
					if ruleStateEmpty {
						// for the rules gotten from prometheus or thanos ruler with a lower version, they may not contain
						// the state property, so compute the rule state by states of its alerts
						if alertState(rule.State) < alertState(aState) {
							rule.State = aState
						}
					}
					rule.Alerts = append(rule.Alerts, &v2alpha1.Alert{
						ActiveAt:    a.ActiveAt,
						Labels:      a.Labels,
						Annotations: a.Annotations,
						State:       aState,
						Value:       a.Value,

						RuleId:   rule.Id,
						RuleName: rule.Name,
					})
				}
				ret = append(ret, rule)
			}
		}
	}
	return ret, nil
}

var (
	statePendingString  = rules.StatePending.String()
	stateFiringString   = rules.StateFiring.String()
	stateInactiveString = rules.StateInactive.String()
)

func alertState(state string) rules.AlertState {
	switch state {
	case statePendingString:
		return rules.StatePending
	case stateFiringString:
		return rules.StateFiring
	case stateInactiveString:
		return rules.StateInactive
	}
	return rules.StateInactive
}
