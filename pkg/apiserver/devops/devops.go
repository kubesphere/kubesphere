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
	log "github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"net/http"
)

func GetPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

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
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.SearchPipelineRuns(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetBranchPipelineRun(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetPipelineRunNodesbyBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetPipelineRunNodesbyBranch(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")
	stepId := req.PathParameter("stepId")

	res, err := devops.GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func GetStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")
	stepId := req.PathParameter("stepId")

	res, err := devops.GetStepLog(projectName, pipelineName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func Validate(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scmId")

	res, err := devops.Validate(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetSCMOrg(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scmId")

	res, err := devops.GetSCMOrg(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetOrgRepo(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scmId")
	organizationId := req.PathParameter("organizationId")

	res, err := devops.GetOrgRepo(scmId, organizationId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func StopBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.StopBranchPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func StopPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")

	res, err := devops.StopPipeline(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func ReplayBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.ReplayBranchPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func ReplayPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")

	res, err := devops.ReplayPipeline(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetBranchRunLog(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func GetRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")

	res, err := devops.GetRunLog(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func GetBranchArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetBranchArtifacts(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")

	res, err := devops.GetArtifacts(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetPipeBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.GetPipeBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func CheckBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")
	stepId := req.PathParameter("stepId")

	res, err := devops.CheckBranchPipeline(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func CheckPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")
	stepId := req.PathParameter("stepId")

	res, err := devops.CheckPipeline(projectName, pipelineName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func GetConsoleLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.GetConsoleLog(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func ScanBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.ScanBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Write(res)
}

func RunBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")

	res, err := devops.RunBranchPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func RunPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.RunPipeline(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchStepsStatus(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")

	res, err := devops.GetBranchStepsStatus(projectName, pipelineName, branchName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetStepsStatus(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")

	res, err := devops.GetStepsStatus(projectName, pipelineName, runId, nodeId, req.Request)
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
	resBody, err := devops.CheckScriptCompile(req.Request)
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
	res, err := devops.CheckCron(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.WriteAsJson(res)
}

func GetPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")

	res, err := devops.GetPipelineRun(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")

	res, err := devops.GetBranchPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetPipelineRunNodes(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")

	res, err := devops.GetPipelineRunNodes(projectName, pipelineName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetBranchNodeSteps(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")

	res, err := devops.GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	resp.Write(res)
}

func GetNodeSteps(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")

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

func GithubWebhook(req *restful.Request, resp *restful.Response) {
	res, err := devops.GithubWebhook(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.Write(res)
}

func GetBranchNodesDetail(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetBranchNodesDetail(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.WriteAsJson(res)
}

func GetNodesDetail(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	runId := req.PathParameter("runId")

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
