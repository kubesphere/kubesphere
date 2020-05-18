package v1

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/kapis/generic"
)

// there are no versions specified cause we want to proxy all versions of requests to backend service
var GroupVersion = schema.GroupVersion{Group: "alerting.kubesphere.io", Version: ""}

func AddToContainer(container *restful.Container, endpoint string) error {
	proxy, err := generic.NewGenericProxy(endpoint, GroupVersion.Group, GroupVersion.Version)
	if err != nil {
		return nil
	}

	return proxy.AddToContainer(container)
}
