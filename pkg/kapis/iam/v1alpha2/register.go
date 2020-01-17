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
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/api/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/iam"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	iam2 "kubesphere.io/kubesphere/pkg/models/iam"
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

type ClusterRoleList struct {
	Items      []rbacv1.ClusterRole `json:"items" description:"paging data"`
	TotalCount int                  `json:"total_count" description:"total count"`
}

type LoginLog struct {
	LoginTime string `json:"login_time" description:"last login time"`
	LoginIP   string `json:"login_ip" description:"last login ip"`
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

	handler := newIAMHandler()

	ws.Route(ws.POST("/authenticate").
		To(handler.TokenReviewHandler).
		Doc("TokenReview attempts to authenticate a token to a known user. Note: TokenReview requests may be cached by the webhook token authenticator plugin in the kube-apiserver.").
		Reads(iamv1alpha2.TokenReview{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.TokenReview{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.POST("/login").
		To(handler.Login).
		Doc("KubeSphere APIs support token-based authentication via the Authtoken request header. The POST Login API is used to retrieve the authentication token. After the authentication token is obtained, it must be inserted into the Authtoken header for all requests.").
		Reads(iamv1alpha2.LoginRequest{}).
		Returns(http.StatusOK, api.StatusOK, models.AuthGrantResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.POST("/users").
		To(handler.CreateUser).
		Doc("Create a user account.").
		Reads(iamv1alpha2.CreateUserRequest{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.UserDetail{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.DELETE("/users/{user}").
		To(handler.DeleteUser).
		Doc("Delete the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.PUT("/users/{user}").
		To(handler.ModifyUser).
		Doc("Update information about the specified user.").
		Param(ws.PathParameter("user", "username")).
		Reads(iamv1alpha2.ModifyUserRequest{}).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.GET("/users/{user}").
		To(handler.DescribeUser).
		Doc("Describe the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.UserDetail{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.GET("/users").
		To(handler.ListUsers).
		Doc("List all users.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.ListUserResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.IdentityManagementTag}))
	ws.Route(ws.GET("/users/{user}/roles").
		To(handler.ListUserRoles).
		Doc("Retrieve all the roles that are assigned to the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, []*rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(handler.ListRoles).
		Doc("Retrieve the roles that are assigned to the user in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles").
		To(handler.ListClusterRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, api.StatusOK, ClusterRoleList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/users").
		To(handler.ListRoleUsers).
		Doc("Retrieve the users that are bound to the role in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, []iamv1alpha2.ListUserResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(handler.ListNamespaceUsers).
		Doc("List all users in the specified namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Returns(http.StatusOK, api.StatusOK, []iamv1alpha2.ListUserResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/users").
		To(handler.ListClusterRoleUsers).
		Doc("List all users that are bound to the specified cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, api.StatusOK, []iamv1alpha2.ListUserResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}/rules").
		To(handler.ListClusterRoleRules).
		Doc("List all policy rules of the specified cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, api.StatusOK, []iam2.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/rules").
		To(handler.ListRoleRules).
		Doc("List all policy rules of the specified role in the given namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, []iam2.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}/rules").
		To(iam.ListDevopsRoleRules).
		Doc("List all policy rules of the specified role in the given devops project.").
		Param(ws.PathParameter("devops", "devops project ID")).
		Param(ws.PathParameter("role", "devops role name")).
		Returns(http.StatusOK, api.StatusOK, []iam2.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/rulesmapping/clusterroles").
		To(handler.ClusterRulesMapping).
		Doc("Get the mapping relationships between cluster roles and policy rules.").
		Returns(http.StatusOK, api.StatusOK, policy.ClusterRoleRuleMapping).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/rulesmapping/roles").
		To(handler.RulesMapping).
		Doc("Get the mapping relationships between namespaced roles and policy rules.").
		Returns(http.StatusOK, api.StatusOK, policy.RoleRuleMapping).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	c.Add(ws)
	return nil
}
