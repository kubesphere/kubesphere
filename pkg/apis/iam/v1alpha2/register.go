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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/iam"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"net/http"
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

	ok := "ok"
	pageableUserList := struct {
		Items      []models.User `json:"items"`
		TotalCount int           `json:"total_count"`
	}{}

	ws.Route(ws.POST("/authenticate").
		To(iam.TokenReviewHandler).
		Doc("TokenReview attempts to authenticate a token to a known user. Note: TokenReview requests may be cached by the webhook token authenticator plugin in the kube-apiserver.").
		Reads(iam.TokenReview{}).
		Returns(http.StatusOK, ok, iam.TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/login").
		To(iam.LoginHandler).
		Doc("KubeSphere APIs support token-based authentication via the Authtoken request header. The POST Login API is used to retrieve the authentication token. After the authentication token is obtained, it must be inserted into the Authtoken header for all requests.").
		Reads(iam.LoginRequest{}).
		Returns(http.StatusOK, ok, models.Token{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users/{username}").
		To(iam.DescribeUser).
		Doc("Describes the specified user.").
		Param(ws.PathParameter("username", "username")).
		Returns(http.StatusOK, ok, models.User{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/users").
		To(iam.CreateUser).
		Doc("Create a user account.").
		Reads(models.User{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/users/{name}").
		To(iam.DeleteUser).
		Doc("Remove a specified user.").
		Param(ws.PathParameter("name", "username")).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.PUT("/users/{name}").
		To(iam.UpdateUser).
		Doc("Updates information about the specified user.").
		Param(ws.PathParameter("name", "username")).
		Reads(models.User{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users/{name}/log").
		To(iam.UserLoginLog).
		Doc("This method is used to retrieve the \"login logs\" for the specified user.").
		Param(ws.PathParameter("name", "username")).
		Returns(http.StatusOK, ok, struct {
			LoginTime string `json:"login_time"`
			LoginIP   string `json:"login_ip"`
		}{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users").
		To(iam.ListUsers).
		Doc("List all users.").
		Returns(http.StatusOK, ok, pageableUserList).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/groups").
		To(iam.ListGroups).
		Doc("List all user groups.").
		Returns(http.StatusOK, ok, []models.Group{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/groups/{path}").
		To(iam.DescribeGroup).
		Doc("Describes the specified user group.").
		Param(ws.PathParameter("path", "user group path separated by colon.")).
		Returns(http.StatusOK, ok, models.Group{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/groups/{path}/users").
		To(iam.ListGroupUsers).
		Doc("List all users in the specified user group.").
		Param(ws.PathParameter("path", "user group path separated by colon.")).
		Returns(http.StatusOK, ok, []models.User{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/groups").
		To(iam.CreateGroup).
		Doc("Create a user group.").
		Reads(models.Group{}).
		Returns(http.StatusOK, ok, models.Group{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/groups/{path}").
		To(iam.DeleteGroup).
		Doc("Delete a user group.").
		Param(ws.PathParameter("path", "user group path separated by colon.")).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.PUT("/groups/{path}").
		To(iam.UpdateGroup).
		Doc("Updates information about the user group.").
		Param(ws.PathParameter("path", "user group path separated by colon.")).
		Reads(models.Group{}).
		Returns(http.StatusOK, ok, models.Group{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/users/{username}/roles").
		To(iam.ListUserRoles).
		Doc("This method is used to retrieve all the roles that are assigned to the user.").
		Param(ws.PathParameter("username", "username")).
		Returns(http.StatusOK, ok, iam.RoleList{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(iam.ListRoles).
		Doc("This method is used to retrieve the roles that are assigned to the user in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Returns(http.StatusOK, ok, struct {
			Items      []rbacv1.Role `json:"items"`
			TotalCount int           `json:"total_count"`
		}{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/clusterroles").
		To(iam.ListClusterRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, ok, struct {
			Items      []rbacv1.ClusterRole `json:"items"`
			TotalCount int                  `json:"total_count"`
		}{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/users").
		To(iam.ListRoleUsers).
		Doc("This method is used to retrieve the users that are bind the role in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, ok, []models.User{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(iam.ListNamespaceUsers).
		Doc("List all users in the specified namespace").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Returns(http.StatusOK, ok, []models.User{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/users").
		To(iam.ListClusterRoleUsers).
		Doc("List all users that are bind the cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, ok, pageableUserList).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/rules").
		To(iam.ListClusterRoleRules).
		Doc("List all policy rules of the specified cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, ok, []models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/rules").
		To(iam.ListRoleRules).
		Doc("List all policy rules of the specified role.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, ok, []models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}/rules").
		To(iam.ListDevopsRoleRules).
		Doc("List all policy rules of the specified role.").
		Param(ws.PathParameter("devops", "devops project id")).
		Param(ws.PathParameter("role", "devops role name")).
		Returns(http.StatusOK, ok, []models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/rulesmapping/clusterroles").
		To(iam.ClusterRulesMapping).
		Doc("Get the mapping relationships between cluster roles and policy rules.").
		Returns(http.StatusOK, ok, policy.ClusterRoleRuleMapping).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/rulesmapping/roles").
		To(iam.RulesMapping).
		Doc("Get the mapping relationships between namespaced roles and policy rules.").
		Returns(http.StatusOK, ok, policy.RoleRuleMapping).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/roles").
		To(iam.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, ok, struct {
			Items      []rbacv1.ClusterRole `json:"items"`
			TotalCount int                  `json:"total_count"`
		}{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/roles/{role}").
		To(iam.DescribeWorkspaceRole).
		Doc("Describes the workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("role", "workspace role name")).
		Returns(http.StatusOK, ok, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/roles/{role}/rules").
		To(iam.ListWorkspaceRoleRules).
		Doc("List all policy rules of the specified workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("role", "workspace role name")).
		Returns(http.StatusOK, ok, []models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/members").
		To(iam.ListWorkspaceUsers).
		Doc("List all members in the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, ok, pageableUserList).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.POST("/workspaces/{workspace}/members").
		To(iam.InviteUser).
		Doc("Invite members to a workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Reads(models.User{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.DELETE("/workspaces/{workspace}/members/{username}").
		To(iam.RemoveUser).
		Doc("Remove members from workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("name", "username")).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{username}").
		To(iam.DescribeWorkspaceUser).
		Doc("Describes the specified user.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("username", "username")).
		Returns(http.StatusOK, ok, models.User{}).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	c.Add(ws)
	return nil
}
