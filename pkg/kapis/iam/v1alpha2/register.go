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
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
)

const (
	GroupName = "iam.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, im im.IdentityManagementInterface, am am.AccessManagementInterface, options *authoptions.AuthenticationOptions) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newIAMHandler(im, am, options)

	// users
	ws.Route(ws.POST("/users").
		To(handler.CreateUser).
		Doc("Create user in global scope.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Reads(iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/users/{user}").
		To(handler.DeleteUser).
		Doc("Delete user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/users/{user}").
		To(handler.UpdateUser).
		Doc("Update user info.").
		Reads(iamv1alpha2.User{}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/users/{user}/password").
		To(handler.ModifyPassword).
		Doc("Modify user's password.").
		Reads(iam.PasswordReset{}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/users/{user}").
		To(handler.DescribeUser).
		Doc("Retrieve user details.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/users").
		To(handler.ListUsers).
		Doc("List all users in global scope.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/users/{user}/loginrecords").
		To(handler.ListUserLoginRecords).
		Param(ws.PathParameter("user", "username of the user")).
		Doc("List user's login records.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.LoginRecord{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// clustermembers
	ws.Route(ws.POST("/clustermembers").
		To(handler.CreateClusterMembers).
		Doc("Add user to current cluster.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/clustermembers/{clustermember}").
		To(handler.RemoveClusterMember).
		Doc("Delete member in cluster scope.").
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/clustermembers/{clustermember}").
		To(handler.UpdateClusterMember).
		Doc("Update cluster member role bind.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clustermembers/{clustermember}").
		To(handler.DescribeClusterMember).
		Doc("Retrieve member details in cluster.").
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clustermembers").
		To(handler.ListClusterMembers).
		Doc("List all members in cluster.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers").
		To(handler.ListWorkspaceMembers).
		Doc("List all members in the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(handler.DescribeWorkspaceMember).
		Doc("Retrieve workspace member details.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspace member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/workspacemembers").
		To(handler.CreateWorkspaceMembers).
		Doc("Batch add workspace members.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(handler.UpdateWorkspaceMember).
		Doc("Update member in workspace.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspace member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(handler.RemoveWorkspaceMember).
		Doc("Delete member in workspace scope.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspace member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/members").
		To(handler.ListNamespaceMembers).
		Doc("List all members in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/members/{member}").
		To(handler.DescribeNamespaceMember).
		Doc("Retrieve namespace member details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.POST("/namespaces/{namespace}/members").
		To(handler.CreateNamespaceMembers).
		Doc("Batch add namespace members.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/namespaces/{namespace}/members/{member}").
		To(handler.UpdateNamespaceMember).
		Doc("Update member in namespace.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/namespaces/{namespace}/members/{member}").
		To(handler.RemoveNamespaceMember).
		Doc("Delete member in namespace scope.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/devops/{devops}/members").
		To(handler.ListNamespaceMembers).
		Doc("List all members in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/members/{member}").
		To(handler.DescribeNamespaceMember).
		Doc("Retrieve devops project member details.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.POST("/devops/{devops}/members").
		To(handler.CreateNamespaceMembers).
		Doc("Batch add devops project members.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/devops/{devops}/members/{member}").
		To(handler.UpdateNamespaceMember).
		Doc("Update member in devops project.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/devops/{devops}/members/{member}").
		To(handler.RemoveNamespaceMember).
		Doc("Remove member in namespace.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// globalroles
	ws.Route(ws.POST("/globalroles").
		To(handler.CreateGlobalRole).
		Doc("Create global role. Automatically aggregate policy rules according to annotation.").
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/globalroles/{globalrole}").
		To(handler.DeleteGlobalRole).
		Doc("Delete global role.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/globalroles/{globalrole}").
		To(handler.UpdateGlobalRole).
		Doc("Update global role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PATCH("/globalroles/{globalrole}").
		To(handler.PatchGlobalRole).
		Doc("Patch global role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/globalroles").
		To(handler.ListGlobalRoles).
		Doc("List all global roles.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.GlobalRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/globalroles/{globalrole}").
		To(handler.DescribeGlobalRole).
		Param(ws.PathParameter("globalrole", "global role name")).
		Doc("Retrieve global role details.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// clusterroles
	ws.Route(ws.POST("/clusterroles").
		To(handler.CreateClusterRole).
		Doc("Create cluster role. Automatically aggregate policy rules according to annotation.").
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/clusterroles/{clusterrole}").
		To(handler.DeleteClusterRole).
		Doc("Delete cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/clusterroles/{clusterrole}").
		To(handler.UpdateClusterRole).
		Doc("Update cluster role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PATCH("/clusterroles/{clusterrole}").
		To(handler.PatchClusterRole).
		Doc("Patch cluster role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles").
		To(handler.ListClusterRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.ClusterRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}").
		To(handler.DescribeClusterRole).
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Doc("Retrieve cluster role details.").
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// workspaceroles
	ws.Route(ws.POST("/workspaces/{workspace}/workspaceroles").
		To(handler.CreateWorkspaceRole).
		Doc("Create workspace role. Automatically aggregate policy rules according to annotation.").
		Reads(iamv1alpha2.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.WorkspaceRole{}).
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.DeleteWorkspaceRole).
		Doc("Delete workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PATCH("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.PatchWorkspaceRole).
		Doc("Patch workspace role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Reads(iamv1alpha2.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.UpdateWorkspaceRole).
		Doc("Update workspace role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Reads(iamv1alpha2.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.WorkspaceRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles").
		To(handler.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.WorkspaceRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.DescribeWorkspaceRole).
		Doc("Retrieve workspace role details.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.WorkspaceRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// roles
	ws.Route(ws.POST("/namespaces/{namespace}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified namespace. Automatically aggregate policy rules according to annotation.").
		Reads(rbacv1.Role{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/namespaces/{namespace}/roles/{role}").
		To(handler.DeleteNamespaceRole).
		Doc("Delete role in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/namespaces/{namespace}/roles/{role}").
		To(handler.UpdateNamespaceRole).
		Doc("Update namespace role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PATCH("/namespaces/{namespace}/roles/{role}").
		To(handler.PatchNamespaceRole).
		Doc("Patch namespace role.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve role details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// roles
	ws.Route(ws.POST("/devops/{devops}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified devops project. Automatically aggregate policy rules according to annotation.").
		Reads(rbacv1.Role{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/devops/{devops}/roles/{role}").
		To(handler.DeleteNamespaceRole).
		Doc("Delete role in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/devops/{devops}/roles/{role}").
		To(handler.UpdateNamespaceRole).
		Doc("Update devops project role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PATCH("/devops/{devops}/roles/{role}").
		To(handler.PatchNamespaceRole).
		Doc("Patch devops project role. Automatically aggregate policy rules according to annotation.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve devops project role details.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/users/{user}/globalroles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve user's global role templates.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.GlobalRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/clustermembers/{clustermember}/clusterroles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve user's role templates in cluster.").
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.ClusterRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}/workspaceroles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve member's role templates in workspace.").
		Param(ws.PathParameter("workspace", "workspace")).
		Param(ws.PathParameter("workspacemember", "workspace member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.WorkspaceRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/members/{member}/roles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve member's role templates in namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/members/{member}/roles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve member's role templates in devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	container.Add(ws)
	return nil
}
