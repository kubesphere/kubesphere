package storage

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {
	glog.Infof("Run Register, subPath=%s", subPath)

	ws.Route(ws.GET(subPath+"/storageclasses/{storageclass}/persistentvolumeclaims").
		To(models.GetStorageListClaims).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}
