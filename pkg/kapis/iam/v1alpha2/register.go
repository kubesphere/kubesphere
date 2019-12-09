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
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
	"time"
)

const GroupName = "iam.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

type UserUpdateRequest struct {
	Username        string `json:"username" description:"username"`
	Email           string `json:"email" description:"email address"`
	Lang            string `json:"lang" description:"user's language setting, default is zh-CN"`
	Description     string `json:"description" description:"user's description"`
	Password        string `json:"password,omitempty" description:"this is necessary if you need to change your password"`
	CurrentPassword string `json:"current_password,omitempty" description:"this is necessary if you need to change your password"`
	ClusterRole     string `json:"cluster_role" description:"user's cluster role"`
}

type CreateUserRequest struct {
	Username    string `json:"username" description:"username"`
	Email       string `json:"email" description:"email address"`
	Lang        string `json:"lang,omitempty" description:"user's language setting, default is zh-CN"`
	Description string `json:"description" description:"user's description"`
	Password    string `json:"password" description:"password'"`
	ClusterRole string `json:"cluster_role" description:"user's cluster role"`
}

type UserList struct {
	Items []struct {
		Username      string    `json:"username" description:"username"`
		Email         string    `json:"email" description:"email address"`
		Lang          string    `json:"lang,omitempty" description:"user's language setting, default is zh-CN"`
		Description   string    `json:"description" description:"user's description"`
		ClusterRole   string    `json:"cluster_role" description:"user's cluster role"`
		CreateTime    time.Time `json:"create_time" description:"user creation time"`
		LastLoginTime time.Time `json:"last_login_time" description:"last login time"`
	} `json:"items" description:"paging data"`
	TotalCount int `json:"total_count" description:"total count"`
}
type NamespacedUser struct {
	Username      string    `json:"username" description:"username"`
	Email         string    `json:"email" description:"email address"`
	Lang          string    `json:"lang,omitempty" description:"user's language setting, default is zh-CN"`
	Description   string    `json:"description" description:"user's description"`
	Role          string    `json:"role" description:"user's role in the specified namespace"`
	RoleBinding   string    `json:"role_binding" description:"user's role binding name in the specified namespace"`
	RoleBindTime  string    `json:"role_bind_time" description:"user's role binding time"`
	CreateTime    time.Time `json:"create_time" description:"user creation time"`
	LastLoginTime time.Time `json:"last_login_time" description:"last login time"`
}

type ClusterRoleList struct {
	Items      []rbacv1.ClusterRole `json:"items" description:"paging data"`
	TotalCount int                  `json:"total_count" description:"total count"`
}

type LoginLog struct {
	LoginTime string `json:"login_time" description:"last login time"`
	LoginIP   string `json:"login_ip" description:"last login ip"`
}

type RoleList struct {
	Items      []rbacv1.Role `json:"items" description:"paging data"`
	TotalCount int           `json:"total_count" description:"total count"`
}

type InviteUserRequest struct {
	Username      string `json:"username" description:"username"`
	WorkspaceRole string `json:"workspace_role" description:"user's workspace role'"`
}

type DescribeWorkspaceUserResponse struct {
	Username      string    `json:"username" description:"username"`
	Email         string    `json:"email" description:"email address"`
	Lang          string    `json:"lang" description:"user's language setting, default is zh-CN"`
	Description   string    `json:"description" description:"user's description"`
	ClusterRole   string    `json:"cluster_role" description:"user's cluster role"`
	WorkspaceRole string    `json:"workspace_role" description:"user's workspace role"`
	CreateTime    time.Time `json:"create_time" description:"user creation time"`
	LastLoginTime time.Time `json:"last_login_time" description:"last login time"`
}

