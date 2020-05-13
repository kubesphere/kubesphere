package v1

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/kapis/generic"
)

var GroupVersion = schema.GroupVersion{Group: "notification.kubesphere.io", Version: "v1"}

func AddToContainer(container *restful.Container, endpoint string) error {

	proxy, err := generic.NewGenericProxy(endpoint, GroupVersion.Group, GroupVersion.Version)
	if err != nil {
		return err
	}

	return proxy.AddToContainer(container)
}
