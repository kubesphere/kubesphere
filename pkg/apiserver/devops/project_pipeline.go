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
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
)

func CreateDevOpsProjectPipelineHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	var pipeline *devops.ProjectPipeline
	err := request.ReadEntity(&pipeline)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	xuseranme, xpassword, ok := request.Request.BasicAuth()
	if !ok {
		err := fmt.Errorf("basic auth not found")
		klog.Error("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	server, err := devops.NewServer(xuseranme, xpassword)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	pipelineName, err := server.CreateProjectPipeline(projectId, pipeline)

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
	pipelineId := request.PathParameter("pipeline")

	xuseranme, xpassword, ok := request.Request.BasicAuth()
	if !ok {
		err := fmt.Errorf("basic auth not found")
		klog.Error("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	server, err := devops.NewServer(xuseranme, xpassword)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	pipelineName, err := server.DeleteProjectPipeline(projectId, pipelineId)

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
	pipelineId := request.PathParameter("pipeline")
	var pipeline *devops.ProjectPipeline
	err := request.ReadEntity(&pipeline)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	xuseranme, xpassword, ok := request.Request.BasicAuth()
	if !ok {
		err := fmt.Errorf("basic auth not found")
		klog.Error("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	server, err := devops.NewServer(xuseranme, xpassword)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	pipelineName, err := server.UpdateProjectPipeline(projectId, pipelineId, pipeline)

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
	pipelineId := request.PathParameter("pipeline")

	xuseranme, xpassword, ok := request.Request.BasicAuth()
	if !ok {
		err := fmt.Errorf("basic auth not found")
		klog.Error("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	server, err := devops.NewServer(xuseranme, xpassword)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	pipeline, err := server.GetProjectPipeline(projectId, pipelineId)

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
	pipelineId := request.PathParameter("pipeline")
	xuseranme, xpassword, ok := request.Request.BasicAuth()
	if !ok {
		err := fmt.Errorf("basic auth not found")
		klog.Error("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	server, err := devops.NewServer(xuseranme, xpassword)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	sonarStatus, err := server.GetPipelineSonar(projectId, pipelineId)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(sonarStatus)
}

func GetMultiBranchesPipelineSonarStatusHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	pipelineId := request.PathParameter("pipeline")
	branchId := request.PathParameter("branch")
	xuseranme, xpassword, ok := request.Request.BasicAuth()
	if !ok {
		err := fmt.Errorf("basic auth not found")
		klog.Error("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	server, err := devops.NewServer(xuseranme, xpassword)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	sonarStatus, err := server.GetMultiBranchPipelineSonar(projectId, pipelineId, branchId)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(sonarStatus)
}
