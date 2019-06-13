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
package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/monitoring"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"net/http"
)

const (
	GroupName = "monitoring.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	tags := []string{"Monitoring"}

	ws.Route(ws.GET("/cluster").To(monitoring.MonitorCluster).
		Doc("Get cluster-level metrics.").
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. cluster_cpu|cluster_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes").To(monitoring.MonitorAllNodes).
		Doc("Get all nodes' metrics.").
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. node_cpu|node_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Node filter in regexp pattern, eg. i-caojnter|i-cmu82ogj.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort nodes by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}").To(monitoring.MonitorSpecificNode).
		Doc("Get specific node metrics.").
		Param(ws.PathParameter("node", "Specify the target node.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. node_cpu|node_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces").To(monitoring.MonitorAllNamespaces).
		Doc("Get namespace-level metrics.").
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. namespace_cpu|namespace_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Namespace filter in regexp pattern, eg. namespace-1|namespace-2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort namespaces by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}").To(monitoring.MonitorSpecificNamespace).
		Doc("Get specific namespace metrics.").
		Param(ws.PathParameter("namespace", "Specify the target namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. namespace_cpu|namespace_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods").To(monitoring.MonitorAllPodsOfSpecificNamespace).
		Doc("Get all pod-level metrics of a given namespace.").
		Param(ws.PathParameter("namespace", "Specify the target namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. pod_cpu|pod_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Pods filter in regexp pattern, eg. coredns-77b8449dc9-hd6gd|coredns-77b8449dc9-b4n74.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort pods by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}").To(monitoring.MonitorSpecificPodOfSpecificNamespace).
		Doc("Get specific pod metrics.").
		Param(ws.PathParameter("namespace", "Specify the target namespace.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Specify the target pod.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. pod_cpu|pod_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods").To(monitoring.MonitorAllPodsOnSpecificNode).
		Doc("Get metrics of all pods on a specific node.").
		Param(ws.PathParameter("node", "Specify the target node.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. node_cpu|node_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Pod filter in regexp pattern, eg. coredns-77b8449dc9-hd6gd|coredns-77b8449dc9-b4n74.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort pods by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}").To(monitoring.MonitorSpecificPodOnSpecificNode).
		Doc("Get specific pod metrics on a specified node.").
		Param(ws.PathParameter("node", "Specify the target node.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Specify the target pod.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. pod_cpu|pod_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}/containers").To(monitoring.MonitorAllContainersOnSpecificNode).
		Doc("Get container-level metrics under a given node and pod.").
		Param(ws.PathParameter("node", "Specify the target node.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Specify the target pod.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. container_cpu|container_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Container filter in regexp pattern.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort containers by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers").To(monitoring.MonitorAllContainersOfSpecificNamespace).
		Doc("Get all container-level metrics of a given pod.").
		Param(ws.PathParameter("namespace", "Specify the target namespace.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Specify the target pod.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. container_cpu|container_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Container filter in regexp pattern.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort containers by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/{container}").To(monitoring.MonitorSpecificContainerOfSpecificNamespace).
		Doc("Get specific container metrics of a given pod.").
		Param(ws.PathParameter("namespace", "Specify the target namespace.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Specify the target pod.").DataType("string").Required(true)).
		Param(ws.PathParameter("container", "Specify the target container.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. container_cpu|container_memory.").DataType("string").Required(false)).Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	// Only use this api to monitor status of pods under the {workload}
	// To monitor a specific workload, try the next two apis with "resources_filter"
	ws.Route(ws.GET("/namespaces/{namespace}/workloads/{workload_kind}/{workload}").To(monitoring.MonitorSpecificWorkload).
		Doc("Get specific workload metrics under a given namespace.").
		Param(ws.PathParameter("namespace", "Specify the target namespace.").DataType("string").Required(true)).
		Param(ws.PathParameter("workload_kind", "Specify the target workload kind. One of deployment, daemonset, statefulset. Other values will be interpreted as any of three.").DataType("string").Required(true).DefaultValue("(.*)")).
		Param(ws.PathParameter("workload", "Specify the target workload.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. workload_cpu|workload_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/{workload_kind}").To(monitoring.MonitorAllWorkloadsOfSpecificKind).
		Doc("Get all workload-level metrics of a specific workload kind under a given namespace.").
		Param(ws.PathParameter("namespace", "Specify the target namespace.").DataType("string").Required(true)).
		Param(ws.PathParameter("workload_kind", "Specify the target workload kind. One of deployment, daemonset, statefulset. Other values will be interpreted as any of three.").DataType("string").Required(true).DefaultValue("(.*)")).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. node_cpu|node_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Workload filter in regexp pattern.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort workloads by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads").To(monitoring.MonitorAllWorkloadsOfSpecificNamespace).
		Doc("Get all workload-level metrics of a given namespace.").
		Param(ws.PathParameter("namespace", "Specify the target namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. workload_cpu|workload_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Workload filter in regexp pattern.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort workloads by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	// list all namespace in this workspace by selected metrics
	ws.Route(ws.GET("/workspaces/{workspace}").To(monitoring.MonitorSpecificWorkspace).
		Doc("Get specific workspace metrics.").
		Param(ws.PathParameter("workspace", "Specify the target workspace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. workspace_cpu|workspace_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces").To(monitoring.MonitorAllWorkspaces).
		Doc("Get workspace-level metrics.").
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. workspace_cpu|workspace_memory.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "Workspace filter in regexp pattern, eg. workspace_1|workspace_2.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort workspaces by the specified metric. Valid only if type is rank.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sorting order, one of asc, desc. Valid only if type is rank.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The number of paged results per metric. Default to return the whole metrics.").DataType("integer").Required(false).DefaultValue("1")).
		Param(ws.QueryParameter("limit", "Max count of items per page.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("type", "Additional operation to the result. One of rank, statistic.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/components/{component}").To(monitoring.MonitorComponent).
		Doc("Get service component-level metrics.").
		Param(ws.PathParameter("component", "Specify the target component. One of etcd, apiserver, scheduler, controller_manager, coredns, prometheus.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "Metrics filter in regexp pattern, eg. etcd_server_list|coredns_proxy.").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Used to get metrics over a range of time. Query resolution step. eg. 10m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Used to get metrics over a range of time. Start time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "Used to get metrics over a range of time. End time of query range. eg. 1559762729.").DataType("string").Required(false)).
		Param(ws.QueryParameter("time", "Used to get metrics at a given time point. eg. 1559762729.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(metrics.FormatedLevelMetric{}).
		Returns(http.StatusOK, RespOK, metrics.FormatedLevelMetric{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	c.Add(ws)
	return nil
}
