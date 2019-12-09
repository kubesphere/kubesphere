package install

import (
	"github.com/emicklei/go-restful"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/kapis/servicemesh/metrics/v1alpha2"
)

func init() {
	Install(runtime.Container)
}

func Install(c *restful.Container) {
	urlruntime.Must(v1alpha2.AddToContainer(c))
}
