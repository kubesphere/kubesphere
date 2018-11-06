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

package components

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {
	ws.Route(ws.GET(subPath).To(handleGetComponents).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET(subPath+"/{namespace}/{componentName}").To(handleGetComponentStatus).
		Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

// get a specific component status
func handleGetComponentStatus(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	componentName := request.PathParameter("componentName")

	if component, err := models.GetComponentStatus(namespace, componentName); err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
	} else {
		response.WriteAsJson(component)
	}
}

// get all components
func handleGetComponents(request *restful.Request, response *restful.Response) {

	result, err := models.GetAllComponentsStatus()

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})
	} else {
		response.WriteAsJson(result)
	}

}
