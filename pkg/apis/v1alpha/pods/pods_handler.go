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

	ws.Route(ws.GET("/namespaces/{namespace}/pods").To(handlePodsUnderNameSpace).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

func handleAllPods(request *restful.Request, response *restful.Response) {
	var result constants.ResultMessage
	var resultNameSpaces []models.ResultNameSpace
	var resultNameSpace models.ResultNameSpace

	namespaces := models.GetNameSpaces()

	for _, namespace := range namespaces {

		resultNameSpace = models.FormatNameSpaceMetrics(namespace)
		resultNameSpaces = append(resultNameSpaces, resultNameSpace)

	}

	result.Data = resultNameSpaces
	response.WriteAsJson(result)
}

func handlePodsUnderNameSpace(request *restful.Request, response *restful.Response) {
	var result constants.ResultMessage
	var resultNameSpaces []models.ResultNameSpace
	var resultNameSpace models.ResultNameSpace

	resultNameSpace = models.FormatNameSpaceMetrics(request.PathParameter("namespace"))

	resultNameSpaces = append(resultNameSpaces, resultNameSpace)

	result.Data = resultNameSpaces
	response.WriteAsJson(result)
}
