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
		To(handler.CreateUserOrClusterMembers).
		Doc("Create user in global scope.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/users/{user}").
		To(handler.DeleteUserOrClusterMember).
		Doc("Delete user.").
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/users/{user}").
		To(handler.UpdateUserOrClusterMember).
		Doc("Update user info.").
		Reads(iamv1alpha2.User{}).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/users/{user}").
		To(handler.DescribeUserOrClusterMember).
		Doc("Retrieve user details.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.User{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/users").
		To(handler.ListUsersOrClusterMembers).
		Doc("List all users.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1alpha2.User{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/workspaces/{workspace}/users").
		To(handler.ListWorkspaceMembers).
		Doc("List all members in the specified workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/users/{user}").
		To(handler.DescribeWorkspaceMember).
		Doc("Retrieve workspace member details.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.POST("/workspaces/{workspace}/users").
		To(handler.CreateWorkspaceMembers).
		Doc("Batch add workspace members.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/workspaces/{workspace}/users/{user}").
		To(handler.UpdateWorkspaceMember).
		Doc("Update member in workspace.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/users/{user}").
		To(handler.RemoveWorkspaceMember).
		Doc("Remove member in workspace.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/namespaces/{namespace}/users").
		To(handler.ListNamespaceMembers).
		Doc("List all members in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/users/{user}").
		To(handler.DescribeNamespaceMember).
		Doc("Retrieve namespace member details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.POST("/namespaces/{namespace}/users").
		To(handler.CreateNamespaceMembers).
		Doc("Batch add namespace members.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Param(ws.PathParameter("namespace", "namespace")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/namespaces/{namespace}/users/{user}").
		To(handler.UpdateNamespaceMember).
		Doc("Update member in namespace.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/namespaces/{namespace}/users/{user}").
		To(handler.RemoveNamespaceMember).
		Doc("Remove member in namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/devops/{devops}/users").
		To(handler.ListNamespaceMembers).
		Doc("List all members in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/users/{user}").
		To(handler.DescribeNamespaceMember).
		Doc("Retrieve namespace member details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.POST("/devops/{devops}/users").
		To(handler.CreateNamespaceMembers).
		Doc("Batch add namespace members.").
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Param(ws.PathParameter("namespace", "namespace")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/devops/{devops}/users/{user}").
		To(handler.UpdateNamespaceMember).
		Doc("Update member in namespace.").
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/devops/{devops}/users/{user}").
		To(handler.RemoveNamespaceMember).
		Doc("Remove member in namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("user", "username")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// globalroles
	ws.Route(ws.POST("/globalroles").
		To(handler.CreateGlobalRole).
		Doc("Create global role.").
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
		Doc("Update global role.").
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
		Doc("Retrieve global role details.").
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// clusterroles
	ws.Route(ws.POST("/clusterroles").
		To(handler.CreateClusterRole).
		Doc("Create cluster role.").
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
		Doc("Update cluster role.").
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
		Doc("Retrieve cluster role details.").
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// workspaceroles
	ws.Route(ws.POST("/workspaces/{workspace}/workspaceroles").
		To(handler.CreateWorkspaceRole).
		Doc("Create workspace role.").
		Reads(iamv1alpha2.WorkspaceRole{}).
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.DeleteWorkspaceRole).
		Doc("Delete workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.UpdateWorkspaceRole).
		Doc("Update workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles").
		To(handler.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.DescribeWorkspaceRole).
		Doc("Retrieve workspace role details.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// roles
	ws.Route(ws.POST("/namespaces/{namespace}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified namespace.").
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
		Doc("Update namespace role.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
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
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	// roles
	ws.Route(ws.POST("/devops/{devops}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified  devops project.").
		Reads(rbacv1.Role{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.DELETE("/devops/{devops}/roles/{role}").
		To(handler.DeleteNamespaceRole).
		Doc("Delete role in the specified devops project.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.PUT("/devops/{devops}/roles/{role}").
		To(handler.UpdateNamespaceRole).
		Doc("Update devops project role.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(rbacv1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{rbacv1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve role details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	ws.Route(ws.GET("/users/{user}/globalrole").
		To(handler.RetrieveMemberRole).
		Doc("Retrieve user's global role.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/users/{user}/clusterrole").
		To(handler.RetrieveMemberRole).
		Doc("Retrieve user's role in cluster.").
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/users/{user}/workspacerole").
		To(handler.RetrieveMemberRole).
		Doc("Retrieve member's role in workspace.").
		Param(ws.PathParameter("workspace", "workspace")).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1alpha2.WorkspaceRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/users/{user}/role").
		To(handler.RetrieveMemberRole).
		Doc("Retrieve member's role in namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, rbacv1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AccessManagementTag}))

	container.Add(ws)
	return nil
}
