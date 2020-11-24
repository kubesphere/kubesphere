package customalerting

import (
	"context"
	"sort"
	"strings"

	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prominformersv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions/monitoring/v1"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreinformersv1 "k8s.io/client-go/informers/core/v1"
	"kubesphere.io/kubesphere/pkg/api/customalerting/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/customalerting/rules"
	"kubesphere.io/kubesphere/pkg/simple/client/customalerting"
)

const (
	rulerNamespace                  = constants.KubeSphereMonitoringNamespace
	customRuleGroupDefault          = "alerting.custom.defaults"
	customRuleResourceLabelKeyLevel = "custom-alerting-rule-level"
)

var (
	maxSecretSize        = corev1.MaxSecretSize
	maxConfigMapDataSize = int(float64(maxSecretSize) * 0.45)
)

// Operator contains all operations to alerting rules. The operations may involve manipulations of prometheusrule
// custom resources where the rules are persisted, and querying the rules state from prometheus endpoint and
// thanos ruler endpoint.
// For the following apis, if namespace is empty, do operations to alerting rules with cluster level,
// or do operations only to rules of the specified namespaces.
// All custom rules will be configured for thanos ruler, so the operations to custom alerting rule can not be done
// if thanos ruler is not enabled.
type Operator interface {
	// ListCustomAlertingRules lists the custom alerting rules.
	ListCustomAlertingRules(ctx context.Context, namespace string,
		queryParams *v1alpha1.AlertingRuleQueryParams) (*v1alpha1.GettableAlertingRuleList, error)
	// ListCustomRulesAlerts lists the alerts of the custom alerting rules.
	ListCustomRulesAlerts(ctx context.Context, namespace string,
		queryParams *v1alpha1.AlertQueryParams) (*v1alpha1.AlertList, error)
	// GetCustomAlertingRule gets the custom alerting rule with the given name.
	GetCustomAlertingRule(ctx context.Context, namespace, ruleName string) (*v1alpha1.GettableAlertingRule, error)
	// ListCustomSpecifiedRuleAlerts lists the alerts of the custom alerting rule with the given name.
	ListCustomSpecifiedRuleAlerts(ctx context.Context, namespace, ruleName string) ([]*v1alpha1.Alert, error)
	// CreateCustomAlertingRule creates a custom alerting rule.
	CreateCustomAlertingRule(ctx context.Context, namespace string, rule *v1alpha1.PostableAlertingRule) error
	// UpdateCustomAlertingRule updates the custom alerting rule with the given name.
	UpdateCustomAlertingRule(ctx context.Context, namespace, ruleName string, rule *v1alpha1.PostableAlertingRule) error
	// DeleteCustomAlertingRule deletes the custom alerting rule with the given name.
	DeleteCustomAlertingRule(ctx context.Context, namespace, ruleName string) error

	// ListBuiltinAlertingRules lists the builtin(non-custom) alerting rules
	ListBuiltinAlertingRules(ctx context.Context,
		queryParams *v1alpha1.AlertingRuleQueryParams) (*v1alpha1.GettableAlertingRuleList, error)
	// ListBuiltinRulesAlerts lists the alerts of the builtin(non-custom) alerting rules
	ListBuiltinRulesAlerts(ctx context.Context,
		queryParams *v1alpha1.AlertQueryParams) (*v1alpha1.AlertList, error)
	// GetBuiltinAlertingRule gets the builtin(non-custom) alerting rule with the given id
	GetBuiltinAlertingRule(ctx context.Context, ruleId string) (*v1alpha1.GettableAlertingRule, error)
	// ListBuiltinSpecifiedRuleAlerts lists the alerts of the builtin(non-custom) alerting rule with the given id
	ListBuiltinSpecifiedRuleAlerts(ctx context.Context, ruleId string) ([]*v1alpha1.Alert, error)
}

