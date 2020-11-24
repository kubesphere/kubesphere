package rules

import (
	"context"
	"fmt"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prominformersv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions/monitoring/v1"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/api/customalerting/v1alpha1"
)

const (
	customAlertingRuleResourcePrefix = "custom-alerting-rule-"
)

var (
	maxSecretSize        = corev1.MaxSecretSize
	maxConfigMapDataSize = int(float64(maxSecretSize) * 0.3)

	errOutOfConfigMapSize = errors.New("out of config map size")
)

type Ruler interface {
	Namespace() string
	RuleResourceNamespaceSelector() (labels.Selector, error)
	RuleResourceSelector(extraRuleResourceSelector labels.Selector) (labels.Selector, error)
	ExternalLabels() func() map[string]string

	ListRuleResources(ruleNamespace *corev1.Namespace, extraRuleResourceSelector labels.Selector) (
		[]*promresourcesv1.PrometheusRule, error)

	AddAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace, extraRuleResourceSelector labels.Selector,
		group string, rule *promresourcesv1.Rule, ruleResourceLabels map[string]string) error
	UpdateAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace, extraRuleResourceSelector labels.Selector,
		group string, rule *promresourcesv1.Rule, ruleResourceLabels map[string]string) error
	DeleteAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace, extraRuleResourceSelector labels.Selector,
		group string, name string) error
}

type ruleResource promresourcesv1.PrometheusRule

// deleteAlertingRule deletes the rules with the given name.
// If the rule is deleted, return true to indicate the resource should be updated.
func (r *ruleResource) deleteAlertingRule(name string) (bool, error) {
	var (
		nGroups []promresourcesv1.RuleGroup
		ok      bool
	)

	for _, g := range r.Spec.Groups {
		var rules []promresourcesv1.Rule
		for _, gr := range g.Rules {
			if gr.Alert != "" && gr.Alert == name {
				ok = true
				continue
			}
			rules = append(rules, gr)
		}
		if len(rules) > 0 {
			nGroups = append(nGroups, promresourcesv1.RuleGroup{
				Name:                    g.Name,
				Interval:                g.Interval,
				PartialResponseStrategy: g.PartialResponseStrategy,
				Rules:                   rules,
			})
		}
	}

	if ok {
		r.Spec.Groups = nGroups
	}
	return ok, nil
}

// updateAlertingRule updates the rule with the given group.
// If the rule is updated, return true to indicate the resource should be updated.
func (r *ruleResource) updateAlertingRule(groupName string, rule *promresourcesv1.Rule) (bool, error) {
	var (
		ok       bool
		pr       = (promresourcesv1.PrometheusRule)(*r)
		npr      = pr.DeepCopy()
		groupMap = make(map[string]*promresourcesv1.RuleGroup)
	)

	for _, g := range npr.Spec.Groups {
		var rules []promresourcesv1.Rule
		for i, gr := range g.Rules {
			if gr.Alert != "" && gr.Alert == rule.Alert {
				ok = true
				continue
			}
			rules = append(rules, g.Rules[i])
		}
		if len(rules) > 0 {
			groupMap[g.Name] = &promresourcesv1.RuleGroup{
				Name:                    g.Name,
				Interval:                g.Interval,
				PartialResponseStrategy: g.PartialResponseStrategy,
				Rules:                   rules,
			}
		}
	}

	if ok {
		if g, exist := groupMap[groupName]; exist {
			g.Rules = append(g.Rules, *rule)
		} else {
			groupMap[groupName] = &promresourcesv1.RuleGroup{
				Name:  groupName,
				Rules: []promresourcesv1.Rule{*rule},
			}
		}

		var groups []promresourcesv1.RuleGroup
		for _, g := range groupMap {
			groups = append(groups, *g)
		}

		npr.Spec.Groups = groups
		content, err := yaml.Marshal(npr)
		if err != nil {
			return false, errors.Wrap(err, "failed to unmarshal content")
		}

		if len(string(content)) < maxConfigMapDataSize { // check size limit
			r.Spec.Groups = groups
			return true, nil
		}
		return false, errOutOfConfigMapSize
	}
	return false, nil
}

