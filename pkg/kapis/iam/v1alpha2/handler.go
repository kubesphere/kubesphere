package v1alpha2

import (
	"fmt"
	"github.com/emicklei/go-restful"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	apirequeset "kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
	"strings"
)

type iamHandler struct {
	am am.AccessManagementInterface
	im im.IdentityManagementInterface
}

func newIAMHandler(im im.IdentityManagementInterface, am am.AccessManagementInterface, options *authoptions.AuthenticationOptions) *iamHandler {
	return &iamHandler{
		am: am,
		im: im,
	}
}

type Member struct {
	Username string `json:"username"`
	RoleRef  string `json:"roleRef"`
}

func (h *iamHandler) DescribeUserOrClusterMember(request *restful.Request, response *restful.Response) {
	requestInfo, ok := apirequeset.RequestInfoFrom(request.Request.Context())

	if ok && requestInfo.ResourceScope == apirequeset.ClusterScope {
		h.DescribeClusterMember(request, response)
		return
	}

	username := request.PathParameter("user")

	user, err := h.im.DescribeUser(username)

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	globalRole, err := h.am.GetGlobalRoleOfUser(username)

	if err != nil && !errors.IsNotFound(err) {
		api.HandleInternalError(response, request, err)
		return
	}

	if globalRole != nil {
		if user.Annotations == nil {
			user.Annotations = make(map[string]string, 0)
		}
		user.Annotations[iamv1alpha2.GlobalRoleAnnotation] = globalRole.Name
	}

	response.WriteEntity(user)
}

func (h *iamHandler) RetrieveMemberRole(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")

	if strings.HasSuffix(req.Request.URL.Path, iamv1alpha2.ResourcesSingularGlobalRole) {
		globalRole, err := h.am.GetGlobalRoleOfUser(username)

		if err != nil {
			api.HandleInternalError(resp, req, err)
			return
		}
		resp.WriteEntity(globalRole)
		return
	}

	if strings.HasSuffix(req.Request.URL.Path, iamv1alpha2.ResourcesSingularClusterRole) {
		clusterRole, err := h.am.GetClusterRoleOfUser(username)

		if err != nil {
			api.HandleInternalError(resp, req, err)
			return
		}
		resp.WriteEntity(clusterRole)
		return
	}

	if strings.HasSuffix(req.Request.URL.Path, iamv1alpha2.ResourcesSingularWorkspaceRole) {
		workspace := req.PathParameter("workspace")

		workspaceRole, err := h.am.GetWorkspaceRoleOfUser(username, workspace)

		if err != nil {
			api.HandleInternalError(resp, req, err)
			return
		}

		resp.WriteEntity(workspaceRole)
		return
	}

	if strings.HasSuffix(req.Request.URL.Path, iamv1alpha2.ResourcesSingularRole) {
		namespace := req.PathParameter("namespace")

		role, err := h.am.GetNamespaceRoleOfUser(username, namespace)

		if err != nil {
			api.HandleInternalError(resp, req, err)
			return
		}
		resp.WriteEntity(role)
		return
	}
}