func NewOperator(informers informers.InformerFactory,
	promResourceClient promresourcesclient.Interface, ruleClient customalerting.RuleClient,
	option *customalerting.Options) Operator {
	o := operator{
		namespaceInformer: informers.KubernetesSharedInformerFactory().Core().V1().Namespaces(),

		promResourceClient: promResourceClient,

		prometheusInformer:   informers.PrometheusSharedInformerFactory().Monitoring().V1().Prometheuses(),
		thanosRulerInformer:  informers.PrometheusSharedInformerFactory().Monitoring().V1().ThanosRulers(),
		ruleResourceInformer: informers.PrometheusSharedInformerFactory().Monitoring().V1().PrometheusRules(),

		ruleClient: ruleClient,

		thanosRuleResourceLabels: make(map[string]string),
	}

	o.resourceRuleCache = rules.NewRuleCache(o.ruleResourceInformer)

	if option != nil && len(option.ThanosRuleResourceLabels) != 0 {
		lblStrings := strings.Split(option.ThanosRuleResourceLabels, ",")
		for _, lblString := range lblStrings {
			lbl := strings.Split(lblString, "=")
			if len(lbl) == 2 {
				o.thanosRuleResourceLabels[lbl[0]] = lbl[1]
			}
		}
	}

	return &o
}

type operator struct {
	ruleClient customalerting.RuleClient

	promResourceClient promresourcesclient.Interface

	prometheusInformer   prominformersv1.PrometheusInformer
	thanosRulerInformer  prominformersv1.ThanosRulerInformer
	ruleResourceInformer prominformersv1.PrometheusRuleInformer

	namespaceInformer coreinformersv1.NamespaceInformer

	resourceRuleCache *rules.RuleCache

	thanosRuleResourceLabels map[string]string
}

func (o *operator) ListCustomAlertingRules(ctx context.Context, namespace string,
	queryParams *v1alpha1.AlertingRuleQueryParams) (*v1alpha1.GettableAlertingRuleList, error) {

	var level v1alpha1.RuleLevel
	if namespace == "" {
		namespace = rulerNamespace
		level = v1alpha1.RuleLevelCluster
	} else {
		level = v1alpha1.RuleLevelNamespace
	}

	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	alertingRules, err := o.listCustomAlertingRules(ctx, ruleNamespace, level)
	if err != nil {
		return nil, err
	}

	return pageAlertingRules(alertingRules, queryParams), nil
}

func (o *operator) ListCustomRulesAlerts(ctx context.Context, namespace string,
	queryParams *v1alpha1.AlertQueryParams) (*v1alpha1.AlertList, error) {

	var level v1alpha1.RuleLevel
	if namespace == "" {
		namespace = rulerNamespace
		level = v1alpha1.RuleLevelCluster
	} else {
		level = v1alpha1.RuleLevelNamespace
	}

	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	alertingRules, err := o.listCustomAlertingRules(ctx, ruleNamespace, level)
	if err != nil {
		return nil, err
	}

	return pageAlerts(alertingRules, queryParams), nil
}

func (o *operator) GetCustomAlertingRule(ctx context.Context, namespace, ruleName string) (
	*v1alpha1.GettableAlertingRule, error) {

	var level v1alpha1.RuleLevel
	if namespace == "" {
		namespace = rulerNamespace
		level = v1alpha1.RuleLevelCluster
	} else {
		level = v1alpha1.RuleLevelNamespace
	}

	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	return o.getCustomAlertingRule(ctx, ruleNamespace, ruleName, level)
}

func (o *operator) ListCustomSpecifiedRuleAlerts(ctx context.Context, namespace, ruleName string) (
	[]*v1alpha1.Alert, error) {

	rule, err := o.GetCustomAlertingRule(ctx, namespace, ruleName)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, v1alpha1.ErrAlertingRuleNotFound
	}
	return rule.Alerts, nil
}

func (o *operator) ListBuiltinAlertingRules(ctx context.Context,
	queryParams *v1alpha1.AlertingRuleQueryParams) (*v1alpha1.GettableAlertingRuleList, error) {

	alertingRules, err := o.listBuiltinAlertingRules(ctx)
	if err != nil {
		return nil, err
	}

	return pageAlertingRules(alertingRules, queryParams), nil
}

func (o *operator) ListBuiltinRulesAlerts(ctx context.Context,
	queryParams *v1alpha1.AlertQueryParams) (*v1alpha1.AlertList, error) {
	alertingRules, err := o.listBuiltinAlertingRules(ctx)
	if err != nil {
		return nil, err
	}

	return pageAlerts(alertingRules, queryParams), nil
}

func (o *operator) GetBuiltinAlertingRule(ctx context.Context, ruleId string) (
	*v1alpha1.GettableAlertingRule, error) {

	return o.getBuiltinAlertingRule(ctx, ruleId)
}

