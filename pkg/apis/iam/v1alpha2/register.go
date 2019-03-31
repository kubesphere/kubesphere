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
	"kubesphere.io/kubesphere/pkg/apiserver/iam"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
)

const GroupName = "iam.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {
	tags := []string{"IAM"}
	ws := runtime.NewWebService(GroupVersion)

	ws.Route(ws.POST("/authenticate").
		To(iam.TokenReviewHandler).
		Doc("Token review").
		Reads(iam.TokenReview{}).
		Writes(iam.TokenReview{}).
		Doc("k8s token review").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/login").
		To(iam.LoginHandler).
		Doc("User login").
		Reads(iam.LoginRequest{}).
		Writes(models.Token{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users/{username}").
		To(iam.DescribeUser).
		Doc("User detail").
		Param(ws.PathParameter("username", "username")).
		Writes(models.User{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/users").
		To(iam.CreateUser).
		Reads(models.User{}).
		Writes(errors.Error{}).
		Doc("Create user").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/users/{name}").
		To(iam.DeleteUser).
		Param(ws.PathParameter("name", "username")).
		Doc("Delete user").
		Writes(errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.PUT("/users/{name}").
		To(iam.UpdateUser).
		Param(ws.PathParameter("name", "username")).
		Reads(models.User{}).
		Writes(errors.Error{}).
		Doc("Update user").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/users/{name}/log").
		To(iam.UserLoginLog).
		Param(ws.PathParameter("name", "username")).
		Doc("User login log").
		Writes([]map[string]string{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users").
		To(iam.ListUsers).
		Doc("User list").
		Writes(models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/groups").
		To(iam.ListGroups).
		Writes([]models.Group{}).
		Doc("User group list").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/groups/{path}").
		To(iam.DescribeGroup).
		Param(ws.PathParameter("path", "group path")).
		Doc("User group detail").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/groups/{path}/users").
		To(iam.ListGroupUsers).
		Param(ws.PathParameter("path", "group path")).
		Doc("Group user list").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/groups").
		To(iam.CreateGroup).
		Reads(models.Group{}).
		Doc("Create user group").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/groups/{path}").
		To(iam.DeleteGroup).
		Param(ws.PathParameter("path", "group path")).
		Doc("Delete user group").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.PUT("/groups/{path}").
		To(iam.UpdateGroup).
		Param(ws.PathParameter("path", "group path")).
		Doc("Update user group").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users/{username}/roles").
		To(iam.ListUserRoles).
		Param(ws.PathParameter("username", "username")).
		Doc("Get user role list").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(iam.ListRoles).
		Param(ws.PathParameter("namespace", "namespace")).
		Doc("Get role list").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/clusterroles").
		To(iam.ListClusterRoles).
		Doc("Get cluster role list").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/users").
		To(iam.ListRoleUsers).
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Doc("Get user list by role").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(iam.ListNamespaceUsers).
		Param(ws.PathParameter("namespace", "namespace")).
		Doc("Get user list by namespace").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/users").
		To(iam.ListClusterRoleUsers).
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Doc("Get user list by cluster role").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/rules").
		To(iam.ListClusterRoleRules).
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Doc("Get cluster role detail").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/rulesmapping/clusterroles").
		To(iam.ClusterRulesMapping).
		Doc("Get cluster role policy rules mapping").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/rulesmapping/roles").
		To(iam.RulesMapping).
		Doc("Get role policy rules mapping").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/workspaces/{workspace}/roles").
		To(iam.ListWorkspaceRoles).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("List workspace role").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/roles/{role}").
		To(iam.DescribeWorkspaceRole).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("role", "workspace role name")).
		Doc("Describe workspace role").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/roles/{role}/rules").
		To(iam.ListWorkspaceRoleRules).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("role", "workspace role name")).
		Doc("Get workspace role policy rules").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/members").
		To(iam.ListWorkspaceUsers).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Get workspace member list").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/workspaces/{workspace}/members").
		To(iam.InviteUser).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Add user to workspace").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/workspaces/{workspace}/members").
		To(iam.RemoveUser).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Remove user from workspace").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{username}").
		To(iam.DescribeWorkspaceUser).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("username", "username")).
		Doc("Describe user in workspace").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/rules").
		To(iam.ListRoleRules).
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Doc("Get namespace role policy rules").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}/rules").
		To(iam.ListDevopsRoleRules).
		Param(ws.PathParameter("devops", "devops project id")).
		Param(ws.PathParameter("role", "devops role name")).
		Doc("Get devops role policy rules").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	c.Add(ws)
	return nil
}