func (h *iamHandler) ListUsersOrClusterMembers(request *restful.Request, response *restful.Response) {
	requestInfo, ok := apirequeset.RequestInfoFrom(request.Request.Context())

	if ok && requestInfo.ResourceScope == apirequeset.ClusterScope {
		h.ListClusterMembers(request, response)
		return
	}

	queryParam := query.ParseQueryParameter(request)
	result, err := h.im.ListUsers(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	for i, item := range result.Items {
		user := item.(*iamv1alpha2.User)
		user = user.DeepCopy()
		globalRole, err := h.am.GetGlobalRoleOfUser(user.Name)

		if err != nil && !errors.IsNotFound(err) {
			klog.Error(err)
			api.HandleInternalError(response, request, err)
			return
		}

		if globalRole != nil {

			if user.Annotations == nil {
				user.Annotations = make(map[string]string, 0)
			}
			user.Annotations[iamv1alpha2.GlobalRoleAnnotation] = globalRole.Name
		}

		result.Items[i] = user
	}
	response.WriteEntity(result)
}

func (h *iamHandler) ListRoles(request *restful.Request, response *restful.Response) {

	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	queryParam := query.ParseQueryParameter(request)
	result, err := h.am.ListRoles(namespace, queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	response.WriteEntity(result)
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

func (h *iamHandler) ListGlobalRoles(req *restful.Request, resp *restful.Response) {
	queryParam := query.ParseQueryParameter(req)
	result, err := h.am.ListGlobalRoles(queryParam)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	resp.WriteEntity(result)
}

func (h *iamHandler) ListNamespaceMembers(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	queryParam.Filters[iamv1alpha2.ScopeNamespace] = query.Value(namespace)

	result, err := h.im.ListUsers(queryParam)

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *iamHandler) DescribeNamespaceMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	queryParam := query.New()
	queryParam.Filters[query.FieldName] = query.Value(username)
	queryParam.Filters[iamv1alpha2.ScopeNamespace] = query.Value(namespace)

	result, err := h.im.ListUsers(queryParam)

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if len(result.Items) == 0 {
		err := errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularUser), username)
		api.HandleNotFound(response, request, err)
		return
	}

	response.WriteEntity(result.Items[0])
}

func (h *iamHandler) ListWorkspaceRoles(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	workspace := request.PathParameter("workspace")

	queryParam.Filters[iamv1alpha2.ScopeWorkspace] = query.Value(workspace)

	// shared workspace role template
	if string(queryParam.Filters[query.FieldLabel]) == fmt.Sprintf("%s=%s", iamv1alpha2.RoleTemplateLabel, "true") ||
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

func (h *iamHandler) ListWorkspaceMembers(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	workspace := request.PathParameter("workspace")
	queryParam.Filters[iamv1alpha2.ScopeWorkspace] = query.Value(workspace)

	result, err := h.im.ListUsers(queryParam)

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *iamHandler) DescribeWorkspaceMember(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	username := request.PathParameter("user")

	queryParam := query.New()
	queryParam.Filters[query.FieldName] = query.Value(username)
	queryParam.Filters[iamv1alpha2.ScopeWorkspace] = query.Value(workspace)

	result, err := h.im.ListUsers(queryParam)

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if len(result.Items) == 0 {
		err := errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularUser), username)
		api.HandleNotFound(response, request, err)
		return
	}

	response.WriteEntity(result.Items[0])
}

