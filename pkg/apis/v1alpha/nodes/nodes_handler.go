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
	"kubesphere.io/kubesphere/pkg/models/metrics"
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
}

func MakeRequest(node string, ch chan<- metrics.NodeMetrics) {
	resultNode := metrics.FormatNodeMetrics(node)

	ch <- resultNode
}

func handleNodes(request *restful.Request, response *restful.Response) {
	var result constants.PageableResponse

	nodes := metrics.GetNodes()

	ch := make(chan metrics.NodeMetrics)
	for _, node := range nodes {
		go MakeRequest(node, ch)
	}

	for _, _ = range nodes {
		result.Items = append(result.Items, <-ch)
	}

	if result.Items == nil {
		result.Items = make([]interface{}, 0)
	}

	result.TotalCount = len(result.Items)
	response.WriteAsJson(result)
}

func handleSingleNode(request *restful.Request, response *restful.Response) {
	nodeName := request.PathParameter("nodename")
	var resultNode metrics.NodeMetrics

	resultNode = metrics.FormatNodeMetrics(nodeName)

	response.WriteAsJson(resultNode)
}

func handleDrainNode(request *restful.Request, response *restful.Response) {

	nodeName := request.PathParameter("nodename")

	result, err := metrics.DrainNode(nodeName)

	if err != nil {

		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})

	} else {

		response.WriteAsJson(result)

	}

}
