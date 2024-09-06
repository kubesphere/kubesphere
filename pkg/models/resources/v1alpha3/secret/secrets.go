/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package secret

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"

	"github.com/oliveagle/jsonpath"
	v1 "k8s.io/api/core/v1"
)

type secretSearcher struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &secretSearcher{cache: cache}
}

func (s *secretSearcher) Get(namespace, name string) (runtime.Object, error) {
	secret := &v1.Secret{}
	return secret, s.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, secret)
}

func (s *secretSearcher) List(namespace string, query *query.Query) (*api.ListResult, error) {
	secrets := &v1.SecretList{}
	if err := s.cache.List(context.Background(), secrets, client.InNamespace(namespace), client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range secrets.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, query, s.compare, s.filter), nil
}

func (s *secretSearcher) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftSecret, ok := left.(*v1.Secret)
	if !ok {
		return false
	}

	rightSecret, ok := right.(*v1.Secret)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftSecret.ObjectMeta, rightSecret.ObjectMeta, field)
}

func (s *secretSearcher) filter(object runtime.Object, filter query.Filter) bool {
	secret, ok := object.(*v1.Secret)
	if !ok {
		return false
	}

	if filter.Field == query.ParameterFieldSelector {
		return contains(secret, filter.Value)
	}

	return v1alpha3.DefaultObjectMetaFilter(secret.ObjectMeta, filter)
}

// implement a generic query filter to support multiple field selectors with "jsonpath.JsonPathLookup"
// https://github.com/oliveagle/jsonpath/blob/master/readme.md
func contains(secret *v1.Secret, queryValue query.Value) bool {
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
		data, err := json.Marshal(secret)
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
