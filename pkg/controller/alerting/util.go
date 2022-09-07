/*
Copyright 2019 The KubeSphere Authors.
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

package alerting

import (
	"context"
	"reflect"
	"sort"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/prometheus-community/prom-label-proxy/injectproxy"
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promlabels "github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	alertingv2beta1 "kubesphere.io/api/alerting/v2beta1"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	RuleLevelNamesapce RuleLevel = "namespace"
	RuleLevelCluster   RuleLevel = "cluster"
	RuleLevelGlobal    RuleLevel = "global"

	// for rule.labels
	RuleLabelKeyRuleLevel         = "rule_level"
	RuleLabelKeyRuleGroup         = "rule_group"
	RuleLabelKeyCluster           = "cluster"
	RuleLabelKeyNamespace         = "namespace"
	RuleLabelKeySeverity          = "severity"
	RuleLabelKeyAlertType         = "alerttype"
	RuleLabelValueAlertTypeMetric = "metric"

	// label keys in RuleGroup/ClusterRuleGroup/GlobalRuleGroup.metadata.labels
	SourceGroupResourceLabelKeyEnable        = "alerting.kubesphere.io/enable"
	SourceGroupResourceLabelValueEnableTrue  = "true"
	SourceGroupResourceLabelValueEnableFalse = "false"

	// for PrometheusRule.metadata.labels
	PrometheusRuleResourceLabelKeyOwnerNamespace = "alerting.kubesphere.io/owner_namespace"
	PrometheusRuleResourceLabelKeyOwnerCluster   = "alerting.kubesphere.io/owner_cluster"
	PrometheusRuleResourceLabelKeyRuleLevel      = "alerting.kubesphere.io/rule_level"
	PrometheusRuleResourceLabelKeyBuiltin        = "alerting.kubesphere.io/builtin"
	PrometheusRuleResourceLabelValueBuiltinTrue  = "true"
	PrometheusRuleResourceLabelValueBuiltinFalse = "false"

	// name prefix for PrometheusRule
	PrometheusRulePrefix               = "alertrules-"
	PrometheusRulePrefixNamespaceLevel = PrometheusRulePrefix + "ns-"
	PrometheusRulePrefixClusterLevel   = PrometheusRulePrefix + "cl-"
	PrometheusRulePrefixGlobalLevel    = PrometheusRulePrefix + "gl-"

	PrometheusRuleNamespace = constants.KubeSphereMonitoringNamespace
)

type RuleLevel string

var maxConfigMapDataSize = int(float64(corev1.MaxSecretSize) * 0.5)

type enforceRuleFunc func(rule *promresourcesv1.Rule) error

type EnforceExprFunc func(expr string) (string, error)

var emptyEnforceExprFunc = func(expr string) (string, error) {
	return expr, nil
}

func CreateEnforceExprFunc(enforceRuleMatchers []*promlabels.Matcher) EnforceExprFunc {
	if len(enforceRuleMatchers) > 0 {
		enforcer := injectproxy.NewEnforcer(enforceRuleMatchers...)
		return func(expr string) (string, error) {
			parsedExpr, err := parser.ParseExpr(expr)
			if err != nil {
				return expr, err
			}
			if err := enforcer.EnforceNode(parsedExpr); err != nil {
				return expr, err
			}
			return parsedExpr.String(), nil
		}
	}
	return emptyEnforceExprFunc
}

func createEnforceRuleFuncs(enforceRuleMatchers []*promlabels.Matcher, enforceRuleLabels map[string]string) []enforceRuleFunc {
	var enforceFuncs []enforceRuleFunc
	// enforce func for rule.expr
	if len(enforceRuleMatchers) > 0 {
		enforceExprFunc := CreateEnforceExprFunc(enforceRuleMatchers)
		enforceFuncs = append(enforceFuncs, func(rule *promresourcesv1.Rule) error {
			expr, err := enforceExprFunc(rule.Expr.String())
			if err != nil {
				return err
			}
			rule.Expr = intstr.FromString(expr)
			return nil
		})
	}
	// enforce func for rule.labels
	if len(enforceRuleLabels) > 0 {
		enforceFuncs = append(enforceFuncs, func(rule *promresourcesv1.Rule) error {
			if rule.Labels == nil {
				rule.Labels = make(map[string]string)
			}
			for n, v := range enforceRuleLabels {
				rule.Labels[n] = v
			}
			return nil
		})
	}
	return enforceFuncs
}

func makePrometheusRuleGroups(log logr.Logger, groupList client.ObjectList,
	commonEnforceFuncs ...enforceRuleFunc) ([]*promresourcesv1.RuleGroup, error) {
	var rulegroups []*promresourcesv1.RuleGroup

	convertRule := func(rule *alertingv2beta1.Rule, groupName string, enforceFuncs ...enforceRuleFunc) (*promresourcesv1.Rule, error) {
		if rule.Disable { // ignoring disabled rule
			return nil, nil
		}

		rule = rule.DeepCopy()

		if rule.Labels == nil {
			rule.Labels = make(map[string]string)
		}

		if rule.Severity != "" {
			rule.Labels[RuleLabelKeySeverity] = string(rule.Severity)
		}

		prule := promresourcesv1.Rule{
			Alert:       rule.Alert,
			For:         string(rule.For),
			Expr:        rule.Expr,
			Labels:      rule.Labels,
			Annotations: rule.Annotations,
		}

		enforceFuncs = append(enforceFuncs, commonEnforceFuncs...)
		// enforce rule group label and alert type label
		enforceFuncs = append(enforceFuncs, func(rule *promresourcesv1.Rule) error {
			if rule.Labels == nil {
				rule.Labels = make(map[string]string)
			}
			rule.Labels[RuleLabelKeyRuleGroup] = groupName
			rule.Labels[RuleLabelKeyAlertType] = RuleLabelValueAlertTypeMetric
			return nil
		})

		for _, f := range enforceFuncs {
			if f == nil {
				continue
			}
			err := f(&prule)
			if err != nil {
				return nil, errors.Wrapf(err, "alert: %s", rule.Alert)
			}
		}

		return &prule, nil
	}

	switch list := groupList.(type) {
	case *alertingv2beta1.RuleGroupList:
		for _, group := range list.Items {
			var prules []promresourcesv1.Rule
			for _, rule := range group.Spec.Rules {
				prule, err := convertRule(&rule.Rule, group.Name)
				if err != nil {
					log.WithValues("rulegroup", group.Namespace+"/"+group.Name).Error(err, "failed to convert")
					continue
				}
				if prule != nil {
					prules = append(prules, *prule)
				}
			}
			rulegroups = append(rulegroups, &promresourcesv1.RuleGroup{
				Name:                    group.Name,
				Interval:                group.Spec.Interval,
				PartialResponseStrategy: group.Spec.PartialResponseStrategy,
				Rules:                   prules,
			})
		}
	case *alertingv2beta1.ClusterRuleGroupList:
		for _, group := range list.Items {
			var prules []promresourcesv1.Rule
			for _, rule := range group.Spec.Rules {
				prule, err := convertRule(&rule.Rule, group.Name)
				if err != nil {
					log.WithValues("clusterrulegroup", group.Name).Error(err, "failed to convert")
					continue
				}
				if prule != nil {
					prules = append(prules, *prule)
				}
			}
			rulegroups = append(rulegroups, &promresourcesv1.RuleGroup{
				Name:                    group.Name,
				Interval:                group.Spec.Interval,
				PartialResponseStrategy: group.Spec.PartialResponseStrategy,
				Rules:                   prules,
			})
		}
	case *alertingv2beta1.GlobalRuleGroupList:
		for _, group := range list.Items {
			var prules []promresourcesv1.Rule
			for _, rule := range group.Spec.Rules {

				prule, err := convertRule(&rule.Rule, group.Name,
					createEnforceRuleFuncs(ParseGlobalRuleEnforceMatchers(&rule), nil)...)
				if err != nil {
					log.WithValues("globalrulegroup", group.Name).Error(err, "failed to convert")
					continue
				}
				if prule != nil {
					prules = append(prules, *prule)
				}
			}
			rulegroups = append(rulegroups, &promresourcesv1.RuleGroup{
				Name:                    group.Name,
				Interval:                group.Spec.Interval,
				PartialResponseStrategy: group.Spec.PartialResponseStrategy,
				Rules:                   prules,
			})
		}
	}

	return rulegroups, nil
}

func ParseGlobalRuleEnforceMatchers(rule *alertingv2beta1.GlobalRule) []*promlabels.Matcher {
	var enforceRuleMatchers []*promlabels.Matcher
	if rule.ClusterSelector != nil {
		matcher := rule.ClusterSelector.ParseToMatcher(RuleLabelKeyCluster)
		if matcher != nil {
			enforceRuleMatchers = append(enforceRuleMatchers, matcher)
		}
	}
	if rule.NamespaceSelector != nil {
		matcher := rule.NamespaceSelector.ParseToMatcher(RuleLabelKeyNamespace)
		if matcher != nil {
			enforceRuleMatchers = append(enforceRuleMatchers, matcher)
		}
	}
	return enforceRuleMatchers
}

func makePrometheusRuleResources(rulegroups []*promresourcesv1.RuleGroup, namespace, namePrefix string,
	labels map[string]string, ownerReferences []metav1.OwnerReference) ([]*promresourcesv1.PrometheusRule, error) {

	promruleSpecs, err := makePrometheusRuleSpecs(rulegroups)
	if err != nil {
		return nil, err
	}
	var ps = make([]*promresourcesv1.PrometheusRule, len(promruleSpecs))
	for i := range promruleSpecs {
		ps[i] = &promresourcesv1.PrometheusRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       namespace,
				Name:            namePrefix + strconv.Itoa(i),
				Labels:          labels,
				OwnerReferences: ownerReferences,
			},
			Spec: *promruleSpecs[i],
		}
	}

	return ps, nil
}

type rulegroupsWrapper struct {
	rulegroups []*promresourcesv1.RuleGroup
	by         func(g1, g2 *promresourcesv1.RuleGroup) bool
}

func (w rulegroupsWrapper) Len() int {
	return len(w.rulegroups)
}

func (w rulegroupsWrapper) Swap(i, j int) {
	w.rulegroups[i], w.rulegroups[j] = w.rulegroups[j], w.rulegroups[i]
}

func (w rulegroupsWrapper) Less(i, j int) bool {
	return w.by(w.rulegroups[i], w.rulegroups[j])
}

func makePrometheusRuleSpecs(rulegroups []*promresourcesv1.RuleGroup) ([]*promresourcesv1.PrometheusRuleSpec, error) {
	sort.Sort(rulegroupsWrapper{
		rulegroups: rulegroups,
		by: func(g1, g2 *promresourcesv1.RuleGroup) bool {
			return g1.Name < g2.Name
		},
	})

	var (
		pSpecs []*promresourcesv1.PrometheusRuleSpec
		pSpec  promresourcesv1.PrometheusRuleSpec
		size   int
	)

	for i := range rulegroups {
		rulegroup := rulegroups[i]

		content, err := yaml.Marshal(rulegroup)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal content")
		}
		contentLen := len(string(content))
		size += contentLen
		if size > maxConfigMapDataSize*80/100 { // leave space for enforcing possiable label matchers into expr
			pSpecs = append(pSpecs, &pSpec)
			// reinit
			size = contentLen
			pSpec = promresourcesv1.PrometheusRuleSpec{}
		}

		pSpec.Groups = append(pSpec.Groups, *rulegroup)
	}
	if len(pSpec.Groups) > 0 {
		pSpecs = append(pSpecs, &pSpec)
	}

	return pSpecs, nil
}

func bulkUpdatePrometheusRuleResources(client client.Client, ctx context.Context, current, desired []*promresourcesv1.PrometheusRule) error {

	var (
		currentMap = make(map[string]*promresourcesv1.PrometheusRule)
		desiredMap = make(map[string]*promresourcesv1.PrometheusRule)
		err        error
	)
	for i := range current {
		promrule := current[i]
		currentMap[promrule.Namespace+"/"+promrule.Name] = promrule
	}
	for i := range desired {
		promrule := desired[i]
		desiredMap[promrule.Namespace+"/"+promrule.Name] = promrule
	}

	// update if exists in current PrometheusRules, or create
	for name, desired := range desiredMap {
		if current, ok := currentMap[name]; ok {
			if !reflect.DeepEqual(current.Spec, desired.Spec) ||
				!reflect.DeepEqual(current.Labels, desired.Labels) ||
				!reflect.DeepEqual(current.OwnerReferences, desired.OwnerReferences) {
				desired.SetResourceVersion(current.ResourceVersion)
				err = client.Update(ctx, desired)
				if err != nil {
					return err
				}
			}
		} else {
			err = client.Create(ctx, desired)
			if err != nil {
				return err
			}
		}
	}
	// delete if not in desired PrometheusRules
	for name, current := range currentMap {
		if _, ok := desiredMap[name]; !ok {
			err = client.Delete(ctx, current)
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}
		}
	}

	return nil
}
