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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"net/http"
)

const (
	GroupName = "tenant.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, k8sClient k8s.Client, factory informers.InformerFactory) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newTenantHandler(k8sClient, factory)

	ws.Route(ws.GET("/workspaces").
		To(handler.ListWorkspaces).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(handler.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	c.Add(ws)
	return nil
}
