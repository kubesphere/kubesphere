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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	yaml "gopkg.in/yaml.v3"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var rulegrouplog = logf.Log.WithName("rulegroup")

const RuleLabelKeyRuleId = "rule_id"

func (r *RuleGroup) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Defaulter = &RuleGroup{}

func (r *RuleGroup) Default() {
	log := rulegrouplog.WithValues("name", r.Namespace+"/"+r.Name)
	log.Info("default")
	for i := range r.Spec.Rules {
		rule := r.Spec.Rules[i]
		if rule.ExprBuilder != nil {
			if rule.ExprBuilder.Workload != nil {
				rule.Expr = intstr.FromString(rule.ExprBuilder.Workload.Build())
			}
		}
		setRuleId(&rule.Rule)
		r.Spec.Rules[i] = rule
	}
}

func setRuleId(rule *Rule) {
	var setRuleId = true
	if len(rule.Labels) > 0 {
		if _, ok := rule.Labels[RuleLabelKeyRuleId]; ok {
			setRuleId = false
		}
	}
	if setRuleId {
		if rule.Labels == nil {
			rule.Labels = make(map[string]string)
		}
		rule.Labels[RuleLabelKeyRuleId] = string(uuid.NewUUID())
	}
}

var _ webhook.Validator = &RuleGroup{}

func (r *RuleGroup) ValidateCreate() error {
	return r.Validate()
}
func (r *RuleGroup) ValidateUpdate(old runtime.Object) error {
	return r.Validate()
}
func (r *RuleGroup) ValidateDelete() error {
	return nil
}
func (r *RuleGroup) Validate() error {
	log := rulegrouplog.WithValues("name", r.Namespace+"/"+r.Name)
	log.Info("validate")

	var rules []Rule
	for _, r := range r.Spec.Rules {
		rules = append(rules, r.Rule)
	}
	var err = validateRules(log, r.Name, r.Spec.Interval, rules)
	if err == errorEmptyExpr {
		return fmt.Errorf("one of 'expr' and 'exprBuilder.workload' must be set for a RuleGroup")
	}
	return err
}

type ruleGroup struct {
	Name     string         `yaml:"name"`
	Interval model.Duration `yaml:"interval,omitempty"`
	Rules    []rulefmt.Rule `yaml:"rules"`
}

type ruleGroups struct {
	Groups []ruleGroup `yaml:"groups"`
}

var errorEmptyExpr = fmt.Errorf("'expr' is empty")

func validateRules(log logr.Logger, groupName, groupInterval string, rules []Rule) error {
	durationStr := groupInterval
	if durationStr == "" {
		durationStr = "1m"
	}
	interval, err := model.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid 'interval': %s", durationStr)
	}

	var g = ruleGroup{
		Name:     groupName,
		Interval: interval,
	}

	for i := range rules {
		rule := rules[i]
		if rule.Alert == "" {
			return fmt.Errorf("'alert' cannot be empty")
		}
		if rule.Expr.String() == "" {
			return errorEmptyExpr
		}
		durationStr := string(rule.For)
		if durationStr == "" {
			durationStr = "0"
		}
		forDuration, err := model.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid 'for': %s", durationStr)
		}
		g.Rules = append(g.Rules, rulefmt.Rule{
			Alert:       rule.Alert,
			Expr:        rule.Expr.String(),
			For:         forDuration,
			Labels:      rule.Labels,
			Annotations: rule.Annotations,
		})
	}

	var gs = ruleGroups{}
	gs.Groups = append(gs.Groups, g)

	content, err := yaml.Marshal(gs)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %v", err)
	}
	_, errs := rulefmt.Parse(content)

	if len(errs) > 0 {
		for _, err := range errs {
			log.Info(fmt.Sprintf("invalid rule: %v", err))
		}
		return fmt.Errorf("invalid rules: %v", errs)
	}
	return nil
}

var clusterrulegrouplog = logf.Log.WithName("clusterrulegroup")

func (r *ClusterRuleGroup) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Defaulter = &ClusterRuleGroup{}

func (r *ClusterRuleGroup) Default() {
	log := clusterrulegrouplog.WithValues("name", r.Name)
	log.Info("default")

	for i := range r.Spec.Rules {
		rule := r.Spec.Rules[i]
		if rule.ExprBuilder != nil {
			if rule.ExprBuilder.Node != nil {
				rule.Expr = intstr.FromString(rule.ExprBuilder.Node.Build())
			}
		}
		setRuleId(&rule.Rule)
		r.Spec.Rules[i] = rule
	}
}

var _ webhook.Validator = &ClusterRuleGroup{}

func (r *ClusterRuleGroup) ValidateCreate() error {
	return r.Validate()
}
func (r *ClusterRuleGroup) ValidateUpdate(old runtime.Object) error {
	return r.Validate()
}
func (r *ClusterRuleGroup) ValidateDelete() error {
	return nil
}
func (r *ClusterRuleGroup) Validate() error {
	log := clusterrulegrouplog.WithValues("name", r.Name)
	log.Info("validate")

	var rules []Rule
	for _, r := range r.Spec.Rules {
		rules = append(rules, r.Rule)
	}
	var err = validateRules(log, r.Name, r.Spec.Interval, rules)
	if err == errorEmptyExpr {
		return fmt.Errorf("one of 'expr' and 'exprBuilder.node' must be set for a ClusterRuleGroup")
	}
	return err
}

var globalrulegrouplog = logf.Log.WithName("globalrulegroup")

func (r *GlobalRuleGroup) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Defaulter = &GlobalRuleGroup{}

func (r *GlobalRuleGroup) Default() {
	log := globalrulegrouplog.WithValues("name", r.Name)
	log.Info("default")

	for i := range r.Spec.Rules {
		rule := r.Spec.Rules[i]
		if rule.ExprBuilder != nil {
			if rule.ExprBuilder.Node != nil {
				rule.Expr = intstr.FromString(rule.ExprBuilder.Node.Build())
			} else if rule.ExprBuilder.Workload != nil {
				rule.Expr = intstr.FromString(rule.ExprBuilder.Workload.Build())
			}
		}
		setRuleId(&rule.Rule)
		r.Spec.Rules[i] = rule
	}
}

var _ webhook.Validator = &GlobalRuleGroup{}

func (r *GlobalRuleGroup) ValidateCreate() error {
	return r.Validate()
}
func (r *GlobalRuleGroup) ValidateUpdate(old runtime.Object) error {
	return r.Validate()
}
func (r *GlobalRuleGroup) ValidateDelete() error {
	return nil
}
func (r *GlobalRuleGroup) Validate() error {
	log := globalrulegrouplog.WithValues("name", r.Name)
	log.Info("validate")

	var rules []Rule
	for _, r := range r.Spec.Rules {
		if r.ClusterSelector != nil {
			if err := r.ClusterSelector.Validate(); err != nil {
				return err
			}
		}
		if r.NamespaceSelector != nil {
			if err := r.NamespaceSelector.Validate(); err != nil {
				return err
			}
		}
		rules = append(rules, r.Rule)
	}
	var err = validateRules(log, r.Name, r.Spec.Interval, rules)
	if err == errorEmptyExpr {
		return fmt.Errorf("'expr' must be set for a GlobalRuleGroup")
	}
	return err
}
