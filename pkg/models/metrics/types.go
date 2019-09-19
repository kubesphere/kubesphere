/*

 Copyright 2019 The KubeSphere Authors.

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

package metrics

import (
	"kubesphere.io/kubesphere/pkg/api/monitoring/v1alpha2"
	"net/url"
)

type RequestParams struct {
	QueryParams      url.Values
	QueryType        string
	SortMetric       string
	SortType         string
	PageNum          string
	LimitNum         string
	Type             string
	MetricsFilter    string
	ResourcesFilter  string
	NodeName         string
	WorkspaceName    string
	NamespaceName    string
	WorkloadKind     string
	WorkloadName     string
	PodName          string
	ContainerName    string
	PVCName          string
	StorageClassName string
	ComponentName    string
}

type APIResponse struct {
	MetricName string `json:"metric_name,omitempty" description:"metric name, eg. scheduler_up_sum"`
	v1alpha2.APIResponse
}

type Response struct {
	MetricsLevel string        `json:"metrics_level" description:"metric level, eg. cluster"`
	Results      []APIResponse `json:"results" description:"actual array of results"`
	CurrentPage  int           `json:"page,omitempty" description:"current page returned"`
	TotalPage    int           `json:"total_page,omitempty" description:"total number of pages"`
	TotalItem    int           `json:"total_item,omitempty" description:"page size"`
}
