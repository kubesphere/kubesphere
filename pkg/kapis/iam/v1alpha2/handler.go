package v1alpha2

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
	"github.com/go-ldap/ldap"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/api/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha2"
	apierr "kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldappool "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
	"kubesphere.io/kubesphere/pkg/utils/jwtutil"
	"net/http"
)

type iamHandler struct {
	amOperator iam.AccessManagementInterface
	imOperator iam.IdentityManagementInterface
}

func newIAMHandler(k8sClient k8s.Client, ldapClient ldappool.Client, options iam.Config) *iamHandler {
	factory := informers.NewInformerFactories(k8sClient.Kubernetes(), k8sClient.KubeSphere(), k8sClient.S2i(), k8sClient.Application())
	return &iamHandler{
		amOperator: iam.NewAMOperator(factory.KubernetesSharedInformerFactory()),
		imOperator: iam.NewIMOperator(ldapClient, options),
	}
}

// Implement webhook authentication interface
// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
func (h *iamHandler) TokenReviewHandler(req *restful.Request, resp *restful.Response) {
	var tokenReview iamv1alpha2.TokenReview

	err := req.ReadEntity(&tokenReview)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, err)
		return
	}

	if err = tokenReview.Validate(); err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, err)
		return
	}

	token, err := jwtutil.ValidateToken(tokenReview.Spec.Token)

	if err != nil {
		failed := iamv1alpha2.TokenReview{APIVersion: tokenReview.APIVersion,
			Kind: iam.KindTokenReview,
			Status: &iamv1alpha2.Status{
				Authenticated: false,
			},
		}
		resp.WriteEntity(failed)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		api.HandleBadRequest(resp, errors.New("invalid token"))
		return
	}

	username, ok := claims["username"].(string)

	if !ok {
		api.HandleBadRequest(resp, errors.New("invalid token"))
		return
	}

	user, err := h.imOperator.DescribeUser(username)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	success := iamv1alpha2.TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: iam.KindTokenReview,
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
		if err == iam.AuthRateLimitExceeded {
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
	var createRequest iamv1alpha2.CreateUserRequest
	err := req.ReadEntity(&createRequest)
	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, err)
		return
	}

	if err := createRequest.Validate(); err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, err)
		return
	}

	created, err := h.imOperator.CreateUser(createRequest.User)

	if err != nil {
		if err == iam.UserAlreadyExists {
			klog.V(4).Infoln(err)
			resp.WriteHeaderAndEntity(http.StatusConflict, err)
			return
		}
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	err = h.amOperator.CreateClusterRoleBinding(created.Username, createRequest.ClusterRole)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(created)
}

func (h *iamHandler) DeleteUser(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")
	operator := req.HeaderParameter(constants.UserNameHeader)

	if operator == username {
		err := errors.New("cannot delete yourself")
		klog.V(4).Infoln(err)
		api.HandleForbidden(resp, err)
		return
	}

	err := h.amOperator.UnBindAllRoles(username)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	err = h.imOperator.DeleteUser(username)

	// TODO release user resources
	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(apierr.None)
}

func (h *iamHandler) ModifyUser(request *restful.Request, response *restful.Response) {

	username := request.PathParameter("user")
	operator := request.HeaderParameter(constants.UserNameHeader)
	var modifyUserRequest iamv1alpha2.ModifyUserRequest

	err := request.ReadEntity(&modifyUserRequest)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(response, err)
		return
	}

	if username != modifyUserRequest.Username {
		err = fmt.Errorf("the name of user (%s) does not match the name on the URL (%s)", modifyUserRequest.Username, username)
		klog.V(4).Infoln(err)
		api.HandleBadRequest(response, err)
		return
	}

	if err = modifyUserRequest.Validate(); err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(response, err)
		return
	}

	// change password by self
	if operator == modifyUserRequest.Username && modifyUserRequest.Password != "" {

	}

	result, err := h.imOperator.ModifyUser(modifyUserRequest.User)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(response, err)
		return
	}

	// TODO modify cluster role

	response.WriteEntity(result)
}

