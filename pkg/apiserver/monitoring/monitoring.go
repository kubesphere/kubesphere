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
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"kubesphere.io/kubesphere/pkg/simple/client/prometheus"
)

func MonitorPod(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)
	podName := requestParams.PodName
	metricName := requestParams.MetricsName
	if podName != "" {
		// single pod single metric
		queryType, params, nullRule := metrics.AssemblePodMetricRequestInfo(requestParams, metricName)
		var res *metrics.FormatedMetric
		if !nullRule {
			res = metrics.GetMetric(queryType, params, metricName)
		}
		response.WriteAsJson(res)

	} else {
		// multiple
		rawMetrics := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelPod)
		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelPodName)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	}
}

func MonitorContainer(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)
	metricName := requestParams.MetricsName
	if requestParams.MetricsFilter != "" {
		rawMetrics := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelContainer)
		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelContainerName)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)

	} else {
		res := metrics.MonitorContainer(requestParams, metricName)
		response.WriteAsJson(res)
	}

}

func MonitorWorkload(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	rawMetrics := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelWorkload)

	var sortedMetrics *metrics.FormatedLevelMetric
	var maxMetricCount int

	wlKind := requestParams.WorkloadKind

	// sorting
	if wlKind == "" {

		sortedMetrics, maxMetricCount = metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelWorkload)
	} else {

		sortedMetrics, maxMetricCount = metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelPodName)
	}

	// paging
	pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

	response.WriteAsJson(pagedMetrics)

}

func MonitorAllWorkspaces(request *restful.Request, response *restful.Response) {

	requestParams := prometheus.ParseMonitoringRequestParams(request)

	tp := requestParams.Tp
	if tp == "_statistics" {
		// merge multiple metric: all-devops, all-roles, all-projects...this api is designed for admin
		res := metrics.MonitorAllWorkspacesStatistics()

		response.WriteAsJson(res)

	} else if tp == "rank" {
		rawMetrics := metrics.MonitorAllWorkspaces(requestParams)
		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelWorkspace)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	} else {
		res := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelWorkspace)
		response.WriteAsJson(res)
	}
}

func MonitorOneWorkspace(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	tp := requestParams.Tp
	if tp == "rank" {
		// multiple
		rawMetrics := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelWorkspace)
		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelNamespace)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)

	} else if tp == "_statistics" {
		wsName := requestParams.WsName

		// merge multiple metric: devops, roles, projects...
		res := metrics.MonitorOneWorkspaceStatistics(wsName)
		response.WriteAsJson(res)
	} else {
		res := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelWorkspace)
		response.WriteAsJson(res)
	}
}

func MonitorNamespace(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)
	metricName := requestParams.MetricsName
	nsName := requestParams.NsName
	if nsName != "" {
		// single
		queryType, params := metrics.AssembleNamespaceMetricRequestInfo(requestParams, metricName)
		res := metrics.GetMetric(queryType, params, metricName)
		response.WriteAsJson(res)
	} else {
		// multiple
		rawMetrics := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelNamespace)
		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelNamespace)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	}
}

func MonitorCluster(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	metricName := requestParams.MetricsName
	if metricName != "" {
		// single
		queryType, params := metrics.AssembleClusterMetricRequestInfo(requestParams, metricName)
		res := metrics.GetMetric(queryType, params, metricName)

		response.WriteAsJson(res)
	} else {
		// multiple
		res := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelCluster)
		response.WriteAsJson(res)
	}
}

func MonitorNode(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	metricName := requestParams.MetricsName
	if metricName != "" {
		// single
		queryType, params := metrics.AssembleNodeMetricRequestInfo(requestParams, metricName)
		res := metrics.GetMetric(queryType, params, metricName)
		nodeAddress := metrics.GetNodeAddressInfo()
		metrics.AddNodeAddressMetric(res, nodeAddress)
		response.WriteAsJson(res)
	} else {
		// multiple
		rawMetrics := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelNode)
		nodeAddress := metrics.GetNodeAddressInfo()

		for i := 0; i < len(rawMetrics.Results); i++ {
			metrics.AddNodeAddressMetric(&rawMetrics.Results[i], nodeAddress)
		}

		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelNode)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	}
}

// k8s component(controller, scheduler, etcd) status
func MonitorComponentStatus(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	status := metrics.MonitorComponentStatus(requestParams)
	response.WriteAsJson(status)
}
