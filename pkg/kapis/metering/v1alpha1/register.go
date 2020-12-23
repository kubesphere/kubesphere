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
package v1alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	monitoringv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha3"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

const (
	groupName = "metering.kubesphere.io"
	respOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha1"}

func AddToContainer(c *restful.Container, k8sClient kubernetes.Interface, meteringClient monitoring.Interface, factory informers.InformerFactory, cache cache.Cache) error {
	ws := runtime.NewWebService(GroupVersion)

	h := newHandler(k8sClient, meteringClient, factory, resourcev1alpha3.NewResourceGetter(factory, cache))

	ws.Route(ws.GET("/cluster").
		To(h.HandleClusterMetersQuery).
		Doc("Get cluster-level meter data.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which meter data to return. For example, the following filter matches both cluster CPU usage and disk usage: `meter_cluster_cpu_usage|meter_cluster_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes").
		To(h.HandleNodeMetersQuery).
		Doc("Get node-level meter data of all nodes.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which meter data to return. For example, the following filter matches both node CPU usage and disk usage: `meter_node_cpu_usage|meter_node_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The node filter consists of a regexp pattern. It specifies which node data to return. For example, the following filter matches both node i-caojnter and i-cmu82ogj: `i-caojnter|i-cmu82ogj`.").DataType("string").Required(false)).
		Param(ws.PathParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVCs filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort nodes by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NodeMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}").
		To(h.HandleNodeMetersQuery).
		Doc("Get node-level meter data of the specific node.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("node", "Node name.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which meter data to return. For example, the following filter matches both node CPU usage and disk usage: `meter_node_cpu_usage|meter_node_memory_usage`.").DataType("string").Required(false)).
		Param(ws.PathParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVCs filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NodeMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces").
		To(h.HandleWorkspaceMetersQuery).
		Doc("Get workspace-level meter data of all workspaces.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both workspace CPU usage and memory usage: `meter_workspace_cpu_usage|meter_workspace_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The workspace filter consists of a regexp pattern. It specifies which workspace data to return.").DataType("string").Required(false)).
		Param(ws.PathParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVC filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort workspaces by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/{workspace}").
		To(h.HandleWorkspaceMetersQuery).
		Doc("Get workspace-level meter data of a specific workspace.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("workspace", "Workspace name.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both workspace CPU usage and memory usage: `meter_workspace_cpu_usage|meter_workspace_memory_usage`.").DataType("string").Required(false)).
		Param(ws.PathParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVC filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("type", "Additional operations. Currently available types is statistics. It retrieves the total number of namespaces, devops projects, members and roles in this workspace at the moment.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(h.HandleNamespaceMetersQuery).
		Doc("Get namespace-level meter data of a specific workspace.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("workspace", "Workspace name.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both namespace CPU usage and memory usage: `meter_namespace_cpu_usage|meter_namespace_memory_usage_wo_cache`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The namespace filter consists of a regexp pattern. It specifies which namespace data to return. For example, the following filter matches both namespace test and kube-system: `test|kube-system`.").DataType("string").Required(false)).
		Param(ws.PathParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVC filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort namespaces by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces").
		To(h.HandleNamespaceMetersQuery).
		Doc("Get namespace-level meter data of all namespaces.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both namespace CPU usage and memory usage: `meter_namespace_cpu_usage|meter_namespace_memory_usage_wo_cache`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The namespace filter consists of a regexp pattern. It specifies which namespace data to return. For example, the following filter matches both namespace test and kube-system: `test|kube-system`.").DataType("string").Required(false)).
		Param(ws.PathParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVC filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort namespaces by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}").
		To(h.HandleNamespaceMetersQuery).
		Doc("Get namespace-level meter data of the specific namespace.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("namespace", "The name of the namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both namespace CPU usage and memory usage: `meter_namespace_cpu_usage|meter_namespace_memory_usage_wo_cache`.").DataType("string").Required(false)).
		Param(ws.PathParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVC filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads").
		To(h.HandleWorkloadMetersQuery).
		Doc("Get workload-level meter data of all workloads which belongs to a specific kind.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("namespace", "The name of the namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("kind", "Workload kind. One of deployment, daemonset, statefulset.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both workload CPU usage and memory usage: `meter_workload_cpu_usage|meter_workload_memory_usage_wo_cache`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The workload filter consists of a regexp pattern. It specifies which workload data to return. For example, the following filter matches any workload whose name begins with prometheus: `prometheus.*`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort workloads by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkloadMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/applications").
		To(h.HandleApplicationMetersQuery).
		Doc("Get app-level meter data of a specific application. Navigate to the app by the app's namespace.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("namespace", "The name of the namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("applications", "Appliction names, format app_name[:app_version](such as nginx:v1, nignx) which are joined by \"|\" ").DataType("string").Required(false)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both pod CPU usage and memory usage: `meter_application_cpu_usage|meter_application_memory_usage_wo_cache`.").DataType("string").Required(false)).
		Param(ws.PathParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort pods by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.PodMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods").
		To(h.HandlePodMetersQuery).
		Doc("Get pod-level meter data of the specific namespace's pods.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("namespace", "The name of the namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both pod CPU usage and memory usage: `meter_pod_cpu_usage|meter_pod_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The pod filter consists of a regexp pattern. It specifies which pod data to return. For example, the following filter matches any pod whose name begins with redis: `redis.*`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort pods by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.PodMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}").
		To(h.HandlePodMetersQuery).
		Doc("Get pod-level meter data of a specific pod. Navigate to the pod by the pod's namespace.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("namespace", "The name of the namespace.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Pod name.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both pod CPU usage and memory usage: `meter_pod_cpu_usage|_meter_pod_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.PodMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/{workload}/pods").
		To(h.HandlePodMetersQuery).
		Doc("Get pod-level meter data of a specific workload's pods. Navigate to the workload by the namespace.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("namespace", "The name of the namespace.").DataType("string").Required(true)).
		Param(ws.PathParameter("workload", "Workload name.").DataType("string").Required(true)).
		Param(ws.QueryParameter("kind", "Workload kind. One of deployment, daemonset, statefulset.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both pod CPU usage and memory usage: `meter_pod_cpu_usage|meter_pod_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The pod filter consists of a regexp pattern. It specifies which pod data to return. For example, the following filter matches any pod whose name begins with redis: `redis.*`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort pods by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.PodMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods").
		To(h.HandlePodMetersQuery).
		Doc("Get pod-level meter data of all pods on a specific node.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("node", "Node name.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both pod CPU usage and memory usage: `meter_pod_cpu_usage|meter_pod_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The pod filter consists of a regexp pattern. It specifies which pod data to return. For example, the following filter matches any pod whose name begins with redis: `redis.*`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort pods by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.PodMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/nodes/{node}/pods/{pod}").
		To(h.HandlePodMetersQuery).
		Doc("Get pod-level meter data of a specific pod. Navigate to the pod by the node where it is scheduled.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("node", "Node name.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Pod name.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both pod CPU usage and memory usage: `meter_pod_cpu_usage|meter_pod_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.PodMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/services").
		To(h.HandleServiceMetersQuery).
		Doc("Get service-level meter data of the specific namespace's pods.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("namespace", "The name of the namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both pod CPU usage and memory usage: `meter_pod_cpu_usage|meter_pod_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("services", "Services which are joined by \"|\".").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort pods by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ServiceMetricsTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	c.Add(ws)
	return nil
}
