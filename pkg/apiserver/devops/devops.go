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

	_ = resp.WriteAsJson(res)
}

func SearchPipelines(req *restful.Request, resp *restful.Response) {
	res, err := devops.SearchPipelines(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func SearchPipelineRuns(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.SearchPipelineRuns(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func GetPipelineRun(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetPipelineRun(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func GetPipelineRunNodes(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetPipelineRunNodes(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func GetStepLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")
	stepId := req.PathParameter("stepId")

	res, err := devops.GetStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}
	_, _ = resp.Write(res)
}

func Validate(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scmId")

	res, err := devops.Validate(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_, _ = resp.Write(res)
}

func GetSCMOrg(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scmId")

	res, err := devops.GetSCMOrg(scmId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func GetOrgRepo(req *restful.Request, resp *restful.Response) {
	scmId := req.PathParameter("scmId")
	organizationId := req.PathParameter("organizationId")

	res, err := devops.GetOrgRepo(scmId, organizationId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func StopPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.StopPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func ReplayPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.ReplayPipeline(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func GetRunLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetRunLog(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_, _ = resp.Write(res)
}

func GetArtifacts(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")

	res, err := devops.GetArtifacts(projectName, pipelineName, branchName, runId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func GetPipeBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.GetPipeBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func CheckPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")
	stepId := req.PathParameter("stepId")

	res, err := devops.CheckPipeline(projectName, pipelineName, branchName, runId, nodeId, stepId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_, _ = resp.Write(res)
}

func GetConsoleLog(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.GetConsoleLog(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_, _ = resp.Write(res)
}

func ScanBranch(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")

	res, err := devops.ScanBranch(projectName, pipelineName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_, _ = resp.Write(res)
}

func RunPipeline(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")

	res, err := devops.RunPipeline(projectName, pipelineName, branchName, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func GetStepsStatus(req *restful.Request, resp *restful.Response) {
	projectName := req.PathParameter("projectName")
	pipelineName := req.PathParameter("pipelineName")
	branchName := req.PathParameter("branchName")
	runId := req.PathParameter("runId")
	nodeId := req.PathParameter("nodeId")

	res, err := devops.GetStepsStatus(projectName, pipelineName, branchName, runId, nodeId, req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func GetCrumb(req *restful.Request, resp *restful.Response) {

	res, err := devops.GetCrumb(req.Request)
	if err != nil {
		parseErr(err, resp)
		return
	}

	_ = resp.WriteAsJson(res)
}

func parseErr(err error, resp *restful.Response) {
	log.Error(err)
	if jErr, ok := err.(*devops.JkError); ok {
		_ = resp.WriteError(jErr.Code, err)
	} else {
		_ = resp.WriteError(http.StatusInternalServerError, err)
	}
	return
}