func (r *ruleResource) addAlertingRule(group string, rule *promresourcesv1.Rule) (bool, error) {
	var (
		err error
		pr  = (promresourcesv1.PrometheusRule)(*r)
		npr = pr.DeepCopy()
		ok  bool
	)

	for i := 0; i < len(npr.Spec.Groups); i++ {
		if npr.Spec.Groups[i].Name == group {
			npr.Spec.Groups[i].Rules = append(npr.Spec.Groups[i].Rules, *rule)
			ok = true
			break
		}
	}
	if !ok { // add a group when there is no group with the specified group name
		npr.Spec.Groups = append(npr.Spec.Groups, promresourcesv1.RuleGroup{
			Name:  group,
			Rules: []promresourcesv1.Rule{*rule},
		})
	}

	content, err := yaml.Marshal(npr)
	if err != nil {
		return false, errors.Wrap(err, "failed to unmarshal content")
	}

	if len(string(content)) < maxConfigMapDataSize { // check size limit
		r.Spec.Groups = npr.Spec.Groups
		return true, nil
	} else {
		return false, errOutOfConfigMapSize
	}
}

func (r *ruleResource) commit(ctx context.Context, prometheusResourceClient promresourcesclient.Interface) error {
	var pr = (promresourcesv1.PrometheusRule)(*r)
	if len(pr.Spec.Groups) == 0 {
		return prometheusResourceClient.MonitoringV1().PrometheusRules(r.Namespace).Delete(ctx, r.Name, metav1.DeleteOptions{})
	}
	newPr, err := prometheusResourceClient.MonitoringV1().PrometheusRules(r.Namespace).Update(ctx, &pr, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	newPr.DeepCopyInto(&pr)
	return nil
}

type PrometheusRuler struct {
	resource *promresourcesv1.Prometheus
	informer prominformersv1.PrometheusRuleInformer
	client   promresourcesclient.Interface
}

func NewPrometheusRuler(resource *promresourcesv1.Prometheus, informer prominformersv1.PrometheusRuleInformer,
	client promresourcesclient.Interface) Ruler {
	return &PrometheusRuler{
		resource: resource,
		informer: informer,
		client:   client,
	}
}

func (r *PrometheusRuler) Namespace() string {
	return r.resource.Namespace
}

func (r *PrometheusRuler) RuleResourceNamespaceSelector() (labels.Selector, error) {
	if r.resource.Spec.RuleNamespaceSelector == nil {
		return nil, nil
	}
	return metav1.LabelSelectorAsSelector(r.resource.Spec.RuleNamespaceSelector)
}

func (r *PrometheusRuler) RuleResourceSelector(extraRuleResourceSelector labels.Selector) (labels.Selector, error) {
	rSelector, err := metav1.LabelSelectorAsSelector(r.resource.Spec.RuleSelector)
	if err != nil {
		return nil, err
	}
	if extraRuleResourceSelector != nil {
		if requirements, ok := extraRuleResourceSelector.Requirements(); ok {
			rSelector = rSelector.Add(requirements...)
		}
	}
	return rSelector, nil
}

func (r *PrometheusRuler) ExternalLabels() func() map[string]string {
	// ignoring the external labels because rules gotten from prometheus endpoint do not include them
	return nil
}

func (r *PrometheusRuler) ListRuleResources(ruleNamespace *corev1.Namespace, extraRuleResourceSelector labels.Selector) (
	[]*promresourcesv1.PrometheusRule, error) {
	selected, err := ruleNamespaceSelected(r, ruleNamespace)
	if err != nil {
		return nil, err
	}
	if !selected {
		return nil, nil
	}
	rSelector, err := r.RuleResourceSelector(extraRuleResourceSelector)
	if err != nil {
		return nil, err
	}
	return r.informer.Lister().PrometheusRules(ruleNamespace.Name).List(rSelector)
}

func (r *PrometheusRuler) AddAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector,
	group string, rule *promresourcesv1.Rule, ruleResourceLabels map[string]string) error {
	return errors.New("not supported to add rules for prometheus")
}

