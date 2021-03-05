package alerting

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prominformersv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions/monitoring/v1"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreinformersv1 "k8s.io/client-go/informers/core/v1"
	"kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/alerting/rules"
	"kubesphere.io/kubesphere/pkg/simple/client/alerting"
)

const (
	rulerNamespace = constants.KubeSphereMonitoringNamespace

	customRuleResourceLabelKeyLevel = "custom-alerting-rule-level"
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
		queryParams *v2alpha1.AlertingRuleQueryParams) (*v2alpha1.GettableAlertingRuleList, error)
	// ListCustomRulesAlerts lists the alerts of the custom alerting rules.
	ListCustomRulesAlerts(ctx context.Context, namespace string,
		queryParams *v2alpha1.AlertQueryParams) (*v2alpha1.AlertList, error)
	// GetCustomAlertingRule gets the custom alerting rule with the given name.
	GetCustomAlertingRule(ctx context.Context, namespace, ruleName string) (*v2alpha1.GettableAlertingRule, error)
	// ListCustomRuleAlerts lists the alerts of the custom alerting rule with the given name.
	ListCustomRuleAlerts(ctx context.Context, namespace, ruleName string) (*v2alpha1.AlertList, error)
	// CreateCustomAlertingRule creates a custom alerting rule.
	CreateCustomAlertingRule(ctx context.Context, namespace string, rule *v2alpha1.PostableAlertingRule) error
	// UpdateCustomAlertingRule updates the custom alerting rule with the given name.
	UpdateCustomAlertingRule(ctx context.Context, namespace, ruleName string, rule *v2alpha1.PostableAlertingRule) error
	// DeleteCustomAlertingRule deletes the custom alerting rule with the given name.
	DeleteCustomAlertingRule(ctx context.Context, namespace, ruleName string) error

	// CreateOrUpdateCustomAlertingRules creates or updates custom alerting rules in bulk.
	CreateOrUpdateCustomAlertingRules(ctx context.Context, namespace string, rs []*v2alpha1.PostableAlertingRule) (*v2alpha1.BulkResponse, error)
	// DeleteCustomAlertingRules deletes a batch of custom alerting rules.
	DeleteCustomAlertingRules(ctx context.Context, namespace string, ruleNames []string) (*v2alpha1.BulkResponse, error)

	// ListBuiltinAlertingRules lists the builtin(non-custom) alerting rules
	ListBuiltinAlertingRules(ctx context.Context,
		queryParams *v2alpha1.AlertingRuleQueryParams) (*v2alpha1.GettableAlertingRuleList, error)
	// ListBuiltinRulesAlerts lists the alerts of the builtin(non-custom) alerting rules
	ListBuiltinRulesAlerts(ctx context.Context,
		queryParams *v2alpha1.AlertQueryParams) (*v2alpha1.AlertList, error)
	// GetBuiltinAlertingRule gets the builtin(non-custom) alerting rule with the given id
	GetBuiltinAlertingRule(ctx context.Context, ruleId string) (*v2alpha1.GettableAlertingRule, error)
	// ListBuiltinRuleAlerts lists the alerts of the builtin(non-custom) alerting rule with the given id
	ListBuiltinRuleAlerts(ctx context.Context, ruleId string) (*v2alpha1.AlertList, error)
}

func NewOperator(informers informers.InformerFactory,
	promResourceClient promresourcesclient.Interface, ruleClient alerting.RuleClient,
	option *alerting.Options) Operator {
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
			lbl := strings.Split(strings.TrimSpace(lblString), "=")
			if len(lbl) == 2 {
				o.thanosRuleResourceLabels[lbl[0]] = lbl[1]
			}
		}
	}

	return &o
}

type operator struct {
	ruleClient alerting.RuleClient

	promResourceClient promresourcesclient.Interface

	prometheusInformer   prominformersv1.PrometheusInformer
	thanosRulerInformer  prominformersv1.ThanosRulerInformer
	ruleResourceInformer prominformersv1.PrometheusRuleInformer

	namespaceInformer coreinformersv1.NamespaceInformer

	resourceRuleCache *rules.RuleCache

	thanosRuleResourceLabels map[string]string
}

