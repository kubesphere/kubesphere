package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"sort"
	"strings"
)

type Interface interface {
	// Get retrieves a single object by its namespace and name
	Get(namespace, name string) (runtime.Object, error)

	// List retrieves a collection of objects matches given query
	List(namespace string, query *query.Query) (*api.ListResult, error)
}

type CompareFunc func(runtime.Object, runtime.Object, query.Field) bool

type FilterFunc func(runtime.Object, query.Filter) bool

func DefaultList(objects []runtime.Object, query *query.Query, compareFunc CompareFunc, filterFunc FilterFunc) *api.ListResult {
	// selected matched ones
	var filtered []runtime.Object
	for _, object := range objects {
		selected := true
		for _, filter := range query.Filters {
			if !filterFunc(object, filter) {
				selected = false
				break
			}
		}

		if selected {
			filtered = append(filtered, object)
		}
	}

	// sort by sortBy field
	sort.Slice(filtered, func(i, j int) bool {
		if !query.Ascending {
			return !compareFunc(filtered[i], filtered[j], query.SortBy)
		}
		return compareFunc(filtered[i], filtered[j], query.SortBy)
	})

	total := len(filtered)
	start, end := query.Pagination.GetValidPagination(total)

	return &api.ListResult{
		TotalItems: len(filtered),
		Items:      objectsToInterfaces(filtered[start:end]),
	}
}

func DefaultObjectMetaCompare(left, right metav1.ObjectMeta, sortBy query.Field) bool {
	switch sortBy {
	// ?sortBy=name
	case query.FieldName:
		return strings.Compare(left.Name, right.Name) > 0
	//	?sortBy=creationTimestamp
	case query.FieldCreationTimeStamp:
		return left.CreationTimestamp.After(right.CreationTimestamp.Time)
	default:
		return false
	}
}

//  Default metadata filter
func DefaultObjectMetaFilter(item metav1.ObjectMeta, filter query.Filter) bool {
	switch filter.Field {
	// /namespaces?page=1&limit=10&name=default
	case query.FieldName:
		return strings.Contains(item.Name, string(filter.Value))
		// /namespaces?page=1&limit=10&uid=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldUID:
		return strings.Compare(string(item.UID), string(filter.Value)) == 0
		// /deployments?page=1&limit=10&namespace=kubesphere-system
	case query.FieldNamespace:
		return strings.Compare(item.Namespace, string(filter.Value)) == 0
		// /namespaces?page=1&limit=10&ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldOwnerReference:
		for _, ownerReference := range item.OwnerReferences {
			if strings.Compare(string(ownerReference.UID), string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&ownerKind=Workspace
	case query.FieldOwnerKind:
		for _, ownerReference := range item.OwnerReferences {
			if strings.Compare(ownerReference.Kind, string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&annotation=openpitrix_runtime
	case query.FieldAnnotation:
		return containsAnyValue(item.Annotations, string(filter.Value))
		// /namespaces?page=1&limit=10&label=kubesphere.io/workspace:system-workspace
	case query.FieldLabel:
		return containsAnyValue(item.Labels, string(filter.Value))
	default:
		return false
	}
}

// Filter format <key:><value>,if the key is defined, the key must match exactly, value match according to strings.Contains.
func containsAnyValue(keyValues map[string]string, filter string) bool {
	fields := strings.SplitN(filter, ":", 2)
	var keyFilter, valueFilter string
	if len(fields) == 2 {
		keyFilter = fields[0]
		valueFilter = fields[1]
	} else {
		valueFilter = fields[0]
	}
	for key, value := range keyValues {
		if (key == keyFilter || keyFilter == "") && strings.Contains(value, valueFilter) {
			return true
		}
	}
	return false
}

func objectsToInterfaces(objs []runtime.Object) []interface{} {
	res := make([]interface{}, 0)
	for _, obj := range objs {
		res = append(res, obj)
	}
	return res
}
