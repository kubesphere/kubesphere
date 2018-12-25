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
package logging

import (
	//"strings"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/log"
)

func (u LoggingResource) loggingQueryCluster(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelCluster, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryWorkspace(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelWorkspace, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryNamespace(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelNamespace, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryWorkload(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelWorkload, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryPod(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelPod, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryContainer(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelContainer, request)
	response.WriteAsJson(res)
}

type LoggingResource struct {
}

func Register(ws *restful.WebService, subPath string) {
	tags := []string{"logging apis"}
	u := LoggingResource{}

	ws.Route(ws.GET(subPath+"/v1alpha1/logging").To(u.loggingQueryCluster).
		Filter(route.RouteLogging).
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
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/v1alpha1/logging").To(u.loggingQueryWorkspace).
		Filter(route.RouteLogging).
		Doc("workspace level log query").
		Param(ws.PathParameter("workspace_name", "workspace specify").DataType("string").Required(true)).
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
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/namespaces/{namespace_name}/v1alpha1/logging").To(u.loggingQueryNamespace).
		Filter(route.RouteLogging).
		Doc("namespace level log query").
		Param(ws.PathParameter("workspace_name", "workspace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("namespace_name", "namespace specify").DataType("string").Required(true)).
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
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/namespaces/{namespace_name}/workloads/{workload_name}/v1alpha1/logging").To(u.loggingQueryWorkload).
		Filter(route.RouteLogging).
		Doc("workload level log query").
		Param(ws.PathParameter("workspace_name", "workspace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("namespace_name", "namespace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("workload_name", "workload specify").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("pods", "pods specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod_query", "pod query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("containers", "containers specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "container query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/namespaces/{namespace_name}/workloads/{workload_name}/pods/{pod_name}/v1alpha1/logging").To(u.loggingQueryPod).
		Filter(route.RouteLogging).
		Doc("pod level log query").
		Param(ws.PathParameter("workspace_name", "workspace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("namespace_name", "namespace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("workload_name", "workload specify").DataType("string").Required(true)).
		Param(ws.PathParameter("pod_name", "pod specify").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("containers", "containers specify").DataType("string").Required(false)).
		Param(ws.QueryParameter("container_query", "container query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/namespaces/{namespace_name}/workloads/{workload_name}/pods/{pod_name}/containers/{container_name}/v1alpha1/logging").To(u.loggingQueryContainer).
		Filter(route.RouteLogging).
		Doc("container level log query").
		Param(ws.PathParameter("workspace_name", "workspace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("namespace_name", "namespace specify").DataType("string").Required(true)).
		Param(ws.PathParameter("workload_name", "workload specify").DataType("string").Required(true)).
		Param(ws.PathParameter("pod_name", "pod specify").DataType("string").Required(true)).
		Param(ws.PathParameter("container_name", "container specify").DataType("string").Required(true)).
		Param(ws.QueryParameter("operation", "operation: query statistics").DataType("string").Required(true)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("interval", "interval of time histogram").DataType("string").Required(false)).
		Param(ws.QueryParameter("start_time", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end_time", "range end time").DataType("string").Required(false)).
		Param(ws.QueryParameter("from", "begin index of result returned").DataType("int").Required(true)).
		Param(ws.QueryParameter("size", "size of result returned").DataType("int").Required(true)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}