func (o *operator) ListCustomAlertingRules(ctx context.Context, namespace string,
	queryParams *v2alpha1.AlertingRuleQueryParams) (*v2alpha1.GettableAlertingRuleList, error) {

	var level v2alpha1.RuleLevel
	if namespace == "" {
		namespace = rulerNamespace
		level = v2alpha1.RuleLevelCluster
	} else {
		level = v2alpha1.RuleLevelNamespace
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
	queryParams *v2alpha1.AlertQueryParams) (*v2alpha1.AlertList, error) {

	var level v2alpha1.RuleLevel
	if namespace == "" {
		namespace = rulerNamespace
		level = v2alpha1.RuleLevelCluster
	} else {
		level = v2alpha1.RuleLevelNamespace
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
	*v2alpha1.GettableAlertingRule, error) {

	var level v2alpha1.RuleLevel
	if namespace == "" {
		namespace = rulerNamespace
		level = v2alpha1.RuleLevelCluster
	} else {
		level = v2alpha1.RuleLevelNamespace
	}

	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	return o.getCustomAlertingRule(ctx, ruleNamespace, ruleName, level)
}

func (o *operator) ListCustomRuleAlerts(ctx context.Context, namespace, ruleName string) (
	*v2alpha1.AlertList, error) {

	rule, err := o.GetCustomAlertingRule(ctx, namespace, ruleName)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, v2alpha1.ErrAlertingRuleNotFound
	}
	return &v2alpha1.AlertList{
		Total: len(rule.Alerts),
		Items: rule.Alerts,
	}, nil
}

func (o *operator) ListBuiltinAlertingRules(ctx context.Context,
	queryParams *v2alpha1.AlertingRuleQueryParams) (*v2alpha1.GettableAlertingRuleList, error) {

	alertingRules, err := o.listBuiltinAlertingRules(ctx)
	if err != nil {
		return nil, err
	}

	return pageAlertingRules(alertingRules, queryParams), nil
}

func (o *operator) ListBuiltinRulesAlerts(ctx context.Context,
	queryParams *v2alpha1.AlertQueryParams) (*v2alpha1.AlertList, error) {
	alertingRules, err := o.listBuiltinAlertingRules(ctx)
	if err != nil {
		return nil, err
	}

	return pageAlerts(alertingRules, queryParams), nil
}

func (o *operator) GetBuiltinAlertingRule(ctx context.Context, ruleId string) (
	*v2alpha1.GettableAlertingRule, error) {

	return o.getBuiltinAlertingRule(ctx, ruleId)
}

func (o *operator) ListBuiltinRuleAlerts(ctx context.Context, ruleId string) (*v2alpha1.AlertList, error) {
	rule, err := o.getBuiltinAlertingRule(ctx, ruleId)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, v2alpha1.ErrAlertingRuleNotFound
	}
	return &v2alpha1.AlertList{
		Total: len(rule.Alerts),
		Items: rule.Alerts,
	}, nil
}

func (o *operator) listCustomAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	level v2alpha1.RuleLevel) ([]*v2alpha1.GettableAlertingRule, error) {

	ruler, err := o.getThanosRuler()
	if err != nil {
		return nil, err
	}
	if ruler == nil {
		return nil, v2alpha1.ErrThanosRulerNotEnabled
	}

	resourceRulesMap, err := o.resourceRuleCache.ListRules(ruler, ruleNamespace,
		labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)}))
	if err != nil {
		return nil, err
	}

	if len(resourceRulesMap) == 0 {
		return nil, nil
	}

	ruleGroups, err := o.ruleClient.ThanosRules(ctx)
	if err != nil {
		return nil, err
	}

	return rules.GetAlertingRulesStatus(ruleNamespace.Name, &rules.ResourceRuleChunk{
		ResourceRulesMap: resourceRulesMap,
		Custom:           true,
		Level:            level,
	}, ruleGroups, ruler.ExternalLabels())
}

