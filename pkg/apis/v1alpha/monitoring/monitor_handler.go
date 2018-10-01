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
package monitoring

import (
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/metrics"
)

func (u MonitorResource) monitorPod(request *restful.Request, response *restful.Response) {
	podName := strings.Trim(request.PathParameter("pod_name"), " ")
	if podName != "" {
		// single pod single metric
		metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
		res := metrics.MonitorPodSingleMetric(request, metricsName)
		response.WriteAsJson(res)
	} else {
		// multiple pod multiple metric
		res := metrics.MonitorAllMetrics(request)
		response.WriteAsJson(res)
	}
}

func (u MonitorResource) monitorContainer(request *restful.Request, response *restful.Response) {
	metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
	promql := metrics.MakeContainerPromQL(request)
	res := client.SendPrometheusRequest(request, promql)
	cleanedJson := metrics.ReformatJson(res, metricsName)
	response.WriteAsJson(cleanedJson)
}

func (u MonitorResource) monitorWorkload(request *restful.Request, response *restful.Response) {
	wlKind := request.PathParameter("workload_kind")
	if strings.Trim(wlKind, " ") == "" {
		// count all workloads figure
		//metricName := "workload_count"
		res := metrics.MonitorWorkloadCount(request)
		response.WriteAsJson(res)
	} else {
		res := metrics.MonitorAllMetrics(request)
		response.WriteAsJson(res)
	}
}

// merge multiple metric: all-devops, all-roles, all-projects...this api is designed for admin
func (u MonitorResource) monitorWorkspaceUserInfo(request *restful.Request, response *restful.Response) {
	res := metrics.MonitorWorkspaceUserInfo(request)
	response.WriteAsJson(res)
}

// merge multiple metric: devops, roles, projects...
func (u MonitorResource) monitorWorkspaceResourceLevelMetrics(request *restful.Request, response *restful.Response) {
	res := metrics.MonitorWorkspaceResourceLevelMetrics(request)
	response.WriteAsJson(res)
}

func (u MonitorResource) monitorWorkspacePodLevelMetrics(request *restful.Request, response *restful.Response) {
	res := metrics.MonitorAllMetrics(request)
	response.WriteAsJson(res)
}

func (u MonitorResource) monitorNamespace(request *restful.Request, response *restful.Response) {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	if nsName != "" {
		// single
		metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
		res := metrics.MonitorNamespaceSingleMetric(request, metricsName)
		response.WriteAsJson(res)
	} else {
		// multiple
		res := metrics.MonitorAllMetrics(request)
		response.WriteAsJson(res)
	}
}

func (u MonitorResource) monitorNodeorCluster(request *restful.Request, response *restful.Response) {
	metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
	//var res *metrics.FormatedMetric
	if metricsName != "" {
		// single
		res := metrics.MonitorNodeorClusterSingleMetric(request, metricsName)
		response.WriteAsJson(res)
	} else {
		// multiple
		res := metrics.MonitorAllMetrics(request)
		response.WriteAsJson(res)
	}
}

type MonitorResource struct {
}

func Register(ws *restful.WebService, subPath string) {
	tags := []string{"monitoring apis"}
	u := MonitorResource{}

	ws.Route(ws.GET(subPath+"/clusters").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor cluster level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("cluster_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/nodes").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor nodes level metrics").
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("node_cpu_utilisation")).
		Param(ws.QueryParameter("nodes_filter", "node re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/nodes/{node_id}").To(u.monitorNodeorCluster).
		Filter(route.RouteLogging).
		Doc("monitor specific node level metrics").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("node_cpu_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces").To(u.monitorNamespace).
		Filter(route.RouteLogging).
		Doc("monitor namespaces level metrics").
		Param(ws.QueryParameter("namespaces_filter", "namespaces re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("namespace_memory_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}").To(u.monitorNamespace).
		Filter(route.RouteLogging).
		Doc("monitor specific namespace level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("namespace_memory_utilisation")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/pods").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor pods level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/pods/{pod_name}").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor specific pod level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/nodes/{node_id}/pods").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor pods level metrics by nodeid").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.QueryParameter("pods_filter", "pod re2 expression filter").DataType("string").Required(false).DefaultValue("openpitrix.*")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...in re2 regex").DataType("string").Required(false).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/nodes/{node_id}/pods/{pod_name}").To(u.monitorPod).
		Filter(route.RouteLogging).
		Doc("monitor specific pod level metrics by nodeid").
		Param(ws.PathParameter("node_id", "specific node").DataType("string").Required(true).DefaultValue("i-k89a62il")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("pod_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/pods/{pod_name}/containers").To(u.monitorContainer).
		Filter(route.RouteLogging).
		Doc("monitor containers level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("containers_filter", "container re2 expression filter").DataType("string").Required(false).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/pods/{pod_name}/containers/{container_name}").To(u.monitorContainer).
		Filter(route.RouteLogging).
		Doc("monitor specific container level metrics").
		Param(ws.PathParameter("ns_name", "specific namespace").DataType("string").Required(true).DefaultValue("monitoring")).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true).DefaultValue("")).
		Param(ws.PathParameter("container_name", "specific container").DataType("string").Required(true).DefaultValue("")).
		Param(ws.QueryParameter("metrics_name", "metrics name cpu memory...").DataType("string").Required(true).DefaultValue("container_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/workloads/{workload_kind}").To(u.monitorWorkload).
		Filter(route.RouteLogging).
		Doc("monitor specific workload level metrics").
		Param(ws.PathParameter("ns_name", "namespace").DataType("string").Required(true).DefaultValue("kube-system")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false)).
		Param(ws.PathParameter("workload_kind", "workload kind").DataType("string").Required(false).DefaultValue("daemonset")).
		Param(ws.QueryParameter("workload_name", "workload name").DataType("string").Required(true).DefaultValue("")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/namespaces/{ns_name}/workloads").To(u.monitorWorkload).
		Filter(route.RouteLogging).
		Doc("monitor all workload level metrics").
		Param(ws.PathParameter("ns_name", "namespace").DataType("string").Required(true).DefaultValue("kube-system")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/pods").To(u.monitorWorkspacePodLevelMetrics).
		Filter(route.RouteLogging).
		Doc("monitor specific workspace level metrics").
		Param(ws.PathParameter("workspace_name", "workspace name").DataType("string").Required(true)).
		Param(ws.QueryParameter("namespaces_filter", "namespaces filter").DataType("string").Required(false).DefaultValue("k.*")).
		Param(ws.QueryParameter("metrics_filter", "metrics name cpu memory...").DataType("string").Required(false).DefaultValue("tenant_memory_utilisation_wo_cache")).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}").To(u.monitorWorkspaceResourceLevelMetrics).
		Filter(route.RouteLogging).
		Doc("monitor specific workspace level metrics").
		Param(ws.PathParameter("workspace_name", "workspace name").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces").To(u.monitorWorkspaceUserInfo).
		Filter(route.RouteLogging).
		Doc("monitor specific workspace level metrics").
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

}
