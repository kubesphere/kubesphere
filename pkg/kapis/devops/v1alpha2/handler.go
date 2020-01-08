package v1alpha2

import (
	"kubesphere.io/kubesphere/pkg/models/devops"
)

type ProjectPipelineHandler struct {
	projectCredentialOperator devops.ProjectCredentialOperator
	projectMemberOperator     devops.ProjectMemberOperator
	projectPipelineOperator   devops.ProjectPipelineOperator
	devopsOperator            devops.DevopsOperator
}

type PipelineSonarHandler struct {
	pipelineSonarGetter devops.PipelineSonarGetter
}

type S2iHandler struct {
	s2iUploader devops.S2iBinaryUploader
}

func New() *ProjectPipelineHandler {
	handler := &ProjectPipelineHandler{
		projectCredentialOperator: devops.NewProjectCredentialOperator(),
		projectMemberOperator:     devops.NewProjectMemberOperator(),
		projectPipelineOperator:   devops.NewProjectPipelineOperator(),
		devopsOperator:            devops.NewDevopsOperator(),
	}

	return handler
}
