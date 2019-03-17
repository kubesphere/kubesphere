package logging

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/log"
)

func LoggingQueryCluster(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelCluster, request)
	response.WriteAsJson(res)
}

func LoggingQueryWorkspace(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelWorkspace, request)
	response.WriteAsJson(res)
}

func LoggingQueryNamespace(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelNamespace, request)
	response.WriteAsJson(res)
}

func LoggingQueryWorkload(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelWorkload, request)
	response.WriteAsJson(res)
}

func LoggingQueryPod(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelPod, request)
	response.WriteAsJson(res)
}

func LoggingQueryContainer(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(constants.QueryLevelContainer, request)
	response.WriteAsJson(res)
}

func LoggingQueryFluentbitFilters(request *restful.Request, response *restful.Response) {
	res := log.FluentbitFiltersQuery(request)
	response.WriteAsJson(res)
}

func LoggingUpdateFluentbitFilters(request *restful.Request, response *restful.Response) {
	res := log.FluentbitFiltersUpdate(request)
	response.WriteAsJson(res)
}

func LoggingQueryFluentbitOutputs(request *restful.Request, response *restful.Response) {
	res := log.FluentbitOutputsQuery(request)
	response.WriteAsJson(res)
}

func LoggingInsertFluentbitOutput(request *restful.Request, response *restful.Response) {
	res := log.FluentbitOutputInsert(request)
	response.WriteAsJson(res)
}

func LoggingUpdateFluentbitOutput(request *restful.Request, response *restful.Response) {
	res := log.FluentbitOutputUpdate(request)
	response.WriteAsJson(res)
}

func LoggingDeleteFluentbitOutput(request *restful.Request, response *restful.Response) {
	res := log.FluentbitOutputDelete(request)
	response.WriteAsJson(res)
}
