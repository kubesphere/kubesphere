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
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
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

func GetPipeline(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipelineUrl, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelines(req *http.Request) ([]byte, error) {
	baseUrl := JenkinsUrl + SearchPipelineUrl + req.URL.RawQuery
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelineRuns(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+SearchPipelineRunUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipeBranchRunUrl, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetBranchPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetStepLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func Validate(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+ValidateUrl, scmId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetSCMOrg(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetSCMOrgUrl+req.URL.RawQuery, scmId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetOrgRepo(scmId, organizationId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetOrgRepoUrl+req.URL.RawQuery, scmId, organizationId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func StopPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+StopPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ReplayPipeline(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+ReplayPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetRunLog(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetRunLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetArtifacts(projectName, pipelineName, branchName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipeBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipeBranchUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func CheckPipeline(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+CheckPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetConsoleLog(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetConsoleLogUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func ScanBranch(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+ScanBranchUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func RunPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+RunPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetStepsStatus(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetStepsStatusUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetCrumb(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl + GetCrumbUrl)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func CheckScriptCompile(req *http.Request) ([]byte, error) {
	baseUrl := JenkinsUrl + CheckScriptCompileUrl
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func CheckCron(req *http.Request) (*CheckCronRes, error) {
	baseUrl := JenkinsUrl + CheckCronUrl + req.URL.RawQuery
	log.Infof("Jenkins-url: " + baseUrl)
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
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipelineRunUrl, projectName, pipelineName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetBranchPipeUrl, projectName, pipelineName, branchName)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ToJenkinsfile(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl + ToJenkinsfileUrl)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func ToJson(req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl + ToJsonUrl)
	log.Infof("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
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
