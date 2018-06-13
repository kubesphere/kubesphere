package storage

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
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

	ws.Route(ws.GET(subPath+"/storageclasses/metrics").
		To(GetScMetricsList).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

// List all PersistentVolumeClaims of a specific StorageClass
// Extended API URL: "GET /api/v1alpha1/storage/storageclasses/{storageclass}/persistentvolumeclaims"
func GetPvcListBySc(request *restful.Request, response *restful.Response) {
	scName := request.PathParameter("storageclass")
	claims, err := models.GetPvcListBySc(scName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	result := models.PvcListBySc{scName, claims}

	response.WriteAsJson(result)
}

// Get StorageClass item
// Extended API URL: "GET /api/v1alpha1/storage/storageclasses/{storageclass}/metrics"
func GetScMetrics(request *restful.Request, response *restful.Response) {
	scName := request.PathParameter("storageclass")
	result, err := models.GetScItemMetrics(scName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	response.WriteAsJson(result)
}

// Get StorageClass item list
// Extended API URL: "GET /api/v1alpha1/storage/storageclasses/metrics"
func GetScMetricsList(request *restful.Request, response *restful.Response) {
	result, err := models.GetScItemMetricsList()
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	response.WriteAsJson(result)
}
