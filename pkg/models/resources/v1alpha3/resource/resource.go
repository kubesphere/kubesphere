package resource

import (
	"errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/deployment"
	"sort"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

type NamespacedResourceGetter struct {
	getters map[schema.GroupVersionResource]v1alpha3.Interface
}

func New(informers informers.SharedInformerFactory) *NamespacedResourceGetter {
	getters := make(map[schema.GroupVersionResource]v1alpha3.Interface)

	getters[schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}] = deployment.New(informers)

	return &NamespacedResourceGetter{
		getters: getters,
	}
}

// tryResource will retrieve a getter with resource name, it doesn't guarantee find resource with correct group version
// need to refactor this use schema.GroupVersionResource
func (r *NamespacedResourceGetter) tryResource(resource string) v1alpha3.Interface {
	for k, v := range r.getters {
		if k.Resource == resource {
			return v
		}
	}

	return nil
}

func (r *NamespacedResourceGetter) Get(resource, namespace, name string) (interface{}, error) {
	getter := r.tryResource(resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}

	return getter.Get(namespace, name)
}

func (r *NamespacedResourceGetter) List(resource, namespace string, query *query.Query) (*api.ListResult, error) {
	getter := r.tryResource(resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}

	all, err := getter.List(namespace)
	if err != nil {
		return nil, err
	}

	// selected matched ones
	var filtered []interface{}
	for _, deploy := range all {
		for _, filter := range query.Filters {
			if getter.Filter(deploy, filter) {
				filtered = append(filtered, deploy)
			}
		}
	}

	// sort
	sort.Slice(filtered, func(i, j int) bool {
		if !query.Ascending {
			return !getter.Compare(filtered[i], filtered[j], query.SortBy)
		}
		return getter.Compare(filtered[i], filtered[j], query.SortBy)
	})

	start, end := query.Pagination.GetPaginationSettings(len(filtered))
	if query.Pagination.IsPageAvailable(len(filtered), start) {
		var result []interface{}

		for i := start; i < end; i++ {
			result = append(result, filtered[i])
		}

		return &api.ListResult{
			Items:      result,
			TotalItems: len(filtered),
		}, nil
	}

	return &api.ListResult{
		Items:      nil,
		TotalItems: len(filtered),
	}, nil
}
