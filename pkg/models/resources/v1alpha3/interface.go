package v1alpha3

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"sort"
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

	start, end := query.Pagination.GetPaginationSettings(len(filtered))
	if !query.Pagination.IsPageAvailable(len(filtered), start) {
		return &api.ListResult{
			Items:      nil,
			TotalItems: 0,
		}
	}

	// sort by sortBy field
	sort.Slice(filtered, func(i, j int) bool {
		if !query.Ascending {
			return !compareFunc(filtered[i], filtered[j], query.SortBy)
		}
		return compareFunc(filtered[i], filtered[j], query.SortBy)
	})

	return &api.ListResult{
		Items:      objectsToInterfaces(filtered[start:end]),
		TotalItems: len(filtered),
	}
}

func objectsToInterfaces(objs []runtime.Object) []interface{} {
	var res []interface{}
	for _, obj := range objs {
		res = append(res, obj)
	}
	return res
}