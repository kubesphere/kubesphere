/*
Copyright 2020 The KubeSphere Authors.

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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/apiserver/pkg/authentication/user"
	log "k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/devops"
	clientDevOps "kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"strings"
)

const jenkinsHeaderPre = "X-"

func (h *ProjectPipelineHandler) GetPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.GetPipeline(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ListPipelines(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.ListPipelines(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetPipelineRun(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ListPipelineRuns(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.ListPipelineRuns(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) StopPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.StopPipeline(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ReplayPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.ReplayPipeline(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) RunPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.RunPipeline(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetArtifacts(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetRunLog(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	res, header, err := h.devopsOperator.GetStepLog(projectName, pipelineName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	for k, v := range header {
		if strings.HasPrefix(k, jenkinsHeaderPre) {
			resp.AddHeader(k, v[0])
		}
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetNodeSteps(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")

	res, err := h.devopsOperator.GetNodeSteps(projectName, pipelineName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetPipelineRunNodes(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetPipelineRunNodes(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) approvableCheck(nodes []clientDevOps.NodesDetail, req *restful.Request) {
	currentUserName, roleName := h.getCurrentUser(req)
	// check if current user belong to the admin group, grant it if it's true
	isAdmin := roleName == iamv1alpha2.PlatformAdmin

	for i, node := range nodes {
		if node.State != clientDevOps.StatePaused {
			continue
		}

		for j, step := range node.Steps {
			if step.State != clientDevOps.StatePaused || step.Input == nil {
				continue
			}

			nodes[i].Steps[j].Approvable = isAdmin || step.Input.Approvable(currentUserName)
		}
	}
}

func (h *ProjectPipelineHandler) createdBy(projectName string, pipelineName string, currentUserName string) bool {
	if pipeline, err := h.devopsOperator.GetPipelineObj(projectName, pipelineName); err == nil {
		if creator, ok := pipeline.Annotations[constants.CreatorAnnotationKey]; ok {
			return creator == currentUserName
		}
	} else {
		log.V(4).Infof("cannot get pipeline %s/%s, error %#v", projectName, pipelineName, err)
	}
	return false
}

func (h *ProjectPipelineHandler) getCurrentUser(req *restful.Request) (username, roleName string) {
	var userInfo user.Info
	var ok bool
	var err error

	ctx := req.Request.Context()
	if userInfo, ok = request.UserFrom(ctx); ok {
		var role *iamv1alpha2.GlobalRole
		username = userInfo.GetName()
		if role, err = h.amInterface.GetGlobalRoleOfUser(username); err == nil {
			roleName = role.Name
		}
	}
	return
}

func (h *ProjectPipelineHandler) hasSubmitPermission(req *restful.Request) (hasPermit bool, err error) {
	currentUserName, roleName := h.getCurrentUser(req)
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	// check if current user belong to the admin group or he's the owner, grant it if it's true
	if roleName == iamv1alpha2.PlatformAdmin || h.createdBy(projectName, pipelineName, currentUserName) {
		hasPermit = true
		return
	}

	// step 2, check if current user if was addressed
	httpReq := &http.Request{
		URL:      req.Request.URL,
		Header:   req.Request.Header,
		Form:     req.Request.Form,
		PostForm: req.Request.PostForm,
	}

	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	// check if current user can approve this input
	var res []clientDevOps.NodesDetail
	if res, err = h.devopsOperator.GetNodesDetail(projectName, pipelineName, runId, httpReq); err == nil {
		for _, node := range res {
			if node.ID != nodeId {
				continue
			}

			for _, step := range node.Steps {
				if step.ID != stepId || step.Input == nil {
					continue
				}

				hasPermit = step.Input.Approvable(currentUserName)
				break
			}
			break
		}
	} else {
		log.V(4).Infof("cannot get nodes detail, error: %v", err)
		err = errors.New("cannot get the submitters of current pipeline run")
		return
	}
	return
}

func (h *ProjectPipelineHandler) SubmitInputStep(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	var response []byte
	var err error
	var ok bool

	if ok, err = h.hasSubmitPermission(req); !ok || err != nil {
		msg := map[string]string{
			"allow":   "false",
			"message": fmt.Sprintf("%v", err),
		}

		response, _ = json.Marshal(msg)
	} else {
		response, err = h.devopsOperator.SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId, req.Request)
		if err != nil {
			parseErr(err, resp)
			return
		}
	}
	resp.Write(response)
}

func (h *ProjectPipelineHandler) GetNodesDetail(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetNodesDetail(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	h.approvableCheck(res, req)

	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")

	res, err := h.devopsOperator.GetBranchPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchPipelineRun(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) StopBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.StopBranchPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ReplayBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.ReplayBranchPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) RunBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")

	res, err := h.devopsOperator.RunBranchPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchArtifacts(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchRunLog(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetBranchStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	res, header, err := h.devopsOperator.GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)

	if err != nil {
		parseErr(err, resp)
		return
	}
	for k, v := range header {
		if strings.HasPrefix(k, jenkinsHeaderPre) {
			resp.AddHeader(k, v[0])
		}
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetBranchNodeSteps(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")

	res, err := h.devopsOperator.GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetBranchPipelineRunNodes(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) SubmitBranchInputStep(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	var response []byte
	var err error
	var ok bool

	if ok, err = h.hasSubmitPermission(req); !ok || err != nil {
		msg := map[string]string{
			"allow":   "false",
			"message": fmt.Sprintf("%v", err),
		}

		response, _ = json.Marshal(msg)
	} else {
		response, err = h.devopsOperator.SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)
		if err != nil {
			parseErr(err, resp)
			return
		}
	}

	resp.Write(response)
}

func (h *ProjectPipelineHandler) GetBranchNodesDetail(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := h.devopsOperator.GetBranchNodesDetail(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	h.approvableCheck(res, req)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetPipelineBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.GetPipelineBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ScanBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.ScanBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetConsoleLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := h.devopsOperator.GetConsoleLog(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func (h *ProjectPipelineHandler) GetCrumb(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.GetCrumb(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetSCMServers(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := h.devopsOperator.GetSCMServers(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetSCMOrg(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := h.devopsOperator.GetSCMOrg(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetOrgRepo(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")
	organizationId := req.PathParameter("organization")

	res, err := h.devopsOperator.GetOrgRepo(scmId, organizationId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) CreateSCMServers(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := h.devopsOperator.CreateSCMServers(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) Validate(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := h.devopsOperator.Validate(scmId, req.Request)
	if err != nil {
		log.Error(err)
		if jErr, ok := err.(*devops.JkError); ok {
			if jErr.Code != http.StatusUnauthorized {
				resp.WriteError(jErr.Code, err)
			} else {
				resp.WriteHeader(http.StatusPreconditionRequired)
			}
		} else {
			resp.WriteError(http.StatusInternalServerError, err)
		}
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetNotifyCommit(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.GetNotifyCommit(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) PostNotifyCommit(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.GetNotifyCommit(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) GithubWebhook(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.GithubWebhook(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func (h *ProjectPipelineHandler) CheckScriptCompile(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	resBody, err := h.devopsOperator.CheckScriptCompile(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.WriteAsJson(resBody)
}

func (h *ProjectPipelineHandler) CheckCron(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")

	res, err := h.devopsOperator.CheckCron(projectName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ToJenkinsfile(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.ToJenkinsfile(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) ToJson(req *restful.Request, resp *restful.Response) {
	res, err := h.devopsOperator.ToJson(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func (h *ProjectPipelineHandler) GetProjectCredentialUsage(req *restful.Request, resp *restful.Response) {
	projectId := req.PathParameter("devops")
	credentialId := req.PathParameter("credential")
	response, err := h.projectCredentialGetter.GetProjectCredentialUsage(projectId, credentialId)
	if err != nil {
		log.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteAsJson(response)
	return

}

func parseErr(err error, resp *restful.Response) {
	log.Error(err)
	if jErr, ok := err.(*devops.JkError); ok {
		resp.WriteError(jErr.Code, err)
	} else {
		resp.WriteError(http.StatusInternalServerError, err)
	}
	return
}
