/*
Copyright 2018 The KubeSphere Authors.

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

package v1beta1

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/util"

	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/metrics"
)

var groupVersion = util.GroupVersion{Group: "monitoring", Version: "v1beta1"}

type ResourceHandlerSwagger struct {
	handler func(*restful.Request, *restful.Response)
	params  []*restful.Parameter
}

var ResourceHandlerMap = map[string]ResourceHandlerSwagger{
	"clusters": {
		handler: MonitorCluster,
		params: []*restful.Parameter{
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("cluster_cpu_utilisation"),
		},
	},

	"nodes": {
		handler: monitorNode,
		params: []*restful.Parameter{
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("node_cpu_utilisation"),
			restful.QueryParameter("nodes_filter", "node re2 expression filter").DataType("string").Required(false).DefaultValue(""),
			restful.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false),
			restful.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false),
			restful.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1"),
			restful.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("4"),
		},
	},

	"nodes/{node_id}": {
		handler: monitorNode,
		params: []*restful.Parameter{
			restful.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue(""),
			restful.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("node_cpu_utilisation"),
		},
	},

	"namespaces": {
		handler: monitorNamespace,
		params: []*restful.Parameter{
			restful.QueryParameter("namespaces_filter", "namespaces re2 expression filter").DataType("string").Required(false).DefaultValue(".*"),
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("namespace_memory_utilisation"),
			restful.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false),
			restful.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false),
			restful.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1"),
			restful.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("5"),
		},
	},

	"namespaces/{ns_name}": {
		handler: monitorNamespace,
		params: []*restful.Parameter{
			restful.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring"),
			restful.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("namespace_memory_utilisation"),
		},
	},

	"namespaces/{ns_name}/pods": {
		handler: monitorPod,
		params: []*restful.Parameter{
			restful.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring"),
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache"),
			restful.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue(".*"),
			restful.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false),
			restful.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false),
			restful.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1"),
			restful.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("5"),
		},
	},

	"namespaces/{ns_name}/pods/{pod_name}": {
		handler: monitorPod,
		params: []*restful.Parameter{
			restful.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring"),
			restful.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue(""),
			restful.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache"),
		},
	},

	"nodes/{node_id}/pods": {
		handler: monitorPod,
		params: []*restful.Parameter{
			restful.PathParameter("node_id", "specific node").DataType("string").Required(true),
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache"),
			restful.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false),
			restful.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false),
			restful.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false),
			restful.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1"),
			restful.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("5"),
		},
	},

	"nodes/{node_id}/pods/{pod_name}": {
		handler: monitorPod,
		params: []*restful.Parameter{
			restful.PathParameter("node_id", "specific node").DataType("string").Required(true),
			restful.PathParameter("pod_name", "specific pod").DataType("string").Required(true),
			restful.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache"),
		},
	},

	"namespaces/{ns_name}/pods/{pod_name}/containers": {
		handler: monitorContainer,
		params: []*restful.Parameter{
			restful.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring"),
			restful.PathParameter("pod_name", "specific pod").DataType("string").Required(true),
			restful.QueryParameter("containers_filter", "container re2 expression filter").DataType("string").Required(false),
			restful.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache"),
		},
	},

	"namespaces/{ns_name}/pods/{pod_name}/containers/{container_name}": {
		handler: monitorContainer,
		params: []*restful.Parameter{
			restful.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring"),
			restful.PathParameter("pod_name", "specific pod").DataType("string").Required(true),
			restful.PathParameter("container_name", "specific container").DataType("string").Required(true),
			restful.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache"),
		},
	},

	"namespaces/{ns_name}/workloads/{workload_kind}": {
		handler: monitorWorkload,
		params: []*restful.Parameter{
			restful.PathParameter("ns_name", "namespace").DataType("string").Required(true).DefaultValue("kube-system"),
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false),
			restful.PathParameter("workload_kind", "workload kind").DataType("string").Required(false).DefaultValue("daemonset"),
			restful.QueryParameter("workload_name", "workload name").DataType("string").Required(true),
		},
	},

	"namespaces/{ns_name}/workloads": {
		handler: monitorWorkload,
		params: []*restful.Parameter{
			restful.PathParameter("ns_name", "namespace").DataType("string").Required(true).DefaultValue("kube-system"),
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false),
		},
	},

	"workspaces/{workspace_name}": {
		handler: monitorOneWorkspace,
		params: []*restful.Parameter{
			restful.PathParameter("workspace_name", "workspace name").DataType("string").Required(true),
			restful.QueryParameter("namespaces_filter", "namespaces filter").DataType("string").Required(false),
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("namespace_memory_utilisation_wo_cache"),
			restful.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false),
			restful.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false),
			restful.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1"),
			restful.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("5"),
			restful.QueryParameter("type", "page number").DataType("string").Required(false).DefaultValue("rank"),
		},
	},

	"workspaces": {
		handler: monitorAllWorkspaces,
		params: []*restful.Parameter{
			restful.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("workspace_memory_utilisation"),
			restful.QueryParameter("workspaces_filter", "workspaces re2 expression filter").DataType("string").Required(false).DefaultValue(".*"),
			restful.QueryParameter("sort_metric", "sort metric").DataType("string").Required(false),
			restful.QueryParameter("sort_type", "ascending descending order").DataType("string").Required(false),
			restful.QueryParameter("page", "page number").DataType("string").Required(false).DefaultValue("1"),
			restful.QueryParameter("limit", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("5"),
			restful.QueryParameter("type", "page number").DataType("string").Required(false).DefaultValue("rank"),
		},
	},

	"events": {
		handler: monitorEvents,
		params: []*restful.Parameter{
			restful.QueryParameter("namespaces_filter", "namespaces filter").DataType("string").Required(false).DefaultValue(".*"),
		},
	},

	"components": {
		handler: monitorComponentStatus,
		params:  []*restful.Parameter{},
	},
}

func Register(ws *restful.WebService) {

	tags := []string{"monitoring apis"}
	for _pattern, _handlerAndSwagger := range ResourceHandlerMap {
		var apiPath = groupVersion.String() + "/" + _pattern
		router := ws.GET(apiPath).To(_handlerAndSwagger.handler)

		// add swagger doc for this router builder
		for i := 0; i < len(_handlerAndSwagger.params); i++ {
			router.Param(_handlerAndSwagger.params[i])
		}

		ws.Route(router.
			Filter(route.RouteLogging).
			Metadata(restfulspec.KeyOpenAPITags, tags)).
			Consumes(restful.MIME_JSON, restful.MIME_XML).
			Produces(restful.MIME_JSON)
	}
}

func monitorPod(request *restful.Request, response *restful.Response) {
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

func monitorContainer(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)
	res := metrics.MonitorContainer(requestParams)

	response.WriteAsJson(res)
}

func monitorWorkload(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)
	wlKind := requestParams.WorkloadKind
	if wlKind == "" {
		// count all workloads figure
		res := metrics.MonitorWorkloadCount(requestParams.NsName)
		response.WriteAsJson(res)
	} else {
		res := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelWorkload)
		response.WriteAsJson(res)
	}
}

func monitorAllWorkspaces(request *restful.Request, response *restful.Response) {

	requestParams := client.ParseMonitoringRequestParams(request)

	if requestParams.Tp == "_statistics" {
		// merge multiple metric: all-devops, all-roles, all-projects...this api is designed for admin
		res := metrics.MonitorAllWorkspacesStatistics()

		response.WriteAsJson(res)
	} else {
		rawMetrics := metrics.MonitorAllWorkspaces(requestParams)
		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics, metrics.MetricLevelWorkspace)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	}
}

func monitorOneWorkspace(request *restful.Request, response *restful.Response) {
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

func monitorNamespace(request *restful.Request, response *restful.Response) {
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

func MonitorCluster(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)

	metricName := requestParams.MetricsName
	if metricName != "" {
		// single
		queryType, params := metrics.AssembleClusterMetricRequestInfo(requestParams, metricName)
		res := metrics.GetMetric(queryType, params, metricName)

		if metricName == metrics.MetricNameWorkspaceAllProjectCount {
			res = metrics.MonitorWorkspaceNamespaceHistory(res)
		}

		response.WriteAsJson(res)
	} else {
		// multiple
		res := metrics.MonitorAllMetrics(requestParams, metrics.MetricLevelCluster)
		response.WriteAsJson(res)
	}
}

func monitorNode(request *restful.Request, response *restful.Response) {
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
func monitorComponentStatus(request *restful.Request, response *restful.Response) {
	requestParams := client.ParseMonitoringRequestParams(request)

	status := metrics.MonitorComponentStatus(requestParams)
	response.WriteAsJson(status)
}

func monitorEvents(request *restful.Request, response *restful.Response) {
	// k8s component healthy status
	requestParams := client.ParseMonitoringRequestParams(request)

	nsFilter := requestParams.NsFilter
	events := metrics.MonitorEvents(nsFilter)
	response.WriteAsJson(events)
}
