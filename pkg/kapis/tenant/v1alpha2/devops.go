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
	devopsv1alpha2 "kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
)

func (h *tenantHandler) DeleteDevOpsProjectHandler(req *restful.Request, resp *restful.Response) {
	projectId := req.PathParameter("devops")
	workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	_, err := h.tenant.GetWorkspace(workspaceName)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	err = h.tenant.DeleteDevOpsProject(projectId, username)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(errors.None)
}

func (h *tenantHandler) CreateDevOpsProjectHandler(req *restful.Request, resp *restful.Response) {

	workspaceName := req.PathParameter("workspace")
	username := req.HeaderParameter(constants.UserNameHeader)

	var devops devopsv1alpha2.DevOpsProject

	err := req.ReadEntity(&devops)

	if err != nil {
		klog.Infof("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	klog.Infoln("create workspace", username, workspaceName, devops)
	project, err := h.tenant.CreateDevOpsProject(username, workspaceName, &devops)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
}

func (h *tenantHandler) GetDevOpsProjectsCountHandler(req *restful.Request, resp *restful.Response) {
	username := req.HeaderParameter(constants.UserNameHeader)

	result, err := h.tenant.GetDevOpsProjectsCount(username)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(struct {
		Count uint32 `json:"count"`
	}{Count: result})
}

func (h *tenantHandler) ListDevOpsProjectsHandler(req *restful.Request, resp *restful.Response) {

	workspace := req.PathParameter("workspace")
	username := req.PathParameter("member")
	if username == "" {
		username = req.HeaderParameter(constants.UserNameHeader)
	}
	orderBy := req.QueryParameter(params.OrderByParam)
	reverse := params.GetBoolValueWithDefault(req, params.ReverseParam, false)
	limit, offset := params.ParsePaging(req)
	conditions, err := params.ParseConditions(req)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}

	result, err := h.tenant.ListDevOpsProjects(workspace, username, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(result)
}

func (h *tenantHandler) ListDevOpsRules(req *restful.Request, resp *restful.Response) {

	devops := req.PathParameter("devops")
	username := req.HeaderParameter(constants.UserNameHeader)

	rules, err := h.tenant.GetUserDevOpsSimpleRules(username, devops)

	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(rules)
}

func (h *tenantHandler) ListDevopsRules(req *restful.Request, resp *restful.Response) {

	devops := req.PathParameter("devops")
	username := req.HeaderParameter(constants.UserNameHeader)

	rules, err := h.tenant.GetUserDevOpsSimpleRules(username, devops)

	if err != nil {
		api.HandleInternalError(resp, err)
		return
	}

	resp.WriteAsJson(rules)
}