func (o *operator) ListBuiltinSpecifiedRuleAlerts(ctx context.Context, ruleId string) ([]*v1alpha1.Alert, error) {
	rule, err := o.getBuiltinAlertingRule(ctx, ruleId)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, v1alpha1.ErrAlertingRuleNotFound
	}
	return rule.Alerts, nil
}

func (o *operator) ListClusterAlertingRules(ctx context.Context, customFlag string,
	queryParams *v1alpha1.AlertingRuleQueryParams) (*v1alpha1.GettableAlertingRuleList, error) {

	namespace := rulerNamespace
	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	alertingRules, err := o.listCustomAlertingRules(ctx, ruleNamespace, v1alpha1.RuleLevelCluster)
	if err != nil {
		return nil, err
	}

	return pageAlertingRules(alertingRules, queryParams), nil
}

func (o *operator) ListClusterRulesAlerts(ctx context.Context,
	queryParams *v1alpha1.AlertQueryParams) (*v1alpha1.AlertList, error) {

	namespace := rulerNamespace
	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	alertingRules, err := o.listCustomAlertingRules(ctx, ruleNamespace, v1alpha1.RuleLevelCluster)
	if err != nil {
		return nil, err
	}

	return pageAlerts(alertingRules, queryParams), nil
}

func (o *operator) listCustomAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	level v1alpha1.RuleLevel) ([]*v1alpha1.GettableAlertingRule, error) {

	ruler, err := o.getThanosRuler()
	if err != nil {
		return nil, err
	}
	if ruler == nil {
		return nil, v1alpha1.ErrThanosRulerNotEnabled
	}

	resourceRulesMap, err := o.resourceRuleCache.ListRules(ruler, ruleNamespace,
		labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)}))
	if err != nil {
		return nil, err
	}

	ruleGroups, err := o.ruleClient.ThanosRules(ctx)
	if err != nil {
		return nil, err
	}

	return rules.MixAlertingRules(ruleNamespace.Name, &rules.ResourceRuleChunk{
		ResourceRulesMap: resourceRulesMap,
		Custom:           true,
		Level:            level,
	}, ruleGroups, ruler.ExternalLabels())
}

func (o *operator) getCustomAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	ruleName string, level v1alpha1.RuleLevel) (*v1alpha1.GettableAlertingRule, error) {

	ruler, err := o.getThanosRuler()
	if err != nil {
		return nil, err
	}
	if ruler == nil {
		return nil, v1alpha1.ErrThanosRulerNotEnabled
	}

	resourceRule, err := o.resourceRuleCache.GetRule(ruler, ruleNamespace,
		labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)}), ruleName)
	if err != nil {
		return nil, err
	}
	if resourceRule == nil {
		return nil, v1alpha1.ErrAlertingRuleNotFound
	}

	ruleGroups, err := o.ruleClient.ThanosRules(ctx)
	if err != nil {
		return nil, err
	}

	return rules.MixAlertingRule(ruleNamespace.Name, &rules.ResourceRuleSole{
		ResourceRule: *resourceRule,
		Custom:       true,
		Level:        level,
	}, ruleGroups, ruler.ExternalLabels())
}

func (o *operator) listBuiltinAlertingRules(ctx context.Context) (
	[]*v1alpha1.GettableAlertingRule, error) {

	ruler, err := o.getPrometheusRuler()
	if err != nil {
		return nil, err
	}

	ruleGroups, err := o.ruleClient.PrometheusRules(ctx)
	if err != nil {
		return nil, err
	}

	if ruler == nil {
		// for out-cluster prometheus
		return rules.ParseAlertingRules(ruleGroups, false, v1alpha1.RuleLevelCluster,
			func(group, id string, rule *customalerting.AlertingRule) bool {
				return true
			})
	}

	namespace := rulerNamespace
	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	resourceRulesMap, err := o.resourceRuleCache.ListRules(ruler, ruleNamespace, nil)
	if err != nil {
		return nil, err
	}

	return rules.MixAlertingRules(ruleNamespace.Name, &rules.ResourceRuleChunk{
		ResourceRulesMap: resourceRulesMap,
		Custom:           false,
		Level:            v1alpha1.RuleLevelCluster,
	}, ruleGroups, ruler.ExternalLabels())
}

