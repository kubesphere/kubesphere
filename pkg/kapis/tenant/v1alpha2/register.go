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
	devopsv1alpha2 "kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"

	"net/http"
)

const (
	GroupName = "tenant.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, k8sClient k8s.Client, factory informers.InformerFactory, db *mysql.Database) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newTenantHandler(k8sClient, factory, db)

	ws.Route(ws.GET("/workspaces").
		To(handler.ListWorkspaces).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List all workspaces that belongs to the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}").
		To(handler.DescribeWorkspace).
		Doc("Describe the specified workspace").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.Workspace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(handler.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces of the specified workspace for the current user").
		Returns(http.StatusOK, api.StatusOK, []v1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{member}/namespaces").
		To(handler.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "workspace member's username")).
		Doc("List the namespaces for the workspace member").
		Returns(http.StatusOK, api.StatusOK, []v1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/namespaces").
		To(handler.CreateNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create a namespace in the specified workspace").
		Returns(http.StatusOK, api.StatusOK, []v1.Namespace{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/namespaces/{namespace}").
		To(handler.DeleteNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "the name of the namespace")).
		Doc("Delete the specified namespace from the workspace").
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/devops").
		To(handler.ListDevopsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(ws.QueryParameter(params.ConditionsParam, "query conditions").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Doc("List devops projects for the current user").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{member}/devops").
		To(handler.ListDevopsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "workspace member's username")).
		Param(ws.QueryParameter(params.PagingParam, "page").
			Required(false).
			DataFormat("limit=%d,page=%d").
			DefaultValue("limit=10,page=1")).
		Param(ws.QueryParameter(params.ConditionsParam, "query conditions").
			Required(false).
			DataFormat("key=%s,key~%s")).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Doc("List the devops projects for the workspace member").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.GET("/devopscount").
		To(handler.GetDevOpsProjectsCount).
		Returns(http.StatusOK, api.StatusOK, struct {
			Count uint32 `json:"count"`
		}{}).
		Doc("Get the devops projects count for the member").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/devops").
		To(handler.CreateDevopsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create a devops project in the specified workspace").
		Reads(devopsv1alpha2.DevOpsProject{}).
		Returns(http.StatusOK, api.StatusOK, devopsv1alpha2.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/devops/{devops}").
		To(handler.DeleteDevopsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("devops", "devops project ID")).
		Doc("Delete the specified devops project from the workspace").
		Returns(http.StatusOK, api.StatusOK, devopsv1alpha2.DevOpsProject{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TenantResourcesTag}))

	c.Add(ws)
	return nil
}
