/*
Copyright 2020 KubeSphere Authors

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

package jenkins

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Pipeline struct {
	HttpParameters *devops.HttpParameters
	Jenkins        *Jenkins
	Path           string
}

const (
	GetPipelineUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/"
	ListPipelinesUrl       = "/blue/rest/search/?"
	GetPipelineRunUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/"
	ListPipelineRunUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/?"
	StopPipelineUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/stop/?"
	ReplayPipelineUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/replay/"
	RunPipelineUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/"
	GetArtifactsUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/artifacts/?"
	GetRunLogUrl           = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/log/?"
	GetStepLogUrl          = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/%s/log/?"
	GetPipelineRunNodesUrl = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/?"
	SubmitInputStepUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/%s/"
	GetNodeStepsUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/nodes/%s/steps/?"

	GetBranchPipelineUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/"
	GetBranchPipelineRunUrl  = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/"
	StopBranchPipelineUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/stop/?"
	ReplayBranchPipelineUrl  = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/replay/"
	RunBranchPipelineUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/"
	GetBranchArtifactsUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/artifacts/?"
	GetBranchRunLogUrl       = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/log/?"
	GetBranchStepLogUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/log/?"
	GetBranchNodeStepsUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/?"
	GetBranchPipeRunNodesUrl = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/?"
	CheckBranchPipelineUrl   = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/"
	GetPipeBranchUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/?"
	ScanBranchUrl            = "/job/%s/job/%s/build?"
	GetConsoleLogUrl         = "/job/%s/job/%s/indexing/consoleText"
	GetCrumbUrl              = "/crumbIssuer/api/json/"
	GetSCMServersUrl         = "/blue/rest/organizations/jenkins/scm/%s/servers/"
	GetSCMOrgUrl             = "/blue/rest/organizations/jenkins/scm/%s/organizations/?"
	GetOrgRepoUrl            = "/blue/rest/organizations/jenkins/scm/%s/organizations/%s/repositories/?"
	CreateSCMServersUrl      = "/blue/rest/organizations/jenkins/scm/%s/servers/"
	ValidateUrl              = "/blue/rest/organizations/jenkins/scm/%s/validate"

	GetNotifyCommitUrl    = "/git/notifyCommit/?"
	GithubWebhookUrl      = "/github-webhook/"
	CheckScriptCompileUrl = "/job/%s/job/%s/descriptorByName/org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition/checkScriptCompile"

	CheckPipelienCronUrl = "/job/%s/job/%s/descriptorByName/hudson.triggers.TimerTrigger/checkSpec?value=%s"
	CheckCronUrl         = "/job/%s/descriptorByName/hudson.triggers.TimerTrigger/checkSpec?value=%s"
	ToJenkinsfileUrl     = "/pipeline-model-converter/toJenkinsfile"
	ToJsonUrl            = "/pipeline-model-converter/toJson"

	cronJobLayout = "Monday, January 2, 2006 15:04:05 PM"
)

func (p *Pipeline) GetPipeline() (*devops.Pipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var pipeline devops.Pipeline

	err = json.Unmarshal(res, &pipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &pipeline, err
}

func (p *Pipeline) ListPipelines() (*devops.PipelineList, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	count, err := p.searchPipelineCount()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	pipelienList := devops.PipelineList{Total: count}
	err = json.Unmarshal(res, &pipelienList.Items)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &pipelienList, err
}

func (p *Pipeline) searchPipelineCount() (int, error) {
	query, _ := parseJenkinsQuery(p.HttpParameters.Url.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")

	formatUrl := ListPipelinesUrl + query.Encode()

	res, err := p.Jenkins.SendPureRequest(formatUrl, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	var pipelines []devops.Pipeline
	err = json.Unmarshal(res, &pipelines)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	return len(pipelines), nil
}

func (p *Pipeline) GetPipelineRun() (*devops.PipelineRun, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var pipelineRun devops.PipelineRun
	err = json.Unmarshal(res, &pipelineRun)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &pipelineRun, err
}

func (p *Pipeline) ListPipelineRuns() (*devops.PipelineRunList, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var pipelineRunList devops.PipelineRunList
	err = json.Unmarshal(res, &pipelineRunList.Items)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	total, err := p.searchPipelineRunsCount()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	pipelineRunList.Total = total
	return &pipelineRunList, err
}

func (p *Pipeline) searchPipelineRunsCount() (int, error) {
	query, _ := parseJenkinsQuery(p.HttpParameters.Url.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")
	p.HttpParameters.Url.RawQuery = query.Encode()
	u, err := url.Parse(p.Path)
	if err != nil {
		return 0, err
	}
	res, err := p.Jenkins.SendPureRequest(u.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	var runs []devops.PipelineRun
	err = json.Unmarshal(res, &runs)
	if err != nil {
		klog.Error(err)
		return 0, err
	}
	return len(runs), nil
}

func (p *Pipeline) StopPipeline() (*devops.StopPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var stopPipeline devops.StopPipeline
	err = json.Unmarshal(res, &stopPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &stopPipeline, err
}

func (p *Pipeline) ReplayPipeline() (*devops.ReplayPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var replayPipeline devops.ReplayPipeline
	err = json.Unmarshal(res, &replayPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &replayPipeline, err
}

func (p *Pipeline) RunPipeline() (*devops.RunPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var runPipeline devops.RunPipeline
	err = json.Unmarshal(res, &runPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &runPipeline, err
}

func (p *Pipeline) GetArtifacts() ([]devops.Artifacts, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var artifacts []devops.Artifacts
	err = json.Unmarshal(res, &artifacts)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return artifacts, err
}

func (p *Pipeline) GetRunLog() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GetStepLog() ([]byte, http.Header, error) {
	res, header, err := p.Jenkins.SendPureRequestWithHeaderResp(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, header, err
}

func (p *Pipeline) GetNodeSteps() ([]devops.NodeSteps, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	var nodeSteps []devops.NodeSteps
	err = json.Unmarshal(res, &nodeSteps)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return nodeSteps, err
}

func (p *Pipeline) GetPipelineRunNodes() ([]devops.PipelineRunNodes, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var pipelineRunNodes []devops.PipelineRunNodes
	err = json.Unmarshal(res, &pipelineRunNodes)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return pipelineRunNodes, err
}

func (p *Pipeline) SubmitInputStep() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GetBranchPipeline() (*devops.BranchPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var branchPipeline devops.BranchPipeline
	err = json.Unmarshal(res, &branchPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchPipeline, err
}

func (p *Pipeline) GetBranchPipelineRun() (*devops.PipelineRun, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var branchPipelineRun devops.PipelineRun
	err = json.Unmarshal(res, &branchPipelineRun)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchPipelineRun, err
}

func (p *Pipeline) StopBranchPipeline() (*devops.StopPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var branchStopPipeline devops.StopPipeline
	err = json.Unmarshal(res, &branchStopPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchStopPipeline, err
}

func (p *Pipeline) ReplayBranchPipeline() (*devops.ReplayPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var branchReplayPipeline devops.ReplayPipeline
	err = json.Unmarshal(res, &branchReplayPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchReplayPipeline, err
}

func (p *Pipeline) RunBranchPipeline() (*devops.RunPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var branchRunPipeline devops.RunPipeline
	err = json.Unmarshal(res, &branchRunPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchRunPipeline, err
}

func (p *Pipeline) GetBranchArtifacts() ([]devops.Artifacts, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var artifacts []devops.Artifacts
	err = json.Unmarshal(res, &artifacts)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return artifacts, err
}

func (p *Pipeline) GetBranchRunLog() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GetBranchStepLog() ([]byte, http.Header, error) {
	res, header, err := p.Jenkins.SendPureRequestWithHeaderResp(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, header, err
}

func (p *Pipeline) GetBranchNodeSteps() ([]devops.NodeSteps, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var branchNodeSteps []devops.NodeSteps
	err = json.Unmarshal(res, &branchNodeSteps)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return branchNodeSteps, err
}

func (p *Pipeline) GetBranchPipelineRunNodes() ([]devops.BranchPipelineRunNodes, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var branchPipelineRunNodes []devops.BranchPipelineRunNodes
	err = json.Unmarshal(res, &branchPipelineRunNodes)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return branchPipelineRunNodes, err
}

func (p *Pipeline) SubmitBranchInputStep() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GetPipelineBranch() (*devops.PipelineBranch, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var pipelineBranch devops.PipelineBranch
	err = json.Unmarshal(res, &pipelineBranch)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &pipelineBranch, err
}

func (p *Pipeline) ScanBranch() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GetConsoleLog() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GetCrumb() (*devops.Crumb, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var crumb devops.Crumb
	err = json.Unmarshal(res, &crumb)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &crumb, err
}

func (p *Pipeline) GetSCMServers() ([]devops.SCMServer, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var SCMServer []devops.SCMServer
	err = json.Unmarshal(res, &SCMServer)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return SCMServer, err
}

func (p *Pipeline) GetSCMOrg() ([]devops.SCMOrg, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var SCMOrg []devops.SCMOrg
	err = json.Unmarshal(res, &SCMOrg)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return SCMOrg, err
}

func (p *Pipeline) GetOrgRepo() (devops.OrgRepo, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return devops.OrgRepo{}, err
	}
	var OrgRepo devops.OrgRepo
	err = json.Unmarshal(res, &OrgRepo)
	if err != nil {
		klog.Error(err)
		return devops.OrgRepo{}, err
	}

	return OrgRepo, err
}

func (p *Pipeline) CreateSCMServers() (*devops.SCMServer, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	var SCMServer devops.SCMServer
	err = json.Unmarshal(res, &SCMServer)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &SCMServer, err
}

func (p *Pipeline) GetNotifyCommit() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GithubWebhook() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) Validate() (*devops.Validates, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var validates devops.Validates
	err = json.Unmarshal(res, &validates)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &validates, err
}

func (p *Pipeline) CheckScriptCompile() (*devops.CheckScript, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// Jenkins will return different struct according to different results.
	var checkScript devops.CheckScript
	ok := json.Unmarshal(res, &checkScript)
	if ok != nil {
		var resJson []*devops.CheckScript
		err := json.Unmarshal(res, &resJson)
		if err != nil {
			klog.Error(err)
			return nil, err
		}

		return resJson[0], nil
	}

	return &checkScript, err

}

func (p *Pipeline) CheckCron() (*devops.CheckCronRes, error) {

	var res = new(devops.CheckCronRes)

	Url, err := url.Parse(p.Jenkins.Server + p.Path)

	reqJenkins := &http.Request{
		Method: http.MethodGet,
		URL:    Url,
		Header: p.HttpParameters.Header,
	}

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(reqJenkins)

	if resp != nil && resp.StatusCode != http.StatusOK {
		resBody, _ := getRespBody(resp)
		return &devops.CheckCronRes{
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

func (p *Pipeline) ToJenkinsfile() (*devops.ResJenkinsfile, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var jenkinsfile devops.ResJenkinsfile
	err = json.Unmarshal(res, &jenkinsfile)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &jenkinsfile, err
}

func (p *Pipeline) ToJson() (*devops.ResJson, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var toJson devops.ResJson
	err = json.Unmarshal(res, &toJson)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &toJson, err
}
