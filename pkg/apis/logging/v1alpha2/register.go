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
)

const GroupName = "logging.kubesphere.io"

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
		Doc("cluster level log query").
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("workspaces", "workspaces specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("workspace_query", "workspace query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespaces", "namespaces specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespace_query", "namespace query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("workloads", "workloads specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("workload_query", "workload query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("pods", "pods specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "pod query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "containers specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "container query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "sort method").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/workspaces/{workspace}").To(logging.LoggingQueryWorkspace).
		Filter(filter.Logging).
		Doc("workspace level log query").
		Param(ws.PathParameter("workspace", "workspace specify").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("namespaces", "namespaces specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespace_query", "namespace query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("workloads", "workloads specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("workload_query", "workload query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("pods", "pods specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "pod query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "containers specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "container query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "sort method").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}").To(logging.LoggingQueryNamespace).
		Filter(filter.Logging).
		Doc("namespace level log query").
		Param(ws.PathParameter("namespace", "namespace specify").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("workloads", "workloads specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("workload_query", "workload query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("pods", "pods specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "pod query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "containers specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "container query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "sort method").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/workloads/{workload}").To(logging.LoggingQueryWorkload).
		Filter(filter.Logging).
		Doc("workload level log query").
		Param(ws.PathParameter("namespace", "namespace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("workload", "workload specify").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("pods", "pods specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "pod query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "containers specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "container query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "sort method").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}").To(logging.LoggingQueryPod).
		Filter(filter.Logging).
		Doc("pod level log query").
		Param(ws.PathParameter("namespace", "namespace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "pod specify").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("containers", "containers specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "container query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "sort method").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/{container}").To(logging.LoggingQueryContainer).
		Filter(filter.Logging).
		Doc("container level log query").
		Param(ws.PathParameter("namespace", "namespace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("pod", "pod specify").DataType("string").Required(true)).
		Param(ws.PathParameter("container", "container specify").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort", "sort method").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/fluentbit/filters").To(logging.LoggingQueryFluentbitFilters).
		Filter(filter.Logging).
		Doc("log fluent-bit filters query").
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/fluentbit/filters").To(logging.LoggingUpdateFluentbitFilters).
		Filter(filter.Logging).
		Doc("log fluent-bit filters update").
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/fluentbit/outputs").To(logging.LoggingQueryFluentbitOutputs).
		Filter(filter.Logging).
		Doc("log fluent-bit outputs query").
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/fluentbit/outputs").To(logging.LoggingInsertFluentbitOutput).
		Filter(filter.Logging).
		Doc("log fluent-bit outputs insert").
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/fluentbit/outputs/{output}").To(logging.LoggingUpdateFluentbitOutput).
		Filter(filter.Logging).
		Doc("log fluent-bit outputs update").
		Param(ws.PathParameter("output", "output id").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.DELETE("/fluentbit/outputs/{output}").To(logging.LoggingDeleteFluentbitOutput).
		Filter(filter.Logging).
		Doc("log fluent-bit outputs delete").
		Param(ws.PathParameter("output", "output id").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	c.Add(ws)
	return nil
}
