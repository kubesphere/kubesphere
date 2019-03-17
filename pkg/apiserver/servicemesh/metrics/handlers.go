package metrics

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/kiali/kiali/handlers"
)

// Get app metrics
func GetAppMetrics(request *restful.Request, response *restful.Response) {
	handlers.AppMetrics(response.ResponseWriter, request.Request)
}

// Get workload metrics
func GetWorkloadMetrics(request *restful.Request, response *restful.Response) {
	handlers.WorkloadMetrics(response.ResponseWriter, request.Request)
}

// Get service metrics
func GetServiceMetrics(request *restful.Request, response *restful.Response) {
	handlers.ServiceMetrics(response.ResponseWriter, request.Request)
}

// Get namespace metrics
func GetNamespaceMetrics(request *restful.Request, response *restful.Response) {
	handlers.NamespaceMetrics(response.ResponseWriter, request.Request)
}

// Get service graph for namespace
func GetNamespaceGraph(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	if len(namespace) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s", request.Request.URL.RawQuery, namespace)
	}

	handlers.GraphNamespaces(response.ResponseWriter, request.Request)
}

// Get service graph for namespaces
func GetNamespacesGraph(request *restful.Request, response *restful.Response) {
	handlers.GraphNamespaces(response.ResponseWriter, request.Request)
}
