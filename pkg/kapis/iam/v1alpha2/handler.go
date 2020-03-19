package v1alpha2

import (
	"errors"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/api/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldappool "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
	"net/http"

	iamapi "kubesphere.io/kubesphere/pkg/api/iam"
)

const (
	kindTokenReview = "TokenReview"
)

type iamHandler struct {
	amOperator am.AccessManagementInterface
	imOperator im.IdentityManagementInterface
}

func newIAMHandler(k8sClient k8s.Client, factory informers.InformerFactory, ldapClient ldappool.Interface, cacheClient cache.Interface, options *iamapi.AuthenticationOptions) *iamHandler {
	return &iamHandler{
		amOperator: am.NewAMOperator(k8sClient.Kubernetes(), factory.KubernetesSharedInformerFactory()),
		imOperator: im.NewIMOperator(ldapClient, cacheClient, options),
	}
}

// Implement webhook authentication interface
// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
func (h *iamHandler) TokenReviewHandler(req *restful.Request, resp *restful.Response) {
	var tokenReview iamv1alpha2.TokenReview

	err := req.ReadEntity(&tokenReview)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	if err = tokenReview.Validate(); err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	user, err := h.imOperator.VerifyToken(tokenReview.Spec.Token)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, req, err)
		return
	}

	success := iamv1alpha2.TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: kindTokenReview,
		Status: &iamv1alpha2.Status{
			Authenticated: true,
			User:          map[string]interface{}{"username": user.Username, "uid": user.Username, "groups": user.Groups},
		},
	}

	resp.WriteEntity(success)
}

func (h *iamHandler) Login(req *restful.Request, resp *restful.Response) {
	var loginRequest iamv1alpha2.LoginRequest

	err := req.ReadEntity(&loginRequest)

	if err != nil || loginRequest.Username == "" || loginRequest.Password == "" {
		err = errors.New("incorrect username or password")
		klog.V(4).Infoln(err)
		resp.WriteHeaderAndEntity(http.StatusUnauthorized, err)
		return
	}

	ip := iputil.RemoteIp(req.Request)

	token, err := h.imOperator.Login(loginRequest.Username, loginRequest.Password, ip)

	if err != nil {
		if err == im.AuthRateLimitExceeded {
			klog.V(4).Infoln(err)
			resp.WriteHeaderAndEntity(http.StatusTooManyRequests, err)
			return
		}
		klog.V(4).Infoln(err)
		resp.WriteHeaderAndEntity(http.StatusUnauthorized, err)
		return
	}

	resp.WriteEntity(token)
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

	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, v1alpha2.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, true)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, nil, err)
		return
	}

	result, err := h.imOperator.ListUsers(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
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