func (r *PrometheusRuler) UpdateAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector,
	group string, rule *promresourcesv1.Rule, ruleResourceLabels map[string]string) error {
	return errors.New("not supported to update rules for prometheus")
}

func (r *PrometheusRuler) DeleteAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector,
	group string, name string) error {
	return errors.New("not supported to update rules for prometheus")
}

type ThanosRuler struct {
	resource *promresourcesv1.ThanosRuler
	informer prominformersv1.PrometheusRuleInformer
	client   promresourcesclient.Interface
}

func NewThanosRuler(resource *promresourcesv1.ThanosRuler, informer prominformersv1.PrometheusRuleInformer,
	client promresourcesclient.Interface) Ruler {
	return &ThanosRuler{
		resource: resource,
		informer: informer,
		client:   client,
	}
}

func (r *ThanosRuler) Namespace() string {
	return r.resource.Namespace
}

func (r *ThanosRuler) RuleResourceNamespaceSelector() (labels.Selector, error) {
	if r.resource.Spec.RuleNamespaceSelector == nil {
		return nil, nil
	}
	return metav1.LabelSelectorAsSelector(r.resource.Spec.RuleNamespaceSelector)
}

func (r *ThanosRuler) RuleResourceSelector(extraRuleSelector labels.Selector) (labels.Selector, error) {
	rSelector, err := metav1.LabelSelectorAsSelector(r.resource.Spec.RuleSelector)
	if err != nil {
		return nil, err
	}
	if extraRuleSelector != nil {
		if requirements, ok := extraRuleSelector.Requirements(); ok {
			rSelector = rSelector.Add(requirements...)
		}
	}
	return rSelector, nil
}

func (r *ThanosRuler) ExternalLabels() func() map[string]string {
	// rules gotten from thanos ruler endpoint include the labels
	lbls := make(map[string]string)
	if ls := r.resource.Spec.Labels; ls != nil {
		for k, v := range ls {
			lbls[k] = v
		}
	}
	return func() map[string]string {
		return lbls
	}
}

func (r *ThanosRuler) ListRuleResources(ruleNamespace *corev1.Namespace, extraRuleSelector labels.Selector) (
	[]*promresourcesv1.PrometheusRule, error) {
	selected, err := ruleNamespaceSelected(r, ruleNamespace)
	if err != nil {
		return nil, err
	}
	if !selected {
		return nil, nil
	}
	rSelector, err := r.RuleResourceSelector(extraRuleSelector)
	if err != nil {
		return nil, err
	}
	return r.informer.Lister().PrometheusRules(ruleNamespace.Name).List(rSelector)
}

func (r *ThanosRuler) AddAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector,
	group string, rule *promresourcesv1.Rule, ruleResourceLabels map[string]string) error {

	prometheusRules, err := r.ListRuleResources(ruleNamespace, extraRuleResourceSelector)
	if err != nil {
		return err
	}

	return r.addAlertingRule(ctx, ruleNamespace, prometheusRules, nil, group, rule, ruleResourceLabels)
}

