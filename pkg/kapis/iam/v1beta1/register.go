/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1beta1

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"

	"kubesphere.io/kubesphere/pkg/api"
	apiserverruntime "kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

var GroupVersion = schema.GroupVersion{Group: "iam.kubesphere.io", Version: "v1beta1"}

func (h *handler) AddToContainer(container *restful.Container) error {
	ws := apiserverruntime.NewWebService(GroupVersion)

	ws.Route(ws.POST("/users").
		To(h.CreateUser).
		Doc("Create user").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagIdentityManagement}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.User{}).
		Reads(iamv1beta1.User{}))
	ws.Route(ws.PUT("/users/{user}").
		To(h.UpdateUser).
		Doc("Update user").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagIdentityManagement}).
		Reads(iamv1beta1.User{}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.User{}))
	ws.Route(ws.DELETE("/users/{user}").
		To(h.DeleteUser).
		Doc("Delete user").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagIdentityManagement}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.None))
	ws.Route(ws.PUT("/users/{user}/password").
		To(h.ModifyPassword).
		Doc("Reset password").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagIdentityManagement}).
		Reads(PasswordReset{}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, errors.None))
	ws.Route(ws.GET("/users/{user}").
		To(h.DescribeUser).
		Doc("Get user").
		Notes("Retrieve user details.").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagIdentityManagement}).
		Param(ws.PathParameter("user", "username")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.User{}))
	ws.Route(ws.GET("/users").
		To(h.ListUsers).
		Doc("List users").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagIdentityManagement}).
		Param(ws.QueryParameter("globalrole", "specific golalrole name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []runtime.Object{&iamv1beta1.User{}}}))
	ws.Route(ws.GET("/users/{user}/loginrecords").
		To(h.ListUserLoginRecords).
		Doc("List login records").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagIdentityManagement}).
		Param(ws.PathParameter("user", "username of the user")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []runtime.Object{&iamv1beta1.LoginRecord{}}}))

	// members
	ws.Route(ws.GET("/clustermembers").
		To(h.ListClusterMembers).
		Doc("List all members of cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.QueryParameter("clusterrole", "specific the cluster role name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []runtime.Object{&iamv1beta1.User{}}}))
	ws.Route(ws.POST("/clustermembers").
		To(h.CreateClusterMembers).
		Doc("Add members to cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}))
	ws.Route(ws.DELETE("/clustermembers/{clustermember}").
		To(h.RemoveClusterMember).
		Doc("Delete member from cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("clustermember", "cluster member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None))
	ws.Route(ws.PUT("/clustermembers/{clustermember}").
		To(h.UpdateClusterMember).
		Doc("Update member from the cluster").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("clustermember", "the member name from cluster")).
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None))

	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers").
		To(h.ListWorkspaceMembers).
		Doc("List all members in the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("workspace", "The specified workspace.")).
		Param(ws.QueryParameter("workspacerole", "specific the workspace role name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []runtime.Object{&iamv1beta1.User{}}}))
	ws.Route(ws.PUT("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(h.UpdateWorkspaceMember).
		Doc("Update member from the workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("workspace", "The specified workspace.")).
		Param(ws.PathParameter("workspacemember", "the member from workspace")).
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None))
	ws.Route(ws.POST("/workspaces/{workspace}/workspacemembers").
		To(h.CreateWorkspaceMembers).
		Doc("Add members to the specified workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("workspace", "The specified workspace.")).
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}))
	ws.Route(ws.DELETE("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(h.RemoveWorkspaceMember).
		Doc("Delete a member from the workspace").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("workspace", "The specified workspace.")).
		Param(ws.PathParameter("workspacemember", "Workspace member's name.")).
		Returns(http.StatusOK, api.StatusOK, errors.None))
	ws.Route(ws.GET("/workspaces/{workspace}/workspacemembers/{workspacemember}").
		To(h.DescribeWorkspaceMember).
		Doc("Get workspace member").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("workspace", "The specified workspace.")).
		Param(ws.PathParameter("workspacemember", "Workspace member's name.")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.User{}))

	ws.Route(ws.GET("/namespaces/{namespace}/namespacemembers").
		To(h.ListNamespaceMembers).
		Doc("List all members in the specified namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.QueryParameter("role", "specific the role name")).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []runtime.Object{&iamv1beta1.User{}}}))
	ws.Route(ws.POST("/namespaces/{namespace}/namespacemembers").
		To(h.CreateNamespaceMembers).
		Doc("Add members to the namespace in bulk.").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Reads([]Member{}).
		Returns(http.StatusOK, api.StatusOK, []Member{}).
		Param(ws.PathParameter("namespace", "The specified namespace.")))
	ws.Route(ws.DELETE("/namespaces/{namespace}/namespacemembers/{member}").
		To(h.RemoveNamespaceMember).
		Doc("Delete a member from the namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("member", "namespace member's username")).
		Returns(http.StatusOK, api.StatusOK, errors.None))
	ws.Route(ws.PUT("/namespaces/{namespace}/namespacemembers/{namespacemember}").
		To(h.UpdateNamespaceMember).
		Doc("Update member from the namespace").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("namespace", "The specified namespace.")).
		Param(ws.PathParameter("namespacemember", "the member from namespace")).
		Reads(Member{}).
		Returns(http.StatusOK, api.StatusOK, errors.None))

	ws.Route(ws.GET("/users/{username}/roletemplates").
		To(h.ListRoleTemplateOfUser).
		Doc("List all role templates of the specified user").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Param(ws.PathParameter("username", "the name of the specified user")).
		Param(ws.QueryParameter("scope", "the scope of role templates")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []runtime.Object{&iamv1beta1.RoleTemplate{}}}))

	ws.Route(ws.POST("/subjectaccessreviews").
		To(h.CreateSubjectAccessReview).
		Doc("Create subject access review").
		Notes("Evaluates all of the request attributes against all policies and allows or denies the request.").
		Metadata(restfulspec.KeyOpenAPITags, []string{api.TagAccessManagement}).
		Reads(iamv1beta1.SubjectAccessReview{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.SubjectAccessReview{}))

	container.Add(ws)
	return nil
}
