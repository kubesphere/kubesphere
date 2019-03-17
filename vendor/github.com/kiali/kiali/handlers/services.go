package handlers

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus"
	"github.com/kiali/kiali/util"
)

// ServiceList is the API handler to fetch the list of services in a given namespace
func ServiceList(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	// Get business layer
	business, err := business.Get()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Services initialization error: "+err.Error())
		return
	}
	namespace := params["namespace"]

	// Fetch and build services
	serviceList, err := business.Svc.GetServiceList(namespace)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, serviceList)
}

// ServiceMetrics is the API handler to fetch metrics to be displayed, related to a single service
func ServiceMetrics(w http.ResponseWriter, r *http.Request) {
	getServiceMetrics(w, r, defaultPromClientSupplier, defaultK8SClientSupplier)
}

// getServiceMetrics (mock-friendly version)
func getServiceMetrics(w http.ResponseWriter, r *http.Request, promSupplier promClientSupplier, k8sSupplier k8sClientSupplier) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	service := vars["service"]

	prom, _, namespaceInfo := initClientsForMetrics(w, promSupplier, k8sSupplier, namespace)
	if prom == nil {
		// any returned value nil means error & response already written
		return
	}

	params := prometheus.IstioMetricsQuery{Namespace: namespace, Service: service}
	err := extractIstioMetricsQueryParams(r, &params, namespaceInfo)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics := prom.GetMetrics(&params)
	RespondWithJSON(w, http.StatusOK, metrics)
}

// ServiceDetails is the API handler to fetch full details of an specific service
func ServiceDetails(w http.ResponseWriter, r *http.Request) {
	// Get business layer
	business, err := business.Get()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Services initialization error: "+err.Error())
		return
	}

	// Rate interval is needed to fetch request rates based health
	rateInterval := defaultHealthRateInterval
	queryParams := r.URL.Query()
	if rateIntervals, ok := queryParams["rateInterval"]; ok && len(rateIntervals) > 0 {
		rateInterval = rateIntervals[0]
	}

	includeValidations := false
	if _, found := queryParams["validate"]; found {
		includeValidations = true
	}

	params := mux.Vars(r)
	namespace := params["namespace"]
	service := params["service"]
	queryTime := util.Clock.Now()
	rateInterval, err = adjustRateInterval(business, namespace, rateInterval, queryTime)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Adjust rate interval error: "+err.Error())
		return
	}

	var istioConfigValidations = models.IstioValidations{}

	wg := sync.WaitGroup{}
	if includeValidations {
		wg.Add(1)
		go func(istioConfigValidations *models.IstioValidations, err *error) {
			defer wg.Done()
			istioConfigValidationResults, errValidations := business.Validations.GetValidations(namespace, service)
			if errValidations != nil && err == nil {
				*err = errValidations
			} else {
				*istioConfigValidations = istioConfigValidationResults
			}
		}(&istioConfigValidations, &err)
	}

	serviceDetails, err := business.Svc.GetService(namespace, service, rateInterval, queryTime)
	if includeValidations && err == nil {
		wg.Wait()
		serviceDetails.Validations = istioConfigValidations
	}

	if err != nil {
		if errors.IsNotFound(err) {
			RespondWithError(w, http.StatusNotFound, err.Error())
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			RespondWithError(w, http.StatusInternalServerError, statusError.ErrStatus.Message)
		} else {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, serviceDetails)
}

// ServiceDashboard is the API handler to fetch Istio dashboard, related to a single service
func ServiceDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	service := vars["service"]

	prom, _, namespaceInfo := initClientsForMetrics(w, defaultPromClientSupplier, defaultK8SClientSupplier, namespace)
	if prom == nil {
		// any returned value nil means error & response already written
		return
	}

	params := prometheus.IstioMetricsQuery{Namespace: namespace, Service: service}
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
