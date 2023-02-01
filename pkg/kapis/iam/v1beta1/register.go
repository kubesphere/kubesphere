package v1beta1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/iam/v1beta1/am"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

const (
	GroupName = "iam.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1beta1"}

func AddToContainer(container *restful.Container, am am.AccessManagementInterface) error {
	ws := runtime.NewWebService(GroupVersion)
	handler := newIAMHandler(am)

	// category
	ws.Route(ws.POST("/categories").
		To(handler.CreateCategory).
		Doc("Create category.").
		Reads(iamv1beta1.Category{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Category{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.CategoryTag}))
	ws.Route(ws.DELETE("/categories/{category}").
		To(handler.DeleteCategory).
		Doc("Delete category.").
		Param(ws.PathParameter("category", "category name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.CategoryTag}))
	ws.Route(ws.PUT("/categories/{category}").
		To(handler.UpdateCategory).
		Doc("Update category.").
		Param(ws.PathParameter("category", "category name")).
		Reads(iamv1beta1.Category{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Category{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.CategoryTag}))
	ws.Route(ws.GET("/categories").
		To(handler.ListCategories).
		Doc("List all categories.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1beta1.Category{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.CategoryTag}))
	ws.Route(ws.GET("/categories/{category}").
		To(handler.DescribeCategory).
		Param(ws.PathParameter("category", "category name")).
		Doc("Retrieve category details.").
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Category{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.CategoryTag}))

	// globalroles
	ws.Route(ws.POST("/globalroles").
		To(handler.CreateGlobalRole).
		Doc("Create global role.").
		Reads(iamv1beta1.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.GlobalRole{}).
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
		Reads(iamv1beta1.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.PATCH("/globalroles/{globalrole}").
		To(handler.PatchGlobalRole).
		Doc("Patch global role.").
		Param(ws.PathParameter("globalrole", "global role name")).
		Reads(iamv1beta1.GlobalRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.GET("/globalroles").
		To(handler.ListGlobalRoles).
		Doc("List all global roles.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1beta1.GlobalRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))
	ws.Route(ws.GET("/globalroles/{globalrole}").
		To(handler.DescribeGlobalRole).
		Param(ws.PathParameter("globalrole", "global role name")).
		Doc("Retrieve global role details.").
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.GlobalRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.GlobalRoleTag}))

	// clusterroles
	ws.Route(ws.POST("/clusterroles").
		To(handler.CreateClusterRole).
		Doc("Create cluster role.").
		Reads(iamv1beta1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.ClusterRole{}).
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
		Reads(iamv1beta1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.PATCH("/clusterroles/{clusterrole}").
		To(handler.PatchClusterRole).
		Doc("Patch cluster role.").
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Reads(iamv1beta1.ClusterRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.GET("/clusterroles").
		To(handler.ListClusterRoles).
		Doc("List all cluster roles.").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1beta1.ClusterRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))
	ws.Route(ws.GET("/clusterroles/{clusterrole}").
		To(handler.DescribeClusterRole).
		Param(ws.PathParameter("clusterrole", "cluster role name")).
		Doc("Retrieve cluster role details.").
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.ClusterRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ClusterRoleTag}))

	// workspaceroles
	ws.Route(ws.POST("/workspaces/{workspace}/workspaceroles").
		To(handler.CreateWorkspaceRole).
		Doc("Create workspace role.").
		Reads(iamv1beta1.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.WorkspaceRole{}).
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
		Reads(iamv1beta1.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.PUT("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.UpdateWorkspaceRole).
		Doc("Update workspace role.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Reads(iamv1beta1.WorkspaceRole{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.WorkspaceRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles").
		To(handler.ListWorkspaceRoles).
		Doc("List all workspace roles.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1beta1.WorkspaceRole{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))
	ws.Route(ws.GET("/workspaces/{workspace}/workspaceroles/{workspacerole}").
		To(handler.DescribeWorkspaceRole).
		Doc("Retrieve workspace role details.").
		Param(ws.PathParameter("workspace", "workspace name")).
		Param(ws.PathParameter("workspacerole", "workspace role name")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.WorkspaceRole{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceRoleTag}))

	// roles
	ws.Route(ws.POST("/namespaces/{namespace}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified namespace.").
		Reads(iamv1beta1.Role{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Role{}).
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
		Reads(iamv1beta1.Role{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.PATCH("/namespaces/{namespace}/roles/{role}").
		To(handler.PatchNamespaceRole).
		Doc("Patch namespace role.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Reads(iamv1beta1.Role{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1beta1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve role details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.NamespaceRoleTag}))

	// roles
	ws.Route(ws.POST("/devops/{devops}/roles").
		To(handler.CreateNamespaceRole).
		Doc("Create role in the specified devops project.").
		Reads(iamv1beta1.Role{}).
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Role{}).
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
		Reads(iamv1beta1.Role{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.PATCH("/devops/{devops}/roles/{role}").
		To(handler.PatchNamespaceRole).
		Doc("Patch devops project role.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Reads(iamv1beta1.Role{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.GET("/devops/{devops}/roles").
		To(handler.ListRoles).
		Doc("List all roles in the specified devops project.").
		Param(ws.PathParameter("devops", "devops project name")).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1beta1.Role{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))
	ws.Route(ws.GET("/devops/{devops}/roles/{role}").
		To(handler.DescribeNamespaceRole).
		Doc("Retrieve devops project role details.").
		Param(ws.PathParameter("devops", "devops project name")).
		Param(ws.PathParameter("role", "role name")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.Role{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.DevOpsProjectRoleTag}))

	ws.Route(ws.GET("/roletemplates/{roletemplate}").
		To(handler.GetRoleTemplate).
		Doc("Get role template").
		Param(ws.PathParameter("roletemplate", "role template name")).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.RoleTemplate{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.RoleTemplateTag}))

	ws.Route(ws.GET("/roletemplate").
		To(handler.ListRoleTemplates).
		Doc("List role templates").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{iamv1beta1.RoleTemplate{}}}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.RoleTemplateTag}))

	ws.Route(ws.POST("/roletemplate").
		To(handler.CreateRoleTemplate).
		Doc("Create role template.").
		Reads(iamv1beta1.RoleTemplate{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.RoleTemplate{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.RoleTemplateTag}))

	ws.Route(ws.DELETE("/roletemplates/{roletemplate}").
		To(handler.DeleteRoleTemplate).
		Doc("Delete role template.").
		Param(ws.PathParameter("roletemplate", "role template name")).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.RoleTemplateTag}))

	ws.Route(ws.PUT("/roletemplates/{roletemplate}").
		To(handler.UpdateRoleTemplate).
		Doc("Update role template.").
		Param(ws.PathParameter("roletemplate", "role template name")).
		Reads(iamv1beta1.RoleTemplate{}).
		Returns(http.StatusOK, api.StatusOK, iamv1beta1.RoleTemplate{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.RoleTemplateTag}))

	// namespace rolebinding
	ws.Route(ws.POST("/namespaces/{namespace}/rolebindings").
		To(handler.CreateRoleBinding).
		Doc("Create rolebinding in the specified namespace.").
		Reads([]iamv1beta1.RoleBinding{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, []iamv1beta1.RoleBinding{}).
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
		Reads([]iamv1beta1.WorkspaceRoleBinding{}).
		Doc("Create group's workspacerolebindings of the workspace.").
		Returns(http.StatusOK, api.StatusOK, []iamv1beta1.WorkspaceRoleBinding{}).
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
