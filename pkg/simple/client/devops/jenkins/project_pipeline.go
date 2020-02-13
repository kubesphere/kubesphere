package jenkins

import (
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