func (h *iamHandler) DescribeUser(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")

	user, err := h.imOperator.DescribeUser(username)

	if err != nil {
		if err == iam.UserNotExists {
			klog.V(4).Infoln(err)
			api.HandleNotFound(resp, err)
			return
		}
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	// TODO append more user info
	clusterRole, err := h.amOperator.GetClusterRole(username)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	result := iamv1alpha2.UserDetail{
		User:        user,
		ClusterRole: clusterRole.Name,
	}

	resp.WriteEntity(result)
}

func (h *iamHandler) ListUsers(req *restful.Request, resp *restful.Response) {

	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, v1alpha2.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, true)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, err)
		return
	}

	result, err := h.imOperator.ListUsers(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(result)
}

func (h *iamHandler) ListUserRoles(req *restful.Request, resp *restful.Response) {

	username := req.PathParameter("user")

	roles, err := h.imOperator.GetUserRoles(username)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(roles)
}

func (h *iamHandler) ListRoles(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, v1alpha2.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, true)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, err)
		return
	}

	result, err := h.amOperator.ListRoles(namespace, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteAsJson(result)

}
func (h *iamHandler) ListClusterRoles(req *restful.Request, resp *restful.Response) {
	limit, offset := params.ParsePaging(req)
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, v1alpha2.CreateTime)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, true)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.V(4).Infoln(err)
		api.HandleBadRequest(resp, err)
		return
	}

	result, err := h.amOperator.ListClusterRoles(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(result)

}

func (h *iamHandler) ListRoleUsers(req *restful.Request, resp *restful.Response) {
	role := req.PathParameter("role")
	namespace := req.PathParameter("namespace")

	roleBindings, err := h.amOperator.ListRoleBindings(namespace, role)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}
	result := make([]*iam.User, 0)
	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind {
				user, err := h.imOperator.DescribeUser(subject.Name)
				// skip if user not exist
				if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
					continue
				}
				if err != nil {
					klog.Errorln(err)
					api.HandleInternalError(resp, err)
					return
				}
				result = append(result, user)
			}
		}
	}

	resp.WriteEntity(result)
}

// List users by namespace
func (h *iamHandler) ListNamespaceUsers(req *restful.Request, resp *restful.Response) {

	namespace := req.PathParameter("namespace")

	roleBindings, err := h.amOperator.ListRoleBindings(namespace, "")

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	result := make([]*iam.User, 0)
	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind {
				user, err := h.imOperator.DescribeUser(subject.Name)
				// skip if user not exist
				if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
					continue
				}
				if err != nil {
					klog.Errorln(err)
					api.HandleInternalError(resp, err)
					return
				}
				result = append(result, user)
			}
		}
	}

	resp.WriteEntity(result)
}

func (h *iamHandler) ListClusterRoleUsers(req *restful.Request, resp *restful.Response) {
	clusterRole := req.PathParameter("clusterrole")
	clusterRoleBindings, err := h.amOperator.ListClusterRoleBindings(clusterRole)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	result := make([]*iam.User, 0)
	for _, roleBinding := range clusterRoleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind {
				user, err := h.imOperator.DescribeUser(subject.Name)
				// skip if user not exist
				if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
					continue
				}
				if err != nil {
					klog.Errorln(err)
					api.HandleInternalError(resp, err)
					return
				}
				result = append(result, user)
			}
		}
	}

	resp.WriteEntity(result)
}

func (h *iamHandler) RulesMapping(req *restful.Request, resp *restful.Response) {
	rules := policy.RoleRuleMapping
	resp.WriteEntity(rules)
}

func (h *iamHandler) ClusterRulesMapping(req *restful.Request, resp *restful.Response) {
	rules := policy.ClusterRoleRuleMapping
	resp.WriteEntity(rules)
}

func (h *iamHandler) ListClusterRoleRules(req *restful.Request, resp *restful.Response) {
	clusterRole := req.PathParameter("clusterrole")
	rules, err := h.amOperator.GetClusterRoleSimpleRules(clusterRole)
	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}
	resp.WriteEntity(rules)
}

func (h *iamHandler) ListRoleRules(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	role := req.PathParameter("role")

	rules, err := h.amOperator.GetRoleSimpleRules(namespace, role)

	if err != nil {
		klog.Errorln(err)
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteEntity(rules)
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
