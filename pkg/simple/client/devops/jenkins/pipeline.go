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
	GetPipelineUrl    = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/"
	ListPipelinesUrl  = "/blue/rest/search/?"
	GetPipelineRunUrl = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/"
	ListPipelineRunUrl     = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/?"
	StopPipelineUrl          = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/stop/?"
	ReplayPipelineUrl        = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/%s/replay/"
	RunPipelineUrl           = "/blue/rest/organizations/jenkins/pipelines/%s/pipelines/%s/runs/"

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

func (p *Pipeline)GetPipelineRun()([]byte, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (p *Pipeline)ListPipelineRuns()([]byte, error){
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

func (p *Pipeline)StopPipeline()([]byte, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (p *Pipeline)ReplayPipeline()([]byte, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}

func (p *Pipeline)RunPipeline()([]byte, error){
	res, err := p.Jenkins.SendPureRequest(p.Path, p.Request)
	if err != nil {
		klog.Error(err)
	}
	return res, err
}



	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
)

func (j *Jenkins) CreateProjectPipeline(projectId string, pipeline *devops.ProjectPipeline) (string, error) {
	switch pipeline.Type {
	case devops.NoScmPipelineType:

		config, err := createPipelineConfigXml(pipeline.Pipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := j.GetJob(pipeline.Pipeline.Name, projectId)
		if job != nil {
			err := fmt.Errorf("job name [%s] has been used", job.GetName())
			return "", restful.NewError(http.StatusConflict, err.Error())
		}

		if err != nil && GetJenkinsStatusCode(err) != http.StatusNotFound {
			return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}

		_, err = j.CreateJobInFolder(config, pipeline.Pipeline.Name, projectId)
		if err != nil {
			return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}

		return pipeline.Pipeline.Name, nil
	case devops.MultiBranchPipelineType:
		config, err := createMultiBranchPipelineConfigXml(projectId, pipeline.MultiBranchPipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := j.GetJob(pipeline.MultiBranchPipeline.Name, projectId)
		if job != nil {
			err := fmt.Errorf("job name [%s] has been used", job.GetName())
			return "", restful.NewError(http.StatusConflict, err.Error())
		}

		if err != nil && GetJenkinsStatusCode(err) != http.StatusNotFound {
			return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}

		_, err = j.CreateJobInFolder(config, pipeline.MultiBranchPipeline.Name, projectId)
		if err != nil {
			return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}

		return pipeline.MultiBranchPipeline.Name, nil

	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
}

func (j *Jenkins) DeleteProjectPipeline(projectId string, pipelineId string) (string, error) {
	_, err := j.DeleteJob(pipelineId, projectId)
	if err != nil {
		return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	return pipelineId, nil

}
func (j *Jenkins) UpdateProjectPipeline(projectId string, pipeline *devops.ProjectPipeline) (string, error) {
	switch pipeline.Type {
	case devops.NoScmPipelineType:

		config, err := createPipelineConfigXml(pipeline.Pipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := j.GetJob(pipeline.Pipeline.Name, projectId)

		if err != nil {
			return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}

		err = job.UpdateConfig(config)
		if err != nil {
			return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}

		return pipeline.Pipeline.Name, nil
	case devops.MultiBranchPipelineType:

		config, err := createMultiBranchPipelineConfigXml(projectId, pipeline.MultiBranchPipeline)
		if err != nil {
			klog.Errorf("%+v", err)

			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := j.GetJob(pipeline.MultiBranchPipeline.Name, projectId)

		if err != nil {
			return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}

		err = job.UpdateConfig(config)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}

		return pipeline.MultiBranchPipeline.Name, nil

	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
}

func (j *Jenkins) GetProjectPipelineConfig(projectId, pipelineId string) (*devops.ProjectPipeline, error) {
	job, err := j.GetJob(pipelineId, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	switch job.Raw.Class {
	case "org.jenkinsci.plugins.workflow.job.WorkflowJob":
		config, err := job.GetConfig()
		if err != nil {
			return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}
		pipeline, err := parsePipelineConfigXml(config)
		if err != nil {
			return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}
		pipeline.Name = pipelineId
		return &devops.ProjectPipeline{
			Type:     devops.NoScmPipelineType,
			Pipeline: pipeline,
		}, nil

	case "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject":
		config, err := job.GetConfig()
		if err != nil {
			return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}
		pipeline, err := parseMultiBranchPipelineConfigXml(config)
		if err != nil {
			return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}
		pipeline.Name = pipelineId
		return &devops.ProjectPipeline{
			Type:                devops.MultiBranchPipelineType,
			MultiBranchPipeline: pipeline,
		}, nil
	default:
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}
}
