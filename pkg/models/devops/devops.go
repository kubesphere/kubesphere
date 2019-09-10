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
	log "github.com/golang/glog"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/simple/client/admin_jenkins"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const channelMaxCapacity = 100

var jenkins *gojenkins.Jenkins

func JenkinsInit() {
	jenkins = admin_jenkins.GetJenkins()
}

func GetPipeline(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetPipelineUrl, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelines(req *http.Request) ([]byte, error) {
	baseUrl := jenkins.Server + SearchPipelineUrl + req.URL.RawQuery
	klog.V(3).Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	count, err := searchPipelineCount(req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	responseStruct := models.PageableResponse{TotalCount: count}
	err = json.Unmarshal(res, &responseStruct.Items)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	res, err = json.Marshal(responseStruct)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return res, err
}

func searchPipelineCount(req *http.Request) (int, error) {
	query, _ := parseJenkinsQuery(req.URL.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")

	baseUrl := jenkins.Server + SearchPipelineUrl + query.Encode()
	klog.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	var pipelines []Pipeline
	err = json.Unmarshal(res, &pipelines)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	return len(pipelines), nil
}

func searchPipelineRunsCount(projectName, pipelineName string, req *http.Request) (int, error) {
	query, _ := parseJenkinsQuery(req.URL.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")
	baseUrl := fmt.Sprintf(jenkins.Server+SearchPipelineRunUrl, projectName, pipelineName)

	klog.V(3).Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl+query.Encode(), req)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	var runs []PipelineRun
	err = json.Unmarshal(res, &runs)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	return len(runs), nil
}

func SearchPipelineRuns(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+SearchPipelineRunUrl, projectName, pipelineName)

	klog.V(3).Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl+req.URL.RawQuery, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	count, err := searchPipelineRunsCount(projectName, pipelineName, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	responseStruct := models.PageableResponse{TotalCount: count}
	err = json.Unmarshal(res, &responseStruct.Items)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	res, err = json.Marshal(responseStruct)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return res, err
}

func GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetPipeBranchRunUrl, projectName, pipelineName, branchName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodesbyBranch(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchStepLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Info("Jenkins-url: " + baseUrl)

	resBody, header, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}

	return resBody, header, err
}

func GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetStepLogUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId, stepId)
	log.Info("Jenkins-url: " + baseUrl)

	resBody, header, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}

	return resBody, header, err

}

func GetSCMServers(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetSCMServersUrl, scmId)
	log.Info("Jenkins-url: " + baseUrl)
	req.Method = http.MethodGet
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return resBody, err
}

