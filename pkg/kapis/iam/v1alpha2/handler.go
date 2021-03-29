/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"fmt"
	"strings"

	authuser "k8s.io/apiserver/pkg/authentication/user"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/auth"

	"github.com/emicklei/go-restful"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	apirequest "kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/group"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	servererr "kubesphere.io/kubesphere/pkg/server/errors"
)

type Member struct {
	Username string `json:"username"`
	RoleRef  string `json:"roleRef"`
}

type GroupMember struct {
	UserName  string `json:"userName"`
	GroupName string `json:"groupName"`
}

type PasswordReset struct {
	CurrentPassword string `json:"currentPassword"`
	Password        string `json:"password"`
}

type iamHandler struct {
	am         am.AccessManagementInterface
	im         im.IdentityManagementInterface
	group      group.GroupOperator
	authorizer authorizer.Authorizer
}

func newIAMHandler(im im.IdentityManagementInterface, am am.AccessManagementInterface, group group.GroupOperator, authorizer authorizer.Authorizer) *iamHandler {
	return &iamHandler{
		am:         am,
		im:         im,
		group:      group,
		authorizer: authorizer,
	}
}

func (h *iamHandler) DescribeUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	user, err := h.im.DescribeUser(username)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	globalRole, err := h.am.GetGlobalRoleOfUser(username)
	// ignore not found error
	if err != nil && !errors.IsNotFound(err) {
		api.HandleInternalError(response, request, err)
		return
	}
	if globalRole != nil {
		user = appendGlobalRoleAnnotation(user, globalRole.Name)
	}

	response.WriteEntity(user)
}

func (h *iamHandler) RetrieveMemberRoleTemplates(request *restful.Request, response *restful.Response) {

	if strings.HasSuffix(request.Request.URL.Path, iamv1alpha2.ResourcesPluralGlobalRole) {
		username := request.PathParameter("user")

		globalRole, err := h.am.GetGlobalRoleOfUser(username)
		if err != nil {
			// if role binding not exist return empty list
			if errors.IsNotFound(err) {
				response.WriteEntity([]interface{}{})
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}

		result, err := h.am.ListGlobalRoles(&query.Query{
			Pagination: query.NoPagination,
			SortBy:     "",
			Ascending:  false,
			Filters:    map[query.Field]query.Value{iamv1alpha2.AggregateTo: query.Value(globalRole.Name)},
		})
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}

		response.WriteEntity(result.Items)
		return
	}

	if strings.HasSuffix(request.Request.URL.Path, iamv1alpha2.ResourcesPluralClusterRole) {
		username := request.PathParameter("clustermember")
		clusterRole, err := h.am.GetClusterRoleOfUser(username)
		if err != nil {
			// if role binding not exist return empty list
			if errors.IsNotFound(err) {
				response.WriteEntity([]interface{}{})
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}

		result, err := h.am.ListClusterRoles(&query.Query{
			Pagination: query.NoPagination,
			SortBy:     "",
			Ascending:  false,
			Filters:    map[query.Field]query.Value{iamv1alpha2.AggregateTo: query.Value(clusterRole.Name)},
		})
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}

		response.WriteEntity(result.Items)
		return
	}

	if strings.HasSuffix(request.Request.URL.Path, iamv1alpha2.ResourcesPluralWorkspaceRole) {
		workspace := request.PathParameter("workspace")
		username := request.PathParameter("workspacemember")

		user, err := h.im.DescribeUser(username)
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}

		workspaceRoles, err := h.am.GetWorkspaceRoleOfUser(username, user.Spec.Groups, workspace)
		if err != nil {
			// if role binding not exist return empty list
			if errors.IsNotFound(err) {
				response.WriteEntity([]interface{}{})
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}
		templateRoles := make(map[string]*iamv1alpha2.WorkspaceRole)
		for _, role := range workspaceRoles {
			// merge template Role
			result, err := h.am.ListWorkspaceRoles(&query.Query{
				Pagination: query.NoPagination,
				SortBy:     "",
				Ascending:  false,
				Filters:    map[query.Field]query.Value{iamv1alpha2.AggregateTo: query.Value(role.Name)},
			})
			if err != nil {
				api.HandleInternalError(response, request, err)
				return
			}

			for _, obj := range result.Items {
				templateRole := obj.(*iamv1alpha2.WorkspaceRole)
				templateRoles[templateRole.Name] = templateRole
			}
		}

		results := make([]*iamv1alpha2.WorkspaceRole, 0, len(templateRoles))
		for _, value := range templateRoles {
			results = append(results, value)
		}

		response.WriteEntity(results)
		return
	}

	if strings.HasSuffix(request.Request.URL.Path, iamv1alpha2.ResourcesPluralRole) {
		namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
		username := request.PathParameter("member")
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}

		user, err := h.im.DescribeUser(username)
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}

		roles, err := h.am.GetNamespaceRoleOfUser(username, user.Spec.Groups, namespace)
		if err != nil {
			// if role binding not exist return empty list
			if errors.IsNotFound(err) {
				response.WriteEntity([]interface{}{})
				return
			}
			api.HandleInternalError(response, request, err)
			return
		}

		templateRoles := make(map[string]*rbacv1.Role)
		for _, role := range roles {
			// merge template Role
			result, err := h.am.ListRoles(namespace, &query.Query{
				Pagination: query.NoPagination,
				SortBy:     "",
				Ascending:  false,
				Filters:    map[query.Field]query.Value{iamv1alpha2.AggregateTo: query.Value(role.Name)},
			})

			if err != nil {
				api.HandleInternalError(response, request, err)
				return
			}

			for _, obj := range result.Items {
				templateRole := obj.(*rbacv1.Role)
				templateRoles[templateRole.Name] = templateRole
			}
		}

		results := make([]*rbacv1.Role, 0, len(templateRoles))
		for _, value := range templateRoles {
			results = append(results, value)
		}

		response.WriteEntity(results)
		return
	}
}