func (o *operator) getCustomAlertingRule(ctx context.Context, ruleNamespace *corev1.Namespace,
	ruleName string, level v2alpha1.RuleLevel) (*v2alpha1.GettableAlertingRule, error) {

	ruler, err := o.getThanosRuler()
	if err != nil {
		return nil, err
	}
	if ruler == nil {
		return nil, v2alpha1.ErrThanosRulerNotEnabled
	}

	resourceRule, err := o.resourceRuleCache.GetRule(ruler, ruleNamespace,
		labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)}), ruleName)
	if err != nil {
		return nil, err
	}
	if resourceRule == nil {
		return nil, v2alpha1.ErrAlertingRuleNotFound
	}

	ruleGroups, err := o.ruleClient.ThanosRules(ctx)
	if err != nil {
		return nil, err
	}

	return rules.GetAlertingRuleStatus(ruleNamespace.Name, &rules.ResourceRule{
		ResourceRuleItem: *resourceRule,
		Custom:           true,
		Level:            level,
	}, ruleGroups, ruler.ExternalLabels())
}

func (o *operator) listBuiltinAlertingRules(ctx context.Context) (
	[]*v2alpha1.GettableAlertingRule, error) {

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
		return rules.ParseAlertingRules(ruleGroups, false, v2alpha1.RuleLevelCluster,
			func(group, id string, rule *alerting.AlertingRule) bool {
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

	return rules.GetAlertingRulesStatus(ruleNamespace.Name, &rules.ResourceRuleChunk{
		ResourceRulesMap: resourceRulesMap,
		Custom:           false,
		Level:            v2alpha1.RuleLevelCluster,
	}, ruleGroups, ruler.ExternalLabels())
}

func (o *operator) getBuiltinAlertingRule(ctx context.Context, ruleId string) (*v2alpha1.GettableAlertingRule, error) {

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
		alertingRules, err := rules.ParseAlertingRules(ruleGroups, false, v2alpha1.RuleLevelCluster,
			func(group, id string, rule *alerting.AlertingRule) bool {
				return ruleId == id
			})
		if err != nil {
			return nil, err
		}
		if len(alertingRules) == 0 {
			return nil, v2alpha1.ErrAlertingRuleNotFound
		}
		sort.Slice(alertingRules, func(i, j int) bool {
			return v2alpha1.AlertingRuleIdCompare(alertingRules[i].Id, alertingRules[j].Id)
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
		return nil, v2alpha1.ErrAlertingRuleNotFound
	}

	return rules.GetAlertingRuleStatus(ruleNamespace.Name, &rules.ResourceRule{
		ResourceRuleItem: *resourceRule,
		Custom:           false,
		Level:            v2alpha1.RuleLevelCluster,
	}, ruleGroups, ruler.ExternalLabels())
}

func (o *operator) CreateCustomAlertingRule(ctx context.Context, namespace string,
	rule *v2alpha1.PostableAlertingRule) error {
	ruler, err := o.getThanosRuler()
	if err != nil {
		return err
	}
	if ruler == nil {
		return v2alpha1.ErrThanosRulerNotEnabled
	}

	var (
		level              v2alpha1.RuleLevel
		ruleResourceLabels = make(map[string]string)
	)
	for k, v := range o.thanosRuleResourceLabels {
		ruleResourceLabels[k] = v
	}
	if namespace == "" {
		namespace = rulerNamespace
		level = v2alpha1.RuleLevelCluster
	} else {
		level = v2alpha1.RuleLevelNamespace
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
		return v2alpha1.ErrAlertingRuleAlreadyExists
	}

	setRuleUpdateTime(rule, time.Now())

	respItems, err := ruler.AddAlertingRules(ctx, ruleNamespace, extraRuleResourceSelector,
		ruleResourceLabels, &rules.RuleWithGroup{Rule: *parseToPrometheusRule(rule)})
	if err != nil {
		return err
	}
	for _, item := range respItems {
		if item.Status == v2alpha1.StatusError {
			if item.ErrorType == v2alpha1.ErrNotFound {
				return v2alpha1.ErrAlertingRuleNotFound
			}
			return item.Error
		}
	}
	return nil
}

func (o *operator) UpdateCustomAlertingRule(ctx context.Context, namespace, name string,
	rule *v2alpha1.PostableAlertingRule) error {

	rule.Name = name

	ruler, err := o.getThanosRuler()
	if err != nil {
		return err
	}
	if ruler == nil {
		return v2alpha1.ErrThanosRulerNotEnabled
	}

	var (
		level              v2alpha1.RuleLevel
		ruleResourceLabels = make(map[string]string)
	)
	for k, v := range o.thanosRuleResourceLabels {
		ruleResourceLabels[k] = v
	}
	if namespace == "" {
		namespace = rulerNamespace
		level = v2alpha1.RuleLevelCluster
	} else {
		level = v2alpha1.RuleLevelNamespace
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
		return v2alpha1.ErrAlertingRuleNotFound
	}

	setRuleUpdateTime(rule, time.Now())

	respItems, err := ruler.UpdateAlertingRules(ctx, ruleNamespace, extraRuleResourceSelector, ruleResourceLabels,
		&rules.ResourceRuleItem{ResourceName: resourceRule.ResourceName,
			RuleWithGroup: rules.RuleWithGroup{Group: resourceRule.Group, Rule: *parseToPrometheusRule(rule)}})
	if err != nil {
		return err
	}
	for _, item := range respItems {
		if item.Status == v2alpha1.StatusError {
			if item.ErrorType == v2alpha1.ErrNotFound {
				return v2alpha1.ErrAlertingRuleNotFound
			}
			return item.Error
		}
	}
	return nil
}

func (o *operator) DeleteCustomAlertingRule(ctx context.Context, namespace, name string) error {
	ruler, err := o.getThanosRuler()
	if err != nil {
		return err
	}
	if ruler == nil {
		return v2alpha1.ErrThanosRulerNotEnabled
	}

	var (
		level v2alpha1.RuleLevel
	)
	if namespace == "" {
		namespace = rulerNamespace
		level = v2alpha1.RuleLevelCluster
	} else {
		level = v2alpha1.RuleLevelNamespace
	}

	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return err
	}

	extraRuleResourceSelector := labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)})
	resourceRules, err := o.resourceRuleCache.GetRuleByIdOrName(ruler, ruleNamespace, extraRuleResourceSelector, name)
	if err != nil {
		return err
	}
	if len(resourceRules) == 0 {
		return v2alpha1.ErrAlertingRuleNotFound
	}

	respItems, err := ruler.DeleteAlertingRules(ctx, ruleNamespace, resourceRules...)
	if err != nil {
		return err
	}
	for _, item := range respItems {
		if item.Status == v2alpha1.StatusError {
			if item.ErrorType == v2alpha1.ErrNotFound {
				return v2alpha1.ErrAlertingRuleNotFound
			}
			return item.Error
		}
	}
	return nil
}

