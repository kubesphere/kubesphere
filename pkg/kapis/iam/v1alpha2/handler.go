package v1alpha2

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
	"github.com/go-ldap/ldap"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/iam/policy"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/jwtutil"
	"net/http"
	"sort"
)

type iamHandler struct {
	amOperator iam.AccessManagementInterface
	imOperator iam.IdentityManagementInterface
}

func newIAMHandler() *iamHandler {
	return &iamHandler{}
}

// k8s token review
func (h *iamHandler) TokenReviewHandler(req *restful.Request, resp *restful.Response) {
	var tokenReview iam.TokenReview

	err := req.ReadEntity(&tokenReview)

	if err != nil {
		api.HandleBadRequest(resp, err)
		return
	}

	if tokenReview.Spec == nil {
		api.HandleBadRequest(resp, errors.New("token must not be null"))
		return
	}

	uToken := tokenReview.Spec.Token

	token, err := jwtutil.ValidateToken(uToken)

	if err != nil {
		klog.Errorln(err)
		failed := iam.TokenReview{APIVersion: tokenReview.APIVersion,
			Kind: iam.KindTokenReview,
			Status: &iam.Status{
				Authenticated: false,
			},
		}
		resp.WriteAsJson(failed)
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	username, ok := claims["username"].(string)

	if !ok {
		api.HandleBadRequest(resp, errors.New("username not found"))
		return
	}

	user, err := h.imOperator.GetUserInfo(username)

	if err != nil {
		api.HandleInternalError(resp, err)
		return
	}

	groups, err := h.imOperator.GetUserGroups(username)

	if err != nil {
		api.HandleInternalError(resp, err)
		return
	}

	user.Groups = groups

	success := iam.TokenReview{APIVersion: tokenReview.APIVersion,
		Kind: iam.KindTokenReview,
		Status: &iam.Status{
			Authenticated: true,
			User:          map[string]interface{}{"username": user.Username, "uid": user.Username, "groups": user.Groups},
		},
	}

	resp.WriteAsJson(success)
	return
}

func (h *iamHandler) ListRoleUsers(req *restful.Request, resp *restful.Response) {
	role := req.PathParameter("role")
	namespace := req.PathParameter("namespace")

	roleBindings, err := h.amOperator.ListRoleBindings(namespace, role)

	if err != nil {
		api.HandleInternalError(resp, err)
		return
	}
	result := make([]*iam.User, 0)
	for _, roleBinding := range roleBindings {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == rbacv1.UserKind {
				user, err := h.imOperator.GetUserInfo(subject.Name)
				// skip if user not exist
				if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
					continue
				}
				if err != nil {
					api.HandleInternalError(resp, err)
					return
				}
				result = append(result, user)
			}
		}
	}

	resp.WriteEntity(result)
}

func (h *iamHandler) ListClusterRoles(req *restful.Request, resp *restful.Response) {
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := iam.ListClusterRoles(conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)

}

func (h *iamHandler) ListRoles(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := iam.ListRoles(namespace, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)

}

// List users by namespace
func (h *iamHandler) ListNamespaceUsers(req *restful.Request, resp *restful.Response) {

	namespace := req.PathParameter("namespace")

	users, err := iam.NamespaceUsers(namespace)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	// sort by time by default
	sort.Slice(users, func(i, j int) bool {
		return users[i].RoleBindTime.After(*users[j].RoleBindTime)
	})

	resp.WriteAsJson(users)
}

func (h *iamHandler) ListUserRoles(req *restful.Request, resp *restful.Response) {

	username := req.PathParameter("user")

	roles, err := iam.GetUserRoles("", username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	_, clusterRoles, err := iam.GetUserClusterRoles(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	roleList := RoleList{}
	roleList.Roles = roles
	roleList.ClusterRoles = clusterRoles

	resp.WriteAsJson(roleList)
}

func (h *iamHandler) RulesMapping(req *restful.Request, resp *restful.Response) {
	rules := policy.RoleRuleMapping
	resp.WriteAsJson(rules)
}

func (h *iamHandler) ClusterRulesMapping(req *restful.Request, resp *restful.Response) {
	rules := policy.ClusterRoleRuleMapping
	resp.WriteAsJson(rules)
}

func (h *iamHandler) ListClusterRoleRules(req *restful.Request, resp *restful.Response) {
	clusterRoleName := req.PathParameter("clusterrole")
	rules, err := iam.GetClusterRoleSimpleRules(clusterRoleName)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	resp.WriteAsJson(rules)
}

func (h *iamHandler) ListClusterRoleUsers(req *restful.Request, resp *restful.Response) {
	clusterRoleName := req.PathParameter("clusterrole")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := iam.ListClusterRoleUsers(clusterRoleName, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		if k8serr.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, err)
		} else {
			resp.WriteError(http.StatusInternalServerError, err)
		}
		return
	}

	resp.WriteAsJson(result)
}

func (h *iamHandler) ListRoleRules(req *restful.Request, resp *restful.Response) {
	namespaceName := req.PathParameter("namespace")
	roleName := req.PathParameter("role")

	rules, err := iam.GetRoleSimpleRules(namespaceName, roleName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(rules)
}
