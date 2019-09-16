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

package components

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
)

func GetSystemHealthStatus(request *restful.Request, response *restful.Response) {
	result, err := components.GetSystemHealthStatus()

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(result)
}

// get a specific component status
func GetComponentStatus(request *restful.Request, response *restful.Response) {
	component := request.PathParameter("component")

	result, err := components.GetComponentStatus(component)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(result)
}

// get all componentsHandler
func GetComponents(request *restful.Request, response *restful.Response) {

	result, err := components.GetAllComponentsStatus()

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(result)
}
