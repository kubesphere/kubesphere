package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
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

func (h *iamHandler) CreateUser(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) DeleteUser(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ModifyUser(request *restful.Request, response *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) DescribeUser(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")
	user, err := h.im.DescribeUser(username)
	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	globalRole, err := h.am.GetRoleOfUserInTargetScope(iamv1alpha2.GlobalScope, "", username)

	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}
	result := iamv1alpha2.UserDetail{User: user, GlobalRole: globalRole}

	resp.WriteEntity(result)
}

func (h *iamHandler) ListUsers(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListUserRoles(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListRoles(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListRolesOfUser(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")

	var roles []iamv1alpha2.Role
	var err error

	if strings.HasSuffix(req.Request.URL.Path, "workspaceroles") {
		roles, err = h.am.ListRolesOfUser(iamv1alpha2.WorkspaceScope, username)
	} else if strings.HasSuffix(req.Request.URL.Path, "clusterroles") {
		roles, err = h.am.ListRolesOfUser(iamv1alpha2.ClusterScope, username)
	} else if strings.HasSuffix(req.Request.URL.Path, "namespaceroles") {
		roles, err = h.am.ListRolesOfUser(iamv1alpha2.NamespaceScope, username)
	}

	if err != nil {
		api.HandleInternalError(resp, req, err)
		return
	}

	result := iamv1alpha2.RoleList{
		TypeMeta: v1.TypeMeta{
			Kind:       "List",
			APIVersion: "v1",
		},
		ListMeta: v1.ListMeta{},
		Items:    roles,
	}

	resp.WriteEntity(result)
}
func (h *iamHandler) ListClusterRoles(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListRoleUsers(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

// List users by namespace
func (h *iamHandler) ListNamespaceUsers(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListClusterRoleUsers(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListClusterRoleRules(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListRoleRules(req *restful.Request, resp *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListWorkspaceRoles(request *restful.Request, response *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) DescribeWorkspaceRole(request *restful.Request, response *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListWorkspaceRoleRules(request *restful.Request, response *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) ListWorkspaceUsers(request *restful.Request, response *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) InviteUser(request *restful.Request, response *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) RemoveUser(request *restful.Request, response *restful.Response) {
	panic("implement me")
}

func (h *iamHandler) DescribeWorkspaceUser(request *restful.Request, response *restful.Response) {
	panic("implement me")
}
