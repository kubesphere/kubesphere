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
	"encoding/json"
	"github.com/emicklei/go-restful"
	log "k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"net/http"
	"strings"
)

const jenkinsHeaderPre = "X-"

func GetPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := devops.GetPipeline(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func SearchPipelines(req *restful.Request, resp *restful.Response) {
	res, err := devops.SearchPipelines(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func SearchPipelineRuns(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := devops.SearchPipelineRuns(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := devops.GetBranchPipelineRun(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetPipelineRunNodesbyBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := devops.GetPipelineRunNodesbyBranch(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	res, header, err := devops.GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)

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

func GetStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	res, header, err := devops.GetStepLog(projectName, pipelineName, runId, nodeId, stepId, req.Request)
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

func GetSCMServers(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := devops.GetSCMServers(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func CreateSCMServers(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := devops.CreateSCMServers(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func Validate(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := devops.Validate(scmId, req.Request)
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
	resp.Write(res)
}

func GetSCMOrg(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")

	res, err := devops.GetSCMOrg(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetOrgRepo(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scm")
	organizationId := req.PathParameter("organization")

	res, err := devops.GetOrgRepo(scmId, organizationId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func StopBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := devops.StopBranchPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func StopPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := devops.StopPipeline(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func ReplayBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := devops.ReplayBranchPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func ReplayPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := devops.ReplayPipeline(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := devops.GetBranchRunLog(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func GetRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := devops.GetRunLog(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func GetBranchArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := devops.GetBranchArtifacts(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := devops.GetArtifacts(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetPipeBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := devops.GetPipeBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func SubmitBranchInputStep(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	res, err := devops.SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func SubmitInputStep(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")
	stepId := req.PathParameter("step")

	res, err := devops.SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func GetConsoleLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := devops.GetConsoleLog(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func ScanBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := devops.ScanBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func RunBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")

	res, err := devops.RunBranchPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func RunPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	res, err := devops.RunPipeline(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetCrumb(req *restful.Request, resp *restful.Response) {
	res, err := devops.GetCrumb(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func CheckScriptCompile(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")

	resBody, err := devops.CheckScriptCompile(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	// Jenkins will return different struct according to different results.
	var resJson = new(devops.CheckScript)
	if ok := json.Unmarshal(resBody, &resJson); ok != nil {
		var resJson []interface{}
		err := json.Unmarshal(resBody, &resJson)
		if err != nil {
			resp.WriteError(http.StatusInternalServerError, err)
			return
		}
		resp.WriteAsJson(resJson[0])
		return

	}

	resp.WriteAsJson(resJson)
}

func CheckCron(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")

	res, err := devops.CheckCron(projectName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func GetPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := devops.GetPipelineRun(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")

	res, err := devops.GetBranchPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetPipelineRunNodes(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := devops.GetPipelineRunNodes(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchNodeSteps(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")

	res, err := devops.GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetNodeSteps(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")
	nodeId := req.PathParameter("node")

	res, err := devops.GetNodeSteps(projectName, pipelineName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func ToJenkinsfile(req *restful.Request, resp *restful.Response) {
	res, err := devops.ToJenkinsfile(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func ToJson(req *restful.Request, resp *restful.Response) {
	res, err := devops.ToJson(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetNotifyCommit(req *restful.Request, resp *restful.Response) {
	res, err := devops.GetNotifyCommit(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func PostNotifyCommit(req *restful.Request, resp *restful.Response) {
	res, err := devops.GetNotifyCommit(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}
func GithubWebhook(req *restful.Request, resp *restful.Response) {
	res, err := devops.GithubWebhook(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func GetBranchNodesDetail(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	branchName := req.PathParameter("branch")
	runId := req.PathParameter("run")

	res, err := devops.GetBranchNodesDetail(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.WriteAsJson(res)
}

func GetNodesDetail(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("devops")
	pipelineName := req.PathParameter("pipeline")
	runId := req.PathParameter("run")

	res, err := devops.GetNodesDetail(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.WriteAsJson(res)
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
