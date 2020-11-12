package handlers

import (
	"github.com/emicklei/go-restful"
	"net/http"

	"kubesphere.io/kubesphere/pkg/models/servicemesh/metrics/business"
	"kubesphere.io/kubesphere/pkg/models/servicemesh/metrics/log"
	"kubesphere.io/kubesphere/pkg/models/servicemesh/metrics/prometheus"
)

func NamespaceList(w http.ResponseWriter, r *http.Request) {

	business, err := business.Get()
	if err != nil {
		log.Error(err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	namespaces, err := business.Namespace.GetNamespaces()
	if err != nil {
		log.Error(err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, namespaces)
}

// NamespaceMetrics is the API handler to fetch metrics to be displayed, related to all
// services in the namespace
func NamespaceMetrics(request *restful.Request, response *restful.Response) {
	getNamespaceMetrics(request, response, defaultPromClientSupplier, defaultK8SClientSupplier)
}

// getServiceMetrics (mock-friendly version)
func getNamespaceMetrics(request *restful.Request, response *restful.Response, promSupplier promClientSupplier, k8sSupplier k8sClientSupplier) {
	namespace := request.PathParameters()["namespace"]

	prom, _, namespaceInfo := initClientsForMetrics(response.ResponseWriter, promSupplier, k8sSupplier, namespace)
	if prom == nil {
		// any returned value nil means error & response already written
		return
	}

	params := prometheus.IstioMetricsQuery{Namespace: namespace}
	err := extractIstioMetricsQueryParams(request.Request, &params, namespaceInfo)
	if err != nil {
		RespondWithError(response.ResponseWriter, http.StatusBadRequest, err.Error())
		return
	}

	metrics := prom.GetMetrics(&params)
	RespondWithJSON(response.ResponseWriter, http.StatusOK, metrics)
}
