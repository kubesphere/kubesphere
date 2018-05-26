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
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"net/http"
)

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.POST(subPath + "/validation").To(handlerRegistryValidation).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

}

func handlerRegistryValidation(request *restful.Request, response *restful.Response) {

	authinfo := models.AuthInfo{}

	err := request.ReadEntity(&authinfo)

	if err != nil {

		response.WriteError(http.StatusInternalServerError, err)

	}

	result := models.RegistryLoginAuth(authinfo)

	response.WriteAsJson(result)

}


