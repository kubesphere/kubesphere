package jenkins

import (
	"encoding/json"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
)

type Pipeline struct {
	HttpParameters *devops.HttpParameters
	Jenkins        *Jenkins
	Path           string
}

const (
	GetPipelineUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/"
	ListPipelinesUrl   = "/blue/rest/search/?"
	GetPipelineRunUrl  = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/"
	ListPipelineRunUrl = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/?"
	StopPipelineUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/stop/?"
	ReplayPipelineUrl  = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/replay/"
	RunPipelineUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/"
)

func (p *Pipeline) GetPipeline() (*devops.Pipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}
	var pipeline *devops.Pipeline

	err = json.Unmarshal(res, pipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return pipeline, err
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

	var pipelineRun *devops.PipelineRun
	err = json.Unmarshal(res, pipelineRun)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return pipelineRun, err
}

func (p *Pipeline) ListPipelineRuns() (*devops.PipelineRunList, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	var pipelineRunList *devops.PipelineRunList
	err = json.Unmarshal(res, pipelineRunList)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return pipelineRunList, err
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

	var stopPipeline *devops.StopPipeline
	err = json.Unmarshal(res, stopPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return stopPipeline, err
}

func (p *Pipeline) ReplayPipeline() (*devops.ReplayPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	var replayPipeline *devops.ReplayPipeline
	err = json.Unmarshal(res, replayPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return replayPipeline, err
}

func (p *Pipeline) RunPipeline() (*devops.RunPipeline, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.HttpParameters)
	if err != nil {
		klog.Error(err)
	}

	var runPipeline *devops.RunPipeline
	err = json.Unmarshal(res, runPipeline)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return runPipeline, err
}
