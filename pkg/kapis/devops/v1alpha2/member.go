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

package v1alpha2

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"net/http"
)

func (h ProjectPipelineHandler) GetDevOpsProjectMembersHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)

	err := h.projectOperator.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	orderBy := request.QueryParameter(params.OrderByParam)
	reverse := params.GetBoolValueWithDefault(request, params.ReverseParam, false)
	limit, offset := params.ParsePaging(request)
	conditions, err := params.ParseConditions(request)

	project, err := h.projectMemberOperator.GetProjectMembers(projectId, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func (h ProjectPipelineHandler) GetDevOpsProjectMemberHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	member := request.PathParameter("member")

	err := h.projectOperator.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err := h.projectMemberOperator.GetProjectMember(projectId, member)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func (h ProjectPipelineHandler) AddDevOpsProjectMemberHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	member := &devops.ProjectMembership{}
	err := request.ReadEntity(&member)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	if govalidator.IsNull(member.Username) {
		err := fmt.Errorf("error need username")
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	if !reflectutils.In(member.Role, devops.AllRoleSlice) {
		err := fmt.Errorf("err role [%s] not in [%s]", member.Role,
			devops.AllRoleSlice)
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	err = h.projectOperator.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}

	member.GrantBy = username
	project, err := h.projectMemberOperator.AddProjectMember(projectId, member)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func (h ProjectPipelineHandler) UpdateDevOpsProjectMemberHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	member := &devops.ProjectMembership{}
	err := request.ReadEntity(&member)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	member.Username = request.PathParameter("member")
	if govalidator.IsNull(member.Username) {
		err := fmt.Errorf("error need username")
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	if username == member.Username {
		err := fmt.Errorf("you can not change your role")
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	if !reflectutils.In(member.Role, devops.AllRoleSlice) {
		err := fmt.Errorf("err role [%s] not in [%s]", member.Role,
			devops.AllRoleSlice)
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	err = h.projectOperator.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err := h.projectMemberOperator.UpdateProjectMember(projectId, member)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func (h ProjectPipelineHandler) DeleteDevOpsProjectMemberHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	member := request.PathParameter("member")

	err := h.projectOperator.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	username, err = h.projectMemberOperator.DeleteProjectMember(projectId, member)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(struct {
		Username string `json:"username"`
	}{Username: username})
	return
}
