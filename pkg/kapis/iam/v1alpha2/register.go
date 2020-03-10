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
	"kubesphere.io/kubesphere/pkg/api/iam"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/api/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldappool "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"net/http"
)

const groupName = "iam.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, k8sClient k8s.Client, factory informers.InformerFactory, ldapClient ldappool.Interface, cacheClient cache.Interface, options *iam.AuthenticationOptions) error {
	ws := runtime.NewWebService(GroupVersion)

	handler := newIAMHandler(k8sClient, factory, ldapClient, cacheClient, options)

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
		Returns(http.StatusOK, api.StatusOK, models.PageableResponse{}).
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
		Returns(http.StatusOK, api.StatusOK, []policy.SimpleRule{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}/rules").
		To(handler.ListRoleRules).
		Doc("List all policy rules of the specified role in the given namespace.").
		Param(ws.PathParameter("namespace", "kubernetes namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, []policy.SimpleRule{}).
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

	ws.Route(ws.GET("/workspaces/{workspace}/roles").
		To(handler.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/roles/{role}").
		To(handler.DescribeWorkspaceRole).
		Doc("Describe the workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("role", "workspace role name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/roles/{role}/rules").
		To(handler.ListWorkspaceRoleRules).
		Doc("List all policy rules of the specified workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("role", "workspace role name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/members").
		To(handler.ListWorkspaceUsers).
		Doc("List all members in the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/members").
		To(handler.InviteUser).
		Doc("Invite a member to the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/members/{member}").
		To(handler.RemoveUser).
		Doc("Remove the specified member from the workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.Error{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/members/{member}").
		To(handler.DescribeWorkspaceUser).
		Doc("Describe the specified user in the given workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("member", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	c.Add(ws)
	return nil
}
