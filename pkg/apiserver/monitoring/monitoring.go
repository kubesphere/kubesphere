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
			metricsStr := prometheus.SendMonitoringRequest(prometheus.PrometheusEndpoint, queryType, params)
			res = metrics.ReformatJson(metricsStr, metricName, map[string]string{metrics.MetricLevelPodName: ""})
		}
		response.WriteAsJson(res)

	} else {
		// multiple
		rawMetrics := metrics.GetPodLevelMetrics(requestParams)
		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)
		response.WriteAsJson(pagedMetrics)
	}
}

func MonitorContainer(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)
	metricName := requestParams.MetricsName
	if requestParams.MetricsFilter != "" {
		rawMetrics := metrics.GetContainerLevelMetrics(requestParams)
		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
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

	rawMetrics := metrics.GetWorkloadLevelMetrics(requestParams)

	// sorting
	sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)

	// paging
	pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

	response.WriteAsJson(pagedMetrics)

}

func MonitorAllWorkspaces(request *restful.Request, response *restful.Response) {

	requestParams := prometheus.ParseMonitoringRequestParams(request)

	tp := requestParams.Tp
	if tp == "statistics" {
		// merge multiple metric: all-devops, all-roles, all-projects...this api is designed for admin
		res := metrics.GetAllWorkspacesStatistics()
		response.WriteAsJson(res)

	} else if tp == "rank" {
		rawMetrics := metrics.MonitorAllWorkspaces(requestParams)

		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)

		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	} else {
		rawMetrics := metrics.MonitorAllWorkspaces(requestParams)
		response.WriteAsJson(rawMetrics)
	}
}

func MonitorOneWorkspace(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	tp := requestParams.Tp
	if tp == "rank" {
		// multiple
		rawMetrics := metrics.GetWorkspaceLevelMetrics(requestParams)

		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)

		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)
		response.WriteAsJson(pagedMetrics)

	} else if tp == "statistics" {
		wsName := requestParams.WsName

		// merge multiple metric: devops, roles, projects...
		res := metrics.MonitorOneWorkspaceStatistics(wsName)
		response.WriteAsJson(res)
	} else {
		res := metrics.GetWorkspaceLevelMetrics(requestParams)
		response.WriteAsJson(res)
	}
}

func MonitorNamespace(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)
	// multiple
	rawMetrics := metrics.GetNamespaceLevelMetrics(requestParams)

	// sorting
	sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
	// paging
	pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)
	response.WriteAsJson(pagedMetrics)
}

func MonitorCluster(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	metricName := requestParams.MetricsName
	if metricName != "" {
		// single
		queryType, params := metrics.AssembleClusterMetricRequestInfo(requestParams, metricName)
		metricsStr := prometheus.SendMonitoringRequest(prometheus.PrometheusEndpoint, queryType, params)
		res := metrics.ReformatJson(metricsStr, metricName, map[string]string{metrics.MetricLevelCluster: "local"})

		response.WriteAsJson(res)
	} else {
		// multiple
		res := metrics.GetClusterLevelMetrics(requestParams)
		response.WriteAsJson(res)
	}
}

func MonitorNode(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	metricName := requestParams.MetricsName
	if metricName != "" {
		// single
		queryType, params := metrics.AssembleNodeMetricRequestInfo(requestParams, metricName)
		metricsStr := prometheus.SendMonitoringRequest(prometheus.PrometheusEndpoint, queryType, params)
		res := metrics.ReformatJson(metricsStr, metricName, map[string]string{metrics.MetricLevelNode: ""})
		// The raw node-exporter result doesn't include ip address information
		// Thereby, append node ip address to .data.result[].metric

		nodeAddress := metrics.GetNodeAddressInfo()
		metrics.AddNodeAddressMetric(res, nodeAddress)

		response.WriteAsJson(res)
	} else {
		// multiple
		rawMetrics := metrics.GetNodeLevelMetrics(requestParams)
		nodeAddress := metrics.GetNodeAddressInfo()

		for i := 0; i < len(rawMetrics.Results); i++ {
			metrics.AddNodeAddressMetric(&rawMetrics.Results[i], nodeAddress)
		}

		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	}
}

func MonitorComponent(request *restful.Request, response *restful.Response) {
	requestParams := prometheus.ParseMonitoringRequestParams(request)

	if requestParams.MetricsFilter == "" {
		requestParams.MetricsFilter = requestParams.ComponentName + "_.*"
	}

	rawMetrics := metrics.GetComponentLevelMetrics(requestParams)

	response.WriteAsJson(rawMetrics)
}
