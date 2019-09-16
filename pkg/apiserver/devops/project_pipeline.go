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
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
)

func CreateDevOpsProjectPipelineHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	var pipeline *devops.ProjectPipeline
	err := request.ReadEntity(&pipeline)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	err = devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	pipelineName, err := devops.CreateProjectPipeline(projectId, pipeline)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: pipelineName})
	return
}

func DeleteDevOpsProjectPipelineHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")

	err := devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	pipelineName, err := devops.DeleteProjectPipeline(projectId, pipelineId)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: pipelineName})
	return
}

func UpdateDevOpsProjectPipelineHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")
	var pipeline *devops.ProjectPipeline
	err := request.ReadEntity(&pipeline)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	err = devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	pipelineName, err := devops.UpdateProjectPipeline(projectId, pipelineId, pipeline)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: pipelineName})
	return
}

func GetDevOpsProjectPipelineHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")

	err := devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	pipeline, err := devops.GetProjectPipeline(projectId, pipelineId)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(pipeline)
	return
}

func GetPipelineSonarStatusHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")
	err := devops.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	sonarStatus, err := devops.GetPipelineSonar(projectId, pipelineId)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(sonarStatus)
}

func GetMultiBranchesPipelineSonarStatusHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")
	branchId := request.PathParameter("branch")
	err := devops.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	sonarStatus, err := devops.GetMultiBranchPipelineSonar(projectId, pipelineId, branchId)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(sonarStatus)
}