func (o *operator) CreateOrUpdateCustomAlertingRules(ctx context.Context, namespace string,
	rs []*v2alpha1.PostableAlertingRule) (*v2alpha1.BulkResponse, error) {

	if l := len(rs); l == 0 {
		return &v2alpha1.BulkResponse{}, nil
	}

	ruler, err := o.getThanosRuler()
	if err != nil {
		return nil, err
	}
	if ruler == nil {
		return nil, v2alpha1.ErrThanosRulerNotEnabled
	}

	var (
		level              v2alpha1.RuleLevel
		ruleResourceLabels = make(map[string]string)
	)
	for k, v := range o.thanosRuleResourceLabels {
		ruleResourceLabels[k] = v
	}
	if namespace == "" {
		namespace = rulerNamespace
		level = v2alpha1.RuleLevelCluster
	} else {
		level = v2alpha1.RuleLevelNamespace
	}
	ruleResourceLabels[customRuleResourceLabelKeyLevel] = string(level)
	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	extraRuleResourceSelector := labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)})

	resourceRulesMap, err := o.resourceRuleCache.ListRules(ruler, ruleNamespace, extraRuleResourceSelector)
	if err != nil {
		return nil, err
	}
	exists := make(map[string][]*rules.ResourceRuleItem)
	for _, c := range resourceRulesMap {
		for n, items := range c.NameRules {
			exists[n] = append(exists[n], items...)
		}
	}

	// check all the rules
	var (
		br       = &v2alpha1.BulkResponse{}
		nameSet  = make(map[string]struct{})
		invalids = make(map[string]struct{})
	)
	for i := range rs {
		var (
			r    = rs[i]
			name = r.Name
		)

		if _, ok := nameSet[name]; ok {
			br.Items = append(br.Items, &v2alpha1.BulkItemResponse{
				RuleName:  name,
				Status:    v2alpha1.StatusError,
				ErrorType: v2alpha1.ErrDuplicateName,
				Error:     errors.Errorf("There is more than one rule named %s in the bulk request", name),
			})
			invalids[name] = struct{}{}
			continue
		} else {
			nameSet[name] = struct{}{}
		}
		if err := r.Validate(); err != nil {
			br.Items = append(br.Items, &v2alpha1.BulkItemResponse{
				RuleName:  name,
				Status:    v2alpha1.StatusError,
				ErrorType: v2alpha1.ErrBadData,
				Error:     err,
			})
			invalids[name] = struct{}{}
			continue
		}
		if level == v2alpha1.RuleLevelNamespace {
			expr, err := rules.InjectExprNamespaceLabel(r.Query, namespace)
			if err != nil {
				br.Items = append(br.Items, v2alpha1.NewBulkItemErrorServerResponse(name, err))
				invalids[name] = struct{}{}
				continue
			}
			r.Query = expr
		}
	}
	if len(nameSet) == len(invalids) {
		return br.MakeBulkResponse(), nil
	}

	// Confirm whether the rules should be added or updated. For each rule that is committed,
	// it will be added if the rule does not exist, or updated otherwise.
	// If there are rules with the same name in the existing rules to update, the first will be
	// updated but the others will be deleted
	var (
		addRules []*rules.RuleWithGroup
		updRules []*rules.ResourceRuleItem
		delRules []*rules.ResourceRuleItem // duplicate rules that need to deleted in other resources

		updateTime = time.Now()
	)
	for i := range rs {
		r := rs[i]
		if _, ok := invalids[r.Name]; ok {
			continue
		}
		setRuleUpdateTime(r, updateTime)
		if items, ok := exists[r.Name]; ok && len(items) > 0 {
			item := items[0]
			updRules = append(updRules, &rules.ResourceRuleItem{
				ResourceName:  item.ResourceName,
				RuleWithGroup: rules.RuleWithGroup{Group: item.Group, Rule: *parseToPrometheusRule(r)}})
			if len(items) > 1 {
				for j := 1; j < len(items); j++ {
					if items[j].ResourceName != item.ResourceName {
						delRules = append(delRules, items[j])
					}
				}
			}
		} else {
			addRules = append(addRules, &rules.RuleWithGroup{Rule: *parseToPrometheusRule(r)})
		}
	}

	// add rules
	if len(addRules) > 0 {
		resps, err := ruler.AddAlertingRules(ctx, ruleNamespace, extraRuleResourceSelector, ruleResourceLabels, addRules...)
		if err == nil {
			br.Items = append(br.Items, resps...)
		} else {
			for _, rule := range addRules {
				br.Items = append(br.Items, v2alpha1.NewBulkItemErrorServerResponse(rule.Alert, err))
			}
		}
	}
	// update existing rules
	if len(updRules) > 0 {
		resps, err := ruler.UpdateAlertingRules(ctx, ruleNamespace, extraRuleResourceSelector, ruleResourceLabels, updRules...)
		if err == nil {
			br.Items = append(br.Items, resps...)
		} else {
			for _, rule := range updRules {
				br.Items = append(br.Items, v2alpha1.NewBulkItemErrorServerResponse(rule.Alert, err))
			}
		}
	}
	// delete possible duplicate rules
	if len(delRules) > 0 {
		_, err := ruler.DeleteAlertingRules(ctx, ruleNamespace, delRules...)
		if err != nil {
			for _, rule := range delRules {
				br.Items = append(br.Items, v2alpha1.NewBulkItemErrorServerResponse(rule.Alert, err))
			}
		}
	}
	return br.MakeBulkResponse(), nil
}

