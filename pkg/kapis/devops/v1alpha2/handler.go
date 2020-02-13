package v1alpha2

import (
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/devops"
	devopsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
)

type ProjectPipelineHandler struct {
	projectCredentialOperator devops.ProjectCredentialOperator
	projectMemberOperator     devops.ProjectMemberOperator
	projectPipelineOperator   devops.ProjectPipelineOperator
	devopsOperator            devops.DevopsOperator
	projectOperator           devops.ProjectOperator
}

type PipelineSonarHandler struct {
	pipelineSonarGetter devops.PipelineSonarGetter
}

func NewProjectPipelineHandler(devopsClient devopsClient.Interface, dbClient *mysql.Database) ProjectPipelineHandler {
	return ProjectPipelineHandler{
		projectCredentialOperator: devops.NewProjectCredentialOperator(devopsClient, dbClient),
		projectMemberOperator:     devops.NewProjectMemberOperator(devopsClient, dbClient),
		projectPipelineOperator:   devops.NewProjectPipelineOperator(devopsClient),
		devopsOperator:            devops.NewDevopsOperator(devopsClient),
		projectOperator:           devops.NewProjectOperator(dbClient),
	}
}

func NewPipelineSonarHandler(devopsClient devopsClient.Interface, sonarClient sonarqube.SonarInterface) PipelineSonarHandler {
	return PipelineSonarHandler{
		pipelineSonarGetter: devops.NewPipelineSonarGetter(devopsClient, sonarClient),
	}
}

func NewS2iBinaryHandler(client versioned.Interface, informers externalversions.SharedInformerFactory, s3Client s3.Interface) S2iBinaryHandler {
	return S2iBinaryHandler{devops.NewS2iBinaryUploader(client, informers, s3Client)}
}
