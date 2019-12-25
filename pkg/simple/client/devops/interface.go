package devops

type Job struct {
}

type Interface interface {
	GetJob(projectId, pipelineName string) (*Job, error)

	DeleteJob(projectId, pipelineId string) (bool, error)

	CreateJobInFolder()

	GetGlobalRole(roleName string)

	AddGlobalRole(roleName string, permission string)

	GetProjectRole(roleName string)
}
