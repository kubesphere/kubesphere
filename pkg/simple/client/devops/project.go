package devops

/**
project operator, providing API for creating/getting/deleting projects
The actual data of the project is stored in the CRD,
so we only need to create the project with the corresponding ID in the CI/CD system.
*/
type ProjectOperator interface {
	CreateDevOpsProject(projectId string) (string, error)
	DeleteDevOpsProject(projectId string) error
	GetDevOpsProject(projectId string) (string, error)
}
