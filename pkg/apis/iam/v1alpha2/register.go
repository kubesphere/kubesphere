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
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/login").
		To(iam.LoginHandler).
		Doc("User login").
		Reads(iam.LoginRequest{}).
		Writes(models.Token{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/users/{name}").
		To(iam.UserDetail).
		Doc("User detail").
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
		Doc("Delete user").
		Writes(errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.PUT("/users/{name}").
		To(iam.UpdateUser).
		Reads(models.User{}).
		Writes(errors.Error{}).
		Doc("Update user").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/users/{name}/log").
		To(iam.UserLoginLog).
		Doc("User login log").
		Writes([]map[string]string{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users").
		To(iam.UserList).
		Doc("User list").
		Writes(models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/groups").
		To(iam.RootGroupList).
		Writes([]models.Group{}).
		Doc("User group list").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/groups/{path}").
		To(iam.GroupDetail).
		Doc("User group detail").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/groups/{path}/users").
		To(iam.GroupUsers).
		Doc("Group user list").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/groups").
		To(iam.CreateGroup).
		Reads(models.Group{}).
		Doc("Create user group").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/groups/{path}").
		To(iam.DeleteGroup).
		Doc("Delete user group").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.PUT("/groups/{path}").
		To(iam.UpdateGroup).
		Doc("Update user group").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users/{username}/roles").
		To(iam.UserRoles).
		Doc("Get user role list").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/users").
		To(iam.RoleUsers).
		Doc("Get user list by role").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/rules").
		To(iam.RoleRules).
		Doc("Get role detail").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(iam.NamespaceUsers).
		Doc("Get user list by namespace").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/users").
		To(iam.ClusterRoleUsers).
		Doc("Get user list by cluster role").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/rules").
		To(iam.ClusterRoleRules).
		Doc("Get cluster role detail").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/rulesmapping/clusterroles").
		To(iam.ClusterRulesMappingHandler).
		Doc("Get cluster role policy rules mapping").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/rulesmapping/roles").
		To(iam.RulesMappingHandler).
		Doc("Get role policy rules mapping").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.GET("/workspaces/{workspace}/rules").
		To(iam.WorkspaceRulesHandler).
		Doc("Get workspace level policy rules").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/members").
		To(iam.WorkspaceMemberList).
		Doc("Get workspace member list").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/rules").
		To(iam.NamespacesRulesHandler).
		Doc("Get namespace level policy rules").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/devops/{devops}/rules").
		To(iam.DevopsRulesHandler).
		Doc("Get devops project level policy rules").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	tags = []string{"Workspace"}

	ws.Route(ws.GET("/workspaces").
		To(iam.UserWorkspaceListHandler).
		Doc("Get workspace list").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]models.Workspace{}))
	ws.Route(ws.POST("/workspaces").
		To(iam.WorkspaceCreateHandler).
		Doc("Create workspace").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(models.Workspace{}))
	ws.Route(ws.DELETE("/workspaces/{name}").
		To(iam.DeleteWorkspaceHandler).
		Doc("Delete workspace").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(errors.Error{}))
	ws.Route(ws.GET("/workspaces/{name}").
		To(iam.WorkspaceDetailHandler).
		Doc("Get workspace detail").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(models.Workspace{}))
	ws.Route(ws.PUT("/workspaces/{name}").
		To(iam.WorkspaceEditHandler).
		Doc("Update workspace").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(models.Workspace{}))

	ws.Route(ws.GET("/workspaces/{name}/members/{member}").
		To(iam.WorkspaceMemberDetail).
		Doc("Get workspace member detail").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{name}/roles").
		To(iam.WorkspaceRoles).
		Doc("Get workspace roles").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.POST("/workspaces/{name}/members").
		To(iam.WorkspaceMemberInvite).
		Doc("Add user to workspace").
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/workspaces/{name}/members").
		To(iam.WorkspaceMemberRemove).
		Doc("Delete user from workspace").
		Metadata(restfulspec.KeyOpenAPITags, tags))

	c.Add(ws)
	return nil
}
