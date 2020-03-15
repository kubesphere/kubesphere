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
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
)

func (h ProjectPipelineHandler) CreateDevOpsProjectPipelineHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	var pipeline *devops.ProjectPipeline
	err := request.ReadEntity(&pipeline)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	err = h.projectOperator.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleForbidden(resp, nil, err)
		return
	}
	pipelineName, err := h.projectPipelineOperator.CreateProjectPipeline(projectId, pipeline)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: pipelineName})
	return
}

func (h ProjectPipelineHandler) DeleteDevOpsProjectPipelineHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")

	err := h.projectOperator.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleForbidden(resp, request, err)
		return
	}
	pipelineName, err := h.projectPipelineOperator.DeleteProjectPipeline(projectId, pipelineId)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: pipelineName})
	return
}

func (h ProjectPipelineHandler) UpdateDevOpsProjectPipelineHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")
	var pipeline *devops.ProjectPipeline
	err := request.ReadEntity(&pipeline)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleBadRequest(resp, nil, err)
		return
	}
	err = h.projectOperator.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleForbidden(resp, nil, err)
		return
	}
	pipelineName, err := h.projectPipelineOperator.UpdateProjectPipeline(projectId, pipelineId, pipeline)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: pipelineName})
	return
}

func (h ProjectPipelineHandler) GetDevOpsProjectPipelineConfigHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")

	err := h.projectOperator.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleForbidden(resp, nil, err)
		return
	}
	pipeline, err := h.projectPipelineOperator.GetProjectPipelineConfig(projectId, pipelineId)

	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteAsJson(pipeline)
	return
}
