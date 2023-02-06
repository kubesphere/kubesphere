package v1beta1

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oliveagle/jsonpath"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

type resourceCache struct {
	cache cache.Cache
}

func NewResourceCache(cache cache.Cache) Interface {
	return &resourceCache{cache: cache}
}

func (u *resourceCache) Get(name, namespace string, object client.Object) error {
	return u.cache.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, object)
}

func (u *resourceCache) List(namespace string, query *query.Query, list client.ObjectList) error {
	listOpt := &client.ListOptions{
		LabelSelector: query.Selector(),
		Namespace:     namespace,
	}
	err := u.cache.List(context.Background(), list, listOpt)
	if err != nil {
		return err
	}

	extractList, err := meta.ExtractList(list)
	if err != nil {
		return err
	}
	DefaultList(extractList, query, u.compare, u.filter)
	return nil
}

func (u *resourceCache) compare(left, right runtime.Object, field query.Field) bool {
	l, err := meta.Accessor(left)
	if err != nil {
		return false
	}
	r, err := meta.Accessor(right)
	if err != nil {
		return false
	}
	return DefaultObjectMetaCompare(l, r, field)
}

func (u *resourceCache) filter(object runtime.Object, filter query.Filter) bool {
	o, err := meta.Accessor(object)
	if err != nil {
		return false
	}
	return DefaultObjectMetaFilter(o, filter)
}

// implement a generic query filter to support multiple field selectors with "jsonpath.JsonPathLookup"
// https://github.com/oliveagle/jsonpath/blob/master/readme.md
func contains(object runtime.Object, queryValue query.Value) bool {
	// call the ParseSelector function of "k8s.io/apimachinery/pkg/fields/selector.go" to validate and parse the selector
	fieldSelector, err := fields.ParseSelector(string(queryValue))
	if err != nil {
		klog.V(4).Infof("failed parse selector error: %s", err)
		return false
	}
	for _, requirement := range fieldSelector.Requirements() {
		var negative bool
		// supports '=', '==' and '!='.(e.g. ?fieldSelector=key1=value1,key2=value2)
		// fields.ParseSelector(FieldSelector) has handled the case where the operator is '==' and converted it to '=',
		// so case selection.DoubleEquals can be ignored here.
		switch requirement.Operator {
		case selection.NotEquals:
			negative = true
		case selection.Equals:
			negative = false
		}
		key := requirement.Field
		value := requirement.Value

		var input map[string]interface{}
		data, err := json.Marshal(object)
		if err != nil {
			klog.V(4).Infof("failed marshal to JSON string: %s", err)
			return false
		}
		if err = json.Unmarshal(data, &input); err != nil {
			klog.V(4).Infof("failed unmarshal to map object: %s", err)
			return false
		}
		rawValue, err := jsonpath.JsonPathLookup(input, "$."+key)
		if err != nil {
			klog.V(4).Infof("failed to lookup jsonpath: %s", err)
			return false
		}
		if (negative && fmt.Sprintf("%v", rawValue) != value) || (!negative && fmt.Sprintf("%v", rawValue) == value) {
			continue
		} else {
			return false
		}
	}
	return true
}
