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

package logging

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/log"
	es "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch"
	fb "kubesphere.io/kubesphere/pkg/simple/client/fluentbit"
	"net/http"
	"strconv"
)

func LoggingQueryCluster(request *restful.Request, response *restful.Response) {
	res := logQuery(log.QueryLevelCluster, request)

	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}

	response.WriteAsJson(res)
}

func LoggingQueryWorkspace(request *restful.Request, response *restful.Response) {
	res := logQuery(log.QueryLevelWorkspace, request)

	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}

	response.WriteAsJson(res)
}

func LoggingQueryNamespace(request *restful.Request, response *restful.Response) {
	res := logQuery(log.QueryLevelNamespace, request)

	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}

	response.WriteAsJson(res)
}

func LoggingQueryWorkload(request *restful.Request, response *restful.Response) {
	res := logQuery(log.QueryLevelWorkload, request)

	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}

	response.WriteAsJson(res)
}

func LoggingQueryPod(request *restful.Request, response *restful.Response) {
	res := logQuery(log.QueryLevelPod, request)
	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}
	response.WriteAsJson(res)
}

func LoggingQueryContainer(request *restful.Request, response *restful.Response) {
	res := logQuery(log.QueryLevelContainer, request)
	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}
	response.WriteAsJson(res)
}

func LoggingQueryFluentbitFilters(request *restful.Request, response *restful.Response) {
	res := log.FluentbitFiltersQuery()
	response.WriteAsJson(res)
}

func LoggingUpdateFluentbitFilters(request *restful.Request, response *restful.Response) {

	var res *log.FluentbitFiltersResult

	filters := new([]log.FluentbitFilter)

	err := request.ReadEntity(&filters)
	if err != nil {
		res = &log.FluentbitFiltersResult{Status: http.StatusBadRequest}
	} else {
		res = log.FluentbitFiltersUpdate(filters)
	}

	response.WriteAsJson(res)
}

func LoggingQueryFluentbitOutputs(request *restful.Request, response *restful.Response) {
	res := log.FluentbitOutputsQuery()
	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}
	response.WriteAsJson(res)
}

func LoggingInsertFluentbitOutput(request *restful.Request, response *restful.Response) {

	var output fb.OutputPlugin
	var res *log.FluentbitOutputsResult

	err := request.ReadEntity(&output)
	if err != nil {
		glog.Errorln(err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	res = log.FluentbitOutputInsert(output)
	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}

	response.WriteAsJson(res)
}

func LoggingUpdateFluentbitOutput(request *restful.Request, response *restful.Response) {

	var output fb.OutputPlugin

	id := request.PathParameter("output")

	err := request.ReadEntity(&output)
	if err != nil {
		glog.Errorln(err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	res := log.FluentbitOutputUpdate(output, id)

	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}

	response.WriteAsJson(res)
}

func LoggingDeleteFluentbitOutput(request *restful.Request, response *restful.Response) {

	var res *log.FluentbitOutputsResult

	id := request.PathParameter("output")
	res = log.FluentbitOutputDelete(id)

	if res.Status != http.StatusOK {
		response.WriteHeaderAndEntity(res.Status, errors.New(res.Error))
		return
	}

	response.WriteAsJson(res)
}

func logQuery(level log.LogQueryLevel, request *restful.Request) *es.QueryResult {
	var param es.QueryParameters

	param.Operation = request.QueryParameter("operation")

	switch level {
	case log.QueryLevelCluster:
		{
			param.NamespaceFilled, param.Namespaces = log.QueryWorkspace(request.QueryParameter("workspaces"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = log.MatchNamespace(request.QueryParameter("namespaces"), param.NamespaceFilled, param.Namespaces)
			param.NamespaceFilled, param.NamespaceWithCreationTime = log.GetNamespaceCreationTimeMap(param.Namespaces)
			param.NamespaceQuery = request.QueryParameter("namespace_query")
			param.PodFilled, param.Pods = log.QueryWorkload(request.QueryParameter("workloads"), request.QueryParameter("workload_query"), param.Namespaces)
			param.PodFilled, param.Pods = log.MatchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = log.MatchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case log.QueryLevelWorkspace:
		{
			param.NamespaceFilled, param.Namespaces = log.QueryWorkspace(request.PathParameter("workspace"), "")
			param.NamespaceFilled, param.Namespaces = log.MatchNamespace(request.QueryParameter("namespaces"), param.NamespaceFilled, param.Namespaces)
			param.NamespaceFilled, param.NamespaceWithCreationTime = log.GetNamespaceCreationTimeMap(param.Namespaces)
			param.NamespaceQuery = request.QueryParameter("namespace_query")
			param.PodFilled, param.Pods = log.QueryWorkload(request.QueryParameter("workloads"), request.QueryParameter("workload_query"), param.Namespaces)
			param.PodFilled, param.Pods = log.MatchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = log.MatchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case log.QueryLevelNamespace:
		{
			param.NamespaceFilled, param.Namespaces = log.MatchNamespace(request.PathParameter("namespace"), false, nil)
			param.NamespaceFilled, param.NamespaceWithCreationTime = log.GetNamespaceCreationTimeMap(param.Namespaces)
			param.PodFilled, param.Pods = log.QueryWorkload(request.QueryParameter("workloads"), request.QueryParameter("workload_query"), param.Namespaces)
			param.PodFilled, param.Pods = log.MatchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = log.MatchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case log.QueryLevelWorkload:
		{
			param.NamespaceFilled, param.Namespaces = log.MatchNamespace(request.PathParameter("namespace"), false, nil)
			param.NamespaceFilled, param.NamespaceWithCreationTime = log.GetNamespaceCreationTimeMap(param.Namespaces)
			param.PodFilled, param.Pods = log.QueryWorkload(request.PathParameter("workload"), "", param.Namespaces)
			param.PodFilled, param.Pods = log.MatchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = log.MatchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case log.QueryLevelPod:
		{
			param.NamespaceFilled, param.Namespaces = log.MatchNamespace(request.PathParameter("namespace"), false, nil)
			param.NamespaceFilled, param.NamespaceWithCreationTime = log.GetNamespaceCreationTimeMap(param.Namespaces)
			param.PodFilled, param.Pods = log.MatchPod(request.PathParameter("pod"), false, nil)
			param.ContainerFilled, param.Containers = log.MatchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case log.QueryLevelContainer:
		{
			param.NamespaceFilled, param.Namespaces = log.MatchNamespace(request.PathParameter("namespace"), false, nil)
			param.NamespaceFilled, param.NamespaceWithCreationTime = log.GetNamespaceCreationTimeMap(param.Namespaces)
			param.PodFilled, param.Pods = log.MatchPod(request.PathParameter("pod"), false, nil)
			param.ContainerFilled, param.Containers = log.MatchContainer(request.PathParameter("container"))
		}
	}

	if len(param.Namespaces) == 1 {
		param.Workspace = log.GetWorkspaceOfNamesapce(param.Namespaces[0])
	}

	param.Interval = request.QueryParameter("interval")

	param.LogQuery = log.MatchLog(request.QueryParameter("log_query"))
	param.StartTime = request.QueryParameter("start_time")
	param.EndTime = request.QueryParameter("end_time")
	param.Sort = request.QueryParameter("sort")

	var err error
	param.From, err = strconv.ParseInt(request.QueryParameter("from"), 10, 64)
	if err != nil {
		param.From = 0
	}
	param.Size, err = strconv.ParseInt(request.QueryParameter("size"), 10, 64)
	if err != nil {
		param.Size = 10
	}

	return es.Query(param)
}
