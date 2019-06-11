package install

import (
	"github.com/emicklei/go-restful"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"kubesphere.io/kubesphere/pkg/apis/servicemesh/metrics/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

func init() {
	Install(runtime.Container)
}

func Install(c *restful.Container) {
	urlruntime.Must(v1alpha2.AddToContainer(c))
}
