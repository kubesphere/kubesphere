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
package prometheus

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"os"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
)

const (
	DefaultScheme          = "http"
	DefaultPrometheusPort  = "9090"
	PrometheusApiPath      = "/api/v1/"
	DefaultQueryStep       = "10m"
	DefaultQueryTimeout    = "10s"
	RangeQueryType         = "query_range?"
	DefaultQueryType       = "query?"
	PrometheusAPIServerEnv = "PROMETHEUS_API_SERVER"
)

var PrometheusAPIServer = "prometheus-k8s.kubesphere-monitoring-system.svc"
var PrometheusEndpointUrl string

func init() {
	if env := os.Getenv(PrometheusAPIServerEnv); env != "" {
		PrometheusAPIServer = env
	}
	PrometheusEndpointUrl = DefaultScheme + "://" + PrometheusAPIServer + ":" + DefaultPrometheusPort + PrometheusApiPath
}

type MonitoringRequestParams struct {
	Params          url.Values
	QueryType       string
	SortMetricName  string
	SortType        string
	PageNum         string
	LimitNum        string
	Tp              string
	MetricsFilter   string
	ResourcesFilter string
	MetricsName     string
	WorkloadName    string
	NodeId          string
	WsName          string
	NsName          string
	PodName         string
	ContainerName   string
	WorkloadKind    string
}

var client = &http.Client{}

func SendMonitoringRequest(queryType string, params string) string {
	epurl := PrometheusEndpointUrl + queryType + params
	response, err := client.Get(epurl)
	if err != nil {
		glog.Error(err)
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			glog.Error(err)
		}
		return string(contents)
	}
	return ""
}

func ParseMonitoringRequestParams(request *restful.Request) *MonitoringRequestParams {
	instantTime := strings.Trim(request.QueryParameter("time"), " ")
	start := strings.Trim(request.QueryParameter("start"), " ")
	end := strings.Trim(request.QueryParameter("end"), " ")
	step := strings.Trim(request.QueryParameter("step"), " ")
	timeout := strings.Trim(request.QueryParameter("timeout"), " ")

	sortMetricName := strings.Trim(request.QueryParameter("sort_metric"), " ")
	sortType := strings.Trim(request.QueryParameter("sort_type"), " ")
	pageNum := strings.Trim(request.QueryParameter("page"), " ")
	limitNum := strings.Trim(request.QueryParameter("limit"), " ")
	tp := strings.Trim(request.QueryParameter("type"), " ")

	metricsFilter := strings.Trim(request.QueryParameter("metrics_filter"), " ")
	resourcesFilter := strings.Trim(request.QueryParameter("resources_filter"), " ")

	metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
	workloadName := strings.Trim(request.PathParameter("workload"), " ")

	nodeId := strings.Trim(request.PathParameter("node"), " ")
	wsName := strings.Trim(request.PathParameter("workspace"), " ")
	nsName := strings.Trim(request.PathParameter("namespace"), " ")
	podName := strings.Trim(request.PathParameter("pod"), " ")
	containerName := strings.Trim(request.PathParameter("container"), " ")
	workloadKind := strings.Trim(request.PathParameter("workload_kind"), " ")

	var requestParams = MonitoringRequestParams{
		SortMetricName:  sortMetricName,
		SortType:        sortType,
		PageNum:         pageNum,
		LimitNum:        limitNum,
		Tp:              tp,
		MetricsFilter:   metricsFilter,
		ResourcesFilter: resourcesFilter,
		MetricsName:     metricsName,
		WorkloadName:    workloadName,
		NodeId:          nodeId,
		WsName:          wsName,
		NsName:          nsName,
		PodName:         podName,
		ContainerName:   containerName,
		WorkloadKind:    workloadKind,
	}

	if timeout == "" {
		timeout = DefaultQueryTimeout
	}
	if step == "" {
		step = DefaultQueryStep
	}
	// Whether query or query_range request
	u := url.Values{}

	if start != "" && end != "" {
		u.Set("start", convertTimeGranularity(start))
		u.Set("end", convertTimeGranularity(end))
		u.Set("step", step)
		u.Set("timeout", timeout)
		requestParams.QueryType = RangeQueryType
		requestParams.Params = u
		return &requestParams
	}
	if instantTime != "" {
		u.Set("time", instantTime)
		u.Set("timeout", timeout)
		requestParams.QueryType = DefaultQueryType
		requestParams.Params = u
		return &requestParams
	} else {
		//u.Set("time", strconv.FormatInt(int64(time.Now().Unix()), 10))
		u.Set("timeout", timeout)
		requestParams.QueryType = DefaultQueryType
		requestParams.Params = u
		return &requestParams
	}
}

func convertTimeGranularity(ts string) string {
	timeFloat, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		glog.Errorf("convert second timestamp %s to minute timestamp failed", ts)
		return strconv.FormatInt(int64(time.Now().Unix()), 10)
	}
	timeInt := int64(timeFloat)
	// convert second timestamp to minute timestamp
	secondTime := time.Unix(timeInt, 0).Truncate(time.Minute).Unix()
	return strconv.FormatInt(secondTime, 10)
}