func (h *iamHandler) ListUsers(request *restful.Request, response *restful.Response) {
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
		// ignore not found error
		if err != nil && !errors.IsNotFound(err) {
			api.HandleInternalError(response, request, err)
			return
		}
		if globalRole != nil {
			user = appendGlobalRoleAnnotation(user, globalRole.Name)
		}
		result.Items[i] = user
	}
	response.WriteEntity(result)
}

func appendGlobalRoleAnnotation(user *iamv1alpha2.User, globalRole string) *iamv1alpha2.User {
	if user.Annotations == nil {
		user.Annotations = make(map[string]string, 0)
	}
	user.Annotations[iamv1alpha2.GlobalRoleAnnotation] = globalRole
	return user
}

func (h *iamHandler) ListRoles(request *restful.Request, response *restful.Response) {
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
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
		api.HandleError(response, request, err)
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
	username := request.PathParameter("member")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	queryParam := query.New()
	queryParam.Filters[query.FieldNames] = query.Value(username)
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
	username := request.PathParameter("workspacemember")

	queryParam := query.New()
	queryParam.Filters[query.FieldNames] = query.Value(username)
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

	var workspaceRole iamv1alpha2.WorkspaceRole
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

func (h *iamHandler) CreateUser(req *restful.Request, resp *restful.Response) {
	var user iamv1alpha2.User
	err := req.ReadEntity(&user)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	operator, ok := request.UserFrom(req.Request.Context())
	if ok && operator.GetName() == iamv1alpha2.PreRegistrationUser {
		extra := operator.GetExtra()
		// The token used for registration must contain additional information
		if len(extra[iamv1alpha2.ExtraIdentityProvider]) != 1 || len(extra[iamv1alpha2.ExtraUID]) != 1 {
			err = errors.NewBadRequest("invalid registration token")
			api.HandleBadRequest(resp, req, err)
			return
		}
		if user.Labels == nil {
			user.Labels = make(map[string]string)
		}
		user.Labels[iamv1alpha2.IdentifyProviderLabel] = extra[iamv1alpha2.ExtraIdentityProvider][0]
		user.Labels[iamv1alpha2.OriginUIDLabel] = extra[iamv1alpha2.ExtraUID][0]
		// default role
		delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)
	}

	globalRole := user.Annotations[iamv1alpha2.GlobalRoleAnnotation]
	delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)
	if globalRole != "" {
		if _, err = h.am.GetGlobalRole(globalRole); err != nil {
			api.HandleError(resp, req, err)
			return
		}
	}

	created, err := h.im.CreateUser(&user)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if globalRole != "" {
		if err := h.am.CreateGlobalRoleBinding(user.Name, globalRole); err != nil {
			api.HandleError(resp, req, err)
			return
		}
	}

	// ensure encrypted password will not be output
	created.Spec.EncryptedPassword = ""

	resp.WriteEntity(created)
}

