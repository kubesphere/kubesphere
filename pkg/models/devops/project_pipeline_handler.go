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
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/gojenkins/utils"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
)

func CreateProjectPipeline(projectId string, pipeline *ProjectPipeline) (string, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return "", restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	switch pipeline.Type {
	case NoScmPipelineType:

		config, err := createPipelineConfigXml(pipeline.Pipeline)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := jenkinsClient.GetJob(pipeline.Pipeline.Name, projectId)
		if job != nil {
			err := fmt.Errorf("job name [%s] has been used", job.GetName())
			klog.Warning(err.Error())
			return "", restful.NewError(http.StatusConflict, err.Error())
		}

		if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		_, err = jenkinsClient.CreateJobInFolder(config, pipeline.Pipeline.Name, projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		return pipeline.Pipeline.Name, nil
	case MultiBranchPipelineType:
		config, err := createMultiBranchPipelineConfigXml(projectId, pipeline.MultiBranchPipeline)
		if err != nil {
			klog.Errorf("%+v", err)

			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := jenkinsClient.GetJob(pipeline.MultiBranchPipeline.Name, projectId)
		if job != nil {
			err := fmt.Errorf("job name [%s] has been used", job.GetName())
			klog.Warning(err.Error())
			return "", restful.NewError(http.StatusConflict, err.Error())
		}

		if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		_, err = jenkinsClient.CreateJobInFolder(config, pipeline.MultiBranchPipeline.Name, projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		return pipeline.MultiBranchPipeline.Name, nil

	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
}

func DeleteProjectPipeline(projectId string, pipelineId string) (string, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return "", restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	_, err = jenkinsClient.DeleteJob(pipelineId, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	return pipelineId, nil
}

func UpdateProjectPipeline(projectId, pipelineId string, pipeline *ProjectPipeline) (string, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return "", restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	switch pipeline.Type {
	case NoScmPipelineType:

		config, err := createPipelineConfigXml(pipeline.Pipeline)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := jenkinsClient.GetJob(pipelineId, projectId)

		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		err = job.UpdateConfig(config)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		return pipeline.Pipeline.Name, nil
	case MultiBranchPipelineType:

		config, err := createMultiBranchPipelineConfigXml(projectId, pipeline.MultiBranchPipeline)
		if err != nil {
			klog.Errorf("%+v", err)

			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}

		job, err := jenkinsClient.GetJob(pipelineId, projectId)

		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		err = job.UpdateConfig(config)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		return pipeline.MultiBranchPipeline.Name, nil

	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
}

func GetProjectPipeline(projectId, pipelineId string) (*ProjectPipeline, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	job, err := jenkinsClient.GetJob(pipelineId, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	switch job.Raw.Class {
	case "org.jenkinsci.plugins.workflow.job.WorkflowJob":
		config, err := job.GetConfig()
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		pipeline, err := parsePipelineConfigXml(config)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		pipeline.Name = pipelineId
		return &ProjectPipeline{
			Type:     NoScmPipelineType,
			Pipeline: pipeline,
		}, nil

	case "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject":
		config, err := job.GetConfig()
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		pipeline, err := parseMultiBranchPipelineConfigXml(config)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		pipeline.Name = pipelineId
		return &ProjectPipeline{
			Type:                MultiBranchPipelineType,
			MultiBranchPipeline: pipeline,
		}, nil
	default:
		err := fmt.Errorf("error unsupport job type")
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())

	}
}

func GetPipelineSonar(projectId, pipelineId string) ([]*SonarStatus, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	job, err := jenkinsClient.GetJob(pipelineId, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	build, err := job.GetLastBuild()
	if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	} else if err != nil {
		klog.Errorf("%+v", err)
		return nil, nil
	}

	sonarStatus, err := getBuildSonarResults(build)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}
	if len(sonarStatus) == 0 {
		build, err := job.GetLastCompletedBuild()
		if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		} else if err != nil {
			klog.Errorf("%+v", err)
			return nil, nil
		}
		sonarStatus, err = getBuildSonarResults(build)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusBadRequest, err.Error())
		}
	}
	return sonarStatus, nil
}

func GetMultiBranchPipelineSonar(projectId, pipelineId, branchId string) ([]*SonarStatus, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	job, err := jenkinsClient.GetJob(branchId, projectId, pipelineId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	build, err := job.GetLastBuild()
	if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	} else if err != nil {
		klog.Errorf("%+v", err)
		return nil, nil
	}

	sonarStatus, err := getBuildSonarResults(build)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}
	if len(sonarStatus) == 0 {
		build, err := job.GetLastCompletedBuild()
		if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		} else if err != nil {
			klog.Errorf("%+v", err)
			return nil, nil
		}
		sonarStatus, err = getBuildSonarResults(build)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusBadRequest, err.Error())
		}
	}
	return sonarStatus, nil
}
