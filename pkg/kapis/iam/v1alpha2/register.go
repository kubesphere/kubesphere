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
	"net/http"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/group"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

const (
	GroupName = "iam.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(container *restful.Container, im im.IdentityManagementInterface, am am.AccessManagementInterface, group group.GroupOperator, authorizer authorizer.Authorizer) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newIAMHandler(im, am, group, authorizer)

	// users
	ws.Route(ws.POST("/users").
		To(handler.CreateUser).
		Doc("Create a global user account.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Reads(iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))
	ws.Route(ws.DELETE("/users/{user}").
		To(handler.DeleteUser).
		Doc("Delete the specified user.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))
	ws.Route(ws.PUT("/users/{user}").
		To(handler.UpdateUser).
		Doc("Update user profile.").
		Reads(iamv1alpha2.User{}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))
	ws.Route(ws.PUT("/users/{user}/password").
		To(handler.ModifyPassword).
		Doc("Reset password of the specified user.").
		Reads(PasswordReset{}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))
	ws.Route(ws.GET("/users/{user}").
		To(handler.DescribeUser).
		Doc("Retrieve user details.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))
	ws.Route(ws.GET("/users").
		To(handler.ListUsers).
		Doc("List all users.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserTag}))

	ws.Route(ws.GET("/users/{user}/loginrecords").
		To(handler.ListUserLoginRecords).
		Param(ws.PathParameter("user", "username of the user")).
		Doc("List login records of the specified user.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.LoginRecord{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.UserResourceTag}))

	// clustermembers
	ws.Route(ws.POST("/clustermembers").
		To(handler.CreateClusterMembers).
		Doc("Add members to current cluster in bulk.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMemberTag}))
	ws.Route(ws.DELETE("/clustermembers/{clustermember}").
		To(handler.RemoveClusterMember).
		Doc("Delete a member from current cluster.").
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMemberTag}))
	ws.Route(ws.PUT("/clustermembers/{clustermember}").
		To(handler.UpdateClusterMember).
		Doc("Update the cluster role bind of the member.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMemberTag}))
	ws.Route(ws.GET("/clustermembers/{clustermember}").
		To(handler.DescribeClusterMember).
		Doc("Retrieve the cluster role of the specified member.").
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMemberTag}))
	ws.Route(ws.GET("/clustermembers").
		To(handler.ListClusterMembers).
		Doc("List all members in cluster.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterMemberTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers").
		To(handler.ListWorkspaceMembers).
		Doc("List all members in the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMemberTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(handler.DescribeWorkspaceMember).
		Doc("Retrieve the workspace role of the specified member.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspace member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMemberTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/workspacemembers").
		To(handler.CreateWorkspaceMembers).
		Doc("Add members to current cluster in bulk.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMemberTag}))
	ws.Route(ws.PUT("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(handler.UpdateWorkspaceMember).
		Doc("Update the workspace role bind of the member.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspace member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMemberTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(handler.RemoveWorkspaceMember).
		Doc("Delete a member from the workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacemember", "workspace member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMemberTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/members").
		To(handler.ListNamespaceMembers).
		Doc("List all members in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceMemberTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/members/{member}").
		To(handler.DescribeNamespaceMember).
		Doc("Retrieve the role of the specified member.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceMemberTag}))
	ws.Route(ws.POST("/namespaces/{namespace}/members").
		To(handler.CreateNamespaceMembers).
		Doc("Add members to the namespace in bulk.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceMemberTag}))
	ws.Route(ws.PUT("/namespaces/{namespace}/members/{member}").
		To(handler.UpdateNamespaceMember).
		Doc("Update the role bind of the member.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceMemberTag}))
	ws.Route(ws.DELETE("/namespaces/{namespace}/members/{member}").
		To(handler.RemoveNamespaceMember).
		Doc("Delete a member from the namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceMemberTag}))

	ws.Route(ws.GET("/devops/{devops}/members").
		To(handler.ListNamespaceMembers).
		Doc("List all members in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))
	ws.Route(ws.GET("/devops/{devops}/members/{member}").
		To(handler.DescribeNamespaceMember).
		Doc("Retrieve devops project member details.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))
	ws.Route(ws.POST("/devops/{devops}/members").
		To(handler.CreateNamespaceMembers).
		Doc("Add members to the DevOps project in bulk.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))
	ws.Route(ws.PUT("/devops/{devops}/members/{member}").
		To(handler.UpdateNamespaceMember).
		Doc("Update the role bind of the member.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, Member{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))
	ws.Route(ws.DELETE("/devops/{devops}/members/{member}").
		To(handler.RemoveNamespaceMember).
		Doc("Delete a member from the DevOps project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectMemberTag}))

	// globalroles
	ws.Route(ws.POST("/globalroles").
		To(handler.CreateGlobalRole).
		Doc("Create global role.").
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.DELETE("/globalroles/{globalrole}").
		To(handler.DeleteGlobalRole).
		Doc("Delete global role.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.PUT("/globalroles/{globalrole}").
		To(handler.UpdateGlobalRole).
		Doc("Update global role.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.PATCH("/globalroles/{globalrole}").
		To(handler.PatchGlobalRole).
		Doc("Patch global role.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Reads(iamv1alpha2.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.GET("/globalroles").
		To(handler.ListGlobalRoles).
		Doc("List all global roles.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.GlobalRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.GET("/globalroles/{globalrole}").
		To(handler.DescribeGlobalRole).
		Param(ws.PathParameter("globalrole", "global role name")).
		Doc("Retrieve global role details.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))

	// clusterroles
	ws.Route(ws.POST("/clusterroles").
		To(handler.CreateClusterRole).
		Doc("Create cluster role.").
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.DELETE("/clusterroles/{clusterrole}").
		To(handler.DeleteClusterRole).
		Doc("Delete cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.PUT("/clusterroles/{clusterrole}").
		To(handler.UpdateClusterRole).
		Doc("Update cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.PATCH("/clusterroles/{clusterrole}").
		To(handler.PatchClusterRole).
		Doc("Patch cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.GET("/clusterroles").
		To(handler.ListClusterRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.ClusterRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}").
		To(handler.DescribeClusterRole).
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Doc("Retrieve cluster role details.").
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))

	// workspaceroles
	ws.Route(ws.POST("/workspaces/{workspace}/workspaceroles").
		To(handler.CreateWorkspaceRole).
		Doc("Create workspace role.").
		Reads(iamv1alpha2.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.WorkspaceRole{}).
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.DeleteWorkspaceRole).
		Doc("Delete workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.PATCH("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.PatchWorkspaceRole).
		Doc("Patch workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Reads(iamv1alpha2.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.PUT("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.UpdateWorkspaceRole).
		Doc("Update workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Reads(iamv1alpha2.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.WorkspaceRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles").
		To(handler.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.WorkspaceRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.DescribeWorkspaceRole).
		Doc("Retrieve workspace role details.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.WorkspaceRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))

	// roles
	ws.Route(ws.POST("/namespaces/{namespace}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified namespace.").
		Reads(rbacv1.Role{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.DELETE("/namespaces/{namespace}/roles/{role}").
		To(handler.DeleteNamespaceRole).
		Doc("Delete role in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.PUT("/namespaces/{namespace}/roles/{role}").
		To(handler.UpdateNamespaceRole).
		Doc("Update namespace role.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.PATCH("/namespaces/{namespace}/roles/{role}").
		To(handler.PatchNamespaceRole).
		Doc("Patch namespace role.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve role details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))

	// roles
	ws.Route(ws.POST("/devops/{devops}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified devops project.").
		Reads(rbacv1.Role{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.DELETE("/devops/{devops}/roles/{role}").
		To(handler.DeleteNamespaceRole).
		Doc("Delete role in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.PUT("/devops/{devops}/roles/{role}").
		To(handler.UpdateNamespaceRole).
		Doc("Update devops project role.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.PATCH("/devops/{devops}/roles/{role}").
		To(handler.PatchNamespaceRole).
		Doc("Patch devops project role.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.Role{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.GET("/devops/{devops}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve devops project role details.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))

	ws.Route(ws.GET("/users/{user}/globalroles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve user's global role templates.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.GlobalRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.GET("/clustermembers/{clustermember}/clusterroles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve user's role templates in cluster.").
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.ClusterRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}/workspaceroles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve member's role templates in workspace.").
		Param(ws.PathParameter("workspace", "workspace")).
		Param(ws.PathParameter("workspacemember", "workspace member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.WorkspaceRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/members/{member}/roles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve member's role templates in namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.GET("/devops/{devops}/members/{member}/roles").
		To(handler.RetrieveMemberRoleTemplates).
		Doc("Retrieve member's role templates in devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("member", "devops project member's username")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/groups").
		To(handler.ListWorkspaceGroups).
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Doc("List groups of the specified workspace.").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/groups/{group}").
		To(handler.DescribeGroup).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("group", "group name")).
		Doc("Retrieve group details.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.Group{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.DELETE("/workspaces/{workspace}/groups/{group}").
		To(handler.DeleteGroup).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("group", "group name")).
		Doc("Delete group.").
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.POST("/workspaces/{workspace}/groups").
		To(handler.CreateGroup).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create Group").
		Reads(iamv1alpha2.Group{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.Group{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.PUT("/workspaces/{workspace}/groups/{group}/").
		To(handler.UpdateGroup).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("group", "group name")).
		Doc("Update Group").
		Reads(iamv1alpha2.Group{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.Group{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.PATCH("/workspaces/{workspace}/groups/{group}/").
		To(handler.PatchGroup).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Patch Group").
		Reads(iamv1alpha2.Group{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.Group{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/groupbindings").
		To(handler.ListGroupBindings).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("group", "group name")).
		Doc("Retrieve group's members in the workspace.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/rolebindings").
		To(handler.ListGroupRoleBindings).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("group", "group name")).
		Doc("Retrieve group's rolebindings of all projects in the workspace.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacerolebindings").
		To(handler.ListGroupWorkspaceRoleBindings).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("group", "group name")).
		Doc("Retrieve group's workspacerolebindings of the workspace.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.DELETE("/workspaces/{workspace}/groupbindings/{groupbinding}").
		To(handler.DeleteGroupBinding).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("groupbinding", "groupbinding name")).
		Doc("Delete GroupBinding to remove user from the group.").
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.POST("/workspaces/{workspace}/groupbindings").
		To(handler.CreateGroupBinding).
		Param(ws.PathParameter("workspace", "workspace name")).
		Doc("Create GroupBinding to add a user to the group").
		Reads([]GroupMember{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GroupBinding{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	// namespace rolebinding
	ws.Route(ws.POST("/namespaces/{namespace}/rolebindings").
		To(handler.CreateRoleBinding).
		Doc("Create rolebinding in the specified namespace.").
		Reads([]v1.RoleBinding{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, []v1.RoleBinding{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/rolebindings/{rolebinding}").
		To(handler.DeleteRoleBinding).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("namespace", "groupbinding name")).
		Param(ws.PathParameter("rolebinding", "groupbinding name")).
		Doc("Delete rolebinding under namespace.").
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	// workspace rolebinding
	ws.Route(ws.POST("/workspaces/{workspace}/workspacerolebindings").
		To(handler.CreateWorkspaceRoleBinding).
		Param(ws.PathParameter("workspace", "workspace name")).
		Reads([]iamv1alpha2.WorkspaceRoleBinding{}).
		Doc("Create group's workspacerolebindings of the workspace.").
		Returns(http.StatusOK, api.StatusOK, []iamv1alpha2.WorkspaceRoleBinding{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	ws.Route(ws.DELETE("/workspaces/{workspace}/workspacerolebindings/{rolebinding}").
		To(handler.DeleteWorkspaceRoleBinding).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("rolebinding", "groupbinding name")).
		Doc("Delete workspacerolebinding.").
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GroupTag}))

	container.Add(ws)
	return nil
}