func (h *iamHandler) UpdateUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	var user iamv1alpha2.User

	err := request.ReadEntity(&user)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != user.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", user.Name, username)
		api.HandleBadRequest(response, request, err)
		return
	}

	globalRole := user.Annotations[iamv1alpha2.GlobalRoleAnnotation]
	delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)

	updated, err := h.im.UpdateUser(&user)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	operator, ok := apirequest.UserFrom(request.Request.Context())
	if globalRole != "" && ok {
		err = h.updateGlobalRoleBinding(operator, updated, globalRole)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		updated = appendGlobalRoleAnnotation(updated, globalRole)
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) ModifyPassword(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")
	var passwordReset PasswordReset
	err := request.ReadEntity(&passwordReset)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	operator, ok := apirequest.UserFrom(request.Request.Context())

	if !ok {
		err = errors.NewInternalError(fmt.Errorf("cannot obtain user info"))
		api.HandleInternalError(response, request, err)
		return
	}

	userManagement := authorizer.AttributesRecord{
		Resource:        "users/password",
		Verb:            "update",
		ResourceScope:   apirequest.GlobalScope,
		ResourceRequest: true,
		User:            operator,
	}

	decision, _, err := h.authorizer.Authorize(userManagement)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	// only the user manager can modify the password without verifying the old password
	// if old password is defined must be verified
	if decision != authorizer.DecisionAllow || passwordReset.CurrentPassword != "" {
		if err = h.im.PasswordVerify(username, passwordReset.CurrentPassword); err != nil {
			if err == auth.IncorrectPasswordError {
				err = errors.NewBadRequest("incorrect old password")
			}
			api.HandleError(response, request, err)
			return
		}
	}

	err = h.im.ModifyPassword(username, passwordReset.Password)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) DeleteUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")

	err := h.im.DeleteUser(username)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) CreateGlobalRole(request *restful.Request, response *restful.Response) {

	var globalRole iamv1alpha2.GlobalRole
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

	var globalRole iamv1alpha2.GlobalRole
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
	var clusterRole rbacv1.ClusterRole
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

	var clusterRole rbacv1.ClusterRole

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

func (h *iamHandler) DescribeClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")
	clusterRole, err := h.am.GetClusterRole(clusterRoleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(clusterRole)
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

func (h *iamHandler) CreateNamespaceRole(request *restful.Request, response *restful.Response) {

	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var role rbacv1.Role
	err = request.ReadEntity(&role)
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

	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	err = h.am.DeleteNamespaceRole(namespace, role)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateNamespaceRole(request *restful.Request, response *restful.Response) {
	roleName := request.PathParameter("role")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var role rbacv1.Role
	err = request.ReadEntity(&role)
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

func (h *iamHandler) CreateWorkspaceMembers(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")

	var members []Member
	err := request.ReadEntity(&members)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	for _, member := range members {
		err := h.am.CreateUserWorkspaceRoleBinding(member.Username, workspace, member.RoleRef)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
	}

	response.WriteEntity(members)
}

func (h *iamHandler) RemoveWorkspaceMember(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	username := request.PathParameter("workspacemember")

	err := h.am.RemoveUserFromWorkspace(username, workspace)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateWorkspaceMember(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	username := request.PathParameter("workspacemember")

	var member Member
	err := request.ReadEntity(&member)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != member.Username {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", member.Username, username)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = h.am.CreateUserWorkspaceRoleBinding(member.Username, workspace, member.RoleRef)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(member)
}

func (h *iamHandler) CreateNamespaceMembers(request *restful.Request, response *restful.Response) {

	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var members []Member
	err = request.ReadEntity(&members)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	for _, member := range members {
		err := h.am.CreateNamespaceRoleBinding(member.Username, namespace, member.RoleRef)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
	}

	response.WriteEntity(members)
}

func (h *iamHandler) UpdateNamespaceMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("member")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var member Member
	err = request.ReadEntity(&member)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != member.Username {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", member.Username, username)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = h.am.CreateNamespaceRoleBinding(member.Username, namespace, member.RoleRef)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(member)
}

func (h *iamHandler) RemoveNamespaceMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("member")
	namespace, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	err = h.am.RemoveUserFromNamespace(username, namespace)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) CreateClusterMembers(request *restful.Request, response *restful.Response) {
	var members []Member
	err := request.ReadEntity(&members)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	for _, member := range members {
		err := h.am.CreateClusterRoleBinding(member.Username, member.RoleRef)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
	}

	response.WriteEntity(members)
}

func (h *iamHandler) RemoveClusterMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("clustermember")

	err := h.am.RemoveUserFromCluster(username)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateClusterMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("clustermember")

	var member Member
	err := request.ReadEntity(&member)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if username != member.Username {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", member.Username, username)
		api.HandleBadRequest(response, request, err)
		return
	}

	err = h.am.CreateClusterRoleBinding(member.Username, member.RoleRef)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(member)
}

func (h *iamHandler) DescribeClusterMember(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("clustermember")

	queryParam := query.New()
	queryParam.Filters[query.FieldNames] = query.Value(username)
	queryParam.Filters[iamv1alpha2.ScopeCluster] = "true"

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
	queryParam.Filters[iamv1alpha2.ScopeCluster] = "true"

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
		api.HandleError(response, request, err)
		return
	}

	role, err := h.am.GetNamespaceRole(namespace, roleName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(role)
}

// resolve the namespace which controlled by the devops project
func (h *iamHandler) resolveNamespace(namespace string, devops string) (string, error) {
	if devops == "" {
		return namespace, nil
	}
	return h.am.GetDevOpsRelatedNamespace(devops)
}

func (h *iamHandler) PatchWorkspaceRole(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	workspaceRoleName := request.PathParameter("workspacerole")

	var workspaceRole iamv1alpha2.WorkspaceRole
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

func (h *iamHandler) PatchGlobalRole(request *restful.Request, response *restful.Response) {
	globalRoleName := request.PathParameter("globalrole")

	var globalRole iamv1alpha2.GlobalRole
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

func (h *iamHandler) PatchNamespaceRole(request *restful.Request, response *restful.Response) {
	roleName := request.PathParameter("role")
	namespaceName, err := h.resolveNamespace(request.PathParameter("namespace"), request.PathParameter("devops"))
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	var role rbacv1.Role
	err = request.ReadEntity(&role)
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

func (h *iamHandler) PatchClusterRole(request *restful.Request, response *restful.Response) {
	clusterRoleName := request.PathParameter("clusterrole")

	var clusterRole rbacv1.ClusterRole
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

func (h *iamHandler) updateGlobalRoleBinding(operator authuser.Info, user *iamv1alpha2.User, globalRole string) error {

	oldGlobalRole, err := h.am.GetGlobalRoleOfUser(user.Name)
	if err != nil && !errors.IsNotFound(err) {
		klog.Error(err)
		return err
	}

	if oldGlobalRole != nil && oldGlobalRole.Name == globalRole {
		return nil
	}

	userManagement := authorizer.AttributesRecord{
		Resource:        iamv1alpha2.ResourcesPluralUser,
		Verb:            "update",
		ResourceScope:   apirequest.GlobalScope,
		ResourceRequest: true,
		User:            operator,
	}
	decision, _, err := h.authorizer.Authorize(userManagement)
	if err != nil {
		klog.Error(err)
		return err
	}
	if decision != authorizer.DecisionAllow {
		err = errors.NewForbidden(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularUser),
			user.Name, fmt.Errorf("update global role binding is not allowed"))
		klog.Warning(err)
		return err
	}
	if err := h.am.CreateGlobalRoleBinding(user.Name, globalRole); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (h *iamHandler) ListUserLoginRecords(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("user")
	queryParam := query.ParseQueryParameter(request)
	result, err := h.im.ListLoginRecords(username, queryParam)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}
	response.WriteEntity(result)
}

func (h *iamHandler) ListWorkspaceGroups(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	queryParam := query.ParseQueryParameter(request)
	result, err := h.group.ListGroups(workspaceName, queryParam)

	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *iamHandler) CreateGroup(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	var group iamv1alpha2.Group

	err := request.ReadEntity(&group)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.group.CreateGroup(workspace, &group)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *iamHandler) DescribeGroup(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	groupName := request.PathParameter("group")
	ns, err := h.group.DescribeGroup(workspaceName, groupName)

	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(ns)
}

func (h *iamHandler) DeleteGroup(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	groupName := request.PathParameter("group")

	err := h.group.DeleteGroup(workspaceName, groupName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}

func (h *iamHandler) UpdateGroup(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	groupName := request.PathParameter("group")

	var group iamv1alpha2.Group
	err := request.ReadEntity(&group)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if groupName != group.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", group.Name, groupName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.group.UpdateGroup(workspaceName, &group)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *iamHandler) PatchGroup(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	groupName := request.PathParameter("group")

	var group iamv1alpha2.Group
	err := request.ReadEntity(&group)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	group.Name = groupName
	patched, err := h.group.PatchGroup(workspaceName, &group)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *iamHandler) ListGroupBindings(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	queryParam := query.ParseQueryParameter(request)
	result, err := h.group.ListGroupBindings(workspaceName, queryParam)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *iamHandler) ListGroupRoleBindings(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	queryParam := query.ParseQueryParameter(request)
	result, err := h.am.ListGroupRoleBindings(workspaceName, queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *iamHandler) CreateRoleBinding(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	var roleBindings []rbacv1.RoleBinding
	err := request.ReadEntity(&roleBindings)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	var results []rbacv1.RoleBinding
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

func (h *iamHandler) ListGroupWorkspaceRoleBindings(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	queryParam := query.ParseQueryParameter(request)
	result, err := h.am.ListGroupWorkspaceRoleBindings(workspaceName, queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(result)
}

func (h *iamHandler) CreateWorkspaceRoleBinding(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")

	var roleBindings []iamv1alpha2.WorkspaceRoleBinding
	err := request.ReadEntity(&roleBindings)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	var results []iamv1alpha2.WorkspaceRoleBinding
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

func (h *iamHandler) CreateGroupBinding(request *restful.Request, response *restful.Response) {

	workspace := request.PathParameter("workspace")

	var members []GroupMember
	err := request.ReadEntity(&members)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	var results []iamv1alpha2.GroupBinding
	for _, item := range members {
		b, err := h.group.CreateGroupBinding(workspace, item.GroupName, item.UserName)
		if err != nil {
			api.HandleError(response, request, err)
			return
		}
		results = append(results, *b)
	}

	response.WriteEntity(results)
}

func (h *iamHandler) DeleteGroupBinding(request *restful.Request, response *restful.Response) {
	workspaceName := request.PathParameter("workspace")
	name := request.PathParameter("groupbinding")

	err := h.group.DeleteGroupBinding(workspaceName, name)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}
