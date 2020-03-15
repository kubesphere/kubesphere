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
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
)

func (h ProjectPipelineHandler) GetDevOpsProjectHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)

	err := h.projectOperator.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleForbidden(resp, nil, err)
		return
	}
	project, err := h.projectOperator.GetProject(projectId)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(project)
	return
}

func (h ProjectPipelineHandler) UpdateProjectHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	var project *v1alpha2.DevOpsProject
	err := request.ReadEntity(&project)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(resp, request, err)
		return
	}
	project.ProjectId = projectId
	err = h.projectOperator.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleForbidden(resp, nil, err)
		return
	}
	project, err = h.projectOperator.UpdateProject(project)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(project)
	return
}

func GetDevOpsProjectDefaultRoles(request *restful.Request, resp *restful.Response) {
	resp.WriteAsJson(devops.DefaultRoles)
	return
}
