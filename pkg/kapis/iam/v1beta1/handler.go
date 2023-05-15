package v1beta1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"kubesphere.io/kubesphere/pkg/api"
	resourcev1beta1 "kubesphere.io/kubesphere/pkg/models/resources/v1beta1"

	"github.com/emicklei/go-restful"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
)

type iamHandler struct {
	im              im.IdentityManagementInterface
	am              am.AccessManagementInterface
	resourceManager resourcev1beta1.ResourceManager
}

func newIAMHandler(im im.IdentityManagementInterface, am am.AccessManagementInterface) *iamHandler {
	return &iamHandler{
		im: im,
		am: am,
	}
}

func (h *iamHandler) ListClusterMembers(request *restful.Request, response *restful.Response) {
	bindings, err := h.am.ListClusterRoleBindings("")
	result := &api.ListResult{Items: make([]interface{}, 0)}
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	for _, binding := range bindings {
		for _, subject := range binding.Subjects {
			if subject.Kind == rbacv1.UserKind {
				user, err := h.im.DescribeUser(subject.Name)
				if err != nil {
					api.HandleError(response, request, err)
					return
				}
				result.Items = append(result.Items, user)
				result.TotalItems += 1
			}
		}
	}

	_ = response.WriteEntity(result)
}

func (h *iamHandler) ListWorkspaceMembers(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	bindings, err := h.am.ListWorkspaceRoleBindings("", nil, workspace)
	result := &api.ListResult{Items: make([]interface{}, 0)}
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	for _, binding := range bindings {
		for _, subject := range binding.Subjects {
			if subject.Kind == rbacv1.UserKind {
				user, err := h.im.DescribeUser(subject.Name)
				if err != nil {
					api.HandleError(response, request, err)
					return
				}
				result.Items = append(result.Items, user)
				result.TotalItems += 1
			}
		}
	}

	_ = response.WriteEntity(result)

}

func (h *iamHandler) ListNamespaceMember(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	bindings, err := h.am.ListRoleBindings("", nil, namespace)
	result := &api.ListResult{Items: make([]interface{}, 0)}
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	for _, binding := range bindings {
		for _, subject := range binding.Subjects {
			if subject.Kind == rbacv1.UserKind {
				user, err := h.im.DescribeUser(subject.Name)
				if err != nil {
					api.HandleError(response, request, err)
					return
				}
				result.Items = append(result.Items, user)
				result.TotalItems += 1
			}
		}
	}

	_ = response.WriteEntity(result)
}

func (h *iamHandler) GetGlobalRoleOfUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("username")
	globalRole, err := h.am.GetGlobalRoleOfUser(username)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	_ = response.WriteEntity(globalRole)
}

func (h *iamHandler) GetWorkspaceRoleOfUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("username")
	workspace := request.PathParameter("workspace")
	result := &api.ListResult{Items: make([]interface{}, 0)}

	workspaceRoles, err := h.am.GetWorkspaceRoleOfUser(username, nil, workspace)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	for _, role := range workspaceRoles {
		result.Items = append(result.Items, role)
		result.TotalItems += 1
	}

	_ = response.WriteEntity(result)
}

func (h *iamHandler) GetClusterRoleOfUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("username")

	clusterRole, err := h.am.GetClusterRoleOfUser(username)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	_ = response.WriteEntity(clusterRole)
}

func (h *iamHandler) GetRoleOfUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("username")
	namespace := request.PathParameter("namespace")
	result := &api.ListResult{Items: make([]interface{}, 0)}

	roles, err := h.am.GetNamespaceRoleOfUser(username, nil, namespace)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	for _, role := range roles {
		result.Items = append(result.Items, role)
		result.TotalItems += 1
	}

	_ = response.WriteEntity(result)

}

func (h *iamHandler) ListRoleTemplateOfUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("username")
	scope := request.QueryParameter("scope")
	namespace := request.QueryParameter("namespace")
	workspace := request.QueryParameter("workspace")
	result := &api.ListResult{Items: make([]interface{}, 0)}
	var roleTemplateNames []string

	switch scope {
	case iamv1beta1.ScopeGlobal:
		globalRole, err := h.am.GetGlobalRoleOfUser(username)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}

		roleTemplateNames = globalRole.AggregationRoleTemplates.TemplateNames
	case iamv1beta1.ScopeWorkspace:
		workspaceRoles, err := h.am.GetWorkspaceRoleOfUser(username, nil, workspace)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}

		for _, workspaceRole := range workspaceRoles {
			roleTemplateNames = append(roleTemplateNames, workspaceRole.AggregationRoleTemplates.TemplateNames...)
		}

	case iamv1beta1.ScopeCluster:
		clusterRole, err := h.am.GetClusterRoleOfUser(username)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}

		roleTemplateNames = clusterRole.AggregationRoleTemplates.TemplateNames

	case iamv1beta1.ScopeNamespace:
		roles, err := h.am.GetNamespaceRoleOfUser(username, nil, namespace)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}

		for _, role := range roles {
			roleTemplateNames = append(roleTemplateNames, role.AggregationRoleTemplates.TemplateNames...)
		}
	}

	for _, name := range roleTemplateNames {
		template, err := h.am.GetRoleTemplate(name)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}

		result.Items = append(result.Items, template)
		result.TotalItems += 1
	}

	_ = response.WriteEntity(result)
}
