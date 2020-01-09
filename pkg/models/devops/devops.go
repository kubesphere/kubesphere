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
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/emicklei/go-restful"
	"io"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"

	"k8s.io/klog"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	channelMaxCapacity = 100
	cronJobLayout      = "Monday, January 2, 2006 15:04:05 PM"
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
}

type devopsOperator struct {
	devopsClient devops.Interface
}

func NewDevopsOperator(client jenkins.Client) DevopsOperator {
	return &devopsOperator{}
}

func convertorHttpParameters(req *http.Request) (*devops.HttpParameters) {
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
	//formatUrl := fmt.Sprintf(GetPipelineUrl, projectName, pipelineName)

	res, err := d.devopsClient.GetPipeline(projectName, pipelineName, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) ListPipelines(req *http.Request) (*devops.PipelineList, error) {

	//formatUrl := SearchPipelineUrl + req.URL.RawQuery

	res, err := d.devopsClient.ListPipelines(convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) GetPipelineRun(projectName, pipelineName, runId string, req *http.Request) (*devops.PipelineRun, error) {

	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetPipelineRunUrl, projectName, pipelineName, runId)

	res, err := d.devopsClient.GetPipelineRun(projectName, pipelineName, runId, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) ListPipelineRuns(projectName, pipelineName string, req *http.Request) (*devops.PipelineRunList, error) {

	//formatUrl := fmt.Sprintf(SearchPipelineRunUrl, projectName, pipelineName)

	res, err := d.devopsClient.ListPipelineRuns(projectName, pipelineName, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (d devopsOperator) StopPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.StopPipeline, error) {

	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+StopPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId)

	req.Method = http.MethodPut
	res, err := d.devopsClient.StopPipeline(projectName, pipelineName, runId, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) ReplayPipeline(projectName, pipelineName, runId string, req *http.Request) (*devops.ReplayPipeline, error) {

	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+ReplayPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId)

	res, err := d.devopsClient.ReplayPipeline(projectName, pipelineName, runId, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) RunPipeline(projectName, pipelineName string, req *http.Request) (*devops.RunPipeline, error) {

	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+RunPipelineUrl+req.URL.RawQuery, projectName, pipelineName)

	res, err := d.devopsClient.RunPipeline(projectName, pipelineName, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetArtifacts(projectName, pipelineName, runId string, req *http.Request) ([]devops.Artifacts, error) {
	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, runId)

	res, err := d.devopsClient.GetArtifacts(projectName, pipelineName, runId, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetRunLog(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {

	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetRunLogUrl+req.URL.RawQuery, projectName, pipelineName, runId)

	res, err := d.devopsClient.GetRunLog(projectName, pipelineName, runId, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {

	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetStepLogUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId, stepId)

	resBody, header, err := d.devopsClient.GetStepLog(projectName, pipelineName, runId, nodeId, stepId, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	return resBody, header, err
}

func (d devopsOperator) GetNodeSteps(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]devops.NodeSteps, error) {

	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId)

	res, err := d.devopsClient.GetNodeSteps(projectName, pipelineName, runId, nodeId, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func (d devopsOperator) GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]devops.PipelineRunNodes, error) {
	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, runId)

	res, err := d.devopsClient.GetPipelineRunNodes(projectName, pipelineName, runId, convertorHttpParameters(req))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	fmt.Println()

	return res, err
}

func (d devopsOperator) SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	//baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+CheckPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId, stepId)

	newBody, err := getInputReqBody(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = newBody

	resBody, err := d.devopsClient.SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId, convertorHttpParameters(req))
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

			stepChan <- &devops.NodesStepsIndex{Id:index, Steps:Steps}
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










func GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetBranchPipeUrl, projectName, pipelineName, branchName)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetPipeBranchRunUrl, projectName, pipelineName, branchName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func RunBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+RunBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func StopBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+StopBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)

	req.Method = http.MethodPut
	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+ReplayBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetBranchArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipeBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetPipeBranchUrl, projectName, pipelineName)

	res, err := sendJenkinsRequest(baseUrl+req.URL.RawQuery, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodesbyBranch(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetBranchPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	klog.V(4).Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchNodesDetail(projectName, pipelineName, branchName, runId string, req *http.Request) ([]NodesDetail, error) {
	var wg sync.WaitGroup
	var nodesDetails []NodesDetail
	stepChan := make(chan *NodesStepsIndex, channelMaxCapacity)

	respNodes, err := GetPipelineRunNodesbyBranch(projectName, pipelineName, branchName, runId, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	err = json.Unmarshal(respNodes, &nodesDetails)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// get all steps in nodes.
	for i, v := range nodesDetails {
		wg.Add(1)
		go func(nodeId string, index int) {
			var steps []NodeSteps
			respSteps, err := GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, req)
			if err != nil {
				klog.Error(err)
				return
			}
			err = json.Unmarshal(respSteps, &steps)

			stepChan <- &NodesStepsIndex{index, steps}
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

func GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetBranchStepLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)

	resBody, header, err := jenkinsClient(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	return resBody, header, err
}

func GetSCMServers(scmId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetSCMServersUrl, scmId)
	req.Method = http.MethodGet
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return resBody, err
}

func CreateSCMServers(scmId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	createReq := &CreateScmServerReq{}
	err = json.Unmarshal(requestBody, createReq)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = nil
	byteServers, err := GetSCMServers(scmId, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var servers []*SCMServer
	_ = json.Unmarshal(byteServers, &servers)
	for _, server := range servers {
		if server.ApiURL == createReq.ApiURL {
			return json.Marshal(server)
		}
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(requestBody))

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+CreateSCMServersUrl, scmId)

	req.Method = http.MethodPost
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return resBody, err
}

func Validate(scmId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+ValidateUrl, scmId)

	req.Method = http.MethodPut
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetSCMOrg(scmId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetSCMOrgUrl+req.URL.RawQuery, scmId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetOrgRepo(scmId, organizationId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetOrgRepoUrl+req.URL.RawQuery, scmId, organizationId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetBranchRunLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+CheckBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)

	newBody, err := getInputReqBody(req.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Body = newBody
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func getInputReqBody(reqBody io.ReadCloser) (newReqBody io.ReadCloser, err error) {
	var checkBody CheckPlayload
	var jsonBody []byte
	var workRound struct {
		ID         string                    `json:"id,omitempty" description:"id"`
		Parameters []CheckPlayloadParameters `json:"parameters"`
		Abort      bool                      `json:"abort,omitempty" description:"abort or not"`
	}

	Body, err := ioutil.ReadAll(reqBody)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	err = json.Unmarshal(Body, &checkBody)

	if checkBody.Abort != true && checkBody.Parameters == nil {
		workRound.Parameters = []CheckPlayloadParameters{}
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

func GetConsoleLog(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetConsoleLogUrl+req.URL.RawQuery, projectName, pipelineName)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+ScanBranchUrl+req.URL.RawQuery, projectName, pipelineName)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetCrumb(req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server + GetCrumbUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func CheckScriptCompile(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+CheckScriptCompileUrl, projectName, pipelineName)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func CheckCron(projectName string, req *http.Request) (*CheckCronRes, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	jenkins := jenkins.Jenkins()

	var res = new(CheckCronRes)
	var cron = new(CronData)
	var reader io.ReadCloser
	var baseUrl string

	reader = req.Body
	cronData, err := ioutil.ReadAll(reader)
	json.Unmarshal(cronData, cron)

	if cron.PipelineName != "" {
		baseUrl = fmt.Sprintf(jenkins.Server+CheckPipelienCronUrl, projectName, cron.PipelineName, cron.Cron)
	} else {
		baseUrl = fmt.Sprintf(jenkins.Server+CheckCronUrl, projectName, cron.Cron)
	}

	newUrl, err := url.Parse(baseUrl)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	newUrl.RawQuery = newUrl.Query().Encode()

	reqJenkins := &http.Request{
		Method: http.MethodGet,
		URL:    newUrl,
		Header: req.Header,
	}

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(reqJenkins)

	if resp != nil && resp.StatusCode != http.StatusOK {
		resBody, _ := getRespBody(resp)
		return &CheckCronRes{
			Result:  "error",
			Message: string(resBody),
		}, err
	}
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	doc.Find("div").Each(func(i int, selection *goquery.Selection) {
		res.Message = selection.Text()
		res.Result, _ = selection.Attr("class")
	})
	if res.Result == "ok" {
		res.LastTime, res.NextTime, err = parseCronJobTime(res.Message)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	return res, err
}

func parseCronJobTime(msg string) (string, string, error) {

	times := strings.Split(msg, ";")

	lastTmp := strings.Split(times[0], " ")
	lastCount := len(lastTmp)
	lastTmp = lastTmp[lastCount-7 : lastCount-1]
	lastTime := strings.Join(lastTmp, " ")
	lastUinx, err := time.Parse(cronJobLayout, lastTime)
	if err != nil {
		klog.Error(err)
		return "", "", err
	}
	last := lastUinx.Format(time.RFC3339)

	nextTmp := strings.Split(times[1], " ")
	nextCount := len(nextTmp)
	nextTmp = nextTmp[nextCount-7 : nextCount-1]
	nextTime := strings.Join(nextTmp, " ")
	nextUinx, err := time.Parse(cronJobLayout, nextTime)
	if err != nil {
		klog.Error(err)
		return "", "", err
	}
	next := nextUinx.Format(time.RFC3339)

	return last, next, nil
}

func GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server+GetBranchNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func ToJenkinsfile(req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server + ToJenkinsfileUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func ToJson(req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(jenkins.Jenkins().Server + ToJsonUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetNotifyCommit(req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprint(jenkins.Jenkins().Server, GetNotifyCommitUrl, req.URL.RawQuery)
	req.Method = "GET"

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GithubWebhook(req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprint(jenkins.Jenkins().Server, GithubWebhookUrl, req.URL.RawQuery)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

// create jenkins request
func sendJenkinsRequest(baseUrl string, req *http.Request) ([]byte, error) {
	resBody, _, err := jenkinsClient(baseUrl, req)
	return resBody, err
}

func jenkinsClient(baseUrl string, req *http.Request) ([]byte, http.Header, error) {
	newReqUrl, err := url.Parse(baseUrl)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	newRequest := &http.Request{
		Method:   req.Method,
		URL:      newReqUrl,
		Header:   req.Header,
		Body:     req.Body,
		Form:     req.Form,
		PostForm: req.PostForm,
	}

	resp, err := client.Do(newRequest)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	resBody, _ := getRespBody(resp)
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		klog.Errorf("%+v", string(resBody))
		jkerr := new(JkError)
		jkerr.Code = resp.StatusCode
		jkerr.Message = string(resBody)
		return nil, nil, jkerr
	}

	return resBody, resp.Header, nil

}

// Decompress response.body of JenkinsAPIResponse
func getRespBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(resp.Body)
	} else {
		reader = resp.Body
	}
	resBody, err := ioutil.ReadAll(reader)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return resBody, err

}

// parseJenkinsQuery Parse the special query of jenkins.
// ParseQuery in the standard library makes the query not re-encode
func parseJenkinsQuery(query string) (url.Values, error) {
	m := make(url.Values)
	err := error(nil)
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		m[key] = append(m[key], value)
	}
	return m, err
}
