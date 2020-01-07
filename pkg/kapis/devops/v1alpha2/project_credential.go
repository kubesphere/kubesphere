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
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
)

func (h ProjectPipelineHandler) CreateDevOpsProjectCredentialHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	var credential *devops.Credential
	err := request.ReadEntity(&credential)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	credentialId, err := h.projectCredentialOperator.CreateProjectCredential(projectId, username, credential)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: credentialId})
	return
}

func (h ProjectPipelineHandler) UpdateDevOpsProjectCredentialHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	credentialId := request.PathParameter("credential")
	var credential *devops.Credential
	err := request.ReadEntity(&credential)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	credentialId, err = h.projectCredentialOperator.UpdateProjectCredential(projectId, credentialId, credential)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: credentialId})
	return
}

func (h ProjectPipelineHandler) DeleteDevOpsProjectCredentialHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	credentialId := request.PathParameter("credential")

	credentialId, err := h.projectCredentialOperator.DeleteProjectCredential(projectId, credentialId)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(struct {
		Name string `json:"name"`
	}{Name: credentialId})
	return
}

func (h ProjectPipelineHandler) GetDevOpsProjectCredentialHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	credentialId := request.PathParameter("credential")
	getContent := request.QueryParameter("content")
	response, err := h.projectCredentialOperator.GetProjectCredential(projectId, credentialId, getContent)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(response)
	return
}

func (h ProjectPipelineHandler) GetDevOpsProjectCredentialsHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")

	jenkinsCredentials, err := h.projectCredentialOperator.GetProjectCredentials(projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(jenkinsCredentials)
	return
}
