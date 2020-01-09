package v1alpha2

import "kubesphere.io/kubesphere/pkg/simple/client/monitoring"

type APIResponse struct {
	Results     []monitoring.Metric `json:"results" description:"actual array of results"`
	CurrentPage int                 `json:"page,omitempty" description:"current page returned"`
	TotalPage   int                 `json:"total_page,omitempty" description:"total number of pages"`
	TotalItem   int                 `json:"total_item,omitempty" description:"page size"`
}
