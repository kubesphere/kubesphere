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
package monitoring

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"net/url"
	"strconv"
	"strings"
)

func MonitorCluster(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)

	// TODO: expose kubesphere iam and devops statistics in prometheus format
	var res *metrics.Response
	if r.Type == "statistics" {
		res = metrics.GetClusterStatistics()
	} else {
		res = metrics.GetClusterMetrics(r)
	}

	response.WriteAsJson(res)
}

func MonitorNode(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)
	res := metrics.GetNodeMetrics(r)
	res, metricsNum := res.SortBy(r.SortMetric, r.SortType)
	res = res.Page(r.PageNum, r.LimitNum, metricsNum)
	response.WriteAsJson(res)
}

func MonitorWorkspace(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)

	// TODO: expose kubesphere iam and devops statistics in prometheus format
	var res *metrics.Response
	if r.Type == "statistics" && r.WorkspaceName != "" {
		res = metrics.GetWorkspaceStatistics(r.WorkspaceName)
	} else {
		res = metrics.GetWorkspaceMetrics(r)
		res, metricsNum := res.SortBy(r.SortMetric, r.SortType)
		res = res.Page(r.PageNum, r.LimitNum, metricsNum)
	}

	response.WriteAsJson(res)
}

func MonitorNamespace(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)
	res := metrics.GetNamespaceMetrics(r)
	res, metricsNum := res.SortBy(r.SortMetric, r.SortType)
	res = res.Page(r.PageNum, r.LimitNum, metricsNum)
	response.WriteAsJson(res)
}

func MonitorWorkload(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)
	res := metrics.GetWorkloadMetrics(r)
	res, metricsNum := res.SortBy(r.SortMetric, r.SortType)
	res = res.Page(r.PageNum, r.LimitNum, metricsNum)
	response.WriteAsJson(res)
}

func MonitorPod(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)
	res := metrics.GetPodMetrics(r)
	res, metricsNum := res.SortBy(r.SortMetric, r.SortType)
	res = res.Page(r.PageNum, r.LimitNum, metricsNum)
	response.WriteAsJson(res)
}

func MonitorContainer(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)
	res := metrics.GetContainerMetrics(r)
	res, metricsNum := res.SortBy(r.SortMetric, r.SortType)
	res = res.Page(r.PageNum, r.LimitNum, metricsNum)
	response.WriteAsJson(res)
}

func MonitorPVC(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)
	res := metrics.GetPVCMetrics(r)
	res, metricsNum := res.SortBy(r.SortMetric, r.SortType)
	res = res.Page(r.PageNum, r.LimitNum, metricsNum)
	response.WriteAsJson(res)
}

func MonitorComponent(request *restful.Request, response *restful.Response) {
	r := ParseRequestParams(request)
	res := metrics.GetComponentMetrics(r)
	response.WriteAsJson(res)
}

func ParseRequestParams(request *restful.Request) metrics.RequestParams {
	var requestParams metrics.RequestParams

	queryTime := strings.Trim(request.QueryParameter("time"), " ")
	start := strings.Trim(request.QueryParameter("start"), " ")
	end := strings.Trim(request.QueryParameter("end"), " ")
	step := strings.Trim(request.QueryParameter("step"), " ")
	sortMetric := strings.Trim(request.QueryParameter("sort_metric"), " ")
	sortType := strings.Trim(request.QueryParameter("sort_type"), " ")
	pageNum := strings.Trim(request.QueryParameter("page"), " ")
	limitNum := strings.Trim(request.QueryParameter("limit"), " ")
	tp := strings.Trim(request.QueryParameter("type"), " ")
	metricsFilter := strings.Trim(request.QueryParameter("metrics_filter"), " ")
	resourcesFilter := strings.Trim(request.QueryParameter("resources_filter"), " ")
	nodeName := strings.Trim(request.PathParameter("node"), " ")
	workspaceName := strings.Trim(request.PathParameter("workspace"), " ")
	namespaceName := strings.Trim(request.PathParameter("namespace"), " ")
	workloadKind := strings.Trim(request.PathParameter("kind"), " ")
	workloadName := strings.Trim(request.PathParameter("workload"), " ")
	podName := strings.Trim(request.PathParameter("pod"), " ")
	containerName := strings.Trim(request.PathParameter("container"), " ")
	pvcName := strings.Trim(request.PathParameter("pvc"), " ")
	storageClassName := strings.Trim(request.PathParameter("storageclass"), " ")
	componentName := strings.Trim(request.PathParameter("component"), " ")

	requestParams = metrics.RequestParams{
		SortMetric:       sortMetric,
		SortType:         sortType,
		PageNum:          pageNum,
		LimitNum:         limitNum,
		Type:             tp,
		MetricsFilter:    metricsFilter,
		ResourcesFilter:  resourcesFilter,
		NodeName:         nodeName,
		WorkspaceName:    workspaceName,
		NamespaceName:    namespaceName,
		WorkloadKind:     workloadKind,
		WorkloadName:     workloadName,
		PodName:          podName,
		ContainerName:    containerName,
		PVCName:          pvcName,
		StorageClassName: storageClassName,
		ComponentName:    componentName,
	}

	if metricsFilter == "" {
		requestParams.MetricsFilter = ".*"
	}
	if resourcesFilter == "" {
		requestParams.ResourcesFilter = ".*"
	}

	v := url.Values{}

	if start != "" && end != "" { // range query

		// metrics from a deleted namespace should be hidden
		// therefore, for range query, if range query start time is less than the namespace creation time, set it to creation time
		// it is the same with query at a fixed time point
		if namespaceName != "" {
			nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
			ns, err := nsLister.Get(namespaceName)
			if err == nil {
				creationTime := ns.CreationTimestamp.Time.Unix()
				queryStart, err := strconv.ParseInt(start, 10, 64)
				if err == nil && queryStart < creationTime {
					start = strconv.FormatInt(creationTime, 10)
				}
			}
		}

		v.Set("start", start)
		v.Set("end", end)

		if step == "" {
			v.Set("step", metrics.DefaultQueryStep)
		} else {
			v.Set("step", step)
		}
		requestParams.QueryParams = v
		requestParams.QueryType = metrics.RangeQuery

		return requestParams
	} else if queryTime != "" { // query
		v.Set("time", queryTime)
	}

	requestParams.QueryParams = v
	requestParams.QueryType = metrics.Query
	return requestParams
}
