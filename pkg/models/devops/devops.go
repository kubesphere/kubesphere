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
	"flag"
	"fmt"
	log "github.com/golang/glog"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var JenkinsUrl string

func init() {
	flag.StringVar(&JenkinsUrl, "jenkins-url", "http://ks-jenkins.kubesphere-devops-system.svc.cluster.local:80", "jenkins server host")
}

func GetPipeline(projectName, pipelineName string, req *http.Request) (*Pipeline, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipelineUrl, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)
	var res = new(Pipeline)

	err := jenkinsClient(baseUrl, req, res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelines(req *http.Request) ([]interface{}, error) {
	baseUrl := JenkinsUrl + SearchPipelineUrl + req.URL.RawQuery
	log.Infof("Jenkins-url: " + baseUrl)
	var res []interface{}

	err := jenkinsClient(baseUrl, req, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelineRuns(projectName, pipelineName string, req *http.Request) ([]interface{}, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+SearchPipelineRunUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)
	var res []interface{}

	err := jenkinsClient(baseUrl, req, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) (*PipelineRun, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipelineRunUrl, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)
	var res = new(PipelineRun)

	err := jenkinsClient(baseUrl, req, res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodes(projectName, pipelineName, branchName, runId string, req *http.Request) ([]interface{}, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipelineRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)
	var res []interface{}

	err := jenkinsClient(baseUrl, req, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetStepLog(projectName, pipelineName, branchName, runId, stepId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetStepLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := Client(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func Validate(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+ValidateUrl, scmId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := Client(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetSCMOrg(scmId string, req *http.Request) ([]interface{}, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetSCMOrgUrl+req.URL.RawQuery, scmId)
	log.Infof("Jenkins-url: " + baseUrl)
	var res []interface{}

	err := jenkinsClient(baseUrl, req, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetOrgRepo(scmId, organizationId string, req *http.Request) (*OrgRepo, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetOrgRepoUrl+req.URL.RawQuery, scmId, organizationId)
	log.Infof("Jenkins-url: " + baseUrl)
	var res = new(OrgRepo)

	err := jenkinsClient(baseUrl, req, res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func StopPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*StopPipe, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+StopPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)
	var res = new(StopPipe)

	err := jenkinsClient(baseUrl, req, res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ReplayPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) (*ReplayPipe, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+ReplayPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)
	var res = new(ReplayPipe)

	err := jenkinsClient(baseUrl, req, res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetRunLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := Client(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) ([]interface{}, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)
	var res []interface{}

	err := jenkinsClient(baseUrl, req, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipeBranch(projectName, pipelineName string, req *http.Request) ([]interface{}, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipeBranchUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)
	var res []interface{}

	err := jenkinsClient(baseUrl, req, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func CheckPipeline(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+CheckPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := Client(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetConsoleLog(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetConsoleLogUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := Client(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+ScanBranchUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := Client(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func RunPipeline(projectName, pipelineName, branchName string, req *http.Request) (*QueuedBlueRun, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+RunPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName)
	log.Infof("Jenkins-url: " + baseUrl)
	var res = new(QueuedBlueRun)

	err := jenkinsClient(baseUrl, req, res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetStepsStatus(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) (*NodeStatus, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetStepsStatusUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId)
	log.Infof("Jenkins-url: " + baseUrl)
	var res = new(NodeStatus)

	err := jenkinsClient(baseUrl, req, res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetCrumb(req *http.Request) (*Crumb, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl + GetCrumbUrl)
	log.Infof("Jenkins-url: " + baseUrl)
	var res = new(Crumb)

	err := jenkinsClient(baseUrl, req, res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

// jenkins request and parse response
func jenkinsClient(baseUrl string, req *http.Request, res interface{}) error {
	resBody, err := Client(baseUrl, req)
	if err != nil {
		log.Error(err)
		return err
	}

	err = json.Unmarshal(resBody, res)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// create request
func Client(baseUrl string, req *http.Request) ([]byte, error) {
	newReqUrl, err := url.Parse(baseUrl)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	newRequest := &http.Request{
		Method: req.Method,
		URL:    newReqUrl,
		Header: req.Header,
		Body:   req.Body,
	}

	resp, err := client.Do(newRequest)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	resBody, _ := getRespBody(resp)
	log.Info(string(resBody))
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
