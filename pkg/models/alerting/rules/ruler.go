package rules

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/docker/pkg/locker"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prominformersv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions/monitoring/v1"
	promresourcesclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
)

const (
	customAlertingRuleResourcePrefix = "custom-alerting-rule-"

	customRuleGroupDefaultPrefix = "alerting.custom.defaults."
	customRuleGroupSize          = 20
)

var (
	maxSecretSize        = corev1.MaxSecretSize
	maxConfigMapDataSize = int(float64(maxSecretSize) * 0.45)

	errOutOfConfigMapSize = errors.New("out of config map size")

	ruleResourceLocker locker.Locker
)

type Ruler interface {
	Namespace() string
	RuleResourceNamespaceSelector() (labels.Selector, error)
	RuleResourceSelector(extraRuleResourceSelector labels.Selector) (labels.Selector, error)
	ExternalLabels() func() map[string]string

	ListRuleResources(ruleNamespace *corev1.Namespace, extraRuleResourceSelector labels.Selector) (
		[]*promresourcesv1.PrometheusRule, error)
	AddAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace, extraRuleResourceSelector labels.Selector,
		ruleResourceLabels map[string]string, rules ...*RuleWithGroup) ([]*v2alpha1.BulkItemResponse, error)
	UpdateAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace, extraRuleResourceSelector labels.Selector,
		ruleResourceLabels map[string]string, ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error)
	DeleteAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
		ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error)
}

type ruleResource promresourcesv1.PrometheusRule

// deleteAlertingRules deletes the rules.
// If there are rules to be deleted, return true to indicate the resource should be updated.
func (r *ruleResource) deleteAlertingRules(rules ...*RuleWithGroup) (bool, error) {
	var (
		gs     []promresourcesv1.RuleGroup
		dels   = make(map[string]struct{})
		commit bool
	)

	for _, rule := range rules {
		if rule != nil {
			dels[rule.Alert] = struct{}{}
		}
	}

	for _, g := range r.Spec.Groups {
		var rules []promresourcesv1.Rule
		for _, gr := range g.Rules {
			if gr.Alert != "" {
				if _, ok := dels[gr.Alert]; ok {
					commit = true
					continue
				}
			}
			rules = append(rules, gr)
		}
		if len(rules) > 0 {
			gs = append(gs, promresourcesv1.RuleGroup{
				Name:                    g.Name,
				Interval:                g.Interval,
				PartialResponseStrategy: g.PartialResponseStrategy,
				Rules:                   rules,
			})
		}
	}

	if commit {
		r.Spec.Groups = gs
	}
	return commit, nil
}

// updateAlertingRules updates the rules.
// If there are rules to be updated, return true to indicate the resource should be updated.
func (r *ruleResource) updateAlertingRules(rules ...*RuleWithGroup) (bool, error) {
	var (
		commit     bool
		spec       = r.Spec.DeepCopy()
		ruleMap    = make(map[string]*RuleWithGroup)
		ruleGroups = make(map[string]map[string]struct{}) // mapping of name to group
	)

	if spec == nil {
		return false, nil
	}
	spec.Groups = nil

	for i, rule := range rules {
		if rule != nil {
			ruleMap[rule.Alert] = rules[i]
			ruleGroups[rule.Alert] = make(map[string]struct{})
		}
	}

	// Firstly delete the old rules
	for _, g := range r.Spec.Groups {
		var rules []promresourcesv1.Rule
		for _, r := range g.Rules {
			if r.Alert != "" {
				if _, ok := ruleMap[r.Alert]; ok {
					ruleGroups[r.Alert][g.Name] = struct{}{}
					commit = true
					continue
				}
			}
			rules = append(rules, r)
		}
		if len(rules) > 0 {
			spec.Groups = append(spec.Groups, promresourcesv1.RuleGroup{
				Name:                    g.Name,
				Interval:                g.Interval,
				PartialResponseStrategy: g.PartialResponseStrategy,
				Rules:                   rules,
			})
		}
	}

	addRules := func(g *promresourcesv1.RuleGroup) bool {
		var add bool
		var num = customRuleGroupSize - len(g.Rules)
		if num > 0 {
			for name, rule := range ruleMap {
				if num <= 0 {
					break
				}
				if gNames, ok := ruleGroups[name]; ok {
					// Add a rule to a different group than the group where it resided, to clear its alerts, etc.
					// Because Prometheus may migrate information such as alerts from the old rule into the new rule
					// when updating a rule within a group.
					if _, ok := gNames[g.Name]; !ok {
						g.Rules = append(g.Rules, rule.Rule)
						num--
						delete(ruleMap, name)
						add = true
					}
				}
			}
		}
		return add
	}

	// Then add the new rules
	var groupMax = -1
	for i, g := range spec.Groups {
		if len(ruleMap) == 0 {
			break
		}

		if strings.HasPrefix(g.Name, customRuleGroupDefaultPrefix) {
			suf, err := strconv.Atoi(strings.TrimPrefix(g.Name, customRuleGroupDefaultPrefix))
			if err != nil {
				continue
			}
			if suf > groupMax {
				groupMax = suf
			}
		}

		if addRules(&spec.Groups[i]) {
			commit = true
		}
	}

	for groupMax++; len(ruleMap) > 0; groupMax++ {
		g := promresourcesv1.RuleGroup{Name: fmt.Sprintf("%s%d", customRuleGroupDefaultPrefix, groupMax)}

		if addRules(&g) {
			spec.Groups = append(spec.Groups, g)
			commit = true
		}
	}

	if commit {
		content, err := yaml.Marshal(spec)
		if err != nil {
			return false, errors.Wrap(err, "failed to unmarshal content")
		}
		if len(string(content)) > maxConfigMapDataSize { // check size limit
			return false, errOutOfConfigMapSize
		}
		r.Spec = *spec
	}
	return commit, nil
}

