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
package resources

import (
	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
)

func Route(ws *restful.WebService) {
	ws.Route(ws.GET("/namespaces/{namespace}/{resources}").To(namespaceResourceHandler))
	ws.Route(ws.GET("/{resources}").To(clusterResourceHandler))

	ws.Route(ws.GET("/storageclasses/{storageclass}/persistentvolumeclaims").To(getPvcListBySc))
	ws.Route(ws.GET("/namespaces/{namespace}/persistentvolumeclaims/{pvc}/pods").To(getPodListByPvc))

	tags := []string{"users"}
	ws.Route(ws.GET("/users/{username}/kubectl").Doc("get user's kubectl pod").Param(ws.PathParameter("username",
		"username").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).To(getKubectl))
	ws.Route(ws.GET("/users/{username}/kubeconfig").Doc("get users' kubeconfig").Param(ws.PathParameter("username",
		"username").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).To(getKubeconfig))
}
