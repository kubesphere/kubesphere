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

package registries

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/registries"
)

func V1Alpha2(ws *restful.WebService) {

	ws.Route(ws.POST("registries/verify").To(registryVerify))

}

func registryVerify(request *restful.Request, response *restful.Response) {

	authInfo := registries.AuthInfo{}

	err := request.ReadEntity(&authInfo)

	if errors.HandlerError(err, response) {
		return
	}

	err = registries.RegistryVerify(authInfo)

	if errors.HandlerError(err, response) {
		return
	}

	response.WriteAsJson(errors.None)
}

//func (c *registriesHandler) handlerImageSearch(request *restful.Request, response *restful.Response) {
//
//	registry := request.PathParameter("name")
//	searchWord := request.PathParameter("searchWord")
//	namespace := request.PathParameter("namespace")
//
//	res := c.registries.ImageSearch(namespace, registry, searchWord)
//
//	response.WriteAsJson(res)
//
//}
//
//func (c *registriesHandler) handlerGetImageTags(request *restful.Request, response *restful.Response) {
//
//	registry := request.PathParameter("name")
//	image := request.QueryParameter("image")
//	namespace := request.PathParameter("namespace")
//
//	res := c.registries.GetImageTags(namespace, registry, image)
//
//	response.WriteAsJson(res)
//}
