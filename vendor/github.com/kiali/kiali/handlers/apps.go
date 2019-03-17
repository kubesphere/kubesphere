package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/prometheus"
)

// AppList is the API handler to fetch all the apps to be displayed, related to a single namespace
func AppList(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	// Get business layer
	business, err := business.Get()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Apps initialization error: "+err.Error())
		return
	}
	namespace := params["namespace"]

	// Fetch and build apps
	appList, err := business.App.GetAppList(namespace)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, appList)
}

// AppDetails is the API handler to fetch all details to be displayed, related to a single app
func AppDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	// Get business layer
	business, err := business.Get()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Services initialization error: "+err.Error())
		return
	}
	namespace := params["namespace"]
	app := params["app"]

	// Fetch and build app
	appDetails, err := business.App.GetApp(namespace, app)
	if err != nil {
		if errors.IsNotFound(err) {
			RespondWithError(w, http.StatusNotFound, err.Error())
		} else {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, appDetails)
}

// AppMetrics is the API handler to fetch metrics to be displayed, related to an app-label grouping
func AppMetrics(w http.ResponseWriter, r *http.Request) {
	getAppMetrics(w, r, defaultPromClientSupplier, defaultK8SClientSupplier)
}

// getAppMetrics (mock-friendly version)
func getAppMetrics(w http.ResponseWriter, r *http.Request, promSupplier promClientSupplier, k8sSupplier k8sClientSupplier) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	app := vars["app"]

	prom, _, namespaceInfo := initClientsForMetrics(w, promSupplier, k8sSupplier, namespace)
	if prom == nil {
		// any returned value nil means error & response already written
		return
	}

	params := prometheus.IstioMetricsQuery{Namespace: namespace, App: app}
	err := extractIstioMetricsQueryParams(r, &params, namespaceInfo)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics := prom.GetMetrics(&params)
	RespondWithJSON(w, http.StatusOK, metrics)
}

// CustomDashboard is the API handler to fetch runtime metrics to be displayed, related to a single app
func CustomDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	app := vars["app"]
	template := vars["template"]

	prom, _, namespaceInfo := initClientsForMetrics(w, defaultPromClientSupplier, defaultK8SClientSupplier, namespace)
	if prom == nil {
		// any returned value nil means error & response already written
		return
	}

	monitoringClient, err := kubernetes.NewKialiMonitoringClient()
	if err != nil {
		log.Error(err)
		RespondWithError(w, http.StatusServiceUnavailable, "Kiali monitoring client error: "+err.Error())
		return
	}

	svc := business.NewDashboardsService(monitoringClient, prom)

	params := prometheus.CustomMetricsQuery{Namespace: namespace, App: app}
	err = extractCustomMetricsQueryParams(r, &params, namespaceInfo)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	dashboard, err := svc.GetDashboard(params, template)
	if err != nil {
		if errors.IsNotFound(err) {
			RespondWithError(w, http.StatusNotFound, err.Error())
		} else {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, dashboard)
}

// AppDashboard is the API handler to fetch Istio dashboard, related to a single app
func AppDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	app := vars["app"]

	prom, _, namespaceInfo := initClientsForMetrics(w, defaultPromClientSupplier, defaultK8SClientSupplier, namespace)
	if prom == nil {
		// any returned value nil means error & response already written
		return
	}

	params := prometheus.IstioMetricsQuery{Namespace: namespace, App: app}
	err := extractIstioMetricsQueryParams(r, &params, namespaceInfo)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	svc := business.NewDashboardsService(nil, prom)
	dashboard, err := svc.GetIstioDashboard(params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, dashboard)
}
