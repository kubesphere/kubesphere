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

package containers

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/metrics"
)

// {namespace} namespace name
// {node} node host name
// {pod} pod name
// {container} container name
func Register(ws *restful.WebService) {
	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers/{container}").To(handleContainerUnderNameSpaceAndPod).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/namespaces/{namespace}/pods/{pod}/containers").To(handleContainersUnderNameSpaceAndPod).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/nodes/{node}/namespaces/{namespace}/pods/{pod}/containers").To(handleContainersUnderNodeAndNameSpaceAndPod).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

func handleContainerUnderNameSpaceAndPod(request *restful.Request, response *restful.Response) {
	var resultContainer metrics.ContainerMetrics
	resultContainer = metrics.FormatContainerMetrics(request.PathParameter("namespace"), request.PathParameter("pod"), request.PathParameter("container"))
	resultContainer.NodeName = metrics.GetNodeNameForPod(request.PathParameter("pod"), request.PathParameter("namespace"))
	response.WriteAsJson(resultContainer)
}
func handleContainersUnderNameSpaceAndPod(request *restful.Request, response *restful.Response) {
	var resultNameSpace constants.PageableResponse
	resultNameSpace = metrics.FormatContainersMetrics("", request.PathParameter("namespace"), request.PathParameter("pod"))
	response.WriteAsJson(resultNameSpace)
}

func handleContainersUnderNodeAndNameSpaceAndPod(request *restful.Request, response *restful.Response) {
	var resultNameSpace constants.PageableResponse

	resultNameSpace = metrics.FormatContainersMetrics(request.PathParameter("node"), request.PathParameter("namespace"), request.PathParameter("pod"))

	response.WriteAsJson(resultNameSpace)
}
