package rules

import (
	"sort"
	"sync"

	promresourcesv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prominformersv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"kubesphere.io/kubesphere/pkg/api/alerting/v2alpha1"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

// RuleCache caches all rules from the prometheusrule custom resources
type RuleCache struct {
	lock       sync.RWMutex
	namespaces map[string]*namespaceRuleCache
}

func NewRuleCache(ruleResourceInformer prominformersv1.PrometheusRuleInformer) *RuleCache {
	rc := RuleCache{
		namespaces: make(map[string]*namespaceRuleCache),
	}

	ruleResourceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: rc.addCache,
		UpdateFunc: func(oldObj, newObj interface{}) {
			rc.addCache(newObj)
		},
		DeleteFunc: rc.deleteCache,
	})
	return &rc
}

func (c *RuleCache) addCache(referObj interface{}) {
	pr, ok := referObj.(*promresourcesv1.PrometheusRule)
	if !ok {
		return
	}
	cr := parseRuleResource(pr)

	c.lock.Lock()
	defer c.lock.Unlock()

	cn, ok := c.namespaces[pr.Namespace]
	if !ok || cn == nil {
		cn = &namespaceRuleCache{
			namespace: pr.Namespace,
			resources: make(map[string]*resourceRuleCache),
		}
		c.namespaces[pr.Namespace] = cn
	}
	cn.resources[pr.Name] = cr
}

func (c *RuleCache) deleteCache(referObj interface{}) {
	pr, ok := referObj.(*promresourcesv1.PrometheusRule)
	if !ok {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	cn, ok := c.namespaces[pr.Namespace]
	if !ok {
		return
	}
	delete(cn.resources, pr.Name)
	if len(cn.resources) == 0 {
		delete(c.namespaces, pr.Namespace)
	}
}

func (c *RuleCache) getResourceRuleCaches(ruler Ruler, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector) (map[string]*resourceRuleCache, error) {

	selected, err := ruleNamespaceSelected(ruler, ruleNamespace)
	if err != nil {
		return nil, err
	}
	if !selected {
		return nil, nil
	}
	rSelector, err := ruler.RuleResourceSelector(extraRuleResourceSelector)
	if err != nil {
		return nil, err
	}
	var m = make(map[string]*resourceRuleCache)

	c.lock.RLock()
	defer c.lock.RUnlock()

	cn, ok := c.namespaces[ruleNamespace.Name]
	if ok && cn != nil {
		for _, cr := range cn.resources {
			if rSelector.Matches(labels.Set(cr.Labels)) {
				m[cr.Name] = cr
			}
		}
	}
	return m, nil
}

func (c *RuleCache) GetRule(ruler Ruler, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, idOrName string) (*ResourceRuleItem, error) {

	rules, err := c.GetRuleByIdOrName(ruler, ruleNamespace, extraRuleResourceSelector, idOrName)
	if err != nil {
		return nil, err
	}
	if l := len(rules); l == 0 {
		return nil, nil
	} else if l > 1 {
		// guarantees the stability of the get operations.
		sort.Slice(rules, func(i, j int) bool {
			return v2alpha1.AlertingRuleIdCompare(rules[i].Id, rules[j].Id)
		})
	}
	return rules[0], nil
}

func (c *RuleCache) GetRuleByIdOrName(ruler Ruler, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector, idOrName string) ([]*ResourceRuleItem, error) {

	caches, err := c.getResourceRuleCaches(ruler, ruleNamespace, extraRuleResourceSelector)
	if err != nil {
		return nil, err
	}
	if len(caches) == 0 {
		return nil, nil
	}

	var rules []*ResourceRuleItem
	switch ruler.(type) {
	case *PrometheusRuler:
		for rn, rc := range caches {
			if rule, ok := rc.IdRules[idOrName]; ok {
				rules = append(rules, &ResourceRuleItem{
					RuleWithGroup: RuleWithGroup{
						Group: rule.Group,
						Id:    rule.Id,
						Rule:  *rule.Rule.DeepCopy(),
					},
					ResourceName: rn,
				})
			}
		}
	case *ThanosRuler:
		for rn, rc := range caches {
			if nrules, ok := rc.NameRules[idOrName]; ok {
				for _, nrule := range nrules {
					rules = append(rules, &ResourceRuleItem{
						RuleWithGroup: RuleWithGroup{
							Group: nrule.Group,
							Id:    nrule.Id,
							Rule:  *nrule.Rule.DeepCopy(),
						},
						ResourceName: rn,
					})
				}
			}
		}
	default:
		return nil, errors.New("unsupported ruler type")
	}

	return rules, err
}

func (c *RuleCache) ListRules(ruler Ruler, ruleNamespace *corev1.Namespace,
	extraRuleResourceSelector labels.Selector) (map[string]*ResourceRuleCollection, error) {

	caches, err := c.getResourceRuleCaches(ruler, ruleNamespace, extraRuleResourceSelector)
	if err != nil {
		return nil, err
	}
	if len(caches) == 0 {
		return nil, nil
	}

	ret := make(map[string]*ResourceRuleCollection)
	for rn, rc := range caches {
		rrs := &ResourceRuleCollection{
			GroupSet:  make(map[string]struct{}),
			IdRules:   make(map[string]*ResourceRuleItem),
			NameRules: make(map[string][]*ResourceRuleItem),
		}
		for name, rules := range rc.NameRules {
			for _, rule := range rules {
				rrs.GroupSet[rule.Group] = struct{}{}
				rr := &ResourceRuleItem{
					RuleWithGroup: RuleWithGroup{
						Group: rule.Group,
						Id:    rule.Id,
						Rule:  *rule.Rule.DeepCopy(),
					},
					ResourceName: rn,
				}
				rrs.IdRules[rr.Id] = rr
				rrs.NameRules[name] = append(rrs.NameRules[name], rr)
			}
		}
		if len(rrs.IdRules) > 0 {
			ret[rn] = rrs
		}
	}

	return ret, nil
}

type namespaceRuleCache struct {
	namespace string
	resources map[string]*resourceRuleCache
}

type resourceRuleCache struct {
	Name      string
	Labels    map[string]string
	GroupSet  map[string]struct{}
	IdRules   map[string]*cacheRule
	NameRules map[string][]*cacheRule
}

type cacheRule struct {
	Group string
	Id    string
	Rule  *promresourcesv1.Rule
}

func parseRuleResource(pr *promresourcesv1.PrometheusRule) *resourceRuleCache {
	var (
		groupSet  = make(map[string]struct{})
		idRules   = make(map[string]*cacheRule)
		nameRules = make(map[string][]*cacheRule)
	)
	for i := 0; i < len(pr.Spec.Groups); i++ {
		g := pr.Spec.Groups[i]
		for j := 0; j < len(g.Rules); j++ {
			gr := g.Rules[j]
			if gr.Alert == "" {
				continue
			}
			groupSet[g.Name] = struct{}{}
			cr := &cacheRule{
				Group: g.Name,
				Id:    GenResourceRuleIdIgnoreFormat(g.Name, &gr),
				Rule:  &gr,
			}
			nameRules[cr.Rule.Alert] = append(nameRules[cr.Rule.Alert], cr)
			idRules[cr.Id] = cr
		}
	}
	return &resourceRuleCache{
		Name:      pr.Name,
		Labels:    pr.Labels,
		GroupSet:  groupSet,
		IdRules:   idRules,
		NameRules: nameRules,
	}
}
