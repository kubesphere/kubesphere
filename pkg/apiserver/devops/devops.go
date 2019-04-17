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
	"kubesphere.io/kubesphere/pkg/errors"
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

func parseErr(err error, resp *restful.Response) {
	if jErr, ok := err.(*devops.JkError); ok {
		_ = resp.WriteHeaderAndEntity(jErr.Code, err)
	} else {
		_ = resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
	}
	return
}
