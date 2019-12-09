package v1alpha2

import (
	"github.com/kiali/kiali/graph/cytoscape"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/prometheus"
)

/////////////////////
// SWAGGER RESPONSES
/////////////////////

// NoContent: the response is empty
type NoContent struct {
	Status int32 `json:"status"`
	Reason error `json:"reason"`
}

// BadRequestError: the client request is incorrect
type BadRequestError struct {
	Status int32 `json:"status"`
	Reason error `json:"reason"`
}

// NotFoundError is the error message that is generated when server could not find
// what was requested
type NotFoundError struct {
	Status int32 `json:"status"`
	Reason error `json:"reason"`
}

type GraphResponse struct {
	cytoscape.Config
}

type serviceHealthResponse struct {
	models.ServiceHealth
}

type namespaceAppHealthResponse struct {
	models.NamespaceAppHealth
}

type workloadHealthResponse struct {
	models.WorkloadHealth
}

type appHealthResponse struct {
	models.AppHealth
}

type MetricsResponse struct {
	prometheus.Metrics
}
