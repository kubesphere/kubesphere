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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful"
	"io"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"

	"k8s.io/klog"
	"net/http"
	"sync"
)

const (
	channelMaxCapacity = 100
)

type DevopsOperator interface {
	GetPipeline(projectName, pipelineName string, req *http.Request) (*devops.Pipeline, error)
	ListPipelines(req *http.Request) (*devops.PipelineList, error)
	GetPipelineRun(projectName, pipelineName, runId string, req *http.Request) (*devops.PipelineRun, error)
	ListPipelineRuns(projectName, pipelineName string, req *http.Request) (*devops.PipelineRunList, error)
	StopPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.StopPipeline, error)
	ReplayPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.ReplayPipeline, error)
	RunPipeline(projectName, pipelineName string, req *http.Request) (*devops.RunPipeline, error)
	GetArtifacts(projectName, pipelineName, runId string, req *http.Request) ([]devops.Artifacts, error)
	GetRunLog(projectName, pipelineName, runId string, req *http.Request) ([]byte, error)
	GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error)
	GetNodeSteps(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error)
	GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]devops.PipelineRunNodes, error)
	SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error)
	GetNodesDetail(projectName, pipelineName, runId string, req *http.Request) ([]devops.NodesDetail, error)

	GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) (*devops.BranchPipeline, error)
	GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.PipelineRun, error)
	StopBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.StopPipeline, error)
	ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.ReplayPipeline, error)
	RunBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) (*devops.RunPipeline, error)
	GetBranchArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.Artifacts, error)
	GetBranchRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error)
	GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error)
	GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error)
	GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.BranchPipelineRunNodes, error)
	SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error)
	GetBranchNodesDetail(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.NodesDetail, error)
	GetPipelineBranch(projectName, pipelineName string, req *http.Request) (*devops.PipelineBranch, error)
	ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error)

	GetConsoleLog(projectName, pipelineName string, req *http.Request) ([]byte, error)
	GetCrumb(req *http.Request) (*devops.Crumb, error)

	GetSCMServers(scmId string, req *http.Request) ([]devops.SCMServer, error)
	GetSCMOrg(scmId string, req *http.Request) ([]devops.SCMOrg, error)
	GetOrgRepo(scmId, organizationId string, req *http.Request) ([]devops.OrgRepo, error)
	CreateSCMServers(scmId string, req *http.Request) (*devops.SCMServer, error)
	Validate(scmId string, req *http.Request) (*devops.Validates, error)

	GetNotifyCommit(req *http.Request) ([]byte, error)
	GithubWebhook(req *http.Request) ([]byte, error)

	CheckScriptCompile(projectName, pipelineName string, req *http.Request) (*devops.CheckScript, error)
	CheckCron(projectName string, req *http.Request) (*devops.CheckCronRes, error)

	ToJenkinsfile(req *http.Request) (*devops.ResJenkinsfile, error)
	ToJson(req *http.Request) (*devops.ResJson, error)
}

type devopsOperator struct {
	devopsClient devops.Interface
}

func NewDevopsOperator(client devops.Interface) DevopsOperator {
	return &devopsOperator{devopsClient: client}
}

func convertToHttpParameters(req *http.Request) *devops.HttpParameters {
	httpParameters := devops.HttpParameters{
		Method:   req.Method,
		Header:   req.Header,
		Body:     req.Body,
		Form:     req.Form,
		PostForm: req.PostForm,
		Url:      req.URL,
	}

	return &httpParameters
}

