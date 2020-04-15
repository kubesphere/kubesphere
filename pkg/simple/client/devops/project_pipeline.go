package devops

import "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"

type ProjectPipelineOperator interface {
	CreateProjectPipeline(projectId string, pipeline *v1alpha3.Pipeline) (string, error)
	DeleteProjectPipeline(projectId string, pipelineId string) (string, error)
	UpdateProjectPipeline(projectId string, pipeline *v1alpha3.Pipeline) (string, error)
	GetProjectPipelineConfig(projectId, pipelineId string) (*v1alpha3.Pipeline, error)
}
