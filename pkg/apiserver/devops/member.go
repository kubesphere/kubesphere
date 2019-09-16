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

package devops

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"net/http"
)

func GetDevOpsProjectMembersHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)

	err := devops.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	orderBy := request.QueryParameter(params.OrderByParam)
	reverse := params.ParseReverse(request)
	limit, offset := params.ParsePaging(request.QueryParameter(params.PagingParam))
	conditions, err := params.ParseConditions(request.QueryParameter(params.ConditionsParam))

	project, err := devops.GetProjectMembers(projectId, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func GetDevOpsProjectMemberHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	member := request.PathParameter("member")

	err := devops.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err := devops.GetProjectMember(projectId, member)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func AddDevOpsProjectMemberHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	member := &devops.DevOpsProjectMembership{}
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

	err = devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err := devops.AddProjectMember(projectId, username, member)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func UpdateDevOpsProjectMemberHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	member := &devops.DevOpsProjectMembership{}
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

	err = devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err := devops.UpdateProjectMember(projectId, username, member)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func DeleteDevOpsProjectMemberHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	member := request.PathParameter("member")

	err := devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	username, err = devops.DeleteProjectMember(projectId, member)
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
