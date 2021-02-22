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

// updateAlertingRule updates the rules.
// If there are rules to be updated, return true to indicate the resource should be updated.
func (r *ruleResource) updateAlertingRules(rules ...*RuleWithGroup) (bool, error) {
	var (
		commit  bool
		spec    = r.Spec.DeepCopy()
		ruleMap = make(map[string]*RuleWithGroup)
	)

	if spec == nil {
		return false, nil
	}

	for i, rule := range rules {
		if rule != nil {
			ruleMap[rule.Alert] = rules[i]
		}
	}

	for i, g := range spec.Groups {
		for j, r := range g.Rules {
			if r.Alert == "" {
				continue
			}
			if b, ok := ruleMap[r.Alert]; ok {
				if b == nil {
					spec.Groups[i].Rules = append(g.Rules[:j], g.Rules[j+1:]...)
				} else {
					spec.Groups[i].Rules[j] = b.Rule
					ruleMap[r.Alert] = nil // clear to mark it updated
				}
				commit = true
			}
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

func (r *ruleResource) addAlertingRules(rules ...*RuleWithGroup) (bool, error) {
	var (
		commit   bool
		spec     = r.Spec.DeepCopy()
		groupMax = -1
		cursor   int

		rulesNoGroup   []promresourcesv1.Rule
		rulesWithGroup = make(map[string][]promresourcesv1.Rule)
	)

	for i, rule := range rules {
		if len(strings.TrimSpace(rule.Group)) == 0 {
			rulesNoGroup = append(rulesNoGroup, rules[i].Rule)
		} else {
			rulesWithGroup[rule.Group] = append(rulesWithGroup[rule.Group], rules[i].Rule)
		}
	}

	if spec == nil {
		spec = new(promresourcesv1.PrometheusRuleSpec)
	}

	for i, g := range spec.Groups {
		var (
			gName         = g.Name
			doneNoGroup   = cursor >= len(rulesNoGroup)
			doneWithGroup = len(rulesWithGroup) == 0
		)

		if doneNoGroup && doneWithGroup {
			break
		}

		if !doneWithGroup {
			if _, ok := rulesWithGroup[gName]; ok {
				spec.Groups[i].Rules = append(spec.Groups[i].Rules, rulesWithGroup[gName]...)
				delete(rulesWithGroup, gName)
				commit = true
			}
		}

		g = spec.Groups[i]
		if !doneNoGroup && strings.HasPrefix(gName, customRuleGroupDefaultPrefix) {
			suf, err := strconv.Atoi(strings.TrimPrefix(gName, customRuleGroupDefaultPrefix))
			if err != nil {
				continue
			}
			if suf > groupMax {
				groupMax = suf
			}

			if size := len(g.Rules); size < customRuleGroupSize {
				num := customRuleGroupSize - size
				var limit int
				if limit = cursor + num; limit > len(rulesNoGroup) {
					limit = len(rulesNoGroup)
				}
				spec.Groups[i].Rules = append(spec.Groups[i].Rules, rulesNoGroup[cursor:limit]...)
				cursor = limit
				commit = true
			}
		}
	}

	for groupMax++; cursor < len(rules); groupMax++ {
		g := promresourcesv1.RuleGroup{Name: fmt.Sprintf("%s%d", customRuleGroupDefaultPrefix, groupMax)}
		var limit int
		if limit = cursor + customRuleGroupSize; limit > len(rulesNoGroup) {
			limit = len(rulesNoGroup)
		}
		g.Rules = append(g.Rules, rulesNoGroup[cursor:limit]...)
		spec.Groups = append(spec.Groups, g)
		cursor = limit
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
	return nil, errors.New("not supported to add rules for prometheus")
}

func (r *PrometheusRuler) UpdateAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, ruleResourceLabels map[string]string,
	ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error) {
	return nil, errors.New("not supported to update rules for prometheus")
}

func (r *PrometheusRuler) DeleteAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error) {
	return nil, errors.New("not supported to delete rules for prometheus")
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
			err   error
			num   = len(rules) - cursor
			limit = len(rules)
			rs    []*RuleWithGroup
		)

		for i := 1; i <= 2; i++ {
			limit = cursor + num/i
			rs = rules[cursor:limit]

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
			cursor = limit
		}
	}

	// create new rule resources and add rest rules into them when all existing rule resources are full.
	for cursor < len(rules) {
		var (
			err   error
			num   = len(rules) - cursor
			limit = len(rules)
			rs    []*RuleWithGroup
		)

		for i := 1; ; i++ {
			limit = cursor + num/i
			rs = rules[cursor:limit]

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
		cursor = limit
	}

	return respItems, nil
}

func (r *ThanosRuler) UpdateAlertingRules(ctx context.Context, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, ruleResourceLabels map[string]string,
	ruleItems ...*ResourceRuleItem) ([]*v2alpha1.BulkItemResponse, error) {

	var (
		itemsMap  = make(map[string][]*ResourceRuleItem)
		respItems = make([]*v2alpha1.BulkItemResponse, 0, len(ruleItems))
		successes = make(map[string]struct{})
		moveMap   = make(map[string][]*ResourceRuleItem)
		delMap    = make(map[string][]*ResourceRuleItem) // duplicate rules that need to deleted in the same resource

	)

	for i, item := range ruleItems {
		itemsMap[item.ResourceName] = append(itemsMap[item.ResourceName], ruleItems[i])
	}

	for name, items := range itemsMap {
		var (
			nrules []*RuleWithGroup
			nitems []*ResourceRuleItem
		)

		for i := range items {
			item := items[i]
			if _, ok := successes[item.Alert]; ok {
				delMap[name] = append(delMap[name], item)
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
				successes[item.Alert] = struct{}{}
				respItems = append(respItems, v2alpha1.NewBulkItemSuccessResponse(item.Alert, v2alpha1.ResultUpdated))
			}
		case err == errOutOfConfigMapSize:
			moveMap[name] = append(moveMap[name], nitems...)
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

	for name, items := range moveMap {
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
					delMap[name] = append(delMap[name], item)
				}
			default:
				respItems = append(respItems, resp)
			}
		}
	}

	for _, items := range delMap {
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
