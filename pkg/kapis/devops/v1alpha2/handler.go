package v1alpha2

import (
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/devops"
	devopsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
)

type ProjectPipelineHandler struct {
	devopsOperator          devops.DevopsOperator
	projectCredentialGetter devops.ProjectCredentialGetter
}

type PipelineSonarHandler struct {
	pipelineSonarGetter devops.PipelineSonarGetter
}

func NewProjectPipelineHandler(devopsClient devopsClient.Interface) ProjectPipelineHandler {
	return ProjectPipelineHandler{
		devopsOperator:          devops.NewDevopsOperator(devopsClient, nil, nil, nil, nil),
		projectCredentialGetter: devops.NewProjectCredentialOperator(devopsClient),
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
