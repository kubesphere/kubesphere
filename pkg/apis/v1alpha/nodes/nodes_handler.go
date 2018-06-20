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

package nodes

import (
	"github.com/emicklei/go-restful"

	"net/http"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.GET(subPath).To(handleNodes).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET(subPath+"/{nodename}").To(handleSingleNode).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST(subPath+"/{nodename}/drainage").To(handleDrainNode).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath+"/{nodename}/drainage").To(handleDrainStatus).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

func handleNodes(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse
	var resultNode models.ResultNode

	nodes := models.GetNodes()

	var total_count int
	for i, node := range nodes {
		resultNode = models.FormatNodeMetrics(node)
		result.Items = append(result.Items, resultNode)
		total_count = i
	}
	total_count = total_count + 1

	result.TotalCount = total_count
	response.WriteAsJson(result)
}

func handleSingleNode(request *restful.Request, response *restful.Response) {
	nodeName := request.PathParameter("nodename")
	var resultNode models.ResultNode

	resultNode = models.FormatNodeMetrics(nodeName)

	response.WriteAsJson(resultNode)
}

func handleDrainNode(request *restful.Request, response *restful.Response) {

	nodeName := request.PathParameter("nodename")

	result, err := models.DrainNode(nodeName)

	if err != nil {

		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})

	} else {

		response.WriteAsJson(result)

	}

}

func handleDrainStatus(request *restful.Request, response *restful.Response) {

	nodeName := request.PathParameter("nodename")

	result, err := models.DrainStatus(nodeName)

	if err != nil {

		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})

	} else {

		response.WriteAsJson(result)

	}
}