func (d devopsOperator) GetPipeline(projectName, pipelineName string, req *http.Request) (*devops.Pipeline, error) {

	res, err := d.devopsClient.GetPipeline(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) ListPipelines(req *http.Request) (*devops.PipelineList, error) {

	res, err := d.devopsClient.ListPipelines(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) GetPipelineRun(projectName, pipelineName, runId string, req *http.Request) (*devops.PipelineRun, error) {

	res, err := d.devopsClient.GetPipelineRun(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) ListPipelineRuns(projectName, pipelineName string, req *http.Request) (*devops.PipelineRunList, error) {

	res, err := d.devopsClient.ListPipelineRuns(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) StopPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.StopPipeline, error) {

	req.Method = http.MethodPut
	res, err := d.devopsClient.StopPipeline(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ReplayPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.ReplayPipeline, error) {

	res, err := d.devopsClient.ReplayPipeline(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) RunPipeline(projectName, pipelineName string, req *http.Request) (*devops.RunPipeline, error) {

	res, err := d.devopsClient.RunPipeline(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetArtifacts(projectName, pipelineName, runId string, req *http.Request) ([]devops.Artifacts, error) {

	res, err := d.devopsClient.GetArtifacts(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetRunLog(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {

	res, err := d.devopsClient.GetRunLog(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {

	resBody, header, err := d.devopsClient.GetStepLog(projectName, pipelineName, runId, nodeId, stepId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	return resBody, header, err
}

func (d devopsOperator) GetNodeSteps(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error) {
	res, err := d.devopsClient.GetNodeSteps(projectName, pipelineName, runId, nodeId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]devops.PipelineRunNodes, error) {

	res, err := d.devopsClient.GetPipelineRunNodes(projectName, pipelineName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	fmt.Println()

	return res, err
}

func (d devopsOperator) SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {

	newBody, err := getInputReqBody(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = newBody

	resBody, err := d.devopsClient.SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetNodesDetail(projectName, pipelineName, runId string, req *http.Request) ([]devops.NodesDetail, error) {
	var wg sync.WaitGroup
	var nodesDetails []devops.NodesDetail
	stepChan := make(chan *devops.NodesStepsIndex, channelMaxCapacity)

	respNodes, err := d.GetPipelineRunNodes(projectName, pipelineName, runId, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	Nodes, err := json.Marshal(respNodes)
	err = json.Unmarshal(Nodes, &nodesDetails)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// get all steps in nodes.
	for i, v := range respNodes {
		wg.Add(1)
		go func(nodeId string, index int) {
			Steps, err := d.GetNodeSteps(projectName, pipelineName, runId, nodeId, req)
			if err != nil {
				klog.Error(err)
				return
			}

			stepChan <- &devops.NodesStepsIndex{Id: index, Steps: Steps}
			wg.Done()
		}(v.ID, i)
	}
	wg.Wait()
	close(stepChan)

	for oneNodeSteps := range stepChan {
		if oneNodeSteps != nil {
			nodesDetails[oneNodeSteps.Id].Steps = append(nodesDetails[oneNodeSteps.Id].Steps, oneNodeSteps.Steps...)
		}
	}

	return nodesDetails, err
}

func (d devopsOperator) GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) (*devops.BranchPipeline, error) {

	res, err := d.devopsClient.GetBranchPipeline(projectName, pipelineName, branchName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.PipelineRun, error) {

	res, err := d.devopsClient.GetBranchPipelineRun(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) StopBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.StopPipeline, error) {

	req.Method = http.MethodPut
	res, err := d.devopsClient.StopBranchPipeline(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.ReplayPipeline, error) {

	res, err := d.devopsClient.ReplayBranchPipeline(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) RunBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) (*devops.RunPipeline, error) {

	res, err := d.devopsClient.RunBranchPipeline(projectName, pipelineName, branchName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.Artifacts, error) {

	res, err := d.devopsClient.GetBranchArtifacts(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {

	res, err := d.devopsClient.GetBranchRunLog(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {

	resBody, header, err := d.devopsClient.GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	return resBody, header, err
}

func (d devopsOperator) GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error) {

	res, err := d.devopsClient.GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, req *http.Request) (*devops.BranchPipelineRunNodes, error) {

	res, err := d.devopsClient.GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {

	newBody, err := getInputReqBody(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = newBody
	resBody, err := d.devopsClient.SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetBranchNodesDetail(projectName, pipelineName, branchName, runId string, req *http.Request) ([]devops.NodesDetail, error) {
	var wg sync.WaitGroup
	var nodesDetails []devops.NodesDetail
	stepChan := make(chan *devops.NodesStepsIndex, channelMaxCapacity)

	respNodes, err := d.GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	Nodes, err := json.Marshal(respNodes)
	err = json.Unmarshal(Nodes, &nodesDetails)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// get all steps in nodes.
	for i, v := range nodesDetails {
		wg.Add(1)
		go func(nodeId string, index int) {
			Steps, err := d.GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, req)
			if err != nil {
				klog.Error(err)
				return
			}

			stepChan <- &devops.NodesStepsIndex{Id: index, Steps: Steps}
			wg.Done()
		}(v.ID, i)
	}

	wg.Wait()
	close(stepChan)

	for oneNodeSteps := range stepChan {
		if oneNodeSteps != nil {
			nodesDetails[oneNodeSteps.Id].Steps = append(nodesDetails[oneNodeSteps.Id].Steps, oneNodeSteps.Steps...)
		}
	}

	return nodesDetails, err
}

func (d devopsOperator) GetPipelineBranch(projectName, pipelineName string, req *http.Request) (*devops.PipelineBranch, error) {

	res, err := d.devopsClient.GetPipelineBranch(projectName, pipelineName, convertToHttpParameters(req))
	//baseUrl+req.URL.RawQuery, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {

	resBody, err := d.devopsClient.ScanBranch(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetConsoleLog(projectName, pipelineName string, req *http.Request) ([]byte, error) {

	resBody, err := d.devopsClient.GetConsoleLog(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetCrumb(req *http.Request) (*devops.Crumb, error) {

	res, err := d.devopsClient.GetCrumb(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetSCMServers(scmId string, req *http.Request) ([]devops.SCMServer, error) {

	req.Method = http.MethodGet
	resBody, err := d.devopsClient.GetSCMServers(scmId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return resBody, err
}

func (d devopsOperator) GetSCMOrg(scmId string, req *http.Request) ([]devops.SCMOrg, error) {

	res, err := d.devopsClient.GetSCMOrg(scmId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetOrgRepo(scmId, organizationId string, req *http.Request) ([]devops.OrgRepo, error) {

	res, err := d.devopsClient.GetOrgRepo(scmId, organizationId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) CreateSCMServers(scmId string, req *http.Request) (*devops.SCMServer, error) {

	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	createReq := &devops.CreateScmServerReq{}
	err = json.Unmarshal(requestBody, createReq)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = nil
	servers, err := d.GetSCMServers(scmId, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, server := range servers {
		if server.ApiURL == createReq.ApiURL {
			return &server, nil
		}
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(requestBody))

	req.Method = http.MethodPost
	resBody, err := d.devopsClient.CreateSCMServers(scmId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return resBody, err
}

func (d devopsOperator) Validate(scmId string, req *http.Request) (*devops.Validates, error) {

	req.Method = http.MethodPut
	resBody, err := d.devopsClient.Validate(scmId, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) GetNotifyCommit(req *http.Request) ([]byte, error) {

	req.Method = http.MethodGet

	res, err := d.devopsClient.GetNotifyCommit(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GithubWebhook(req *http.Request) ([]byte, error) {

	res, err := d.devopsClient.GithubWebhook(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) CheckScriptCompile(projectName, pipelineName string, req *http.Request) (*devops.CheckScript, error) {

	resBody, err := d.devopsClient.CheckScriptCompile(projectName, pipelineName, convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func (d devopsOperator) CheckCron(projectName string, req *http.Request) (*devops.CheckCronRes, error) {

	res, err := d.devopsClient.CheckCron(projectName, convertToHttpParameters(req))

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ToJenkinsfile(req *http.Request) (*devops.ResJenkinsfile, error) {

	res, err := d.devopsClient.ToJenkinsfile(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ToJson(req *http.Request) (*devops.ResJson, error) {

	res, err := d.devopsClient.ToJson(convertToHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func getInputReqBody(reqBody io.ReadCloser) (newReqBody io.ReadCloser, err error) {
	var checkBody devops.CheckPlayload
	var jsonBody []byte
	var workRound struct {
		ID         string                           `json:"id,omitempty" description:"id"`
		Parameters []devops.CheckPlayloadParameters `json:"parameters"`
		Abort      bool                             `json:"abort,omitempty" description:"abort or not"`
	}

	Body, err := ioutil.ReadAll(reqBody)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	err = json.Unmarshal(Body, &checkBody)

	if checkBody.Abort != true && checkBody.Parameters == nil {
		workRound.Parameters = []devops.CheckPlayloadParameters{}
		workRound.ID = checkBody.ID
		jsonBody, _ = json.Marshal(workRound)
	} else {
		jsonBody, _ = json.Marshal(checkBody)
	}

	newReqBody = parseBody(bytes.NewBuffer(jsonBody))

	return newReqBody, nil

}

func parseBody(body io.Reader) (newReqBody io.ReadCloser) {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	return rc
}
