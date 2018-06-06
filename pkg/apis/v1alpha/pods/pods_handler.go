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

package pods

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/filter/route"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService) {

	ws.Route(ws.GET("/pods").To(handleAllPods).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/namespaces/{namespace}/pods/{podname}").To(handlePodUnderNameSpace).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/namespaces/{namespace}/pods").To(handlePodsUnderNameSpace).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/nodes/{nodename}/pods").To(handlePodsUnderNode).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/nodes/{nodename}/namespaces/{namespace}/pods").To(handlePodsUnderNodeAndNameSpace).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

func handleAllPods(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse

	namespaces := models.GetNameSpaces()

	var total_count int
	for i, namespace := range namespaces {
		result = models.FormatPodsMetrics("", namespace)
		result.Items = append(result.Items, result)
		total_count = i
	}

	result.TotalCount = total_count

	response.WriteAsJson(result)
}

func handlePodsUnderNameSpace(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse

	result = models.FormatPodsMetrics("", request.PathParameter("namespace"))

	response.WriteAsJson(result)
}
func handlePodsUnderNode(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse
	var resultNameSpace constants.PageableResponse
	namespaces := models.GetNameSpaces()

	var total_count int
	for _, namespace := range namespaces {
		resultNameSpace = models.FormatPodsMetrics(request.PathParameter("nodename"), namespace)

		var sub_total_count int
		for j, pod := range resultNameSpace.Items {
			result.Items = append(result.Items, pod)
			sub_total_count = j
		}
		total_count += sub_total_count
	}
	result.TotalCount = total_count
	response.WriteAsJson(result)
}
func handlePodUnderNameSpace(request *restful.Request, response *restful.Response) {
	var resultPod models.ResultPod

	resultPod = models.FormatPodMetrics(request.PathParameter("namespace"), request.PathParameter("podname"))

	response.WriteAsJson(resultPod)
}

func handlePodsUnderNodeAndNameSpace(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse

	result = models.FormatPodsMetrics(request.PathParameter("nodename"), request.PathParameter("namespace"))

	response.WriteAsJson(result)
}
