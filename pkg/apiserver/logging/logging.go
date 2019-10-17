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
	"bytes"
	"fmt"
	"github.com/emicklei/go-restful"
	"io"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/log"
	"kubesphere.io/kubesphere/pkg/server/errors"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	fb "kubesphere.io/kubesphere/pkg/simple/client/fluentbit"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func LoggingQueryCluster(request *restful.Request, response *restful.Response) {
	param := parseRequest(log.QueryLevelCluster, request)
	if param.Operation == v1alpha2.OperationExport {
		logExport(param, request, response)
	} else {
		logQuery(param, response)
	}
}

func LoggingQueryWorkspace(request *restful.Request, response *restful.Response) {
	param := parseRequest(log.QueryLevelWorkspace, request)
	logQuery(param, response)
}

func LoggingQueryNamespace(request *restful.Request, response *restful.Response) {
	param := parseRequest(log.QueryLevelNamespace, request)
	logQuery(param, response)
}

func LoggingQueryWorkload(request *restful.Request, response *restful.Response) {
	param := parseRequest(log.QueryLevelWorkload, request)
	logQuery(param, response)
}

func LoggingQueryPod(request *restful.Request, response *restful.Response) {
	param := parseRequest(log.QueryLevelPod, request)
	logQuery(param, response)
}

func LoggingQueryContainer(request *restful.Request, response *restful.Response) {
	param := parseRequest(log.QueryLevelContainer, request)
	if param.Operation == v1alpha2.OperationExport {
		logExport(param, request, response)
	} else {
		logQuery(param, response)
	}
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
		klog.Errorln(err)
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
		klog.Errorln(err)
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

func logQuery(param v1alpha2.QueryParameters, response *restful.Response) {
	es, err := cs.ClientSets().ElasticSearch()
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusServiceUnavailable, errors.Wrap(err))
		return
	}

	res, err := es.Query(param)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(res)
}

func logExport(param v1alpha2.QueryParameters, request *restful.Request, response *restful.Response) {
	es, err := cs.ClientSets().ElasticSearch()
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusServiceUnavailable, errors.Wrap(err))
		return
	}

	response.Header().Set(restful.HEADER_ContentType, "text/plain")
	response.Header().Set("Content-Disposition", "attachment")

	// keep search context alive for 1m
	param.ScrollTimeout = time.Minute
	// export 1000 records in every iteration
	param.Size = 1000
	// from is not allowed in a scroll context
	param.From = 0

	var scrollId string
	// limit to retrieve max 100k records
	for i := 0; i < 100; i++ {
		var res *v1alpha2.QueryResult
		if scrollId == "" {
			res, err = es.Query(param)
			if err != nil {
				response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
				return
			}
		} else {
			res, err = es.Scroll(scrollId)
			if err != nil {
				break
			}
		}

		if res.Read == nil || len(res.Read.Records) == 0 {
			break
		}
		output := new(bytes.Buffer)
		for _, r := range res.Read.Records {
			output.WriteString(fmt.Sprintf(`%s`, stringutils.StripAnsi(r.Log)))
		}
		_, err = io.Copy(response, output)
		if err != nil {
			klog.Error(err)
			break
		}

		scrollId = res.Read.ScrollID

		select {
		case <-request.Request.Context().Done():
			break
		default:
		}
	}

	if scrollId != "" {
		es.ClearScroll(scrollId)
	}
}

