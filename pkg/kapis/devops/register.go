package devops

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"kubesphere.io/kubesphere/pkg/kapis/generic"
)

const (
	GroupName = "devops.kubesphere.io"
)

// AddToContainer helps proxy DevOps APIs.
func AddToContainer(container *restful.Container, endpoint string) {
	// Deprecated: It will be replaced by alpha4 in the future
	utilruntime.Must(addToContainer(container, endpoint, schema.GroupVersion{Group: GroupName, Version: "alpha2"}))
	// Deprecated: It will be replaced by alpha4 in the future
	utilruntime.Must(addToContainer(container, endpoint, schema.GroupVersion{Group: GroupName, Version: "alpha3"}))

	utilruntime.Must(addToContainer(container, endpoint, schema.GroupVersion{Group: GroupName, Version: "alpha4"}))
}

func addToContainer(container *restful.Container, endpoint string, gv schema.GroupVersion) error {
	proxy, err := generic.NewGenericProxy(endpoint, gv.Group, gv.Version)
	if err != nil {
		return err
	}

	return proxy.AddToContainer(container)
}