func (o *operator) DeleteCustomAlertingRules(ctx context.Context, namespace string,
	names []string) (*v2alpha1.BulkResponse, error) {

	if l := len(names); l == 0 {
		return &v2alpha1.BulkResponse{}, nil
	}

	ruler, err := o.getThanosRuler()
	if err != nil {
		return nil, err
	}
	if ruler == nil {
		return nil, v2alpha1.ErrThanosRulerNotEnabled
	}

	var (
		level v2alpha1.RuleLevel
	)
	if namespace == "" {
		namespace = rulerNamespace
		level = v2alpha1.RuleLevelCluster
	} else {
		level = v2alpha1.RuleLevelNamespace
	}
	ruleNamespace, err := o.namespaceInformer.Lister().Get(namespace)
	if err != nil {
		return nil, err
	}

	extraRuleResourceSelector := labels.SelectorFromSet(labels.Set{customRuleResourceLabelKeyLevel: string(level)})
	resourceRulesMap, err := o.resourceRuleCache.ListRules(ruler, ruleNamespace, extraRuleResourceSelector)
	if err != nil {
		return nil, err
	}
	exists := make(map[string][]*rules.ResourceRuleItem)
	for _, c := range resourceRulesMap {
		for n, items := range c.NameRules {
			exists[n] = append(exists[n], items...)
		}
	}

	br := &v2alpha1.BulkResponse{}
	var ruleItems []*rules.ResourceRuleItem
	for _, n := range names {
		if items, ok := exists[n]; ok {
			ruleItems = append(ruleItems, items...)
		} else {
			br.Items = append(br.Items, &v2alpha1.BulkItemResponse{
				RuleName:  n,
				Status:    v2alpha1.StatusError,
				ErrorType: v2alpha1.ErrNotFound,
			})
		}
	}

	respItems, err := ruler.DeleteAlertingRules(ctx, ruleNamespace, ruleItems...)
	if err != nil {
		return nil, err
	}
	br.Items = append(br.Items, respItems...)

	return br.MakeBulkResponse(), nil
}

