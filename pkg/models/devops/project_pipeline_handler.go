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
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"net/http"
)

type ProjectPipelineOperator interface {
	CreateProjectPipeline(projectId string, pipeline *devops.ProjectPipeline) (string, error)
	DeleteProjectPipeline(projectId string, pipelineId string) (string, error)
	UpdateProjectPipeline(projectId, pipelineId string, pipeline *devops.ProjectPipeline) (string, error)
	GetProjectPipelineConfig(projectId, pipelineId string) (*devops.ProjectPipeline, error)
}
type projectPipelineOperator struct {
	pipelineOperator devops.ProjectPipelineOperator
}

func NewProjectPipelineOperator(client jenkins.Client) ProjectPipelineOperator {
	return &projectPipelineOperator{}
}

func (o *projectPipelineOperator) CreateProjectPipeline(projectId string, pipeline *devops.ProjectPipeline) (string, error) {
	return o.pipelineOperator.CreateProjectPipeline(projectId, pipeline)
}

func (o *projectPipelineOperator) DeleteProjectPipeline(projectId string, pipelineId string) (string, error) {
	return o.pipelineOperator.DeleteProjectPipeline(projectId, pipelineId)
}

func (o *projectPipelineOperator) UpdateProjectPipeline(projectId, pipelineId string, pipeline *devops.ProjectPipeline) (string, error) {

	switch pipeline.Type {
	case devops.NoScmPipelineType:
		pipeline.Pipeline.Name = pipelineId
	case devops.MultiBranchPipelineType:
		pipeline.MultiBranchPipeline.Name = pipelineId
	default:
		err := fmt.Errorf("error unsupport pipeline type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
	return o.pipelineOperator.UpdateProjectPipeline(projectId, pipeline)
}

func (o *projectPipelineOperator) GetProjectPipelineConfig(projectId, pipelineId string) (*devops.ProjectPipeline, error) {
	return o.pipelineOperator.GetProjectPipelineConfig(projectId, pipelineId)
}
