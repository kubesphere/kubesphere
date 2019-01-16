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
	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/models/metrics"
)

func V1Alpha2(ws *restful.WebService) {
	u := Monitor{}
	tags := []string{"Monitoring"}

	ws.Route(ws.GET("/clusters").To(u.monitorCluster).
		Doc("monitor cluster level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("cluster_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes").To(u.monitorNode).
		Doc("monitor nodes level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("node_cpu_utilisation")).
		Param(ws.QueryParameter("nodes_filter", "node re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes/{node}").To(u.monitorNode).
		Doc("monitor specific node level metrics").
		Param(ws.PathParameter("node", "specific node").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("node_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces").To(u.monitorNamespace).
		Doc("monitor namespaces level metrics").
		Param(ws.QueryParameter("namespaces_filter", "namespaces re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("namespace_memory_utilisation")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}").To(u.monitorNamespace).
		Doc("monitor specific namespace level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("namespace_memory_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/pods").To(u.monitorPod).
		Doc("monitor pods level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}").To(u.monitorPod).
		Doc("monitor specific pod level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes/{node}/pods").To(u.monitorPod).
		Doc("monitor pods level metrics by nodeid").
		Param(ws.PathParameter("node", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("openpitrix.*")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}").To(u.monitorPod).
		Doc("monitor specific pod level metrics by nodeid").
		Param(ws.PathParameter("node", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.PathParameter("pod", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}/containers").To(u.monitorContainer).
		Doc("monitor specific pod level metrics by nodeid").
		Param(ws.PathParameter("node", "specific node").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "specific pod").DataType("string").Required(true)).
		Param(ws.QueryParameter("containers_filter", "container re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false)).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").DataType("string").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers").To(u.monitorContainer).
		Doc("monitor containers level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("containers_filter", "container re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false)).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").DataType("string").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/{container}").To(u.monitorContainer).
		Doc("monitor specific container level metrics").
		Param(ws.PathParameter("namespace", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("container", "specific container").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/{workload}").To(u.monitorWorkload).
		Doc("monitor specific workload level metrics").
		Param(ws.PathParameter("namespace", "namespace").DataType("string").Required(true).DefaultValue("kube-system")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false)).
		Param(ws.PathParameter("workload", "workload kind").DataType("string").Required(false).DefaultValue("daemonset")).
		Param(ws.QueryParameter("workload_name", "workload name").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("openpitrix.*")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "max metric items in a page").DataType("string").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").DataType("string").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/namespaces/{namespace}/workloads").To(u.monitorWorkload).
		Doc("monitor all workload level metrics").
		Param(ws.PathParameter("namespace", "namespace").DataType("string").Required(true).DefaultValue("kube-system")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false)).
		Param(ws.QueryParameter("workloads_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").DataType("string").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	// list all namespace in this workspace by selected metrics
	ws.Route(ws.GET("/workspaces/{workspace}").To(u.monitorOneWorkspace).
		Doc("monitor workspaces level metrics").
		Param(ws.PathParameter("workspace", "workspace name").DataType("string").Required(true)).
		Param(ws.QueryParameter("namespaces_filter", "namespaces filter").DataType("string").Required(false).DefaultValue("k.*")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("namespace_memory_utilisation_wo_cache")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").DataType("string").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/workspaces").To(u.monitorAllWorkspaces).
		Doc("monitor workspaces level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("workspace_memory_utilisation")).
		Param(ws.QueryParameter("workspaces_filter", "workspaces re2 expression filter").DataType("string").Required(false).DefaultValue(".*")).
		Param(ws.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4")).
		Param(ws.QueryParameter("type", "rank, statistic").DataType("string").Required(false).DefaultValue("rank")).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/components").To(u.monitorComponentStatus).
		Doc("monitor k8s components status").
		Metadata(restfulspec.KeyOpenAPITags, tags))

}

func (u Monitor) monitorPod(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)
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

func (u Monitor) monitorContainer(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)
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

func (u Monitor) monitorWorkload(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)

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

func (u Monitor) monitorAllWorkspaces(request *restful.Request, response *restful.Response) {

	requestParams := client.ParseMonitoringRequestParams(request)

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

func (u Monitor) monitorOneWorkspace(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)

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

func (u Monitor) monitorNamespace(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)
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

func (u Monitor) monitorCluster(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)

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

func (u Monitor) monitorNode(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)

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
func (u Monitor) monitorComponentStatus(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)

	status := metrics.MonitorComponentStatus(requestParams)
	response.WriteAsJson(status)
}

type Monitor struct {
}