func (h *iamHandler) UpdateWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	workspaceRoleName := request.PathParameter("workspacerole")

	var workspaceRole iamv1alpha2.WorkspaceRole

	err := request.ReadEntity(&workspaceRole)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if workspaceRoleName != workspaceRole.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", workspaceRole.Name, workspaceRoleName)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateWorkspaceRole(workspace, &workspaceRole)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) CreateWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")

	var workspaceRole iamv1alpha2.WorkspaceRole

	err := request.ReadEntity(&workspaceRole)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateWorkspaceRole(workspace, &workspaceRole)

	if err != nil {
		klog.Error(err)
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	workspaceRoleName := request.PathParameter("workspacerole")

	err := h.am.DeleteWorkspaceRole(workspace, workspaceRoleName)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) CreateUserOrClusterMembers(request *restful.Request, response *restful.Response) {

	requestInfo, ok := apirequeset.RequestInfoFrom(request.Request.Context())

	if ok && requestInfo.ResourceScope == apirequeset.ClusterScope {
		h.CreateClusterMembers(request, response)
		return
	}

	var user iamv1alpha2.User
	err := request.ReadEntity(&user)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	globalRole := user.Annotations[iamv1alpha2.GlobalRoleAnnotation]
	delete(user.Annotations, iamv1alpha2.RoleAnnotation)

	if globalRole != "" {
		if _, err = h.am.GetGlobalRole(globalRole); err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleBadRequest(response, request, err)
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}
	}

	created, err := h.im.CreateUser(&user)

	if err != nil {
		klog.Error(err)
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		if errors.IsAlreadyExists(err) {
			api.HandleConflict(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	if globalRole != "" {
		if err := h.am.CreateOrUpdateGlobalRoleBinding(user.Name, globalRole); err != nil {

			if errors.IsNotFound(err) {
				api.HandleBadRequest(response, request, err)
				return
			}

			api.HandleInternalError(response, request, err)
			return
		}
	}

	// ensure encrypted password will not be output
	created.Spec.EncryptedPassword = ""

	response.WriteEntity(created)
}

func (h *iamHandler) UpdateUserOrClusterMember(request *restful.Request, response *restful.Response) {
	requestInfo, ok := apirequeset.RequestInfoFrom(request.Request.Context())

	if ok && requestInfo.ResourceScope == apirequeset.ClusterScope {
		h.UpdateClusterMember(request, response)
		return
	}

	username := request.PathParameter("user")

	var user iamv1alpha2.User

	err := request.ReadEntity(&user)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != user.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", user.Name, username)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	globalRole := user.Annotations[iamv1alpha2.GlobalRoleAnnotation]
	delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)

	updated, err := h.im.UpdateUser(&user)

	if err != nil {
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	if globalRole != "" {
		if err := h.am.CreateOrUpdateGlobalRoleBinding(user.Name, globalRole); err != nil {

			if errors.IsNotFound(err) {
				api.HandleBadRequest(response, request, err)
				return
			}

			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) DeleteUserOrClusterMember(request *restful.Request, response *restful.Response) {
	requestInfo, ok := apirequeset.RequestInfoFrom(request.Request.Context())

	if ok && requestInfo.ResourceScope == apirequeset.ClusterScope {
		h.RemoveClusterMember(request, response)
		return
	}

	username := request.PathParameter("user")

	err := h.im.DeleteUser(username)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) CreateGlobalRole(request *restful.Request, response *restful.Response) {

	var globalRole iamv1alpha2.GlobalRole

	err := request.ReadEntity(&globalRole)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateGlobalRole(&globalRole)

	if err != nil {
		klog.Error(err)
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteGlobalRole(request *restful.Request, response *restful.Response) {
	globalRole := request.PathParameter("globalrole")

	err := h.am.DeleteGlobalRole(globalRole)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")

	var globalRole iamv1alpha2.GlobalRole

	err := request.ReadEntity(&globalRole)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if globalRoleName != globalRole.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", globalRole.Name, globalRoleName)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateGlobalRole(&globalRole)

	if err != nil {
		klog.Error(err)
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) DescribeGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")
	globalRole, err := h.am.GetGlobalRole(globalRoleName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(globalRole)
}

func (h *iamHandler) CreateClusterRole(request *restful.Request, response *restful.Response) {
	var clusterRole rbacv1.ClusterRole

	err := request.ReadEntity(&clusterRole)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateClusterRole(&clusterRole)

	if err != nil {
		klog.Error(err)
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteClusterRole(request *restful.Request, response *restful.Response) {
	clusterrole := request.PathParameter("clusterrole")

	err := h.am.DeleteClusterRole(clusterrole)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")

	var clusterRole rbacv1.ClusterRole

	err := request.ReadEntity(&clusterRole)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if clusterRoleName != clusterRole.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", clusterRole.Name, clusterRoleName)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateClusterRole(&clusterRole)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) DescribeClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")
	clusterRole, err := h.am.GetClusterRole(clusterRoleName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(clusterRole)
}

func (h *iamHandler) DescribeWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	workspaceRoleName := request.PathParameter("workspacerole")
	workspaceRole, err := h.am.GetWorkspaceRole(workspace, workspaceRoleName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(workspaceRole)
}

func (h *iamHandler) CreateNamespaceRole(request *restful.Request, response *restful.Response) {

	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	var role rbacv1.Role

	err = request.ReadEntity(&role)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.am.CreateOrUpdateNamespaceRole(namespace, &role)

	if err != nil {
		klog.Error(err)
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DeleteNamespaceRole(request *restful.Request, response *restful.Response) {

	role := request.PathParameter("role")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	err = h.am.DeleteNamespaceRole(namespace, role)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateNamespaceRole(request *restful.Request, response *restful.Response) {

	roleName := request.PathParameter("role")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	var role rbacv1.Role

	err = request.ReadEntity(&role)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if roleName != role.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", role.Name, roleName)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.am.CreateOrUpdateNamespaceRole(namespace, &role)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) CreateWorkspaceMembers(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")

	var members []Member

	err := request.ReadEntity(&members)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	for _, member := range members {
		err := h.am.CreateOrUpdateWorkspaceRoleBinding(member.Username, workspace, member.RoleRef)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) RemoveWorkspaceMember(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	username := request.PathParameter("user")

	err := h.am.RemoveUserFromWorkspace(username, workspace)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateWorkspaceMember(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	username := request.PathParameter("user")

	var member Member

	err := request.ReadEntity(&member)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != member.Username {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", member.Username, username)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = h.am.CreateOrUpdateWorkspaceRoleBinding(member.Username, workspace, member.RoleRef)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) CreateNamespaceMembers(request *restful.Request, response *restful.Response) {

	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	var members []Member

	err = request.ReadEntity(&members)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	for _, member := range members {
		err := h.am.CreateOrUpdateNamespaceRoleBinding(member.Username, namespace, member.RoleRef)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateNamespaceMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	var member Member

	err = request.ReadEntity(&member)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != member.Username {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", member.Username, username)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = h.am.CreateOrUpdateNamespaceRoleBinding(member.Username, namespace, member.RoleRef)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) RemoveNamespaceMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	err = h.am.RemoveUserFromNamespace(username, namespace)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) CreateClusterMembers(request *restful.Request, response *restful.Response) {
	var members []Member

	err := request.ReadEntity(&members)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	for _, member := range members {
		err := h.am.CreateOrUpdateClusterRoleBinding(member.Username, member.RoleRef)
		if err != nil {
			klog.Error(err)
			if errors.IsNotFound(err) {
				api.HandleNotFound(response, request, err)
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) RemoveClusterMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	err := h.am.RemoveUserFromCluster(username)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateClusterMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	var member Member

	err := request.ReadEntity(&member)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != member.Username {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", member.Username, username)
		klog.Errorf("%+v", err)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = h.am.CreateOrUpdateClusterRoleBinding(member.Username, member.RoleRef)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		if errors.IsBadRequest(err) {
			api.HandleBadRequest(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) DescribeClusterMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	queryParam := query.New()
	queryParam.Filters[query.FieldName] = query.Value(username)
	queryParam.Filters[iamv1alpha2.ScopeCluster] = iamv1alpha2.LocalCluster

	result, err := h.im.ListUsers(queryParam)

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if len(result.Items) == 0 {
		err := errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularUser), username)
		api.HandleNotFound(response, request, err)
		return
	}

	response.WriteEntity(result.Items[0])
}

func (h *iamHandler) ListClusterMembers(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)

	queryParam.Filters[iamv1alpha2.ScopeCluster] = iamv1alpha2.LocalCluster

	result, err := h.im.ListUsers(queryParam)

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *iamHandler) DescribeNamespaceRole(request *restful.Request, response *restful.Response) {

	roleName := request.PathParameter("role")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	role, err := h.am.GetNamespaceRole(namespace, roleName)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(role)
}

// resolve the namespace which controlled by the devops project
func (h *iamHandler) resolveNamespace(namespace string, devops string) (string, error) {
	if devops == "" {
		return namespace, nil
	}
	return h.am.GetControlledNamespace(devops)
}
