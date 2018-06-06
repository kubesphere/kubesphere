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

package registries

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.POST(subPath + "/validation").To(handlerRegistryValidation).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST(subPath).To(handleCreateRegistries).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET(subPath + "/{project}").To(handleQueryRegistries).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

}

func handlerRegistryValidation(request *restful.Request, response *restful.Response) {

	authinfo := models.AuthInfo{}

	err := request.ReadEntity(&authinfo)

	if err != nil {

		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})

	} else {

		result := models.RegistryLoginAuth(authinfo)

		response.WriteAsJson(result)

	}

}

func handleCreateRegistries(request *restful.Request, response *restful.Response) {

	registries := models.Registries{}

	err := request.ReadEntity(&registries)

	if err != nil {

		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})

	}

	result, err := models.CreateRegistries(registries)

	if err != nil {

		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})

	} else {

		response.WriteAsJson(result)

	}

}

func handleQueryRegistries(request *restful.Request, response *restful.Response) {

	project := request.PathParameter("project")
	result, err := models.QueryRegistries(project)

	if err != nil {

		response.WriteHeaderAndEntity(http.StatusInternalServerError, constants.MessageResponse{Message: err.Error()})

	} else {

		response.WriteAsJson(result)

	}

}