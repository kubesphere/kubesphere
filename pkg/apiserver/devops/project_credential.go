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

func CreateDevOpsProjectCredentialHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	var credential *devops.JenkinsCredential
	err := request.ReadEntity(&credential)
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
	credentialId, err := devops.CreateProjectCredential(projectId, username, credential)

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

func UpdateDevOpsProjectCredentialHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	credentialId := request.PathParameter("credential")
	var credential *devops.JenkinsCredential
	err := request.ReadEntity(&credential)
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
	credentialId, err = devops.UpdateProjectCredential(projectId, credentialId, credential)

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

func DeleteDevOpsProjectCredentialHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	credentialId := request.PathParameter("credential")
	var credential *devops.JenkinsCredential
	err := request.ReadEntity(&credential)
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
	credentialId, err = devops.DeleteProjectCredential(projectId, credentialId, credential)

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

func GetDevOpsProjectCredentialHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	credentialId := request.PathParameter("credential")
	getContent := request.QueryParameter("content")
	domain := request.QueryParameter("domain")

	err := devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	response, err := devops.GetProjectCredential(projectId, credentialId, domain, getContent)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(response)
	return
}

func GetDevOpsProjectCredentialsHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	domain := request.QueryParameter("domain")

	err := devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner, devops.ProjectMaintainer})
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	jenkinsCredentials, err := devops.GetProjectCredentials(projectId, domain)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(jenkinsCredentials)
	return
}