func addWebService(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	ok := "ok"

	ws.Route(ws.POST("/authenticate").
		To(iam.TokenReviewHandler).
		Doc("TokenReview attempts to authenticate a token to a known user. Note: TokenReview requests may be cached by the webhook token authenticator plugin in the kube-apiserver.").
		Reads(iam.TokenReview{}).
		Returns(http.StatusOK, ok, iam.TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.POST("/login").
		To(iam.Login).
		Doc("KubeSphere APIs support token-based authentication via the Authtoken request header. The POST Login API is used to retrieve the authentication token. After the authentication token is obtained, it must be inserted into the Authtoken header for all requests.").
		Reads(iam.LoginRequest{}).
		Returns(http.StatusOK, ok, models.AuthGrantResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.POST("/token").
		To(iam.OAuth).
		Doc("OAuth API,only support resource owner password credentials grant").
		Reads(iam.LoginRequest{}).
		Returns(http.StatusOK, ok, models.AuthGrantResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.GET("/users/{user}").
		To(iam.DescribeUser).
		Doc("Describe the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, ok, models.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.POST("/users").
		To(iam.CreateUser).
		Doc("Create a user account.").
		Reads(CreateUserRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.DELETE("/users/{user}").
		To(iam.DeleteUser).
		Doc("Delete the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.PUT("/users/{user}").
		To(iam.UpdateUser).
		Doc("Update information about the specified user.").
		Param(ws.PathParameter("user", "username")).
		Reads(UserUpdateRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.GET("/users/{user}/logs").
		To(iam.UserLoginLogs).
		Doc("Retrieve the \"login logs\" for the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, ok, LoginLog{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.GET("/users").
		To(iam.ListUsers).
		Doc("List all users.").
		Returns(http.StatusOK, ok, UserList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.GET("/users/{user}/roles").
		To(iam.ListUserRoles).
		Doc("Retrieve all the roles that are assigned to the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, ok, iam.RoleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(iam.ListRoles).
		Doc("Retrieve the roles that are assigned to the user in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Returns(http.StatusOK, ok, RoleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles").
		To(iam.ListClusterRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, ok, ClusterRoleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/users").
		To(iam.ListRoleUsers).
		Doc("Retrieve the users that are bound to the role in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, ok, []NamespacedUser{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(iam.ListNamespaceUsers).
		Doc("List all users in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Returns(http.StatusOK, ok, []NamespacedUser{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/users").
		To(iam.ListClusterRoleUsers).
		Doc("List all users that are bound to the specified cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, ok, UserList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/rules").
		To(iam.ListClusterRoleRules).
		Doc("List all policy rules of the specified cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, ok, []models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/rules").
		To(iam.ListRoleRules).
		Doc("List all policy rules of the specified role in the given namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, ok, []models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}/rules").
		To(iam.ListDevopsRoleRules).
		Doc("List all policy rules of the specified role in the given devops project.").
		Param(ws.PathParameter("devops", "devops project ID")).
		Param(ws.PathParameter("role", "devops role name")).
		Returns(http.StatusOK, ok, []models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/rulesmapping/clusterroles").
		To(iam.ClusterRulesMapping).
		Doc("Get the mapping relationships between cluster roles and policy rules.").
		Returns(http.StatusOK, ok, policy.ClusterRoleRuleMapping).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/rulesmapping/roles").
		To(iam.RulesMapping).
		Doc("Get the mapping relationships between namespaced roles and policy rules.").
		Returns(http.StatusOK, ok, policy.RoleRuleMapping).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/roles").
		To(iam.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, ok, ClusterRoleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/roles/{role}").
		To(iam.DescribeWorkspaceRole).
		Doc("Describe the workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("role", "workspace role name")).
		Returns(http.StatusOK, ok, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/roles/{role}/rules").
		To(iam.ListWorkspaceRoleRules).
		Doc("List all policy rules of the specified workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("role", "workspace role name")).
		Returns(http.StatusOK, ok, []models.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/members").
		To(iam.ListWorkspaceUsers).
		Doc("List all members in the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, ok, UserList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/members").
		To(iam.InviteUser).
		Doc("Invite a member to the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Reads(InviteUserRequest{}).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/members/{member}").
		To(iam.RemoveUser).
		Doc("Remove the specified member from the workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "username")).
		Returns(http.StatusOK, ok, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{member}").
		To(iam.DescribeWorkspaceUser).
		Doc("Describe the specified user in the given workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "username")).
		Returns(http.StatusOK, ok, DescribeWorkspaceUserResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	c.Add(ws)
	return nil
}
