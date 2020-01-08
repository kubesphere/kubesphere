package jenkins

import (
	"encoding/json"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
)

type Pipeline struct {
	Request *http.Request
	Jenkins *Jenkins
	Path    string
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

func (p *Pipeline) GetPipeline() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (p *Pipeline) ListPipelines() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	count, err := p.searchPipelineCount()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	responseStruct := devops.PageableResponse{TotalCount: count}
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

func (p *Pipeline) searchPipelineCount() (int, error) {
	query, _ := parseJenkinsQuery(p.Request.URL.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")

	formatUrl := ListPipelinesUrl + query.Encode()

	res, err := p.Jenkins.SendPureRequest(formatUrl, p.Request)
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

func (p *Pipeline) GetPipelineRun() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (p *Pipeline) ListPipelineRuns() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (p *Pipeline) searchPipelineRunsCount() (int, error) {
	query, _ := parseJenkinsQuery(p.Request.URL.RawQuery)
	query.Set("start", "0")
	query.Set("limit", "1000")
	query.Set("depth", "-1")
	//formatUrl := fmt.Sprintf(SearchPipelineRunUrl, projectName, pipelineName)

	res, err := p.Jenkins.SendPureRequest(ListPipelineRunUrl+query.Encode(), p.Request)
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

func (p *Pipeline) StopPipeline() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (p *Pipeline) ReplayPipeline() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (p *Pipeline) RunPipeline() ([]byte, error) {
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}
