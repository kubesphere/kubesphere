package metrics

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"github.com/kiali/kiali/handlers"
)

// Get app metrics
func GetAppMetrics(request *restful.Request, response *restful.Response) {
	handlers.AppMetrics(request, response)
}

// Get workload metrics
func GetWorkloadMetrics(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	workload := request.PathParameter("workload")

	if len(namespace) > 0 && len(workload) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s&workload=%s", request.Request.URL.RawQuery, namespace, workload)
	}

	handlers.WorkloadMetrics(request, response)
}

// Get service metrics
func GetServiceMetrics(request *restful.Request, response *restful.Response) {
	handlers.ServiceMetrics(request, response)
}

// Get namespace metrics
func GetNamespaceMetrics(request *restful.Request, response *restful.Response) {
	handlers.NamespaceMetrics(request, response)
}

// Get service graph for namespace
func GetNamespaceGraph(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	if len(namespace) > 0 {
		request.Request.URL.RawQuery = fmt.Sprintf("%s&namespaces=%s", request.Request.URL.RawQuery, namespace)
	}

	handlers.GetNamespaceGraph(request, response)
}

// Get service graph for namespaces
func GetNamespacesGraph(request *restful.Request, response *restful.Response) {
	handlers.GraphNamespaces(request, response)
}

// Get namespace health
func GetNamespaceHealth(request *restful.Request, response *restful.Response) {
	handlers.NamespaceHealth(request, response)
}

// Get workload health
func GetWorkloadHealth(request *restful.Request, response *restful.Response) {
	handlers.WorkloadHealth(request, response)
}

// Get app health
func GetAppHealth(request *restful.Request, response *restful.Response) {
	handlers.AppHealth(request, response)
}

// Get service health
func GetServiceHealth(request *restful.Request, response *restful.Response) {
	handlers.ServiceHealth(request, response)
}
