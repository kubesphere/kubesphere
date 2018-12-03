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

	//"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/log"
)

func (u LoggingResource) loggingQueryCluster(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.Cluster, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryWorkspace(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.Workspace, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryProject(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.Project, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryWorkload(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.Workload, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryPod(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.Pod, request)
	response.WriteAsJson(res)
}

func (u LoggingResource) loggingQueryContainer(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.Container, request)
	response.WriteAsJson(res)
}

type LoggingResource struct {
}

func Register(ws *restful.WebService, subPath string) {
	tags := []string{"logging apis"}
	u := LoggingResource{}

	ws.Route(ws.GET(subPath+"/clusters").To(u.loggingQueryCluster).
		Filter(route.RouteLogging).
		Doc("cluster level log query").
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "range end time").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}").To(u.loggingQueryWorkspace).
		Filter(route.RouteLogging).
		Doc("workspace level log query").
		Param(ws.PathParameter("workspace_name", "specific workspace").DataType("string").Required(true)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "range end time").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/projects/{project_name}").To(u.loggingQueryProject).
		Filter(route.RouteLogging).
		Doc("project level log query").
		Param(ws.PathParameter("workspace_name", "specific workspace").DataType("string").Required(true)).
		Param(ws.PathParameter("project_name", "specific project").DataType("string").Required(true)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "range end time").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/projects/{project_name}/workloads/{workload_name}").To(u.loggingQueryWorkload).
		Filter(route.RouteLogging).
		Doc("workload level log query").
		Param(ws.PathParameter("workspace_name", "specific workspace").DataType("string").Required(true)).
		Param(ws.PathParameter("project_name", "specific project").DataType("string").Required(true)).
		Param(ws.PathParameter("workload_name", "specific workload").DataType("string").Required(true)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "range end time").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/projects/{project_name}/workloads/{workload_name}/pods/{pod_name}").To(u.loggingQueryPod).
		Filter(route.RouteLogging).
		Doc("pod level log query").
		Param(ws.PathParameter("workspace_name", "specific workspace").DataType("string").Required(true)).
		Param(ws.PathParameter("project_name", "specific project").DataType("string").Required(true)).
		Param(ws.PathParameter("workload_name", "specific workload").DataType("string").Required(true)).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "range end time").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/workspaces/{workspace_name}/projects/{project_name}/workloads/{workload_name}/pods/{pod_name}/containers/{container_name}").To(u.loggingQueryContainer).
		Filter(route.RouteLogging).
		Doc("container level log query").
		Param(ws.PathParameter("workspace_name", "specific workspace").DataType("string").Required(true)).
		Param(ws.PathParameter("project_name", "specific project").DataType("string").Required(true)).
		Param(ws.PathParameter("workload_name", "specific workload").DataType("string").Required(true)).
		Param(ws.PathParameter("pod_name", "specific pod").DataType("string").Required(true)).
		Param(ws.PathParameter("container_name", "specific container").DataType("string").Required(true)).
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "range end time").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}
