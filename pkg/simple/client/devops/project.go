package devops

import "kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"

type ProjectOperator interface {
	CreateDevOpsProject(username string, project *v1alpha2.DevOpsProject) (*v1alpha2.DevOpsProject, error)
	DeleteDevOpsProject(projectId string) error
}
