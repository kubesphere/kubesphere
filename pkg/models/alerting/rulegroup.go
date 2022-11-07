// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package alerting

import (
	"context"
	"fmt"
	"time"

	promlabels "github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
	promrules "github.com/prometheus/prometheus/rules"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"

	alertingv2beta1 "kubesphere.io/api/alerting/v2beta1"

	"kubesphere.io/kubesphere/pkg/api"
	kapialertingv2beta1 "kubesphere.io/kubesphere/pkg/api/alerting/v2beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	alertinglisters "kubesphere.io/kubesphere/pkg/client/listers/alerting/v2beta1"
	controller "kubesphere.io/kubesphere/pkg/controller/alerting"
	"kubesphere.io/kubesphere/pkg/informers"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
)

type RuleGroupOperator interface {
	ListRuleGroups(ctx context.Context, namespace string, queryParam *query.Query) (*api.ListResult, error)
	GetRuleGroup(ctx context.Context, namespace, name string) (*kapialertingv2beta1.RuleGroup, error)
	ListAlerts(ctx context.Context, namespace string, queryParam *query.Query) (*api.ListResult, error)

	ListGlobalRuleGroups(ctx context.Context, queryParam *query.Query) (*api.ListResult, error)
	GetGlobalRuleGroup(ctx context.Context, name string) (*kapialertingv2beta1.GlobalRuleGroup, error)
	ListGlobalAlerts(ctx context.Context, queryParam *query.Query) (*api.ListResult, error)

	ListClusterRuleGroups(ctx context.Context, queryParam *query.Query) (*api.ListResult, error)
	GetClusterRuleGroup(ctx context.Context, name string) (*kapialertingv2beta1.ClusterRuleGroup, error)
	ListClusterAlerts(ctx context.Context, queryParam *query.Query) (*api.ListResult, error)
}

func NewRuleGroupOperator(informers informers.InformerFactory, ruleClient alerting.RuleClient) RuleGroupOperator {
	return &ruleGroupOperator{
		ruleClient:             ruleClient,
		ruleGroupLister:        informers.KubeSphereSharedInformerFactory().Alerting().V2beta1().RuleGroups().Lister(),
		clusterRuleGroupLister: informers.KubeSphereSharedInformerFactory().Alerting().V2beta1().ClusterRuleGroups().Lister(),
		globalRuleGroupLister:  informers.KubeSphereSharedInformerFactory().Alerting().V2beta1().GlobalRuleGroups().Lister(),
	}
}

type ruleGroupOperator struct {
	ruleClient             alerting.RuleClient
	ruleGroupLister        alertinglisters.RuleGroupLister
	clusterRuleGroupLister alertinglisters.ClusterRuleGroupLister
	globalRuleGroupLister  alertinglisters.GlobalRuleGroupLister
}

