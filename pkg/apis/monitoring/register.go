package monitoring

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/apis/monitoring/v1beta1"
)

func Install(ws *restful.WebService) {
	v1beta1.Register(ws)
}