// addAlertingRules adds the rules.
// If there are rules to be added, return true to indicate the resource should be updated.
func (r *ruleResource) addAlertingRules(rules ...*RuleWithGroup) (bool, error) {
	var (
		commit   bool
		spec     = r.Spec.DeepCopy()
		groupMax = -1

		cursor int // indicates which rule to start adding for the rules with no groups

		unGroupedRules []promresourcesv1.Rule                    // rules that do not specify group names
		groupedRules   = make(map[string][]promresourcesv1.Rule) // rules that have specific group names
	)

	for i, rule := range rules {
		if len(strings.TrimSpace(rule.Group)) == 0 {
			unGroupedRules = append(unGroupedRules, rules[i].Rule)
		} else {
			groupedRules[rule.Group] = append(groupedRules[rule.Group], rules[i].Rule)
		}
	}

	if spec == nil {
		spec = new(promresourcesv1.PrometheusRuleSpec)
	}

	// For the rules that have specific group names, add them to the matched groups.
	// For the rules that do not specify group names, add them to the automatically generated groups until the limit is reached.
	for i, g := range spec.Groups {
		var (
			gName                 = g.Name
			unGroupedRulesDrained = cursor >= len(unGroupedRules) // whether all rules without groups have been added
			groupedRulesDrained   = len(groupedRules) == 0        // whether all rules with groups have been added
		)

		if unGroupedRulesDrained && groupedRulesDrained {
			break
		}

		if !groupedRulesDrained {
			if _, ok := groupedRules[gName]; ok {
				spec.Groups[i].Rules = append(spec.Groups[i].Rules, groupedRules[gName]...)
				delete(groupedRules, gName)
				commit = true
			}
		}

		g = spec.Groups[i]
		if !unGroupedRulesDrained && strings.HasPrefix(gName, customRuleGroupDefaultPrefix) {
			suf, err := strconv.Atoi(strings.TrimPrefix(gName, customRuleGroupDefaultPrefix))
			if err != nil {
				continue
			}
			if suf > groupMax {
				groupMax = suf
			}

			if size := len(g.Rules); size < customRuleGroupSize {
				num := customRuleGroupSize - size
				var stop int
				if stop = cursor + num; stop > len(unGroupedRules) {
					stop = len(unGroupedRules)
				}
				spec.Groups[i].Rules = append(spec.Groups[i].Rules, unGroupedRules[cursor:stop]...)
				cursor = stop
				commit = true
			}
		}
	}

	// If no groups are available, new groups will be created to place the remaining rules.
	for gName := range groupedRules {
		rules := groupedRules[gName]
		if len(rules) == 0 {
			continue
		}
		spec.Groups = append(spec.Groups, promresourcesv1.RuleGroup{Name: gName, Rules: rules})
		commit = true
	}
	for groupMax++; cursor < len(rules); groupMax++ {
		g := promresourcesv1.RuleGroup{Name: fmt.Sprintf("%s%d", customRuleGroupDefaultPrefix, groupMax)}
		var stop int
		if stop = cursor + customRuleGroupSize; stop > len(unGroupedRules) {
			stop = len(unGroupedRules)
		}
		g.Rules = append(g.Rules, unGroupedRules[cursor:stop]...)
		spec.Groups = append(spec.Groups, g)
		cursor = stop
		commit = true
	}

	if commit {
		content, err := yaml.Marshal(spec)
		if err != nil {
			return false, errors.Wrap(err, "failed to unmarshal content")
		}
		if len(string(content)) > maxConfigMapDataSize { // check size limit
			return false, errOutOfConfigMapSize
		}
		r.Spec = *spec
	}
	return commit, nil
}

