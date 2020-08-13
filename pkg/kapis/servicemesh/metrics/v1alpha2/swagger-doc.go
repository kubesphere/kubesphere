/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

type graphResponse struct {
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

type metricsResponse struct {
	prometheus.Metrics
}
