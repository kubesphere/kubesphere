package resource

import (
	"errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/application"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/deployment"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/namespace"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

type ResourceGetter struct {
	getters map[schema.GroupVersionResource]v1alpha3.Interface
}

func NewResourceGetter(factory informers.InformerFactory) *ResourceGetter {
	getters := make(map[schema.GroupVersionResource]v1alpha3.Interface)

	getters[schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}] = deployment.New(factory.KubernetesSharedInformerFactory())
	getters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}] = namespace.New(factory.KubernetesSharedInformerFactory())
	getters[schema.GroupVersionResource{Group: "app.k8s.io", Version: "v1beta1", Resource: "applications"}] = application.New(factory.ApplicationSharedInformerFactory())

	return &ResourceGetter{
		getters: getters,
	}
}

// tryResource will retrieve a getter with resource name, it doesn't guarantee find resource with correct group version
// need to refactor this use schema.GroupVersionResource
func (r *ResourceGetter) tryResource(resource string) v1alpha3.Interface {
	for k, v := range r.getters {
		if k.Resource == resource {
			return v
		}
	}

	return nil
}

func (r *ResourceGetter) Get(resource, namespace, name string) (interface{}, error) {
	getter := r.tryResource(resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.Get(namespace, name)
}

func (r *ResourceGetter) List(resource, namespace string, query *query.Query) (*api.ListResult, error) {
	getter := r.tryResource(resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.List(namespace, query)
}
