package monitoring

import "kubesphere.io/kubesphere/pkg/simple/client/monitoring"

type Metrics struct {
	Results     []monitoring.Metric `json:"results" description:"actual array of results"`
	CurrentPage int                 `json:"page,omitempty" description:"current page returned"`
	TotalPages  int                 `json:"total_page,omitempty" description:"total number of pages"`
	TotalItems  int                 `json:"total_item,omitempty" description:"page size"`
}

type Metadata struct {
	Data []monitoring.Metadata `json:"data" description:"actual array of results"`
}

type MetricLabelSet struct {
	Data []map[string]string `json:"data" description:"actual array of results"`
}
