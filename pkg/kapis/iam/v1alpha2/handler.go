package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/api/auth"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldappool "kubesphere.io/kubesphere/pkg/simple/client/ldap"
)

type iamHandler struct {
	amOperator am.AccessManagementInterface
	imOperator im.IdentityManagementInterface
}

func newIAMHandler(k8sClient k8s.Client, factory informers.InformerFactory, ldapClient ldappool.Interface, cacheClient cache.Interface, options *auth.AuthenticationOptions) *iamHandler {
	return &iamHandler{
		amOperator: am.NewAMOperator(k8sClient.Kubernetes(), factory.KubernetesSharedInformerFactory()),
		imOperator: im.NewLDAPOperator(ldapClient),
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
	panic("implement me")
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
