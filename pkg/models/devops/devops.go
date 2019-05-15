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
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/golang/glog"
	"io"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/admin_jenkins"
	"net/http"
	"net/url"
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
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelineRuns(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+SearchPipelineRunUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetPipeBranchRunUrl, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodesbyBranch(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchStepLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetStepLogUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func Validate(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+ValidateUrl, scmId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetSCMOrg(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetSCMOrgUrl+req.URL.RawQuery, scmId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetOrgRepo(scmId, organizationId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetOrgRepoUrl+req.URL.RawQuery, scmId, organizationId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func StopBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+StopBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func StopPipeline(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+StopPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+ReplayBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ReplayPipeline(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+ReplayPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchRunLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetRunLog(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetRunLogUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetArtifacts(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipeBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetPipeBranchUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func CheckBranchPipeline(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+CheckBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func CheckPipeline(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+CheckPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetConsoleLog(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetConsoleLogUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+ScanBranchUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func RunBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+RunBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func RunPipeline(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+RunPipelineUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchStepsStatus(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchStepsStatusUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetStepsStatus(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetStepsStatusUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetCrumb(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server + GetCrumbUrl)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func CheckScriptCompile(req *http.Request) ([]byte, error) {
	baseUrl := jenkins.Server + CheckScriptCompileUrl
	log.Infof("Jenkins-url: " + baseUrl)
	req.SetBasicAuth(jenkins.Requester.BasicAuth.Username, jenkins.Requester.BasicAuth.Password)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func CheckCron(req *http.Request) (*CheckCronRes, error) {
	baseUrl := jenkins.Server + CheckCronUrl + req.URL.RawQuery
	log.Infof("Jenkins-url: " + baseUrl)
	req.SetBasicAuth(jenkins.Requester.BasicAuth.Username, jenkins.Requester.BasicAuth.Password)
	var res = new(CheckCronRes)

	resp, err := http.Get(baseUrl)
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
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchPipeUrl, projectName, pipelineName, branchName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetBranchNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetNodeSteps(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server+GetNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ToJenkinsfile(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server + ToJenkinsfileUrl)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ToJson(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server + ToJsonUrl)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetNotifyCommit(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server + GetNotifyCommitUrl + req.URL.RawQuery)
	log.Infof("Jenkins-url: " + baseUrl)
	req.Method = "GET"

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GithubWebhook(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(jenkins.Server + GithubWebhookUrl + req.URL.RawQuery)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchNodesDetail(projectName, pipelineName, branchName, runId string, req *http.Request) ([]NodesDetail, error) {
	getNodesUrl := fmt.Sprintf(jenkins.Server+GetBranchPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("getNodesUrl: " + getNodesUrl)
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
	log.Infof("getNodesUrl: " + getNodesUrl)
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
	newReqUrl, err := url.Parse(baseUrl)
	if err != nil {
		log.Error(err)
		return nil, err
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
		return nil, err
	}
	defer resp.Body.Close()

	resBody, _ := getRespBody(resp)
	if resp.StatusCode >= http.StatusBadRequest {
		jkerr := new(JkError)
		jkerr.Code = resp.StatusCode
		jkerr.Message = http.StatusText(resp.StatusCode)
		return nil, jkerr
	}

	return resBody, err
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