func (r *ThanosRuler) addAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	prometheusRules []*promresourcesv1.PrometheusRule, excludeRuleResources map[string]*ruleResource,
	group string, rule *promresourcesv1.Rule, ruleResourceLabels map[string]string) error {

	sort.Slice(prometheusRules, func(i, j int) bool {
		return len(fmt.Sprint(prometheusRules[i])) <= len(fmt.Sprint(prometheusRules[j]))
	})

	for _, prometheusRule := range prometheusRules {
		if len(excludeRuleResources) > 0 {
			if _, ok := excludeRuleResources[prometheusRule.Name]; ok {
				continue
			}
		}
		resource := ruleResource(*prometheusRule)
		if ok, err := resource.addAlertingRule(group, rule); err != nil {
			if err == errOutOfConfigMapSize {
				break
			}
			return err
		} else if ok {
			if err = resource.commit(ctx, r.client); err != nil {
				return err
			}
			return nil
		}
	}
	// create a new rule resource and add rule into it when all existing rule resources are full.
	newPromRule := promresourcesv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    ruleNamespace.Name,
			GenerateName: customAlertingRuleResourcePrefix,
			Labels:       ruleResourceLabels,
		},
		Spec: promresourcesv1.PrometheusRuleSpec{
			Groups: []promresourcesv1.RuleGroup{{
				Name:  group,
				Rules: []promresourcesv1.Rule{*rule},
			}},
		},
	}
	if _, err := r.client.MonitoringV1().
		PrometheusRules(ruleNamespace.Name).Create(ctx, &newPromRule, metav1.CreateOptions{}); err != nil {
		return errors.Wrapf(err, "error creating a prometheusrule resource %s/%s",
			newPromRule.Namespace, newPromRule.Name)
	}
	return nil
}

func (r *ThanosRuler) UpdateAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector,
	group string, rule *promresourcesv1.Rule, ruleResourceLabels map[string]string) error {

	prometheusRules, err := r.ListRuleResources(ruleNamespace, extraRuleResourceSelector)
	if err != nil {
		return err
	}

	var (
		found              bool
		success            bool
		resourcesToDelRule = make(map[string]*ruleResource)
	)
	for _, prometheusRule := range prometheusRules {
		resource := ruleResource(*prometheusRule)
		if success { // If the update has been successful, delete the possible same rule in other resources
			if ok, err := resource.deleteAlertingRule(rule.Alert); err != nil {
				return err
			} else if ok {
				if err = resource.commit(ctx, r.client); err != nil {
					return err
				}
			}
			continue
		}
		if ok, err := resource.updateAlertingRule(group, rule); err != nil {
			if err == errOutOfConfigMapSize {
				// updating the rule in the resource will oversize the size limit,
				// so delete it and then add the new rule to a new resource.
				resourcesToDelRule[resource.Name] = &resource
				found = true
			} else {
				return err
			}
		} else if ok {
			if err = resource.commit(ctx, r.client); err != nil {
				return err
			}
			found = true
			success = true
		}
	}

	if !found {
		return v1alpha1.ErrAlertingRuleNotFound
	}

	if !success {
		err := r.addAlertingRule(ctx, ruleNamespace, prometheusRules, resourcesToDelRule, group, rule, ruleResourceLabels)
		if err != nil {
			return err
		}
	}
	for _, resource := range resourcesToDelRule {
		if ok, err := resource.deleteAlertingRule(rule.Alert); err != nil {
			return err
		} else if ok {
			if err = resource.commit(ctx, r.client); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ThanosRuler) DeleteAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, group string, name string) error {
	prometheusRules, err := r.ListRuleResources(ruleNamespace, extraRuleResourceSelector)
	if err != nil {
		return err
	}
	var success bool
	for _, prometheusRule := range prometheusRules {
		resource := ruleResource(*prometheusRule)
		if ok, err := resource.deleteAlertingRule(name); err != nil {
			return err
		} else if ok {
			if err = resource.commit(ctx, r.client); err != nil {
				return err
			}
			success = true
		}
	}
	if !success {
		return v1alpha1.ErrAlertingRuleNotFound
	}
	return nil
}

func ruleNamespaceSelected(r Ruler, ruleNamespace *corev1.Namespace) (bool, error) {
	rnSelector, err := r.RuleResourceNamespaceSelector()
	if err != nil {
		return false, err
	}
	if rnSelector == nil { // refer to the comment of Prometheus.Spec.RuleResourceNamespaceSelector
		if r.Namespace() != ruleNamespace.Name {
			return false, nil
		}
	} else {
		if !rnSelector.Matches(labels.Set(ruleNamespace.Labels)) {
			return false, nil
		}
	}
	return true, nil
}
