package v1alpha2

import "kubesphere.io/kubesphere/pkg/simple/client/logging"

type APIResponse struct {
	Logs       *logging.Logs       `json:"query,omitempty" description:"query results"`
	Statistics *logging.Statistics `json:"statistics,omitempty" description:"statistics results"`
	Histogram  *logging.Histogram  `json:"histogram,omitempty" description:"histogram results"`
}
