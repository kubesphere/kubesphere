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
	"kubesphere.io/kubesphere/pkg/models"

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

func GetPipeline(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetPipelineUrl, projectName, pipelineName)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func SearchPipelines(req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := devops.Jenkins().Server + SearchPipelineUrl + req.URL.RawQuery

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
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return 0, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	query, _ := parseJenkinsQuery(req.URL.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")

	baseUrl := devops.Jenkins().Server + SearchPipelineUrl + query.Encode()
	klog.V(4).Info("Jenkins-url: " + baseUrl)

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
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return 0, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	query, _ := parseJenkinsQuery(req.URL.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")
	baseUrl := fmt.Sprintf(devops.Jenkins().Server+SearchPipelineRunUrl, projectName, pipelineName)

	klog.V(4).Info("Jenkins-url: " + baseUrl)

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
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+SearchPipelineRunUrl, projectName, pipelineName)

	klog.V(4).Info("Jenkins-url: " + baseUrl)

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
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetPipeBranchRunUrl, projectName, pipelineName, branchName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetBranchPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)
	klog.V(4).Info("Jenkins-url: " + baseUrl)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetBranchStepLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)

	resBody, header, err := jenkinsClient(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	return resBody, header, err
}

func GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, http.Header, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetStepLogUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId, stepId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetSCMServersUrl, scmId)
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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+CreateSCMServersUrl, scmId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+ValidateUrl, scmId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetSCMOrgUrl+req.URL.RawQuery, scmId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetOrgRepoUrl+req.URL.RawQuery, scmId, organizationId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+StopBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)

	req.Method = http.MethodPut
	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func StopPipeline(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+StopPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+ReplayBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func ReplayPipeline(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+ReplayPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetBranchRunLogUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetRunLog(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetRunLogUrl+req.URL.RawQuery, projectName, pipelineName, runId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetBranchArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetArtifacts(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetArtifactsUrl+req.URL.RawQuery, projectName, pipelineName, runId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetPipeBranchUrl, projectName, pipelineName)

	res, err := sendJenkinsRequest(baseUrl+req.URL.RawQuery, req)
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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+CheckBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId, stepId)

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

func SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+CheckPipelineUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId, stepId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetConsoleLogUrl+req.URL.RawQuery, projectName, pipelineName)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+ScanBranchUrl+req.URL.RawQuery, projectName, pipelineName)

	resBody, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return resBody, err
}

func RunBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+RunBranchPipelineUrl+req.URL.RawQuery, projectName, pipelineName, branchName)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func RunPipeline(projectName, pipelineName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	baseUrl := fmt.Sprintf(devops.Jenkins().Server+RunPipelineUrl+req.URL.RawQuery, projectName, pipelineName)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetCrumb(req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server + GetCrumbUrl)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+CheckScriptCompileUrl, projectName, pipelineName)

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

	jenkins := devops.Jenkins()

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

func GetPipelineRun(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetPipelineRunUrl, projectName, pipelineName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchPipeline(projectName, pipelineName, branchName string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetBranchPipeUrl, projectName, pipelineName, branchName)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetPipelineRunNodes(projectName, pipelineName, runId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetPipeRunNodesUrl+req.URL.RawQuery, projectName, pipelineName, runId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetBranchNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, branchName, runId, nodeId)

	res, err := sendJenkinsRequest(baseUrl, req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return res, err
}

func GetNodeSteps(projectName, pipelineName, runId, nodeId string, req *http.Request) ([]byte, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	baseUrl := fmt.Sprintf(devops.Jenkins().Server+GetNodeStepsUrl+req.URL.RawQuery, projectName, pipelineName, runId, nodeId)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server + ToJenkinsfileUrl)

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

	baseUrl := fmt.Sprintf(devops.Jenkins().Server + ToJsonUrl)

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

	baseUrl := fmt.Sprint(devops.Jenkins().Server, GetNotifyCommitUrl, req.URL.RawQuery)
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

	baseUrl := fmt.Sprint(devops.Jenkins().Server, GithubWebhookUrl, req.URL.RawQuery)

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

func GetNodesDetail(projectName, pipelineName, runId string, req *http.Request) ([]NodesDetail, error) {
	var wg sync.WaitGroup
	var nodesDetails []NodesDetail
	stepChan := make(chan *NodesStepsIndex, channelMaxCapacity)

	respNodes, err := GetPipelineRunNodes(projectName, pipelineName, runId, req)
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
			respSteps, err := GetNodeSteps(projectName, pipelineName, runId, nodeId, req)
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