func CreateSCMServers(scmId string, req *http.Request) ([]byte, error) {
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	createReq := &CreateScmServerReq{}
	err = json.Unmarshal(requestBody, createReq)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	req.Body = nil
	byteServers, err := GetSCMServers(scmId, req)
	if err != nil {
		log.Error(err)
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
	baseUrl := fmt.Sprintf(jenkins.Server+CreateSCMServersUrl, scmId)
	log.Info("Jenkins-url: " + baseUrl)
	req.Method = http.MethodPost
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return resBody, err
}

func Validate(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+ValidateUrl, scmId)
	log.Info("Jenkins-url: " + baseUrl)

	req.Method = http.MethodPut
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetSCMOrg(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetSCMOrgUrl+req.URL.RawQuery, scmId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetOrgRepo(scmId, organizationId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetOrgRepoUrl+req.URL.RawQuery, scmId, organizationId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func StopBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+StopBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	req.Method = http.MethodPut
	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func StopPipeline(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+StopPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	req.Method = http.MethodPut
	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+ReplayBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ReplayPipeline(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+ReplayPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchRunLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetRunLog(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetRunLogUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetArtifacts(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipeBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetPipeBranchUrl, projectName, pipelineName)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl+req.URL.RawQuery, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+CheckBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Info("Jenkins-url: " + baseUrl)

	newBody, err := getInputReqBody(req.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	req.Body = newBody
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+CheckPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId, stepId)
	log.Info("Jenkins-url: " + baseUrl)

	newBody, err := getInputReqBody(req.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	req.Body = newBody
	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
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
	baseUrl := fmt.Sprintf(jenkins.Server+GetConsoleLogUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Info("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+ScanBranchUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Info("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func RunBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+RunBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func RunPipeline(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+RunPipelineUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetCrumb(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server + GetCrumbUrl)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func CheckScriptCompile(req *http.Request) ([]byte, error) {
	baseUrl := jenkins.Server + CheckScriptCompileUrl
	log.Info("Jenkins-url: " + baseUrl)
	req.SetBasicAuth(jenkins.Requester.BasicAuth.Username, jenkins.Requester.BasicAuth.Password)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func CheckCron(req *http.Request) (*CheckCronRes, error) {
	newurl, err := url.Parse(jenkins.Server + CheckCronUrl + req.URL.RawQuery)

	reqJenkins := &http.Request{
		Method: http.MethodGet,
		URL:    newurl,
		Header: http.Header{},
	}
	var res = new(CheckCronRes)
	client := &http.Client{Timeout: 30 * time.Second}

	reqJenkins.SetBasicAuth(jenkins.Requester.BasicAuth.Username, jenkins.Requester.BasicAuth.Password)

	resp, err := client.Do(reqJenkins)
	if err != nil {
		log.Error(err)
		return res, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Error(err)
		return res, err
	}
	doc.Find("div").Each(func(i int, selection *goquery.Selection) {
		res.Message = selection.Text()
		res.Result, _ = selection.Attr("class")
	})
	return res, err
}

func GetPipelineRun(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetPipelineRunUrl, projectName, pipelineName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchPipeUrl, projectName, pipelineName, branchName)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetNodeSteps(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ToJenkinsfile(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server + ToJenkinsfileUrl)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ToJson(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server + ToJsonUrl)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetNotifyCommit(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprint(jenkins.Server, GetNotifyCommitUrl, req.URL.RawQuery)
	log.Info("Jenkins-url: " + baseUrl)
	req.Method = "GET"

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GithubWebhook(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprint(jenkins.Server, GithubWebhookUrl, req.URL.RawQuery)
	log.Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchNodesDetail(projectName, pipelineName, branchName, runId string, req *http.Request) ([]NodesDetail, error) {
	getNodesUrl := fmt.Sprintf(jenkins.Server+GetBranchPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Info("getNodesUrl: " + getNodesUrl)
	var wg sync.WaitGroup
	var nodesDetails []NodesDetail
	stepChan := make(chan *NodesStepsIndex, channelMaxCapacity)

	respNodes, err := GetPipelineRunNodesbyBranch(projectName, pipelineName, branchName, runId, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = json.Unmarshal(respNodes, &nodesDetails)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// get all steps in nodes.
	for i, v := range nodesDetails {
		wg.Add(1)
		go func(nodeId string, index int) {
			var steps []NodeSteps
			respSteps, err := GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId, req)
			if err != nil {
				log.Error(err)
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

func GetNodesDetail(projectName, pipelineName, runId string, req *http.Request) ([]NodesDetail, error) {
	getNodesUrl := fmt.Sprintf(jenkins.Server+GetPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Info("getNodesUrl: " + getNodesUrl)
	var wg sync.WaitGroup
	var nodesDetails []NodesDetail
	stepChan := make(chan *NodesStepsIndex, channelMaxCapacity)

	respNodes, err := GetPipelineRunNodes(projectName, pipelineName, runId, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = json.Unmarshal(respNodes, &nodesDetails)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// get all steps in nodes.
	for i, v := range nodesDetails {
		wg.Add(1)
		go func(nodeId string, index int) {
			var steps []NodeSteps
			respSteps, err := GetNodeSteps(projectName, pipelineName, runId, nodeId, req)
			if err != nil {
				log.Error(err)
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

// create jenkins request
func sendJenkinsRequest(baseUrl string, req *http.Request) ([]byte, error) {
	resBody, _, err := jenkinsClient(baseUrl, req)
	return resBody, err
}

func jenkinsClient(baseUrl string, req *http.Request) ([]byte, http.Header, error) {
	newReqUrl, err := url.Parse(baseUrl)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
		return nil, nil, err
	}

	resBody, _ := getRespBody(resp)
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		log.Errorf("%+v", string(resBody))
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
		log.Error(err)
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