func (o *ruleGroupOperator) listRuleGroups(ctx context.Context, namespace string, selector labels.Selector) ([]runtime.Object, error) {
	resourceRuleGroups, err := o.ruleGroupLister.RuleGroups(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	// get rules matching '{rule_level="namespace",namespace="<namespace>"}' from thanos ruler
	matchers := []*promlabels.Matcher{{
		Type:  promlabels.MatchEqual,
		Name:  controller.RuleLabelKeyRuleLevel,
		Value: string(controller.RuleLevelNamesapce),
	}, {
		Type:  promlabels.MatchEqual,
		Name:  controller.RuleLabelKeyNamespace,
		Value: namespace,
	}}
	statusRuleGroups, err := o.ruleClient.ThanosRules(ctx, matchers)
	if err != nil {
		return nil, err
	}
	var statusRuleGroupMap = make(map[string]*alerting.RuleGroup)
	for i := range statusRuleGroups {
		g := statusRuleGroups[i]
		// the matchers only filter rules and all groups still return,
		// and some of them may be with empty rules, so here check them and skip some.
		if len(g.Rules) == 0 {
			continue
		}
		if _, ok := statusRuleGroupMap[g.Name]; !ok {
			statusRuleGroupMap[g.Name] = g
		}
	}

	// copy status info of statusRuleGroups to matched rulegroups
	var groups = make([]runtime.Object, len(resourceRuleGroups))
	for i := range resourceRuleGroups {
		setRuleGroupResourceTypeMeta(resourceRuleGroups[i])
		g := &kapialertingv2beta1.RuleGroup{
			RuleGroup: *resourceRuleGroups[i],
			Status: kapialertingv2beta1.RuleGroupStatus{
				RulesStatus: make([]kapialertingv2beta1.RuleStatus, len(resourceRuleGroups[i].Spec.Rules)),
				RulesStats:  kapialertingv2beta1.RulesStats{},
			},
		}
		statusg, ok := statusRuleGroupMap[g.Name]
		specRules := g.Spec.Rules
		if ok && len(statusg.Rules) > 0 {
			var ruleIds = make([]string, len(specRules))
			var ruleDisableFlags = make([]bool, len(specRules))
			for j := range specRules {
				if specRules[j].Labels != nil {
					ruleIds[j] = specRules[j].Labels[alertingv2beta1.RuleLabelKeyRuleId]
				}
				ruleDisableFlags[j] = specRules[j].Disable
			}

			copyRuleGroupStatus(statusg, &g.Status, ruleIds, ruleDisableFlags)
		} else {
			// for rules not loaded by rule reloader (eg.thanos) yet
			for j := range specRules {
				ruleStatus := kapialertingv2beta1.RuleStatus{Health: string(promrules.HealthUnknown)}
				if specRules[j].Disable {
					ruleStatus.State = stateDisabledString
					g.Status.RulesStats.Disabled++
				} else {
					ruleStatus.State = stateInactiveString
					g.Status.RulesStats.Inactive++
				}
				g.Status.RulesStatus[j] = ruleStatus
			}
		}
		groups[i] = g
	}
	return groups, nil
}

func (o *ruleGroupOperator) ListRuleGroups(ctx context.Context, namespace string,
	queryParam *query.Query) (*api.ListResult, error) {

	groups, err := o.listRuleGroups(ctx, namespace, queryParam.Selector())
	if err != nil {
		return nil, err
	}

	listResult := resources.DefaultList(groups, queryParam, func(left, right runtime.Object, field query.Field) bool {
		hit, great := o.compareRuleGroupStatus(
			&(left.(*kapialertingv2beta1.RuleGroup).Status), &(right.(*kapialertingv2beta1.RuleGroup).Status), field)
		if hit {
			return great
		}
		return resources.DefaultObjectMetaCompare(
			left.(*kapialertingv2beta1.RuleGroup).ObjectMeta, right.(*kapialertingv2beta1.RuleGroup).ObjectMeta, field)
	}, func(obj runtime.Object, filter query.Filter) bool {
		hit, selected := o.filterRuleGroupStatus(&obj.(*kapialertingv2beta1.RuleGroup).Status, filter)
		if hit {
			return selected
		}
		return resources.DefaultObjectMetaFilter(obj.(*kapialertingv2beta1.RuleGroup).ObjectMeta, filter)
	})

	return listResult, nil
}

// compareRuleGroupStatus compare rulegroup status.
// if field in status, return hit(true) and great(true if left great than right, else false).
// if filed not in status, return hit(false) and great(false, should be unuseful).
func (d *ruleGroupOperator) compareRuleGroupStatus(left, right *kapialertingv2beta1.RuleGroupStatus, field query.Field) (hit, great bool) {

	switch field {
	case kapialertingv2beta1.FieldRuleGroupEvaluationTime:
		hit = true
		if left.EvaluationTime == nil {
			great = false
		} else if right.EvaluationTime == nil {
			great = true
		} else {
			great = *left.EvaluationTime > *right.EvaluationTime
		}
	case kapialertingv2beta1.FieldRuleGroupLastEvaluation:
		hit = true
		if left.LastEvaluation == nil {
			great = false
		} else if right.LastEvaluation == nil {
			great = true
		} else {
			great = left.LastEvaluation.After(*right.LastEvaluation)
		}
	}
	return
}

// filterRuleGroupStatus filters by rulegroup status.
// if field in status, return hit(true) and selected(true if match the filter, else false).
// if filed not in status, return hit(false) and selected(false, should be unuseful).
func (d *ruleGroupOperator) filterRuleGroupStatus(status *kapialertingv2beta1.RuleGroupStatus, filter query.Filter) (hit, selected bool) {

	switch filter.Field {
	case kapialertingv2beta1.FieldState:
		hit = true
		switch string(filter.Value) {
		case stateDisabledString:
			selected = status.RulesStats.Disabled > 0
		case stateInactiveString:
			selected = status.RulesStats.Inactive > 0
		case statePendingString:
			selected = status.RulesStats.Pending > 0
		case stateFiringString:
			selected = status.RulesStats.Firing > 0
		case "":
			selected = true
		default:
			selected = false
		}
	}
	return
}

func (o *ruleGroupOperator) ListAlerts(ctx context.Context, namespace string,
	queryParam *query.Query) (*api.ListResult, error) {

	groups, err := o.listRuleGroups(ctx, namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	// encapsulated as runtime.Object for easy comparison and filtering.
	var alerts []runtime.Object
	for i := range groups {
		g := groups[i].(*kapialertingv2beta1.RuleGroup)
		for j := range g.Status.RulesStatus {
			ruleStatus := g.Status.RulesStatus[j]
			for k := range ruleStatus.Alerts {
				alerts = append(alerts, &wrapAlert{Alert: *ruleStatus.Alerts[k]})
			}
		}
	}

	filterAlert, err := o.createFilterAlertFunc(queryParam)
	if err != nil {
		return nil, err
	}
	listResult := resources.DefaultList(alerts, queryParam, func(left, right runtime.Object, field query.Field) bool {
		return o.compareAlert(&left.(*wrapAlert).Alert, &right.(*wrapAlert).Alert, field)
	}, func(obj runtime.Object, filter query.Filter) bool {
		return filterAlert(&obj.(*wrapAlert).Alert, filter)
	})
	for i := range listResult.Items {
		listResult.Items[i] = &listResult.Items[i].(*wrapAlert).Alert
	}
	return listResult, nil
}

func (d *ruleGroupOperator) compareAlert(left, right *kapialertingv2beta1.Alert, field query.Field) bool {
	switch field {
	case kapialertingv2beta1.FieldAlertActiveAt:
		if left.ActiveAt == nil {
			return false
		}
		if right.ActiveAt == nil {
			return true
		}
		return left.ActiveAt.After(*right.ActiveAt)
	}
	return false
}

func (d *ruleGroupOperator) createFilterAlertFunc(queryParam *query.Query) (func(alert *kapialertingv2beta1.Alert, filter query.Filter) bool, error) {
	var labelFilters kapialertingv2beta1.LabelFilters
	var labelMatchers []*promlabels.Matcher
	var err error
	if len(queryParam.Filters) > 0 {
		if filters, ok := queryParam.Filters[kapialertingv2beta1.FieldAlertLabelFilters]; ok {
			labelFilters = kapialertingv2beta1.ParseLabelFilters(string(filters))
		}
		if matcherStr, ok := queryParam.Filters[kapialertingv2beta1.FieldAlertLabelMatcher]; ok {
			labelMatchers, err = parser.ParseMetricSelector(string(matcherStr))
			if err != nil {
				return nil, fmt.Errorf("invalid %s param: %v", kapialertingv2beta1.FieldAlertLabelMatcher, err)
			}
		}
	}
	return func(alert *kapialertingv2beta1.Alert, filter query.Filter) bool {
		switch filter.Field {
		case kapialertingv2beta1.FieldAlertLabelFilters:
			if labelFilters == nil {
				return true
			}
			return labelFilters.Matches(alert.Labels)
		case kapialertingv2beta1.FieldAlertLabelMatcher:
			for _, m := range labelMatchers {
				var v string
				if len(alert.Labels) > 0 {
					v = alert.Labels[m.Name]
				}
				if !m.Matches(v) {
					return false
				}
			}
			return true
		case kapialertingv2beta1.FieldState:
			return alert.State == string(filter.Value)
		}
		return false
	}, nil
}

func (o *ruleGroupOperator) GetRuleGroup(ctx context.Context, namespace, name string) (*kapialertingv2beta1.RuleGroup, error) {
	resourceRuleGroup, err := o.ruleGroupLister.RuleGroups(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	setRuleGroupResourceTypeMeta(resourceRuleGroup)

	ret := &kapialertingv2beta1.RuleGroup{
		RuleGroup: *resourceRuleGroup,
		Status: kapialertingv2beta1.RuleGroupStatus{
			RulesStatus: make([]kapialertingv2beta1.RuleStatus, len(resourceRuleGroup.Spec.Rules)),
			RulesStats:  kapialertingv2beta1.RulesStats{},
		},
	}

	matchers := []*promlabels.Matcher{{
		Type:  promlabels.MatchEqual,
		Name:  controller.RuleLabelKeyRuleLevel,
		Value: string(controller.RuleLevelNamesapce),
	}, {
		Type:  promlabels.MatchEqual,
		Name:  controller.RuleLabelKeyNamespace,
		Value: namespace,
	}}
	statusRuleGroups, err := o.ruleClient.ThanosRules(ctx, matchers)
	if err != nil {
		return nil, err
	}

	var setStatus bool
	specRules := resourceRuleGroup.Spec.Rules
	for _, g := range statusRuleGroups {
		if g.Name == resourceRuleGroup.Name && len(g.Rules) > 0 {
			var ruleIds = make([]string, len(specRules))
			var ruleDisableFlags = make([]bool, len(specRules))
			for j := range specRules {
				if specRules[j].Labels != nil {
					ruleIds[j] = specRules[j].Labels[alertingv2beta1.RuleLabelKeyRuleId]
				}
				ruleDisableFlags[j] = specRules[j].Disable
			}

			copyRuleGroupStatus(g, &ret.Status, ruleIds, ruleDisableFlags)
			setStatus = true
			break
		}
	}
	if !setStatus {
		// for rules not loaded by rule reloader (eg.thanos) yet
		for j := range ret.Spec.Rules {
			ruleStatus := kapialertingv2beta1.RuleStatus{Health: string(promrules.HealthUnknown)}
			if specRules[j].Disable {
				ruleStatus.State = stateDisabledString
				ret.Status.RulesStats.Disabled++
			} else {
				ruleStatus.State = stateInactiveString
				ret.Status.RulesStats.Inactive++
			}
			ret.Status.RulesStatus[j] = ruleStatus
		}
	}

	return ret, nil
}

func (o *ruleGroupOperator) listClusterRuleGroups(ctx context.Context, selector labels.Selector) ([]runtime.Object, error) {
	resourceRuleGroups, err := o.clusterRuleGroupLister.List(selector)
	if err != nil {
		return nil, err
	}

	// get rules matching '{rule_level="cluster"}' from thanos ruler
	matchers := []*promlabels.Matcher{{
		Type:  promlabels.MatchEqual,
		Name:  controller.RuleLabelKeyRuleLevel,
		Value: string(controller.RuleLevelCluster),
	}}
	statusRuleGroups, err := o.ruleClient.ThanosRules(ctx, matchers)
	if err != nil {
		return nil, err
	}
	var statusRuleGroupMap = make(map[string]*alerting.RuleGroup)
	for i := range statusRuleGroups {
		g := statusRuleGroups[i]
		// the matchers only filter rules and all groups still return,
		// and some of them may be with empty rules, so here check them and skip some.
		if len(g.Rules) == 0 {
			continue
		}
		if _, ok := statusRuleGroupMap[g.Name]; !ok {
			statusRuleGroupMap[g.Name] = g
		}
	}
	// copy status info of statusRuleGroups to matched rulegroups
	var groups = make([]runtime.Object, len(resourceRuleGroups))
	for i := range resourceRuleGroups {
		setRuleGroupResourceTypeMeta(resourceRuleGroups[i])
		g := &kapialertingv2beta1.ClusterRuleGroup{
			ClusterRuleGroup: *resourceRuleGroups[i],
			Status: kapialertingv2beta1.RuleGroupStatus{
				RulesStatus: make([]kapialertingv2beta1.RuleStatus, len(resourceRuleGroups[i].Spec.Rules)),
				RulesStats:  kapialertingv2beta1.RulesStats{},
			},
		}
		statusg, ok := statusRuleGroupMap[g.Name]
		specRules := g.Spec.Rules
		if ok && len(statusg.Rules) == len(specRules) {
			var ruleIds = make([]string, len(specRules))
			var ruleDisableFlags = make([]bool, len(specRules))
			for j := range specRules {
				if specRules[j].Labels != nil {
					ruleIds[j] = specRules[j].Labels[alertingv2beta1.RuleLabelKeyRuleId]
				}
				ruleDisableFlags[j] = specRules[j].Disable
			}

			copyRuleGroupStatus(statusg, &g.Status, ruleIds, ruleDisableFlags)
		} else {
			// for rules not loaded by rule reloader (eg.thanos) yet
			for j := range g.Spec.Rules {
				ruleStatus := kapialertingv2beta1.RuleStatus{Health: string(promrules.HealthUnknown)}
				if specRules[j].Disable {
					ruleStatus.State = stateDisabledString
					g.Status.RulesStats.Disabled++
				} else {
					ruleStatus.State = stateInactiveString
					g.Status.RulesStats.Inactive++
				}
				g.Status.RulesStatus[j] = ruleStatus
			}
		}
		groups[i] = g
	}
	return groups, nil
}

func (o *ruleGroupOperator) ListClusterRuleGroups(ctx context.Context,
	queryParam *query.Query) (*api.ListResult, error) {

	groups, err := o.listClusterRuleGroups(ctx, queryParam.Selector())
	if err != nil {
		return nil, err
	}

	listResult := resources.DefaultList(groups, queryParam, func(left, right runtime.Object, field query.Field) bool {
		hit, great := o.compareRuleGroupStatus(
			&(left.(*kapialertingv2beta1.ClusterRuleGroup).Status), &(right.(*kapialertingv2beta1.ClusterRuleGroup).Status), field)
		if hit {
			return great
		}
		return resources.DefaultObjectMetaCompare(
			left.(*kapialertingv2beta1.ClusterRuleGroup).ObjectMeta, right.(*kapialertingv2beta1.ClusterRuleGroup).ObjectMeta, field)
	}, func(obj runtime.Object, filter query.Filter) bool {
		hit, selected := o.filterRuleGroupStatus(&obj.(*kapialertingv2beta1.ClusterRuleGroup).Status, filter)
		if hit {
			return selected
		}
		return resources.DefaultObjectMetaFilter(obj.(*kapialertingv2beta1.ClusterRuleGroup).ObjectMeta, filter)
	})

	return listResult, nil
}

func (o *ruleGroupOperator) ListClusterAlerts(ctx context.Context,
	queryParam *query.Query) (*api.ListResult, error) {

	groups, err := o.listClusterRuleGroups(ctx, labels.Everything())
	if err != nil {
		return nil, err
	}

	// encapsulated as runtime.Object for easy comparison and filtering.
	var alerts []runtime.Object
	for i := range groups {
		g := groups[i].(*kapialertingv2beta1.ClusterRuleGroup)
		for j := range g.Status.RulesStatus {
			ruleStatus := g.Status.RulesStatus[j]
			for k := range ruleStatus.Alerts {
				alerts = append(alerts, &wrapAlert{Alert: *ruleStatus.Alerts[k]})
			}
		}
	}

	filterAlert, err := o.createFilterAlertFunc(queryParam)
	if err != nil {
		return nil, err
	}
	listResult := resources.DefaultList(alerts, queryParam, func(left, right runtime.Object, field query.Field) bool {
		return o.compareAlert(&left.(*wrapAlert).Alert, &right.(*wrapAlert).Alert, field)
	}, func(obj runtime.Object, filter query.Filter) bool {
		return filterAlert(&obj.(*wrapAlert).Alert, filter)
	})
	for i := range listResult.Items {
		listResult.Items[i] = &listResult.Items[i].(*wrapAlert).Alert
	}
	return listResult, nil
}

func (o *ruleGroupOperator) GetClusterRuleGroup(ctx context.Context, name string) (*kapialertingv2beta1.ClusterRuleGroup, error) {
	resourceRuleGroup, err := o.clusterRuleGroupLister.Get(name)
	if err != nil {
		return nil, err
	}

	setRuleGroupResourceTypeMeta(resourceRuleGroup)

	ret := &kapialertingv2beta1.ClusterRuleGroup{
		ClusterRuleGroup: *resourceRuleGroup,
		Status: kapialertingv2beta1.RuleGroupStatus{
			RulesStatus: make([]kapialertingv2beta1.RuleStatus, len(resourceRuleGroup.Spec.Rules)),
			RulesStats:  kapialertingv2beta1.RulesStats{},
		},
	}

	matchers := []*promlabels.Matcher{{
		Type:  promlabels.MatchEqual,
		Name:  controller.RuleLabelKeyRuleLevel,
		Value: string(controller.RuleLevelCluster),
	}}
	statusRuleGroups, err := o.ruleClient.ThanosRules(ctx, matchers)
	if err != nil {
		return nil, err
	}

	var setStatus bool
	specRules := resourceRuleGroup.Spec.Rules
	for _, g := range statusRuleGroups {
		if g.Name == resourceRuleGroup.Name && len(g.Rules) == len(specRules) {
			var ruleIds = make([]string, len(specRules))
			var ruleDisableFlags = make([]bool, len(specRules))
			for j := range specRules {
				if specRules[j].Labels != nil {
					ruleIds[j] = specRules[j].Labels[alertingv2beta1.RuleLabelKeyRuleId]
				}
				ruleDisableFlags[j] = specRules[j].Disable
			}

			copyRuleGroupStatus(g, &ret.Status, ruleIds, ruleDisableFlags)
			setStatus = true
			break
		}
	}
	if !setStatus {
		// for rules not loaded by rule reloader (eg.thanos) yet
		for j := range ret.Spec.Rules {
			ruleStatus := kapialertingv2beta1.RuleStatus{Health: string(promrules.HealthUnknown)}
			if specRules[j].Disable {
				ruleStatus.State = stateDisabledString
				ret.Status.RulesStats.Disabled++
			} else {
				ruleStatus.State = stateInactiveString
				ret.Status.RulesStats.Inactive++
			}
			ret.Status.RulesStatus[j] = ruleStatus
		}
	}

	return ret, nil
}

func (o *ruleGroupOperator) listGlobalRuleGroups(ctx context.Context, selector labels.Selector) ([]runtime.Object, error) {
	resourceRuleGroups, err := o.globalRuleGroupLister.List(selector)
	if err != nil {
		return nil, err
	}

	// get rules matching '{rule_level="global"}' from thanos ruler
	matchers := []*promlabels.Matcher{{
		Type:  promlabels.MatchEqual,
		Name:  controller.RuleLabelKeyRuleLevel,
		Value: string(controller.RuleLevelGlobal),
	}}
	statusRuleGroups, err := o.ruleClient.ThanosRules(ctx, matchers)
	if err != nil {
		return nil, err
	}
	var statusRuleGroupMap = make(map[string]*alerting.RuleGroup)
	for i := range statusRuleGroups {
		g := statusRuleGroups[i]
		// the matchers only filter rules and all groups still return,
		// and some of them may be with empty rules, so here check them and skip some.
		if len(g.Rules) == 0 {
			continue
		}
		if _, ok := statusRuleGroupMap[g.Name]; !ok {
			statusRuleGroupMap[g.Name] = g
		}
	}
	// copy status info of statusRuleGroups to matched rulegroups
	var groups = make([]runtime.Object, len(resourceRuleGroups))
	for i := range resourceRuleGroups {
		setRuleGroupResourceTypeMeta(resourceRuleGroups[i])
		g := &kapialertingv2beta1.GlobalRuleGroup{
			GlobalRuleGroup: *resourceRuleGroups[i],
			Status: kapialertingv2beta1.RuleGroupStatus{
				RulesStatus: make([]kapialertingv2beta1.RuleStatus, len(resourceRuleGroups[i].Spec.Rules)),
				RulesStats:  kapialertingv2beta1.RulesStats{},
			},
		}
		statusg, ok := statusRuleGroupMap[g.Name]
		specRules := g.Spec.Rules
		if ok && len(statusg.Rules) == len(specRules) {
			var ruleIds = make([]string, len(specRules))
			var ruleDisableFlags = make([]bool, len(specRules))
			for j := range specRules {
				if specRules[j].Labels != nil {
					ruleIds[j] = specRules[j].Labels[alertingv2beta1.RuleLabelKeyRuleId]
				}
				ruleDisableFlags[j] = specRules[j].Disable
			}

			copyRuleGroupStatus(statusg, &g.Status, ruleIds, ruleDisableFlags)
			for j := range g.Status.RulesStatus {
				// for rules disabled and etc.
				if g.Status.RulesStatus[j].Expr == "" {
					rule := g.Spec.Rules[j]
					enforceExprFunc := controller.CreateEnforceExprFunc(controller.ParseGlobalRuleEnforceMatchers(&rule))
					expr, err := enforceExprFunc(rule.Expr.String())
					if err != nil {
						return nil, err
					}
					g.Status.RulesStatus[j].Expr = expr
				}
			}
		} else {
			// for rules not loaded by rule reloader (eg.thanos) yet
			for j, rule := range specRules {
				ruleStatus := kapialertingv2beta1.RuleStatus{Health: string(promrules.HealthUnknown)}
				if specRules[j].Disable {
					ruleStatus.State = stateDisabledString
					g.Status.RulesStats.Disabled++
				} else {
					ruleStatus.State = stateInactiveString
					g.Status.RulesStats.Inactive++
				}
				enforceExprFunc := controller.CreateEnforceExprFunc(controller.ParseGlobalRuleEnforceMatchers(&rule))
				expr, err := enforceExprFunc(rule.Expr.String())
				if err != nil {
					return nil, err
				}
				ruleStatus.Expr = expr
				g.Status.RulesStatus[j] = ruleStatus
			}
		}
		groups[i] = g
	}
	return groups, nil
}

func (o *ruleGroupOperator) ListGlobalRuleGroups(ctx context.Context,
	queryParam *query.Query) (*api.ListResult, error) {

	selector := queryParam.Selector()
	if val, ok := queryParam.Filters[kapialertingv2beta1.FieldBuiltin]; ok {
		// add match requirement to the selector to select only builtin or custom rulegroups
		var operator selection.Operator
		if val == controller.PrometheusRuleResourceLabelValueBuiltinTrue {
			operator = selection.Equals
		} else {
			operator = selection.NotEquals
		}
		requirement, _ := labels.NewRequirement(
			controller.PrometheusRuleResourceLabelKeyBuiltin,
			operator,
			[]string{controller.PrometheusRuleResourceLabelValueBuiltinTrue})
		selector = selector.Add(*requirement)
	}
	groups, err := o.listGlobalRuleGroups(ctx, selector)
	if err != nil {
		return nil, err
	}

	listResult := resources.DefaultList(groups, queryParam, func(left, right runtime.Object, field query.Field) bool {
		hit, great := o.compareRuleGroupStatus(
			&(left.(*kapialertingv2beta1.GlobalRuleGroup).Status), &(right.(*kapialertingv2beta1.GlobalRuleGroup).Status), field)
		if hit {
			return great
		}
		return resources.DefaultObjectMetaCompare(
			left.(*kapialertingv2beta1.GlobalRuleGroup).ObjectMeta, right.(*kapialertingv2beta1.GlobalRuleGroup).ObjectMeta, field)
	}, func(obj runtime.Object, filter query.Filter) bool {
		if filter.Field == kapialertingv2beta1.FieldBuiltin { // ignoring this filter because it is filtered at the front
			return true
		}
		hit, selected := o.filterRuleGroupStatus(&obj.(*kapialertingv2beta1.GlobalRuleGroup).Status, filter)
		if hit {
			return selected
		}
		return resources.DefaultObjectMetaFilter(obj.(*kapialertingv2beta1.GlobalRuleGroup).ObjectMeta, filter)
	})

	return listResult, nil
}

func (o *ruleGroupOperator) ListGlobalAlerts(ctx context.Context,
	queryParam *query.Query) (*api.ListResult, error) {

	selector := labels.Everything()
	if val, ok := queryParam.Filters[kapialertingv2beta1.FieldBuiltin]; ok {
		// add match requirement to the selector to select only builtin or custom rulegroups
		var operator selection.Operator
		if val == controller.PrometheusRuleResourceLabelValueBuiltinTrue {
			operator = selection.Equals
		} else {
			operator = selection.NotEquals
		}
		requirement, _ := labels.NewRequirement(
			controller.PrometheusRuleResourceLabelKeyBuiltin,
			operator,
			[]string{controller.PrometheusRuleResourceLabelValueBuiltinTrue})
		selector = selector.Add(*requirement)
	}
	groups, err := o.listGlobalRuleGroups(ctx, selector)
	if err != nil {
		return nil, err
	}

	// encapsulated as runtime.Object for easy comparison and filtering.
	var alerts []runtime.Object
	for i := range groups {
		wrapg := groups[i].(*kapialertingv2beta1.GlobalRuleGroup)
		for j := range wrapg.Status.RulesStatus {
			ruleStatus := wrapg.Status.RulesStatus[j]
			for k := range ruleStatus.Alerts {
				alerts = append(alerts, &wrapAlert{Alert: *ruleStatus.Alerts[k]})
			}
		}
	}

	filterAlert, err := o.createFilterAlertFunc(queryParam)
	if err != nil {
		return nil, err
	}
	listResult := resources.DefaultList(alerts, queryParam, func(left, right runtime.Object, field query.Field) bool {
		return o.compareAlert(&left.(*wrapAlert).Alert, &right.(*wrapAlert).Alert, field)
	}, func(obj runtime.Object, filter query.Filter) bool {
		if filter.Field == kapialertingv2beta1.FieldBuiltin { // ignoring this filter because it is filtered at the front
			return true
		}
		return filterAlert(&obj.(*wrapAlert).Alert, filter)
	})
	for i := range listResult.Items {
		listResult.Items[i] = &listResult.Items[i].(*wrapAlert).Alert
	}
	return listResult, nil
}

func (o *ruleGroupOperator) GetGlobalRuleGroup(ctx context.Context, name string) (*kapialertingv2beta1.GlobalRuleGroup, error) {
	resourceRuleGroup, err := o.globalRuleGroupLister.Get(name)
	if err != nil {
		return nil, err
	}

	setRuleGroupResourceTypeMeta(resourceRuleGroup)

	ret := &kapialertingv2beta1.GlobalRuleGroup{
		GlobalRuleGroup: *resourceRuleGroup,
		Status: kapialertingv2beta1.RuleGroupStatus{
			RulesStatus: make([]kapialertingv2beta1.RuleStatus, len(resourceRuleGroup.Spec.Rules)),
			RulesStats:  kapialertingv2beta1.RulesStats{},
		},
	}

	matchers := []*promlabels.Matcher{{
		Type:  promlabels.MatchEqual,
		Name:  controller.RuleLabelKeyRuleLevel,
		Value: string(controller.RuleLevelGlobal),
	}}
	statusRuleGroups, err := o.ruleClient.ThanosRules(ctx, matchers)
	if err != nil {
		return nil, err
	}

	var setStatus bool
	specRules := resourceRuleGroup.Spec.Rules
	for _, g := range statusRuleGroups {
		if g.Name == resourceRuleGroup.Name && len(g.Rules) == len(specRules) {
			var ruleIds = make([]string, len(specRules))
			var ruleDisableFlags = make([]bool, len(specRules))
			for j := range specRules {
				if specRules[j].Labels != nil {
					ruleIds[j] = specRules[j].Labels[alertingv2beta1.RuleLabelKeyRuleId]
				}
				ruleDisableFlags[j] = specRules[j].Disable
			}

			copyRuleGroupStatus(g, &ret.Status, ruleIds, ruleDisableFlags)
			for j := range ret.Status.RulesStatus {
				// for rules disabled and etc.
				if ret.Status.RulesStatus[j].Expr == "" {
					rule := ret.Spec.Rules[j]
					enforceExprFunc := controller.CreateEnforceExprFunc(controller.ParseGlobalRuleEnforceMatchers(&rule))
					expr, err := enforceExprFunc(rule.Expr.String())
					if err != nil {
						return nil, err
					}
					ret.Status.RulesStatus[j].Expr = expr
				}
			}
			setStatus = true
			break
		}
	}
	if !setStatus {
		// for rules not loaded by rule reloader (eg.thanos) yet
		for j, rule := range ret.Spec.Rules {
			ruleStatus := kapialertingv2beta1.RuleStatus{Health: string(promrules.HealthUnknown)}
			if specRules[j].Disable {
				ruleStatus.State = stateDisabledString
				ret.Status.RulesStats.Disabled++
			} else {
				ruleStatus.State = stateInactiveString
				ret.Status.RulesStats.Inactive++
			}
			ret.Status.RulesStatus = append(ret.Status.RulesStatus, ruleStatus)
			enforceExprFunc := controller.CreateEnforceExprFunc(controller.ParseGlobalRuleEnforceMatchers(&rule))
			expr, err := enforceExprFunc(rule.Expr.String())
			if err != nil {
				return nil, err
			}
			ruleStatus.Expr = expr
			ret.Status.RulesStatus[j] = ruleStatus
		}
	}

	return ret, nil
}

// copyRuleGroupStatus copies group/rule status and alerts from source to target
func copyRuleGroupStatus(source *alerting.RuleGroup, target *kapialertingv2beta1.RuleGroupStatus, ruleIds []string, ruleDisableFlags []bool) {
	target.LastEvaluation = source.LastEvaluation
	if source.EvaluationTime > 0 {
		target.EvaluationTime = &source.EvaluationTime
	}

	sourceRuleMap := make(map[string]*alerting.AlertingRule, len(source.Rules))
	for i := range source.Rules {
		rule := source.Rules[i]
		if len(rule.Labels) > 0 {
			if ruleId, ok := rule.Labels[alertingv2beta1.RuleLabelKeyRuleId]; ok {
				sourceRuleMap[ruleId] = rule
			}
		}
	}
	for i, ruleId := range ruleIds {
		rule, ok := sourceRuleMap[ruleId]
		if !ok {
			ruleStatus := kapialertingv2beta1.RuleStatus{Health: string(promrules.HealthUnknown)}
			if ruleDisableFlags[i] {
				ruleStatus.State = stateDisabledString
				target.RulesStats.Disabled++
			} else {
				// rules are being loaded
				ruleStatus.State = stateInactiveString
				target.RulesStats.Inactive++
			}
			target.RulesStatus[i] = ruleStatus
			continue
		}

		switch rule.State {
		case statePendingString:
			target.RulesStats.Pending++
		case stateFiringString:
			target.RulesStats.Firing++
		case stateInactiveString:
			target.RulesStats.Inactive++
		}

		var ruleActiveAt *time.Time
		alerts := []*kapialertingv2beta1.Alert{}
		for _, alert := range rule.Alerts {
			alerts = append(alerts, &kapialertingv2beta1.Alert{
				ActiveAt:    alert.ActiveAt,
				Annotations: alert.Annotations,
				Labels:      alert.Labels,
				State:       alert.State,
				Value:       alert.Value,
			})
			if alert.ActiveAt != nil && (ruleActiveAt == nil || alert.ActiveAt.Before(*ruleActiveAt)) {
				ruleActiveAt = alert.ActiveAt
			}
		}
		ruleStatus := kapialertingv2beta1.RuleStatus{
			State:          rule.State,
			Health:         rule.Health,
			LastError:      rule.LastError,
			EvaluationTime: rule.EvaluationTime,
			LastEvaluation: rule.LastEvaluation,
			ActiveAt:       ruleActiveAt,
			Alerts:         alerts,
		}
		if len(rule.Labels) > 0 {
			if level, ok := rule.Labels[controller.RuleLabelKeyRuleLevel]; ok &&
				level == string(controller.RuleLevelGlobal) { // provided only for global rules
				ruleStatus.Expr = rule.Query
			}
		}
		target.RulesStatus[i] = ruleStatus
	}
}

var (
	statePendingString  = promrules.StatePending.String()
	stateFiringString   = promrules.StateFiring.String()
	stateInactiveString = promrules.StateInactive.String()
	stateDisabledString = "disabled"
)

type wrapAlert struct {
	kapialertingv2beta1.Alert
	runtime.Object
}

var (
	apiVersion = alertingv2beta1.SchemeGroupVersion.String()
)

// set TypeMeta info because the objects getted by lister.List() and lister.Get() may have no TypeMeta
// related issue: https://github.com/kubernetes/client-go/issues/541
func setRuleGroupResourceTypeMeta(obj runtime.Object) {
	switch o := obj.(type) {
	case *alertingv2beta1.RuleGroup:
		o.APIVersion = apiVersion
		o.Kind = alertingv2beta1.ResourceKindRuleGroup
	case *alertingv2beta1.ClusterRuleGroup:
		o.APIVersion = apiVersion
		o.Kind = alertingv2beta1.ResourceKindClusterRuleGroup
	case *alertingv2beta1.GlobalRuleGroup:
		o.APIVersion = apiVersion
		o.Kind = alertingv2beta1.ResourceKindGlobalRuleGroup
	}
}