func parseRequest(level log.LogQueryLevel, request *restful.Request) v1alpha2.QueryParameters {
	var param v1alpha2.QueryParameters

	switch level {
	case log.QueryLevelCluster:
		var namespaces []string
		param.NamespaceNotFound, namespaces = log.MatchNamespace(stringutils.Split(request.QueryParameter("namespaces"), ","),
			stringutils.Split(strings.ToLower(request.QueryParameter("namespace_query")), ","),
			stringutils.Split(request.QueryParameter("workspaces"), ","),
			stringutils.Split(strings.ToLower(request.QueryParameter("workspace_query")), ","))
		param.NamespaceWithCreationTime = log.MakeNamespaceCreationTimeMap(namespaces)
		param.WorkloadFilter = stringutils.Split(request.QueryParameter("workloads"), ",")
		param.WorkloadQuery = stringutils.Split(request.QueryParameter("workload_query"), ",")
		param.PodFilter = stringutils.Split(request.QueryParameter("pods"), ",")
		param.PodQuery = stringutils.Split(request.QueryParameter("pod_query"), ",")
		param.ContainerFilter = stringutils.Split(request.QueryParameter("containers"), ",")
		param.ContainerQuery = stringutils.Split(request.QueryParameter("container_query"), ",")
	case log.QueryLevelWorkspace:
		var namespaces []string
		param.NamespaceNotFound, namespaces = log.MatchNamespace(stringutils.Split(request.QueryParameter("namespaces"), ","),
			stringutils.Split(strings.ToLower(request.QueryParameter("namespace_query")), ","),
			stringutils.Split(request.PathParameter("workspace"), ","), nil)
		param.NamespaceWithCreationTime = log.MakeNamespaceCreationTimeMap(namespaces)
		param.WorkloadFilter = stringutils.Split(request.QueryParameter("workloads"), ",")
		param.WorkloadQuery = stringutils.Split(request.QueryParameter("workload_query"), ",")
		param.PodFilter = stringutils.Split(request.QueryParameter("pods"), ",")
		param.PodQuery = stringutils.Split(request.QueryParameter("pod_query"), ",")
		param.ContainerFilter = stringutils.Split(request.QueryParameter("containers"), ",")
		param.ContainerQuery = stringutils.Split(request.QueryParameter("container_query"), ",")
	case log.QueryLevelNamespace:
		namespaces := []string{request.PathParameter("namespace")}
		param.NamespaceWithCreationTime = log.MakeNamespaceCreationTimeMap(namespaces)
		param.WorkloadFilter = stringutils.Split(request.QueryParameter("workloads"), ",")
		param.WorkloadQuery = stringutils.Split(request.QueryParameter("workload_query"), ",")
		param.PodFilter = stringutils.Split(request.QueryParameter("pods"), ",")
		param.PodQuery = stringutils.Split(request.QueryParameter("pod_query"), ",")
		param.ContainerFilter = stringutils.Split(request.QueryParameter("containers"), ",")
		param.ContainerQuery = stringutils.Split(request.QueryParameter("container_query"), ",")
	case log.QueryLevelWorkload:
		namespaces := []string{request.PathParameter("namespace")}
		param.NamespaceWithCreationTime = log.MakeNamespaceCreationTimeMap(namespaces)
		param.WorkloadFilter = []string{request.PathParameter("workload")}
		param.PodFilter = stringutils.Split(request.QueryParameter("pods"), ",")
		param.PodQuery = stringutils.Split(request.QueryParameter("pod_query"), ",")
		param.ContainerFilter = stringutils.Split(request.QueryParameter("containers"), ",")
		param.ContainerQuery = stringutils.Split(request.QueryParameter("container_query"), ",")
	case log.QueryLevelPod:
		namespaces := []string{request.PathParameter("namespace")}
		param.NamespaceWithCreationTime = log.MakeNamespaceCreationTimeMap(namespaces)
		param.PodFilter = []string{request.PathParameter("pod")}
		param.ContainerFilter = stringutils.Split(request.QueryParameter("containers"), ",")
		param.ContainerQuery = stringutils.Split(request.QueryParameter("container_query"), ",")
	case log.QueryLevelContainer:
		namespaces := []string{request.PathParameter("namespace")}
		param.NamespaceWithCreationTime = log.MakeNamespaceCreationTimeMap(namespaces)
		param.PodFilter = []string{request.PathParameter("pod")}
		param.ContainerFilter = []string{request.PathParameter("container")}
	}

	param.LogQuery = stringutils.Split(request.QueryParameter("log_query"), ",")
	param.Interval = request.QueryParameter("interval")
	param.StartTime = request.QueryParameter("start_time")
	param.EndTime = request.QueryParameter("end_time")
	param.Sort = request.QueryParameter("sort")
	switch request.QueryParameter("operation") {
	case "statistics":
		param.Operation = v1alpha2.OperationStatistics
	case "histogram":
		param.Operation = v1alpha2.OperationHistogram
	case "export":
		param.Operation = v1alpha2.OperationExport
	default:
		param.Operation = v1alpha2.OperationQuery
	}

	var err error
	param.From, err = strconv.ParseInt(request.QueryParameter("from"), 10, 64)
	if err != nil {
		param.From = 0
	}

	param.Size, err = strconv.ParseInt(request.QueryParameter("size"), 10, 64)
	if err != nil {
		param.Size = 10
	}

	return param
}
