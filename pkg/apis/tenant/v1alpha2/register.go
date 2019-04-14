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
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/apiserver/tenant"
)

const GroupName = "tenant.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	tags := []string{"Tenant"}
	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.GET("/workspaces").
		To(tenant.ListWorkspaces).
		Doc("List workspace by user").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}").
		To(tenant.DescribeWorkspace).
		Doc("Get workspace detail").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/rules").
		To(tenant.ListWorkspaceRules).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the rules for the current user").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/rules").
		To(tenant.ListNamespaceRules).
		Param(ws.PathParameter("namespace", "namespace")).
		Doc("List the rules for the current user").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/devops/{devops}/rules").
		To(tenant.ListDevopsRules).
		Param(ws.PathParameter("devops", "devops project id")).
		Doc("List the rules for the current user").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
		To(tenant.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List the namespaces for the current user").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{username}/namespaces").
		To(tenant.ListNamespaces).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("username", "workspace member's username")).
		Doc("List the namespaces for the workspace member").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/workspaces/{workspace}/namespaces").
		To(tenant.CreateNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create namespace").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/workspaces/{workspace}/namespaces/{namespace}").
		To(tenant.DeleteNamespace).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "namespace")).
		Doc("Delete namespace").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/workspaces/{workspace}/devops").
		To(tenant.ListDevopsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List devops projects for the current user").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{username}/devops").
		To(tenant.ListDevopsProjects).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("username", "workspace member's username")).
		Doc("List the devops projects for the workspace member").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/workspaces/{workspace}/devops").
		To(tenant.CreateDevopsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create devops project").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/workspaces/{workspace}/devops/{id}").
		To(tenant.DeleteDevopsProject).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Delete devops project").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/logging").
		To(tenant.LogQuery).
		Doc("Query cluster-level logs in a multi-tenants environment").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	c.Add(ws)
	return nil
}
