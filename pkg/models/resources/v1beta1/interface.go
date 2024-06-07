/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1beta1

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/oliveagle/jsonpath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/constants"
)

type ResourceManager interface {
	IsServed(schema.GroupVersionResource) (bool, error)
	CreateObjectFromRawData(gvr schema.GroupVersionResource, rawData []byte) (client.Object, error)

	CreateResource(ctx context.Context, object client.Object) error
	UpdateResource(ctx context.Context, object client.Object) error
	PatchResource(ctx context.Context, object client.Object) error
	DeleteResource(ctx context.Context, object client.Object) error
	GetResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, name string) (client.Object, error)
	ListResources(ctx context.Context, gvr schema.GroupVersionResource, namespace string, query *query.Query) (client.ObjectList, error)

	Get(ctx context.Context, namespace, name string, object client.Object) error
	List(ctx context.Context, namespace string, query *query.Query, object client.ObjectList) error
	Create(ctx context.Context, object client.Object) error
	Delete(ctx context.Context, object client.Object) error
	Update(ctx context.Context, object client.Object) error
	Patch(ctx context.Context, old, new client.Object) error
}

type CompareFunc func(runtime.Object, runtime.Object, query.Field) bool

type FilterFunc func(runtime.Object, query.Filter) bool

type TransformFunc func(runtime.Object) runtime.Object

func DefaultList(objects []runtime.Object, q *query.Query, compareFunc CompareFunc, filterFunc FilterFunc, transformFuncs ...TransformFunc) ([]runtime.Object, int, int) {
	// selected matched ones
	var filtered []runtime.Object
	if len(q.Filters) != 0 {
		for _, object := range objects {
			match := true
			for field, value := range q.Filters {
				if match = filterFunc(object, query.Filter{Field: field, Value: value}); !match {
					break
				}
			}
			if match {
				for _, transform := range transformFuncs {
					object = transform(object)
				}
				filtered = append(filtered, object)
			}
		}
	} else {
		filtered = objects
	}

	// sort by sortBy field
	sort.Slice(filtered, func(i, j int) bool {
		if !q.Ascending {
			return compareFunc(filtered[i], filtered[j], q.SortBy)
		}
		return !compareFunc(filtered[i], filtered[j], q.SortBy)
	})

	total := len(filtered)

	if q.Pagination == nil {
		q.Pagination = query.NoPagination
	}

	start, end := q.Pagination.GetValidPagination(total)
	remainingItemCount := total - end
	totalCount := total

	return filtered[start:end], remainingItemCount, totalCount
}

func DefaultObjectMetaCompare(left, right metav1.Object, sortBy query.Field) bool {
	switch sortBy {
	// ?sortBy=name
	case query.FieldName:
		return strings.Compare(left.GetName(), right.GetName()) > 0
	//	?sortBy=creationTimestamp
	default:
		fallthrough
	case query.FieldCreateTime:
		fallthrough
	case query.FieldCreationTimeStamp:
		// compare by name if creation timestamp is equal
		lTime := left.GetCreationTimestamp()
		rTime := right.GetCreationTimestamp()
		if lTime.Equal(&rTime) {
			return strings.Compare(left.GetName(), right.GetName()) > 0
		}
		return left.GetCreationTimestamp().After(right.GetCreationTimestamp().Time)
	}
}

// DefaultObjectMetaFilter filters the metadata of Kubernetes objects based on the given filter conditions.
// Supported filter fields include: FieldNames, FieldName, FieldUID, FieldNamespace,
// FieldOwnerReference, FieldOwnerKind, FieldAnnotation, FieldLabel, and ParameterFieldSelector.
// Returns true if the object satisfies the filter conditions; otherwise, returns false.
//
// Parameters:
//   - item: Metadata of the Kubernetes object to be filtered.
//   - filter: Query object containing filter conditions.
//
// Returns:
//   - bool: True if the object satisfies the filter conditions; false otherwise.
func DefaultObjectMetaFilter(item metav1.Object, filter query.Filter) bool {
	switch filter.Field {
	case query.FieldNames:
		// Check if the object's name matches any name in the filter.
		for _, name := range strings.Split(string(filter.Value), ",") {
			if item.GetName() == name {
				return true
			}
		}
		return false
	// /namespaces?page=1&limit=10&name=default
	case query.FieldName:
		displayName := item.GetAnnotations()[constants.DisplayNameAnnotationKey]
		if displayName != "" && strings.Contains(strings.ToLower(displayName), strings.ToLower(string(filter.Value))) {
			return true
		}
		return strings.Contains(item.GetName(), string(filter.Value))
		// /namespaces?page=1&limit=10&uid=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldUID:
		return strings.Compare(string(item.GetUID()), string(filter.Value)) == 0
		// /deployments?page=1&limit=10&namespace=kubesphere-system
	case query.FieldNamespace:
		return strings.Compare(item.GetNamespace(), string(filter.Value)) == 0
		// /namespaces?page=1&limit=10&ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldOwnerReference:
		for _, ownerReference := range item.GetOwnerReferences() {
			if strings.Compare(string(ownerReference.UID), string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&ownerKind=Workspace
	case query.FieldOwnerKind:
		for _, ownerReference := range item.GetOwnerReferences() {
			if strings.Compare(ownerReference.Kind, string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&annotation=openpitrix_runtime
	case query.FieldAnnotation:
		return labelMatch(item.GetAnnotations(), string(filter.Value))
		// /namespaces?page=1&limit=10&label=kubesphere.io/workspace:system-workspace
	case query.FieldLabel:
		return labelMatch(item.GetLabels(), string(filter.Value))
		// /namespaces?page=1&limit=10&labelSelector=environment in (production, qa)
	case query.ParameterLabelSelector:
		return labelSelectorMatch(item.GetLabels(), string(filter.Value))
	case query.ParameterFieldSelector:
		return contains(item.(runtime.Object), filter.Value)
	default:
		return true
	}
}

func labelSelectorMatch(objLabels map[string]string, filter string) bool {
	selector, err := labels.Parse(filter)
	if err != nil {
		klog.V(4).Infof("failed parse labelSelector error: %s", err)
		return false
	}

	return selector.Matches(labels.Set(objLabels))
}

func labelMatch(labels map[string]string, filter string) bool {
	conditions := strings.SplitN(filter, "=", 2)
	var key, value string
	var opposite bool
	if len(conditions) == 2 {
		key = conditions[0]
		if strings.HasSuffix(key, "!") {
			key = strings.TrimSuffix(key, "!")
			opposite = true
		}
		value = conditions[1]
	} else {
		key = conditions[0]
		value = "*"
	}
	for k, v := range labels {
		if opposite {
			if (k == key) && v != value {
				return true
			}
		} else {
			if (k == key) && (value == "*" || v == value) {
				return true
			}
		}
	}
	return false
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

		// Values prefixed with ~ support case insensitivity. (e.g., a=~b, can hit b, B)
		if strings.HasPrefix(value, "~") {
			value = strings.TrimPrefix(value, "~")
			if (negative && !strings.EqualFold(fmt.Sprintf("%v", rawValue), value)) ||
				(!negative && strings.EqualFold(fmt.Sprintf("%v", rawValue), value)) {
				continue
			} else {
				return false
			}
		}

		if (negative && fmt.Sprintf("%v", rawValue) != value) || (!negative && fmt.Sprintf("%v", rawValue) == value) {
			continue
		} else {
			return false
		}
	}
	return true
}