func (o *operator) getBuiltinAlertingRule(ctx context.Context, ruleId string) (*v1alpha1.GettableAlertingRule, error) {

	ruler, err := o.getPrometheusRuler()
	if err != nil {
		return nil, err
	}

	ruleGroups, err := o.ruleClient.PrometheusRules(ctx)
	if err != nil {
		return nil, err
	}

	if ruler == nil {
		// for out-cluster prometheus
		alertingRules, err := rules.ParseAlertingRules(ruleGroups, false, v1alpha1.RuleLevelCluster,
			func(group, id string, rule *customalerting.AlertingRule) bool {
				return ruleId == id
			})
		if err != nil {
			return nil, err
		}
		if len(alertingRules) == 0 {
			return nil, v1alpha1.ErrAlertingRuleNotFound
		}
		sort.Slice(alertingRules, func(i, j int) bool {
			return v1alpha1.AlertingRuleIdCompare(alertingRules[i].Id, alertingRules[j].Id)
		})
		return alertingRules[0], nil
	}

	namespace := rulerNamespace
	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	resourceRule, err := o.resourceRuleCache.GetRule(ruler, ruleNamespace, nil, ruleId)
	if err != nil {
		return nil, err
	}

	if resourceRule == nil {
		return nil, v1alpha1.ErrAlertingRuleNotFound
	}

	return rules.MixAlertingRule(ruleNamespace.Name, &rules.ResourceRuleSole{
		ResourceRule: *resourceRule,
		Custom:       false,
		Level:        v1alpha1.RuleLevelCluster,
	}, ruleGroups, ruler.ExternalLabels())
}

func (o *operator) CreateCustomAlertingRule(ctx context.Context, namespace string,
	rule *v1alpha1.PostableAlertingRule) error {
	ruler, err := o.getThanosRuler()
	if err != nil {
		return err
	}
	if ruler == nil {
		return v1alpha1.ErrThanosRulerNotEnabled
	}

	var (
		level              v1alpha1.RuleLevel
		ruleResourceLabels = make(map[string]string)
	)
	for k, v := range o.thanosRuleResourceLabels {
		ruleResourceLabels[k] = v
	}
	if namespace == "" {
		namespace = rulerNamespace
		level = v1alpha1.RuleLevelCluster
	} else {
		level = v1alpha1.RuleLevelNamespace
		expr, err := rules.InjectExprNamespaceLabel(rule.Query, namespace)
		if err != nil {
			return err
		}
		rule.Query = expr
	}
	ruleResourceLabels[customRuleResourceLabelKeyLevel] = string(level)

	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return err
	}

	extraRuleResourceSelector := labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)})
	resourceRule, err := o.resourceRuleCache.GetRule(ruler, ruleNamespace, extraRuleResourceSelector, rule.Name)
	if err != nil {
		return err
	}
	if resourceRule != nil {
		return v1alpha1.ErrAlertingRuleAlreadyExists
	}

	return ruler.AddAlertingRule(ctx, ruleNamespace, extraRuleResourceSelector,
		customRuleGroupDefault, parseToPrometheusRule(rule), ruleResourceLabels)
}

func (o *operator) UpdateCustomAlertingRule(ctx context.Context, namespace, name string,
	rule *v1alpha1.PostableAlertingRule) error {

	rule.Name = name

	ruler, err := o.getThanosRuler()
	if err != nil {
		return err
	}
	if ruler == nil {
		return v1alpha1.ErrThanosRulerNotEnabled
	}

	var (
		level              v1alpha1.RuleLevel
		ruleResourceLabels = make(map[string]string)
	)
	for k, v := range o.thanosRuleResourceLabels {
		ruleResourceLabels[k] = v
	}
	if namespace == "" {
		namespace = rulerNamespace
		level = v1alpha1.RuleLevelCluster
	} else {
		level = v1alpha1.RuleLevelNamespace
		expr, err := rules.InjectExprNamespaceLabel(rule.Query, namespace)
		if err != nil {
			return err
		}
		rule.Query = expr
	}
	ruleResourceLabels[customRuleResourceLabelKeyLevel] = string(level)

	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return err
	}

	extraRuleResourceSelector := labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)})
	resourceRule, err := o.resourceRuleCache.GetRule(ruler, ruleNamespace, extraRuleResourceSelector, rule.Name)
	if err != nil {
		return err
	}
	if resourceRule == nil {
		return v1alpha1.ErrAlertingRuleNotFound
	}

	return ruler.UpdateAlertingRule(ctx, ruleNamespace, extraRuleResourceSelector,
		resourceRule.Group, parseToPrometheusRule(rule), ruleResourceLabels)
}

