package storage

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
	"net/http"
)

func Register(ws *restful.WebService, subPath string) {
	ws.Route(ws.GET(subPath+"/storageclasses/{storageclass}/persistentvolumeclaims").
		To(GetPvcListBySc).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/storageclasses/{storageclass}/metrics").
		To(GetScMetrics).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

// List all PersistentVolumeClaims of a specific StorageClass
// Extended API URL: "GET /api/v1alpha/storage/storageclasses/{name}/persistentvolumeclaims"
func GetPvcListBySc(request *restful.Request, response *restful.Response) {
	scName := request.PathParameter("storageclass")
	claims, err := models.GetPvcListBySc(scName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	result := models.PvcListBySc{scName, claims}

	response.WriteAsJson(result)
}

// Get metrics of a specific StorageClass
// Extended API URL: "GET /api/v1alpha/storage/storageclasses/{name}/metrics"
func GetScMetrics(request *restful.Request, response *restful.Response) {
	scName := request.PathParameter("storageclass")
	metrics, err := models.GetScMetrics(scName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	result := models.ScMetrics{Name: scName, Metrics: metrics}
	response.WriteAsJson(result)
}
