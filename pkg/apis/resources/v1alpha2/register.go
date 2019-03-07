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
package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/resources"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const GroupName = "resources.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {

	webservice := runtime.NewWebService(GroupVersion)

	webservice.Route(webservice.GET("/namespaces/{namespace}/{resources}").To(resources.NamespaceResourceHandler))

	webservice.Route(webservice.GET("/{resources}").To(resources.ClusterResourceHandler))

	webservice.Route(webservice.GET("/storageclasses/{storageclass}/persistentvolumeclaims").To(resources.GetPvcListBySc))
	webservice.Route(webservice.GET("/namespaces/{namespace}/persistentvolumeclaims/{pvc}/pods").To(resources.GetPodListByPvc))

	tags := []string{"users"}
	webservice.Route(webservice.GET("/users/{username}/kubectl").Doc("get user's kubectl pod").Param(webservice.PathParameter("username",
		"username").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).To(resources.GetKubectl))
	webservice.Route(webservice.GET("/users/{username}/kubeconfig").Doc("get users' kubeconfig").Param(webservice.PathParameter("username",
		"username").DataType("string")).Metadata(restfulspec.KeyOpenAPITags, tags).To(resources.GetKubeconfig))

	c.Add(webservice)

	return nil
}
