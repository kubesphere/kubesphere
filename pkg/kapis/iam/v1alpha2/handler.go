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

	"kubesphere.io/kubesphere/pkg/models/auth"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
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

func (h *iamHandler) CreateUser(req *restful.Request, resp *restful.Response) {
	//var user iamv1alpha2.User
	//err := req.ReadEntity(&user)
	//if err != nil {
	//	api.HandleBadRequest(resp, req, err)
	//	return
	//}
	//operator, ok := request.UserFrom(req.Request.Context())
	//if ok && operator.GetName() == iamv1alpha2.PreRegistrationUser {
	//	extra := operator.GetExtra()
	//	// The token used for registration must contain additional information
	//	if len(extra[iamv1alpha2.ExtraIdentityProvider]) != 1 || len(extra[iamv1alpha2.ExtraUID]) != 1 {
	//		err = errors.NewBadRequest("invalid registration token")
	//		api.HandleBadRequest(resp, req, err)
	//		return
	//	}
	//	if user.Labels == nil {
	//		user.Labels = make(map[string]string)
	//	}
	//	user.Labels[iamv1alpha2.IdentifyProviderLabel] = extra[iamv1alpha2.ExtraIdentityProvider][0]
	//	user.Labels[iamv1alpha2.OriginUIDLabel] = extra[iamv1alpha2.ExtraUID][0]
	//	// default role
	//	delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)
	//}
	//
	//globalRole := user.Annotations[iamv1alpha2.GlobalRoleAnnotation]
	//delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)
	//if globalRole != "" {
	//	if _, err = h.am.GetGlobalRole(globalRole); err != nil {
	//		api.HandleError(resp, req, err)
	//		return
	//	}
	//}
	//
	//created, err := h.im.CreateUser(&user)
	//if err != nil {
	//	api.HandleError(resp, req, err)
	//	return
	//}
	//
	//if globalRole != "" {
	//	if err := h.am.CreateGlobalRoleBinding(user.Name, globalRole); err != nil {
	//		api.HandleError(resp, req, err)
	//		return
	//	}
	//}
	//
	//// ensure encrypted password will not be output
	//created.Spec.EncryptedPassword = ""
	//
	//resp.WriteEntity(created)

	panic("Need implement")
}

func (h *iamHandler) UpdateUser(request *restful.Request, response *restful.Response) {
	//username := request.PathParameter("user")
	//
	//var user iamv1alpha2.User
	//
	//err := request.ReadEntity(&user)
	//if err != nil {
	//	api.HandleBadRequest(response, request, err)
	//	return
	//}
	//
	//if username != user.Name {
	//	err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", user.Name, username)
	//	api.HandleBadRequest(response, request, err)
	//	return
	//}
	//
	//globalRole := user.Annotations[iamv1alpha2.GlobalRoleAnnotation]
	//delete(user.Annotations, iamv1alpha2.GlobalRoleAnnotation)
	//
	//updated, err := h.im.UpdateUser(&user)
	//if err != nil {
	//	api.HandleError(response, request, err)
	//	return
	//}
	//
	//operator, ok := apirequest.UserFrom(request.Request.Context())
	//if globalRole != "" && ok {
	//	err = h.updateGlobalRoleBinding(operator, updated, globalRole)
	//	if err != nil {
	//		api.HandleError(response, request, err)
	//		return
	//	}
	//	updated = appendGlobalRoleAnnotation(updated, globalRole)
	//}
	//
	//response.WriteEntity(updated)
	panic("Need implement")
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
