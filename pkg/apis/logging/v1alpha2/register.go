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
	"kubesphere.io/kubesphere/pkg/apiserver/logging"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/filter"
	"kubesphere.io/kubesphere/pkg/models/log"
	esclient "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch"
	fluentbitclient "kubesphere.io/kubesphere/pkg/simple/client/fluentbit"
	"net/http"
)

const (
	GroupName = "logging.kubesphere.io"
	RespOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)
	tags := []string{"Logging"}

	ws.Route(ws.GET("/cluster").To(logging.LoggingQueryCluster).
		Filter(filter.Logging).
		Doc("Log query against the cluster.").
		Param(ws.QueryParameter("operation", "Query operation type. One of query, statistics, histogram.").DataType("string").Required(true)).
		Param(ws.QueryParameter("workspaces", "List of workspaces the query will perform against, eg. wk-one,wk-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("workspace_query", "List of keywords for filtering workspaces. Workspaces whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespaces", "List of namespaces the query will perform against, eg. ns-one,ns-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespace_query", "List of keywords for filtering namespaces. Namespaces whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("workloads", "List of workloads the query will perform against, eg. wl-one,wl-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("workload_query", "List of keywords for filtering workloads. Workloads whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pods", "List of pods the query will perform against, eg. pod-one,pod-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "List of keywords for filtering pods. Pods whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "List of containers the query will perform against, eg. container-one,container-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "List of keywords for filtering containers. Containers whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "List of keywords  for filtering logs. The query returns log containing at least one keyword. Non case-sensitive matching. eg. err,INFO.").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "Count logs at intervals. Valid only if operation is histogram. The unit can be ms(milliseconds), s(seconds), m(minutes), h(hours), d(days), w(weeks), M(months), q(quarters), y(years). eg. 30m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "Start time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "End time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "Sort log by time. One of acs, desc.").DataType("string").DefaultValue("desc").Required(false)).
		Param(ws.QueryParameter("from", "Beginning index of result to return. Use this option together with size.").DataType("int").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result to return.").DataType("int").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(esclient.Response{}).
		Returns(http.StatusOK, RespOK, esclient.Response{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/{workspace}").To(logging.LoggingQueryWorkspace).
		Filter(filter.Logging).
		Doc("Log query against a specific workspace.").
		Param(ws.PathParameter("workspace", "Perform query against a specific workspace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "Query operation type. One of query, statistics, histogram.").DataType("string").Required(true)).
		Param(ws.QueryParameter("namespaces", "List of namespaces the query will perform against, eg. ns-one,ns-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespace_query", "List of keywords for filtering namespaces. Namespaces whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("workloads", "List of workloads the query will perform against, eg. wl-one,wl-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("workload_query", "List of keywords for filtering workloads. Workloads whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pods", "List of pods the query will perform against, eg. pod-one,pod-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "List of keywords for filtering pods. Pods whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "List of containers the query will perform against, eg. container-one,container-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "List of keywords for filtering containers. Containers whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "List of keywords  for filtering logs. The query returns log containing at least one keyword. Non case-sensitive matching. eg. err,INFO.").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "Count logs at intervals. Valid only if operation is histogram. The unit can be ms(milliseconds), s(seconds), m(minutes), h(hours), d(days), w(weeks), M(months), q(quarters), y(years). eg. 30m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "Start time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "End time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "Sort log by time. One of acs, desc.").DataType("string").DefaultValue("desc").Required(false)).
		Param(ws.QueryParameter("from", "Beginning index of result to return. Use this option together with size.").DataType("int").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result to return.").DataType("int").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(esclient.Response{}).
		Returns(http.StatusOK, RespOK, esclient.Response{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}").To(logging.LoggingQueryNamespace).
		Filter(filter.Logging).
		Doc("Log query against a specific namespace.").
		Param(ws.PathParameter("namespace", "Perform query against a specific namespace.").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "Query operation type. One of query, statistics, histogram.").DataType("string").Required(true)).
		Param(ws.QueryParameter("workloads", "List of workloads the query will perform against, eg. wl-one,wl-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("workload_query", "List of keywords for filtering workloads. Workloads whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pods", "List of pods the query will perform against, eg. pod-one,pod-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "List of keywords for filtering pods. Pods whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "List of containers the query will perform against, eg. container-one,container-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "List of keywords for filtering containers. Containers whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "List of keywords  for filtering logs. The query returns log containing at least one keyword. Non case-sensitive matching. eg. err,INFO.").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "Count logs at intervals. Valid only if operation is histogram. The unit can be ms(milliseconds), s(seconds), m(minutes), h(hours), d(days), w(weeks), M(months), q(quarters), y(years). eg. 30m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "Start time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "End time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "Sort log by time. One of acs, desc.").DataType("string").DefaultValue("desc").Required(false)).
		Param(ws.QueryParameter("from", "Beginning index of result to return. Use this option together with size.").DataType("int").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result to return.").DataType("int").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(esclient.Response{}).
		Returns(http.StatusOK, RespOK, esclient.Response{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/{workload}").To(logging.LoggingQueryWorkload).
		Filter(filter.Logging).
		Doc("Log query against a specific workload.").
		Param(ws.PathParameter("namespace", "Specify the namespace of the workload.").DataType("string").Required(true)).
		Param(ws.PathParameter("workload", "Perform query against a specific workload.").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "Query operation type. One of query, statistics, histogram.").DataType("string").Required(true)).
		Param(ws.QueryParameter("pods", "List of pods the query will perform against, eg. pod-one,pod-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "List of keywords for filtering pods. Pods whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "List of containers the query will perform against, eg. container-one,container-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "List of keywords for filtering containers. Containers whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "List of keywords  for filtering logs. The query returns log containing at least one keyword. Non case-sensitive matching. eg. err,INFO.").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "Count logs at intervals. Valid only if operation is histogram. The unit can be ms(milliseconds), s(seconds), m(minutes), h(hours), d(days), w(weeks), M(months), q(quarters), y(years). eg. 30m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "Start time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "End time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "Sort log by time. One of acs, desc.").DataType("string").DefaultValue("desc").Required(false)).
		Param(ws.QueryParameter("from", "Beginning index of result to return. Use this option together with size.").DataType("int").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result to return.").DataType("int").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(esclient.Response{}).
		Returns(http.StatusOK, RespOK, esclient.Response{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}").To(logging.LoggingQueryPod).
		Filter(filter.Logging).
		Doc("Log query against a specific pod.").
		Param(ws.PathParameter("namespace", "Specify the namespace of the pod.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Perform query against a specific pod.").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "Query operation type. One of query, statistics, histogram.").DataType("string").Required(true)).
		Param(ws.QueryParameter("containers", "List of containers the query will perform against, eg. container-one,container-two").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "List of keywords for filtering containers. Containers whose name contains at least one keyword will be matched for query. Non case-sensitive matching. eg. one,two.").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "List of keywords  for filtering logs. The query returns log containing at least one keyword. Non case-sensitive matching. eg. err,INFO.").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "Count logs at intervals. Valid only if operation is histogram. The unit can be ms(milliseconds), s(seconds), m(minutes), h(hours), d(days), w(weeks), M(months), q(quarters), y(years). eg. 30m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "Start time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "End time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "Sort log by time. One of acs, desc.").DataType("string").DefaultValue("desc").Required(false)).
		Param(ws.QueryParameter("from", "Beginning index of result to return. Use this option together with size.").DataType("int").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result to return.").DataType("int").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(esclient.Response{}).
		Returns(http.StatusOK, RespOK, esclient.Response{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/{container}").To(logging.LoggingQueryContainer).
		Filter(filter.Logging).
		Doc("Log query against a specific container.").
		Param(ws.PathParameter("namespace", "Specify the namespace of the pod.").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "Specify the pod of the container.").DataType("string").Required(true)).
		Param(ws.PathParameter("container", "Perform query against a specific container.").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "Query operation type. One of query, statistics, histogram.").DataType("string").Required(true)).
		Param(ws.QueryParameter("log_query", "List of keywords  for filtering logs. The query returns log containing at least one keyword. Non case-sensitive matching. eg. err,INFO.").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "Count logs at intervals. Valid only if operation is histogram. The unit can be ms(milliseconds), s(seconds), m(minutes), h(hours), d(days), w(weeks), M(months), q(quarters), y(years). eg. 30m.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "Start time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "End time of query range, eg. 1559664000000.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "Sort log by time. One of acs, desc.").DataType("string").DefaultValue("desc").Required(false)).
		Param(ws.QueryParameter("from", "Beginning index of result to return. Use this option together with size.").DataType("int").DefaultValue("0").Required(false)).
		Param(ws.QueryParameter("size", "Size of result to return.").DataType("int").DefaultValue("10").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(esclient.Response{}).
		Returns(http.StatusOK, RespOK, esclient.Response{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/fluentbit/filters").To(logging.LoggingQueryFluentbitFilters).
		Filter(filter.Logging).
		Doc("List all Fluent bit filter plugins. This API is work-in-process.").
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/fluentbit/filters").To(logging.LoggingUpdateFluentbitFilters).
		Filter(filter.Logging).
		Doc("Add a new Fluent bit filter plugin. This API is work-in-process.").
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/fluentbit/outputs").To(logging.LoggingQueryFluentbitOutputs).
		Filter(filter.Logging).
		Doc("List all Fluent bit output plugins.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(log.FluentbitOutputsResult{}).
		Returns(http.StatusOK, RespOK, log.FluentbitOutputsResult{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/fluentbit/outputs").To(logging.LoggingInsertFluentbitOutput).
		Filter(filter.Logging).
		Doc("Add a new Fluent bit output plugin.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(fluentbitclient.OutputPlugin{}).
		Writes(log.FluentbitOutputsResult{}).
		Returns(http.StatusOK, RespOK, log.FluentbitOutputsResult{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/fluentbit/outputs/{output}").To(logging.LoggingUpdateFluentbitOutput).
		Filter(filter.Logging).
		Doc("Update a Fluent bit output plugin.").
		Param(ws.PathParameter("output", "ID of the output to update.").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(fluentbitclient.OutputPlugin{}).
		Writes(log.FluentbitOutputsResult{}).
		Returns(http.StatusOK, RespOK, log.FluentbitOutputsResult{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/fluentbit/outputs/{output}").To(logging.LoggingDeleteFluentbitOutput).
		Filter(filter.Logging).
		Doc("Delete a Fluent bit output plugin.").
		Param(ws.PathParameter("output", "ID of the output to delete.").DataType("string").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(log.FluentbitOutputsResult{}).
		Returns(http.StatusOK, RespOK, log.FluentbitOutputsResult{})).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	c.Add(ws)
	return nil
}