func (r *ruleResource) commit(ctx context.Context, prometheusResourceClient promresourcesclient.Interface) error {
	var pr = (promresourcesv1.PrometheusRule)(*r)
	if len(pr.Spec.Groups) == 0 {
		return prometheusResourceClient.MonitoringV1().PrometheusRules(r.Namespace).Delete(ctx, r.Name, metav1.DeleteOptions{})
	}
	npr, err := prometheusResourceClient.MonitoringV1().PrometheusRules(r.Namespace).Update(ctx, &pr, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	npr.DeepCopyInto(&pr)
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

func (r *PrometheusRuler) AddAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, ruleResourceLabels map[string]string,
	rules ...*RuleWithGroup) ([]*v2alpha1.BulkItemResponse, error) {
	return nil, errors.New("Adding Prometheus rules not supported")
}

func (r *PrometheusRuler) UpdateAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, ruleResourceLabels map[string]string,
	ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error) {
	return nil, errors.New("Updating Prometheus rules not supported")
}

func (r *PrometheusRuler) DeleteAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error) {
	return nil, errors.New("Deleting Prometheus rules not supported.")
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

func (r *ThanosRuler) AddAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, ruleResourceLabels map[string]string,
	rules ...*RuleWithGroup) ([]*v2alpha1.BulkItemResponse, error) {

	return r.addAlertingRules(ctx, ruleNamespace, extraRuleResourceSelector, nil, ruleResourceLabels, rules...)
}

func (r *ThanosRuler) addAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, excludePrometheusRules map[string]struct{},
	ruleResourceLabels map[string]string, rules ...*RuleWithGroup) ([]*v2alpha1.BulkItemResponse, error) {

	prometheusRules, err := r.ListRuleResources(ruleNamespace, extraRuleResourceSelector)
	if err != nil {
		return nil, err
	}
	// sort by the left space to speed up the hit rate
	sort.Slice(prometheusRules, func(i, j int) bool {
		return len(fmt.Sprint(prometheusRules[i])) <= len(fmt.Sprint(prometheusRules[j]))
	})

	var (
		respItems = make([]*v2alpha1.BulkItemResponse, 0, len(rules))
		cursor    int
	)

	resp := func(rule *RuleWithGroup, err error) *v2alpha1.BulkItemResponse {
		if err != nil {
			return v2alpha1.NewBulkItemErrorServerResponse(rule.Alert, err)
		}
		return v2alpha1.NewBulkItemSuccessResponse(rule.Alert, v2alpha1.ResultCreated)
	}

	for _, pr := range prometheusRules {
		if cursor >= len(rules) {
			break
		}
		if len(excludePrometheusRules) > 0 {
			if _, ok := excludePrometheusRules[pr.Name]; ok {
				continue
			}
		}

		var (
			err  error
			num  = len(rules) - cursor
			stop = len(rules)
			rs   []*RuleWithGroup
		)

		// First add all the rules to this resource,
		// and if the limit is exceeded, add half
		for i := 1; i <= 2; i++ {
			stop = cursor + num/i
			rs = rules[cursor:stop]

			err = r.doRuleResourceOperation(ctx, pr.Namespace, pr.Name, func(pr *promresourcesv1.PrometheusRule) error {
				resource := ruleResource(*pr)
				if ok, err := resource.addAlertingRules(rs...); err != nil {
					return err
				} else if ok {
					if err = resource.commit(ctx, r.client); err != nil {
						return err
					}
				}
				return nil
			})
			if err == errOutOfConfigMapSize && num > 1 {
				continue
			}
			break
		}

		switch {
		case err == errOutOfConfigMapSize:
			break
		case resourceNotFound(err):
			continue
		default:
			for _, rule := range rs {
				respItems = append(respItems, resp(rule, err))
			}
			cursor = stop
		}
	}

	// create new rule resources and add rest rules into them
	// when all existing rule resources are full.
	for cursor < len(rules) {
		var (
			err  error
			num  = len(rules) - cursor
			stop = len(rules)
			rs   []*RuleWithGroup
		)
		// If adding the rules to the new resource exceeds the limit,
		// reduce the amount to 1/2, 1/3... of rest rules until the new resource can accommodate.
		for i := 1; ; i++ {
			stop = cursor + num/i
			rs = rules[cursor:stop]

			pr := &promresourcesv1.PrometheusRule{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    ruleNamespace.Name,
					GenerateName: customAlertingRuleResourcePrefix,
					Labels:       ruleResourceLabels,
				},
			}
			resource := ruleResource(*pr)
			var ok bool
			ok, err = resource.addAlertingRules(rs...)
			if err == errOutOfConfigMapSize {
				continue
			}
			if ok {
				pr.Spec = resource.Spec
				_, err = r.client.MonitoringV1().PrometheusRules(ruleNamespace.Name).Create(ctx, pr, metav1.CreateOptions{})
			}
			break
		}

		for _, rule := range rs {
			respItems = append(respItems, resp(rule, err))
		}
		cursor = stop
	}

	return respItems, nil
}

