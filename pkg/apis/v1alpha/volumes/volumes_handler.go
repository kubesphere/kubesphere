package volumes

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {
	ws.Route(ws.GET(subPath+"/namespaces/{namespace}/persistentvolumeclaims/{pvc}/pods").
		To(GetPodListByPvc).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

// List all pods of a specific PVC
// Extended API URL: "GET /api/v1alpha1/volumes/namespaces/{namespace}/persistentvolumeclaims/{name}/pods"
func GetPodListByPvc(request *restful.Request, response *restful.Response) {
	pvcName := request.PathParameter("pvc")
	nsName := request.PathParameter("namespace")
	pods, err := models.GetPodListByPvc(pvcName, nsName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	result := models.PodListByPvc{Name: pvcName, Namespace: nsName, Pods: pods}
	response.WriteAsJson(result)
}
