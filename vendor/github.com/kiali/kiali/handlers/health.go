package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/util"
)

const defaultHealthRateInterval = "10m"

// NamespaceHealth is the API handler to get app-based health of every services in the given namespace
func NamespaceHealth(w http.ResponseWriter, r *http.Request) {
	// Get business layer
	business, err := business.Get()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Services initialization error: "+err.Error())
		return
	}

	p := namespaceHealthParams{}
	if ok, err := p.extract(r); !ok {
		// Bad request
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// Adjust rate interval
	rateInterval, err := adjustRateInterval(business, p.Namespace, p.RateInterval, p.QueryTime)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Adjust rate interval error: "+err.Error())
		return
	}

	switch p.Type {
	case "app":
		health, err := business.Health.GetNamespaceAppHealth(p.Namespace, rateInterval, p.QueryTime)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Error while fetching app health: "+err.Error())
			return
		}
		RespondWithJSON(w, http.StatusOK, health)
	case "service":
		health, err := business.Health.GetNamespaceServiceHealth(p.Namespace, rateInterval, p.QueryTime)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Error while fetching service health: "+err.Error())
			return
		}
		RespondWithJSON(w, http.StatusOK, health)
	case "workload":
		health, err := business.Health.GetNamespaceWorkloadHealth(p.Namespace, rateInterval, p.QueryTime)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Error while fetching workload health: "+err.Error())
			return
		}
		RespondWithJSON(w, http.StatusOK, health)
	}
}

// AppHealth is the API handler to get health of a single app
func AppHealth(w http.ResponseWriter, r *http.Request) {
	business, err := business.Get()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Services initialization error: "+err.Error())
		return
	}

	p := appHealthParams{}
	p.extract(r)
	rateInterval, err := adjustRateInterval(business, p.Namespace, p.RateInterval, p.QueryTime)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Adjust rate interval error: "+err.Error())
		return
	}

	health, err := business.Health.GetAppHealth(p.Namespace, p.App, rateInterval, p.QueryTime)
	handleHealthResponse(w, health, err)
}

// WorkloadHealth is the API handler to get health of a single workload
func WorkloadHealth(w http.ResponseWriter, r *http.Request) {
	business, err := business.Get()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Services initialization error: "+err.Error())
		return
	}

	p := workloadHealthParams{}
	p.extract(r)
	rateInterval, err := adjustRateInterval(business, p.Namespace, p.RateInterval, p.QueryTime)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Adjust rate interval error: "+err.Error())
		return
	}
	p.RateInterval = rateInterval

	health, err := business.Health.GetWorkloadHealth(p.Namespace, p.Workload, rateInterval, p.QueryTime)
	handleHealthResponse(w, health, err)
}

// ServiceHealth is the API handler to get health of a single service
func ServiceHealth(w http.ResponseWriter, r *http.Request) {
	business, err := business.Get()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Services initialization error: "+err.Error())
		return
	}

	p := serviceHealthParams{}
	p.extract(r)
	rateInterval, err := adjustRateInterval(business, p.Namespace, p.RateInterval, p.QueryTime)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Adjust rate interval error: "+err.Error())
		return
	}

	health, err := business.Health.GetServiceHealth(p.Namespace, p.Service, rateInterval, p.QueryTime)
	handleHealthResponse(w, health, err)
}

func handleHealthResponse(w http.ResponseWriter, health interface{}, err error) {
	if err != nil {
		if errors.IsNotFound(err) {
			RespondWithError(w, http.StatusNotFound, err.Error())
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			RespondWithError(w, http.StatusInternalServerError, statusError.ErrStatus.Message)
		} else {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
	} else {
		RespondWithJSON(w, http.StatusOK, health)
	}
}

type baseHealthParams struct {
	// The namespace scope
	//
	// in: path
	Namespace string `json:"namespace"`
	// The rate interval used for fetching error rate
	//
	// in: query
	// default: 10m
	RateInterval string `json:"rateInterval"`

	// The time to use for the prometheus query
	QueryTime time.Time
}

func (p *baseHealthParams) baseExtract(r *http.Request, vars map[string]string) {
	p.RateInterval = defaultHealthRateInterval
	p.QueryTime = util.Clock.Now()
	queryParams := r.URL.Query()
	if rateIntervals, ok := queryParams["rateInterval"]; ok && len(rateIntervals) > 0 {
		p.RateInterval = rateIntervals[0]
	}
	p.Namespace = vars["namespace"]
}

// namespaceHealthParams holds the path and query parameters for NamespaceHealth
//
// swagger:parameters namespaceHealth
type namespaceHealthParams struct {
	baseHealthParams
	// The type of health, "app", "service" or "workload".
	//
	// in: query
	// pattern: ^(app|service|workload)$
	// default: app
	Type string `json:"type"`
}

func (p *namespaceHealthParams) extract(r *http.Request) (bool, string) {
	vars := mux.Vars(r)
	p.baseExtract(r, vars)
	p.Type = "app"
	queryParams := r.URL.Query()
	if healthTypes, ok := queryParams["type"]; ok && len(healthTypes) > 0 {
		if healthTypes[0] != "app" && healthTypes[0] != "service" && healthTypes[0] != "workload" {
			// Bad request
			return false, "Bad request, query parameter 'type' must be one of ['app','service','workload']"
		}
		p.Type = healthTypes[0]
	}
	return true, ""
}

// appHealthParams holds the path and query parameters for AppHealth
//
// swagger:parameters appHealth
type appHealthParams struct {
	baseHealthParams
	// The target app
	//
	// in: path
	App string `json:"app"`
}

func (p *appHealthParams) extract(r *http.Request) {
	vars := mux.Vars(r)
	p.baseExtract(r, vars)
	p.App = vars["app"]
}

// serviceHealthParams holds the path and query parameters for ServiceHealth
//
// swagger:parameters serviceHealth
type serviceHealthParams struct {
	baseHealthParams
	// The target service
	//
	// in: path
	Service string `json:"service"`
}

func (p *serviceHealthParams) extract(r *http.Request) {
	vars := mux.Vars(r)
	p.baseExtract(r, vars)
	p.Service = vars["service"]
}

// workloadHealthParams holds the path and query parameters for WorkloadHealth
//
// swagger:parameters workloadHealth
type workloadHealthParams struct {
	baseHealthParams
	// The target workload
	//
	// in: path
	Workload string `json:"workload"`
}

func (p *workloadHealthParams) extract(r *http.Request) {
	vars := mux.Vars(r)
	p.baseExtract(r, vars)
	p.Workload = vars["workload"]
}

func adjustRateInterval(business *business.Layer, namespace, rateInterval string, queryTime time.Time) (string, error) {
	namespaceInfo, err := business.Namespace.GetNamespace(namespace)
	if err != nil {
		return "", err
	}
	interval, err := util.AdjustRateInterval(namespaceInfo.CreationTimestamp, queryTime, rateInterval)
	if err != nil {
		return "", err
	}

	if interval != rateInterval {
		log.Debugf("Rate interval for namespace %v was adjusted to %v (original = %v, query time = %v, namespace created = %v)",
			namespace, interval, rateInterval, queryTime, namespaceInfo.CreationTimestamp)
	}

	return interval, nil
}
