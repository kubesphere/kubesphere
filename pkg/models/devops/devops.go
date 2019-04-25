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

func GetPipeline(projectName, pipelineName string, req *http.Request) (*TypePipeline, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipelineUrl, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var res = new(TypePipeline)
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelines(req *http.Request) ([]interface{}, error) {
	baseUrl := JenkinsUrl + SearchPipelineUrl + req.URL.RawQuery
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var res []interface{}
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelineRuns(projectName, pipelineName string, req *http.Request) ([]interface{}, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+SearchPipelineRunUrl+req.URL.RawQuery, projectName, pipelineName)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var res []interface{}
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRun(projectName, pipelineName, branchName, runId string, req *http.Request) (*RunPipeline, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipelineRunUrl, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var res = new(RunPipeline)
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodes(projectName, pipelineName, branchName, runId string, req *http.Request) ([]interface{}, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetPipelineRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var res []interface{}
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetStepLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func Validate(scmId string, req *http.Request) ([]byte, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+ValidateUrl, scmId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return resBody, err
}

func GetSCMOrg(scmId string, req *http.Request) ([]interface{}, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetSCMOrgUrl+req.URL.RawQuery, scmId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var res []interface{}
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

func GetSCMOrgRepo(scmId, organizationId string, req *http.Request) (*TypeSCMOrgRepo, error) {
	baseUrl := fmt.Sprintf(JenkinsUrl+GetSCMOrgRepoUrl+req.URL.RawQuery, scmId, organizationId)
	log.Infof("Jenkins-url: " + baseUrl)

	resBody, err := jenkinsClient(baseUrl, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var res = new(TypeSCMOrgRepo)
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, err
}

// create jenkins request
func jenkinsClient(baseUrl string, req *http.Request) ([]byte, error) {
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
		err = json.Unmarshal(resBody, jkerr)
		if err != nil {
			log.Error(err)
			return nil, err
		}
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
