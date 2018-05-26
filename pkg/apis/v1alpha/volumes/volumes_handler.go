package volumes

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {
	ws.Route(ws.GET(subPath+"/namespaces/{namespace}/persistentvolumeclaims/{pvc}/pods").
		To(models.GetPodListByPvc).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}
