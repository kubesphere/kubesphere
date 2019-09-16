/*

 Copyright 2019 The KubeSphere Authors.

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
package iam

import (
	"github.com/emicklei/go-restful"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/models/workspaces"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
)

func ListWorkspaceRoles(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("workspace")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := iam.ListWorkspaceRoles(workspace, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result.Items)
}

func ListWorkspaceRoleRules(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	role := req.PathParameter("role")

	rules := iam.GetWorkspaceRoleSimpleRules(workspace, role)

	resp.WriteAsJson(rules)
}

func DescribeWorkspaceRole(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	roleName := req.PathParameter("role")

	role, err := iam.GetWorkspaceRole(workspace, roleName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(role)
}

func DescribeWorkspaceUser(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	username := req.PathParameter("member")

	workspaceRole, err := iam.GetUserWorkspaceRole(workspace, username)

	if err != nil {
		if k8serr.IsNotFound(err) {
			resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		} else {
			resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		}

		return
	}

	user, err := iam.GetUserInfo(username)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	user.WorkspaceRole = workspaceRole.Annotations[constants.DisplayNameAnnotationKey]

	resp.WriteAsJson(user)
}

func ListDevopsRoleRules(req *restful.Request, resp *restful.Response) {
	role := req.PathParameter("role")

	rules := iam.GetDevopsRoleSimpleRules(role)

	resp.WriteAsJson(rules)
}

func InviteUser(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	var user models.User
	err := req.ReadEntity(&user)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	err = workspaces.InviteUser(workspace, &user)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func RemoveUser(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	username := req.PathParameter("member")

	err := workspaces.RemoveUser(workspace, username)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(errors.None)
}

func ListWorkspaceUsers(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := iam.ListWorkspaceUsers(workspace, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}
