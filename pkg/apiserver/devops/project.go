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
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
)

func GetDevOpsProjectHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)

	err := devops.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err := devops.GetProject(projectId)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func UpdateProjectHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	var project *v1alpha2.DevOpsProject
	err := request.ReadEntity(&project)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	project.ProjectId = projectId
	err = devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err = devops.UpdateProject(project)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func GetDevOpsProjectDefaultRoles(request *restful.Request, resp *restful.Response) {
	resp.WriteAsJson(devops.DefaultRoles)
	return
}
