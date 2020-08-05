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
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
)

func (j *Jenkins) CreateProjectPipeline(projectId string, pipeline *devopsv1alpha3.Pipeline) (string, error) {
	switch pipeline.Spec.Type {
	case devopsv1alpha3.NoScmPipelineType:

		config, err := createPipelineConfigXml(pipeline.Spec.Pipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := j.GetJob(pipeline.Name, projectId)
		if job != nil {
			err := fmt.Errorf("job name [%s] has been used", job.GetName())
			return "", restful.NewError(http.StatusConflict, err.Error())
		}

		if err != nil && devops.GetDevOpsStatusCode(err) != http.StatusNotFound {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}

		_, err = j.CreateJobInFolder(config, pipeline.Name, projectId)
		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}

		return pipeline.Name, nil
	case devopsv1alpha3.MultiBranchPipelineType:
		config, err := createMultiBranchPipelineConfigXml(projectId, pipeline.Spec.MultiBranchPipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := j.GetJob(pipeline.Name, projectId)
		if job != nil {
			err := fmt.Errorf("job name [%s] has been used", job.GetName())
			return "", restful.NewError(http.StatusConflict, err.Error())
		}

		if err != nil && devops.GetDevOpsStatusCode(err) != http.StatusNotFound {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}

		_, err = j.CreateJobInFolder(config, pipeline.Name, projectId)
		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}

		return pipeline.Name, nil

	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
}

func (j *Jenkins) DeleteProjectPipeline(projectId string, pipelineId string) (string, error) {
	_, err := j.DeleteJob(pipelineId, projectId)
	if err != nil {
		return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
	}
	return pipelineId, nil

}
func (j *Jenkins) UpdateProjectPipeline(projectId string, pipeline *devopsv1alpha3.Pipeline) (string, error) {
	switch pipeline.Spec.Type {
	case devopsv1alpha3.NoScmPipelineType:

		config, err := createPipelineConfigXml(pipeline.Spec.Pipeline)
		if err != nil {
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := j.GetJob(pipeline.Name, projectId)

		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}

		err = job.UpdateConfig(config)
		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}

		return pipeline.Name, nil
	case devopsv1alpha3.MultiBranchPipelineType:

		config, err := createMultiBranchPipelineConfigXml(projectId, pipeline.Spec.MultiBranchPipeline)
		if err != nil {
			klog.Errorf("%+v", err)

			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := j.GetJob(pipeline.Spec.MultiBranchPipeline.Name, projectId)

		if err != nil {
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}

		err = job.UpdateConfig(config)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}

		return pipeline.Name, nil

	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
}

func (j *Jenkins) GetProjectPipelineConfig(projectId, pipelineId string) (*devopsv1alpha3.Pipeline, error) {
	job, err := j.GetJob(pipelineId, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
	}
	switch job.Raw.Class {
	case "org.jenkinsci.plugins.workflow.job.WorkflowJob":
		config, err := job.GetConfig()
		if err != nil {
			return nil, restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}
		pipeline, err := parsePipelineConfigXml(config)
		if err != nil {
			return nil, restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}
		pipeline.Name = pipelineId
		return &devopsv1alpha3.Pipeline{
			Spec: devopsv1alpha3.PipelineSpec{
				Type:     devopsv1alpha3.NoScmPipelineType,
				Pipeline: pipeline,
			},
		}, nil

	case "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject":
		config, err := job.GetConfig()
		if err != nil {
			return nil, restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}
		pipeline, err := parseMultiBranchPipelineConfigXml(config)
		if err != nil {
			return nil, restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
		}
		pipeline.Name = pipelineId
		return &devopsv1alpha3.Pipeline{
			Spec: devopsv1alpha3.PipelineSpec{
				Type:                devopsv1alpha3.MultiBranchPipelineType,
				MultiBranchPipeline: pipeline,
			},
		}, nil
	default:
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}
}