func (r *ThanosRuler) UpdateAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, ruleResourceLabels map[string]string,
	ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error) {

	var (
		itemsMap  = make(map[string][]*ResourceRuleItem)
		respItems = make([]*v2alpha1.BulkItemResponse, 0, len(ruleItems))
		// rules updated successfully. The key is the rule name.
		rulesUpdated = make(map[string]struct{})
		// rules to be moved to other resources. The key is the resource name in which the rules reside.
		rulesToMove = make(map[string][]*ResourceRuleItem)
		// duplicate rules to be deleted
		rulesToDelete = make(map[string][]*ResourceRuleItem)
	)

	for i, item := range ruleItems {
		itemsMap[item.ResourceName] = append(itemsMap[item.ResourceName], ruleItems[i])
	}

	// Update the rules in the resources where the rules reside.
	// If duplicate rules are found, the first will be updated and the others will be deleted.
	// if updating the rules in the original resources causes exceeding size limit,
	// they will be moved to other resources and then be updated.
	for name, items := range itemsMap {
		var (
			nrules []*RuleWithGroup
			nitems []*ResourceRuleItem
		)

		for i := range items {
			item := items[i]
			if _, ok := rulesUpdated[item.Alert]; ok {
				rulesToDelete[name] = append(rulesToDelete[name], item)
				continue
			}
			nrules = append(nrules, &item.RuleWithGroup)
			nitems = append(nitems, item)
		}
		if len(nrules) == 0 {
			continue
		}

		err := r.doRuleResourceOperation(ctx, ruleNamespace.Name, name, func(pr *promresourcesv1.PrometheusRule) error {
			resource := ruleResource(*pr)
			if ok, err := resource.updateAlertingRules(nrules...); err != nil {
				return err
			} else if ok {
				if err = resource.commit(ctx, r.client); err != nil {
					return err
				}
			}
			return nil
		})

		switch {
		case err == nil:
			for _, item := range items {
				rulesUpdated[item.Alert] = struct{}{}
				respItems = append(respItems, v2alpha1.NewBulkItemSuccessResponse(item.Alert, v2alpha1.ResultUpdated))
			}
		case err == errOutOfConfigMapSize: // Cannot update the rules in the original resource
			rulesToMove[name] = append(rulesToMove[name], nitems...)
		case resourceNotFound(err):
			for _, item := range items {
				respItems = append(respItems, &v2alpha1.BulkItemResponse{
					RuleName:  item.Alert,
					Status:    v2alpha1.StatusError,
					ErrorType: v2alpha1.ErrNotFound,
				})
			}
		default:
			for _, item := range items {
				respItems = append(respItems, v2alpha1.NewBulkItemErrorServerResponse(item.Alert, err))
			}
		}
	}

	// The move here is not really move, because the move also requires an update.
	// What really happens is that the new rules will be added in other resources first,
	// and then the old rules will be deleted from the original resources.
	for name, items := range rulesToMove {
		var (
			nrules = make([]*RuleWithGroup, 0, len(items))
			nitems = make(map[string]*ResourceRuleItem, len(items))
		)
		for i := range items {
			item := items[i]
			nrules = append(nrules, &item.RuleWithGroup)
			nitems[item.Alert] = item
		}
		if len(nrules) == 0 {
			continue
		}

		aRespItems, err := r.addAlertingRules(ctx, ruleNamespace, extraRuleResourceSelector,
			map[string]struct{}{name: {}}, ruleResourceLabels, nrules...)
		if err != nil {
			for _, item := range items {
				respItems = append(respItems, v2alpha1.NewBulkItemErrorServerResponse(item.Alert, err))
			}
			continue
		}

		for i := range aRespItems {
			resp := aRespItems[i]
			switch resp.Status {
			case v2alpha1.StatusSuccess:
				if item, ok := nitems[resp.RuleName]; ok {
					rulesToDelete[name] = append(rulesToDelete[name], item)
				}
			default:
				respItems = append(respItems, resp)
			}
		}
	}

	for _, items := range rulesToDelete {
		dRespItems, err := r.DeleteAlertingRules(ctx, ruleNamespace, items...)
		if err != nil {
			for _, item := range items {
				respItems = append(respItems, v2alpha1.NewBulkItemErrorServerResponse(item.Alert, err))
			}
			continue
		}
		for i := range dRespItems {
			resp := dRespItems[i]
			if resp.Status == v2alpha1.StatusSuccess {
				// The delete operation here is for updating, so update the result to v2alpha1.ResultUpdated
				resp.Result = v2alpha1.ResultUpdated
			}
			respItems = append(respItems, resp)
		}
	}

	return respItems, nil
}

