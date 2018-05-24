package storage

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {
	ws.Route(ws.GET(subPath+"/storageclasses/{storageclass}/persistentvolumeclaims").
		To(models.GetPvcListBySc).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}
