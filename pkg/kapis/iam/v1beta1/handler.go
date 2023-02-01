package v1beta1

import (
	"fmt"
	"strings"

	"github.com/emicklei/go-restful"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/iam/v1beta1/am"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
)

type iamHandler struct {
	am am.AccessManagementInterface
}

func newIAMHandler(am am.AccessManagementInterface) *iamHandler {
	return &iamHandler{
		am: am,
	}
}

func (h *iamHandler) CreateCategory(request *restful.Request, response *restful.Response) {
	var category iamv1beta1.Category
	err := request.ReadEntity(&category)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateCategory(&category)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteCategory(request *restful.Request, response *restful.Response) {
	category := request.PathParameter("category")
	err := h.am.DeleteCategory(category)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateCategory(request *restful.Request, response *restful.Response) {
	categoryName := request.PathParameter("category")

	var category iamv1beta1.Category
	err := request.ReadEntity(&category)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if categoryName != category.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", category.Name, categoryName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateCategory(&category)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) ListCategories(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	result, err := h.am.ListCategories(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *iamHandler) DescribeCategory(request *restful.Request, response *restful.Response) {
	categoryName := request.PathParameter("category")
	category, err := h.am.GetCategory(categoryName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(category)
}

func (h *iamHandler) CreateGlobalRole(request *restful.Request, response *restful.Response) {
	var globalRole iamv1beta1.GlobalRole
	err := request.ReadEntity(&globalRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateGlobalRole(&globalRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteGlobalRole(request *restful.Request, response *restful.Response) {
	globalRole := request.PathParameter("globalrole")
	err := h.am.DeleteGlobalRole(globalRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")

	var globalRole iamv1beta1.GlobalRole
	err := request.ReadEntity(&globalRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if globalRoleName != globalRole.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", globalRole.Name, globalRoleName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateGlobalRole(&globalRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) PatchGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")

	var globalRole iamv1beta1.GlobalRole
	err := request.ReadEntity(&globalRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	globalRole.Name = globalRoleName
	patched, err := h.am.PatchGlobalRole(&globalRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *iamHandler) ListGlobalRoles(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	result, err := h.am.ListGlobalRoles(queryParam)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *iamHandler) DescribeGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")
	globalRole, err := h.am.GetGlobalRole(globalRoleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(globalRole)
}

func (h *iamHandler) CreateClusterRole(request *restful.Request, response *restful.Response) {
	var clusterRole iamv1beta1.ClusterRole
	err := request.ReadEntity(&clusterRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateClusterRole(&clusterRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteClusterRole(request *restful.Request, response *restful.Response) {
	clusterrole := request.PathParameter("clusterrole")

	err := h.am.DeleteClusterRole(clusterrole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")

	var clusterRole iamv1beta1.ClusterRole

	err := request.ReadEntity(&clusterRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if clusterRoleName != clusterRole.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", clusterRole.Name, clusterRoleName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateClusterRole(&clusterRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) PatchClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")

	var clusterRole iamv1beta1.ClusterRole
	err := request.ReadEntity(&clusterRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	clusterRole.Name = clusterRoleName
	patched, err := h.am.PatchClusterRole(&clusterRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *iamHandler) ListClusterRoles(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	result, err := h.am.ListClusterRoles(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *iamHandler) DescribeClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")
	clusterRole, err := h.am.GetClusterRole(clusterRoleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(clusterRole)
}

func (h *iamHandler) CreateNamespaceRole(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	var role iamv1beta1.Role
	err := request.ReadEntity(&role)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateNamespaceRole(namespace, &role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteNamespaceRole(request *restful.Request, response *restful.Response) {
	role := request.PathParameter("role")
	namespace := request.PathParameter("namespace")

	err := h.am.DeleteNamespaceRole(namespace, role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) PatchNamespaceRole(request *restful.Request, response *restful.Response) {
	roleName := request.PathParameter("role")
	namespaceName := request.PathParameter("namespace")

	var role iamv1beta1.Role
	err := request.ReadEntity(&role)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	role.Name = roleName
	patched, err := h.am.PatchNamespaceRole(namespaceName, &role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *iamHandler) UpdateNamespaceRole(request *restful.Request, response *restful.Response) {
	roleName := request.PathParameter("role")
	namespace := request.PathParameter("namespace")

	var role iamv1beta1.Role
	err := request.ReadEntity(&role)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if roleName != role.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", role.Name, roleName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateNamespaceRole(namespace, &role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) ListRoles(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	queryParam := query.ParseQueryParameter(request)
	result, err := h.am.ListRoles(namespace, queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *iamHandler) DescribeNamespaceRole(request *restful.Request, response *restful.Response) {
	roleName := request.PathParameter("role")
	namespace := request.PathParameter("namespace")

	role, err := h.am.GetNamespaceRole(namespace, roleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(role)
}

func (h *iamHandler) ListWorkspaceRoles(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	workspace := request.PathParameter("workspace")

	queryParam.Filters[iamv1alpha2.ScopeWorkspace] = query.Value(workspace)
	// shared workspace role template
	if string(queryParam.Filters[query.FieldLabel]) == fmt.Sprintf("%s=%s", iamv1alpha2.RoleTemplateLabel, "true") ||
		strings.Contains(queryParam.LabelSelector, iamv1alpha2.RoleTemplateLabel) ||
		queryParam.Filters[iamv1alpha2.AggregateTo] != "" {
		delete(queryParam.Filters, iamv1alpha2.ScopeWorkspace)
	}

	result, err := h.am.ListWorkspaceRoles(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *iamHandler) PatchWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	workspaceRoleName := request.PathParameter("workspacerole")

	var workspaceRole iamv1beta1.WorkspaceRole
	err := request.ReadEntity(&workspaceRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	workspaceRole.Name = workspaceRoleName
	patched, err := h.am.PatchWorkspaceRole(workspaceName, &workspaceRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *iamHandler) DescribeWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	workspaceRoleName := request.PathParameter("workspacerole")
	workspaceRole, err := h.am.GetWorkspaceRole(workspace, workspaceRoleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(workspaceRole)
}

func (h *iamHandler) UpdateWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	workspaceRoleName := request.PathParameter("workspacerole")

	var workspaceRole iamv1beta1.WorkspaceRole
	err := request.ReadEntity(&workspaceRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if workspaceRoleName != workspaceRole.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", workspaceRole.Name, workspaceRoleName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateWorkspaceRole(workspace, &workspaceRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) CreateWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")

	var workspaceRole iamv1beta1.WorkspaceRole
	err := request.ReadEntity(&workspaceRole)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateWorkspaceRole(workspace, &workspaceRole)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	workspaceRoleName := request.PathParameter("workspacerole")

	err := h.am.DeleteWorkspaceRole(workspace, workspaceRoleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) GetRoleTemplate(request *restful.Request, response *restful.Response) {
}

func (h *iamHandler) ListRoleTemplates(request *restful.Request, response *restful.Response) {

}

func (h *iamHandler) CreateRoleTemplate(request *restful.Request, response *restful.Response) {

}

func (h *iamHandler) DeleteRoleTemplate(request *restful.Request, response *restful.Response) {

}

func (h *iamHandler) UpdateRoleTemplate(request *restful.Request, response *restful.Response) {

}

func (h *iamHandler) CreateRoleBinding(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	var roleBindings []iamv1beta1.RoleBinding
	err := request.ReadEntity(&roleBindings)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	var results []iamv1beta1.RoleBinding
	for _, item := range roleBindings {
		r, err := h.am.CreateRoleBinding(namespace, &item)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		results = append(results, *r)
	}

	response.WriteEntity(results)
}

func (h *iamHandler) DeleteRoleBinding(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("rolebinding")
	namespace := request.PathParameter("namespace")

	err := h.am.DeleteRoleBinding(namespace, name)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) CreateWorkspaceRoleBinding(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")

	var roleBindings []iamv1beta1.WorkspaceRoleBinding
	err := request.ReadEntity(&roleBindings)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	var results []iamv1beta1.WorkspaceRoleBinding
	for _, item := range roleBindings {
		r, err := h.am.CreateWorkspaceRoleBinding(workspaceName, &item)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		results = append(results, *r)
	}

	response.WriteEntity(results)
}

func (h *iamHandler) DeleteWorkspaceRoleBinding(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	name := request.PathParameter("rolebinding")

	err := h.am.DeleteWorkspaceRoleBinding(workspaceName, name)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}