// getPrometheusRuler gets the cluster-in prometheus
func (o *operator) getPrometheusRuler() (rules.Ruler, error) {
	prometheuses, err := o.prometheusInformer.Lister().Prometheuses(rulerNamespace).List(labels.Everything())
	if err != nil {
		return nil, errors.Wrap(err, "error listing prometheuses")
	}
	if len(prometheuses) > 1 {
		// It is not supported to have multiple Prometheus instances in the monitoring namespace for now
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
		// It is not supported to have multiple thanosruler instances in the monitoring namespace for now
		return nil, errors.Errorf(
			"there is more than one thanosruler custom resource in %s", rulerNamespace)
	}
	if len(thanosrulers) == 0 {
		// if there is no thanos ruler, custom rules will not be supported
		return nil, nil
	}

	return rules.NewThanosRuler(thanosrulers[0], o.ruleResourceInformer, o.promResourceClient), nil
}

func parseToPrometheusRule(rule *v2alpha1.PostableAlertingRule) *promresourcesv1.Rule {
	if _, ok := rule.Labels[rules.LabelKeyAlertType]; !ok {
		rule.Labels[rules.LabelKeyAlertType] = rules.LabelValueAlertType
	}
	return &promresourcesv1.Rule{
		Alert:       rule.Name,
		Expr:        intstr.FromString(rule.Query),
		For:         rule.Duration,
		Labels:      rule.Labels,
		Annotations: rule.Annotations,
	}
}

func pageAlertingRules(alertingRules []*v2alpha1.GettableAlertingRule,
	queryParams *v2alpha1.AlertingRuleQueryParams) *v2alpha1.GettableAlertingRuleList {

	alertingRules = queryParams.Filter(alertingRules)
	queryParams.Sort(alertingRules)

	return &v2alpha1.GettableAlertingRuleList{
		Total: len(alertingRules),
		Items: queryParams.Sub(alertingRules),
	}
}

func pageAlerts(alertingRules []*v2alpha1.GettableAlertingRule,
	queryParams *v2alpha1.AlertQueryParams) *v2alpha1.AlertList {

	var alerts []*v2alpha1.Alert
	for _, rule := range alertingRules {
		alerts = append(alerts, queryParams.Filter(rule.Alerts)...)
	}
	queryParams.Sort(alerts)

	return &v2alpha1.AlertList{
		Total: len(alerts),
		Items: queryParams.Sub(alerts),
	}
}

func setRuleUpdateTime(rule *v2alpha1.PostableAlertingRule, t time.Time) {
	if rule.Annotations == nil {
		rule.Annotations = make(map[string]string)
	}
	if t.IsZero() {
		t = time.Now()
	}
	rule.Annotations[v2alpha1.AnnotationKeyRuleUpdateTime] = t.UTC().Format(time.RFC3339)
}
