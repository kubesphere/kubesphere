package jenkins

import (
	"encoding/json"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
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

	GetBranchPipelineUrl   = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/"
	GetBranchPipelineRunUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/"
	StopBranchPipelineUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/stop/?"
	ReplayBranchPipelineUrl  = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/replay/"
	RunBranchPipelineUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/"
	GetBranchRunLogUrl       = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/log/?"
	GetBranchStepLogUrl      = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/log/?"
	GetBranchNodeStepsUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/?"
	GetBranchPipeRunNodesUrl = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/?"
	CheckBranchPipelineUrl   = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/%s/runs/%s/nodes/%s/steps/%s/"
	GetPipeBranchUrl         = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/branches/?"
	ScanBranchUrl            = "/job/%s/job/%s/build?"

)

func (p *Pipeline) GetPipeline() (*devops.Pipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
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
	}

	var pipelineRunList devops.PipelineRunList
	err = json.Unmarshal(res, &pipelineRunList)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &pipelineRunList, err
}

func (p *Pipeline) searchPipelineRunsCount() (int, error) {
	query, _ := parseJenkinsQuery(p.HttpParameters.Url.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")
	//formatUrl := fmt.Sprintf(SearchPipelineRunUrl, projectName, pipelineName)

	res, err := p.Jenkins.SendPureRequest(ListPipelineRunUrl+query.Encode(), p.HttpParameters)
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

func (p *Pipeline) GetNodeSteps() (*devops.NodeSteps, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	var nodeSteps devops.NodeSteps
	err = json.Unmarshal(res, &nodeSteps)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &nodeSteps, err
}

func (p *Pipeline) GetPipelineRunNodes() ([]devops.PipelineRunNodes, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
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
	}
	var branchPipeline devops.BranchPipeline
	err = json.Unmarshal(res, &branchPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchPipeline, err
}

func (p *Pipeline) GetBranchPipelineRun()(*devops.PipelineRun, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}
	var branchPipelineRun devops.PipelineRun
	err = json.Unmarshal(res, &branchPipelineRun)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchPipelineRun, err
}

func (p *Pipeline) StopBranchPipeline()(*devops.StopPipeline, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}
	var branchStopPipeline devops.StopPipeline
	err = json.Unmarshal(res, &branchStopPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchStopPipeline, err
}

func (p *Pipeline) ReplayBranchPipeline()(*devops.ReplayPipeline, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}
	var branchReplayPipeline devops.ReplayPipeline
	err = json.Unmarshal(res, &branchReplayPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchReplayPipeline, err
}

func (p *Pipeline) RunBranchPipeline()(*devops.RunPipeline, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}
	var branchRunPipeline devops.RunPipeline
	err = json.Unmarshal(res, &branchRunPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchRunPipeline, err
}

func (p *Pipeline) GetBranchRunLog()([]byte, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GetBranchStepLog()([]byte, http.Header, error){
	res, header, err := p.Jenkins.SendPureRequestWithHeaderResp(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, header, err
}

func (p *Pipeline) GetBranchNodeSteps()(*devops.NodeSteps, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}
	var branchNodeSteps devops.NodeSteps
	err = json.Unmarshal(res, &branchNodeSteps)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchNodeSteps, err
}

func (p *Pipeline) GetBranchPipelineRunNodes()(*devops.BranchPipelineRunNodes, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}
	var branchPipelineRunNodes devops.BranchPipelineRunNodes
	err = json.Unmarshal(res, &branchPipelineRunNodes)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &branchPipelineRunNodes, err
}

func (p *Pipeline) SubmitBranchInputStep()([]byte, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}

func (p *Pipeline) GetPipelineBranch()(*devops.PipelineBranch, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}
	var pipelineBranch devops.PipelineBranch
	err = json.Unmarshal(res, &pipelineBranch)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &pipelineBranch, err
}

func (p *Pipeline) ScanBranch()([]byte, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	return res, err
}
