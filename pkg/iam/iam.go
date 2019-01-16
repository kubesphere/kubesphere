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
package iam

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"

	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/iam/am"
	"kubesphere.io/kubesphere/pkg/iam/auth"
	"kubesphere.io/kubesphere/pkg/iam/groups"
	"kubesphere.io/kubesphere/pkg/iam/users"
	"kubesphere.io/kubesphere/pkg/iam/workspaces"
	"kubesphere.io/kubesphere/pkg/models"
)

func V1Alpha2(ws *restful.WebService) {
	t1 := []string{"IAM"}

	ws.Route(ws.POST("/authenticate").
		To(auth.TokenReviewHandler).
		Doc("Token review").
		Reads(auth.TokenReview{}).
		Writes(auth.TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.POST("/login").
		To(auth.LoginHandler).
		Doc("User login").
		Reads(auth.LoginRequest{}).
		Writes(auth.Token{}).
		Metadata(restfulspec.KeyOpenAPITags, t1))
	//TODO merge into /users/{name}
	//ws.Route(ws.GET("/users/current").
	//	To(users.CurrentUserHandler))
	ws.Route(ws.GET("/users/{name}").
		To(users.DetailHandler).
		Doc("User detail").
		Writes(models.User{}).
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.POST("/users").
		To(users.CreateHandler).
		Reads(models.User{}).
		Writes(errors.Error{}).
		Doc("Create user").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.DELETE("/users/{name}").
		To(users.DeleteHandler).
		Doc("Delete user").
		Writes(errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.PUT("/users/{name}").
		To(users.EditHandler).
		Reads(models.User{}).
		Writes(errors.Error{}).
		Doc("Update user").
		Metadata(restfulspec.KeyOpenAPITags, t1))

	//ws.Route(ws.GET("/users/{name}/namespaces").To(users.NamespacesListHandler))
	ws.Route(ws.GET("/users/{name}/log").
		To(users.LogHandler).
		Doc("User login log").
		Writes([]map[string]string{}).
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/users").
		To(users.ListHandler).
		Doc("User list").
		Writes(models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/groups").
		To(groups.RootListHandler).
		Writes([]models.Group{}).
		Doc("User group list").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	// TODO  merge into /groups
	//ws.Route(ws.GET("/groups/count"))
	ws.Route(ws.GET("/groups/{path}").
		To(groups.DetailHandler).
		Doc("User group detail").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/groups/{path}/users").
		To(groups.UsersHandler).
		Doc("Group user list").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.POST("/groups").
		To(groups.CreateHandler).
		Reads(models.Group{}).
		Doc("Create user group").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.DELETE("/groups/{path}").
		To(groups.DeleteHandler).
		Doc("Delete user group").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.PUT("/groups/{path}").
		To(groups.EditHandler).
		Doc("Update user group").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/users/{username}/roles").
		To(am.UserRolesHandler).
		Doc("Get user role list").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/users").
		To(am.RoleUsersHandler).
		Doc("Get user list by role").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/rules").
		To(am.RoleRulesHandler).
		Doc("Get role detail").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(am.NamespaceUsersHandler).
		Doc("Get user list by namespace").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/users").
		To(am.ClusterRoleUsersHandler).
		Doc("Get user list by cluster role").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/rules").
		To(am.ClusterRoleRulesHandler).
		Doc("Get cluster role detail").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/rulesmapping/clusterroles").
		To(am.ClusterRulesMappingHandler).
		Doc("Get cluster role policy rules mapping").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/rulesmapping/roles").
		To(am.RulesMappingHandler).
		Doc("Get role policy rules mapping").
		Metadata(restfulspec.KeyOpenAPITags, t1))

	ws.Route(ws.GET("/workspaces/{workspace}/rules").
		To(workspaces.WorkspaceRulesHandler).
		Doc("Get workspace level policy rules").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/workspaces/{workspace}/members").
		To(workspaces.UsersHandler).
		Doc("Get workspace member list").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/namespaces/{namespace}/rules").
		To(workspaces.NamespacesRulesHandler).
		Doc("Get namespace level policy rules").
		Metadata(restfulspec.KeyOpenAPITags, t1))
	ws.Route(ws.GET("/devops/{devops}/rules").
		To(workspaces.DevopsRulesHandler).
		Doc("Get devops project level policy rules").
		Metadata(restfulspec.KeyOpenAPITags, t1))

	t2 := []string{"Workspace"}

	ws.Route(ws.GET("/workspaces").
		To(workspaces.UserWorkspaceListHandler).
		Doc("Get workspace list").
		Metadata(restfulspec.KeyOpenAPITags, t2).
		Writes([]models.Workspace{}))
	ws.Route(ws.POST("/workspaces").
		To(workspaces.WorkspaceCreateHandler).
		Doc("Create workspace").
		Metadata(restfulspec.KeyOpenAPITags, t2).
		Writes(models.Workspace{}))
	ws.Route(ws.DELETE("/workspaces/{name}").
		To(workspaces.DeleteWorkspaceHandler).
		Doc("Delete workspace").
		Metadata(restfulspec.KeyOpenAPITags, t2).
		Writes(errors.Error{}))
	ws.Route(ws.GET("/workspaces/{name}").
		To(workspaces.WorkspaceDetailHandler).
		Doc("Get workspace detail").
		Metadata(restfulspec.KeyOpenAPITags, t2).
		Writes(models.Workspace{}))
	ws.Route(ws.PUT("/workspaces/{name}").
		To(workspaces.WorkspaceEditHandler).
		Doc("Update workspace").
		Metadata(restfulspec.KeyOpenAPITags, t2).
		Writes(models.Workspace{}))
	// TODO move to /apis/resources.kubesphere.io/namespaces
	//ws.Route(ws.GET("/workspaces/{workspace}/namespaces").
	//	To(UserNamespaceListHandler).
	//	Doc("Get namespace list").
	//	Metadata(restfulspec.KeyOpenAPITags, t2).
	//	Writes(models.PageableResponse{}))
	//ws.Route(ws.POST("/workspaces/{name}/namespaces").
	//	To(NamespaceCreateHandler).
	//	Doc("Create namespace").
	//	Metadata(restfulspec.KeyOpenAPITags, t2).
	//	Writes(v1.Namespace{}))
	//ws.Route(ws.DELETE("/workspaces/{name}/namespaces/{namespace}").To(NamespaceDeleteHandler))
	//ws.Route(ws.GET("/workspaces/{name}/namespaces/{namespace}").To(NamespaceCheckHandler))
	//ws.Route(ws.GET("/namespaces/{namespace}").To(NamespaceCheckHandler))
	// TODO move to /apis/resources.kubesphere.io/workspaces/{workspace}/members/{username}
	//ws.Route(ws.GET("/workspaces/{workspace}/members/{username}/namespaces").To(UserNamespaceListHandler))
	//ws.Route(ws.GET("/workspaces/{name}/members/{username}/devops").To(DevOpsProjectHandler))

	// TODO move to /apis/resources.kubesphere.io/devops
	//ws.Route(ws.GET("/workspaces/{name}/devops").To(DevOpsProjectHandler))
	//ws.Route(ws.POST("/workspaces/{name}/devops").To(DevOpsProjectCreateHandler))
	//ws.Route(ws.DELETE("/workspaces/{name}/devops/{id}").To(DevOpsProjectDeleteHandler))

	ws.Route(ws.GET("/workspaces/{name}/members").
		To(workspaces.MembersHandler).
		Doc("Get workspace members").
		Metadata(restfulspec.KeyOpenAPITags, t2))
	ws.Route(ws.GET("/workspaces/{name}/members/{member}").
		To(workspaces.MemberHandler).
		Doc("Get workspace member detail").
		Metadata(restfulspec.KeyOpenAPITags, t2))
	ws.Route(ws.GET("/workspaces/{name}/roles").
		To(workspaces.RolesHandler).
		Doc("Get workspace roles").
		Metadata(restfulspec.KeyOpenAPITags, t2))
	// TODO /workspaces/{name}/roles/{role}
	ws.Route(ws.POST("/workspaces/{name}/members").
		To(workspaces.MembersInviteHandler).
		Doc("Add user to workspace").
		Metadata(restfulspec.KeyOpenAPITags, t2))
	ws.Route(ws.DELETE("/workspaces/{name}/members").
		To(workspaces.MembersRemoveHandler).
		Doc("Delete user from workspace").
		Metadata(restfulspec.KeyOpenAPITags, t2))
}