func (o *operator) DeleteCustomAlertingRule(ctx context.Context, namespace, name string) error {
	ruler, err := o.getThanosRuler()
	if err != nil {
		return err
	}
	if ruler == nil {
		return v1alpha1.ErrThanosRulerNotEnabled
	}

	var (
		level v1alpha1.RuleLevel
	)
	if namespace == "" {
		namespace = rulerNamespace
		level = v1alpha1.RuleLevelCluster
	} else {
		level = v1alpha1.RuleLevelNamespace
	}

	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return err
	}

	extraRuleResourceSelector := labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)})
	resourceRule, err := o.resourceRuleCache.GetRule(ruler, ruleNamespace, extraRuleResourceSelector, name)
	if err != nil {
		return err
	}
	if resourceRule == nil {
		return v1alpha1.ErrAlertingRuleNotFound
	}

	return ruler.DeleteAlertingRule(ctx, ruleNamespace, extraRuleResourceSelector, resourceRule.Group, name)
}

// getPrometheusRuler gets the cluster-in prometheus
func (o *operator) getPrometheusRuler() (rules.Ruler, error) {
	prometheuses, err := o.prometheusInformer.Lister().Prometheuses(rulerNamespace).List(labels.Everything())
	if err != nil {
		return nil, errors.Wrap(err, "error listing prometheuses")
	}
	if len(prometheuses) > 1 {
		// it is not supported temporarily to have multiple prometheuses in the monitoring namespace
		return nil, errors.Errorf(
			"there is more than one prometheus custom resource in %s", rulerNamespace)
	}
	if len(prometheuses) == 0 {
		return nil, nil
	}

	return rules.NewPrometheusRuler(prometheuses[0], o.ruleResourceInformer, o.promResourceClient), nil
}

func (o *operator) getThanosRuler() (rules.Ruler, error) {
	thanosrulers, err := o.thanosRulerInformer.Lister().ThanosRulers(rulerNamespace).List(labels.Everything())
	if err != nil {
		return nil, errors.Wrap(err, "error listing thanosrulers: ")
	}
	if len(thanosrulers) > 1 {
		// it is not supported temporarily to have multiple thanosrulers in the monitoring namespace
		return nil, errors.Errorf(
			"there is more than one thanosruler custom resource in %s", rulerNamespace)
	}
	if len(thanosrulers) == 0 {
		// if there is no thanos ruler, custom rules will not be supported
		return nil, nil
	}

	return rules.NewThanosRuler(thanosrulers[0], o.ruleResourceInformer, o.promResourceClient), nil
}

func parseToPrometheusRule(rule *v1alpha1.PostableAlertingRule) *promresourcesv1.Rule {
	lbls := rule.Labels
	lbls[rules.LabelKeyInternalRuleAlias] = rule.Alias
	lbls[rules.LabelKeyInternalRuleDescription] = rule.Description
	return &promresourcesv1.Rule{
		Alert:       rule.Name,
		Expr:        intstr.FromString(rule.Query),
		For:         rule.Duration,
		Labels:      lbls,
		Annotations: rule.Annotations,
	}
}

func pageAlertingRules(alertingRules []*v1alpha1.GettableAlertingRule,
	queryParams *v1alpha1.AlertingRuleQueryParams) *v1alpha1.GettableAlertingRuleList {

	alertingRules = queryParams.Filter(alertingRules)
	queryParams.Sort(alertingRules)

	return &v1alpha1.GettableAlertingRuleList{
		Total: len(alertingRules),
		Items: queryParams.Sub(alertingRules),
	}
}

func pageAlerts(alertingRules []*v1alpha1.GettableAlertingRule,
	queryParams *v1alpha1.AlertQueryParams) *v1alpha1.AlertList {

	var alerts []*v1alpha1.Alert
	for _, rule := range alertingRules {
		alerts = append(alerts, queryParams.Filter(rule.Alerts)...)
	}
	queryParams.Sort(alerts)

	return &v1alpha1.AlertList{
		Total: len(alerts),
		Items: queryParams.Sub(alerts),
	}
}