func (r *ThanosRuler) DeleteAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error) {

	var (
		itemsMap  = make(map[string][]*ResourceRuleItem)
		respItems = make([]*v2alpha1.BulkItemResponse, 0, len(ruleItems))
	)

	for i, ruleItem := range ruleItems {
		itemsMap[ruleItem.ResourceName] = append(itemsMap[ruleItem.ResourceName], ruleItems[i])
	}

	resp := func(item *ResourceRuleItem, err error) *v2alpha1.BulkItemResponse {
		if err != nil {
			return v2alpha1.NewBulkItemErrorServerResponse(item.Alert, err)
		}
		return v2alpha1.NewBulkItemSuccessResponse(item.Alert, v2alpha1.ResultDeleted)
	}

	for name, items := range itemsMap {
		var rules []*RuleWithGroup
		for i := range items {
			rules = append(rules, &items[i].RuleWithGroup)
		}

		err := r.doRuleResourceOperation(ctx, ruleNamespace.Name, name, func(pr *promresourcesv1.PrometheusRule) error {
			resource := ruleResource(*pr)
			if ok, err := resource.deleteAlertingRules(rules...); err != nil {
				return err
			} else if ok {
				if err = resource.commit(ctx, r.client); err != nil {
					return err
				}
			}
			return nil
		})
		for _, item := range items {
			respItems = append(respItems, resp(item, err))
		}
	}

	return respItems, nil
}

func (r *ThanosRuler) doRuleResourceOperation(ctx context.Context, namespace, name string,
	operation func(pr *promresourcesv1.PrometheusRule) error) error {
	// Lock here is used to lock specific resource in order to prevent frequent conflicts
	key := namespace + "/" + name
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ruleResourceLocker.Lock(key)
		defer ruleResourceLocker.Unlock(key)
		pr, err := r.client.MonitoringV1().PrometheusRules(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return operation(pr)
	})
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

func resourceNotFound(err error) bool {
	switch e := err.(type) {
	case *apierrors.StatusError:
		if e.Status().Code == http.StatusNotFound {
			return true
		}
	}
	return false
}
