package devops

import (
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
)

type Interface interface {
	CreateFolder(name, description string, parents ...string) (*jenkins.Folder, error)

	CreateJobInFolder(config string, jobName string, parentIDs ...string) (*jenkins.Job, error)

	DeleteJob(name string, parentIDs ...string) (bool, error)

	BuildJob(name string, options ...interface{}) (int64, error)

	GetBuild(jobName string, number int64) (*jenkins.Build, error)

	GetJob(id string, parentIDs ...string) (*jenkins.Job, error)

	GetFolder(id string, parents ...string) (*jenkins.Folder, error)

	GetGlobalRole(roleName string) (*jenkins.GlobalRole, error)

	GetProjectRole(roleName string) (*jenkins.ProjectRole, error)

	AddGlobalRole(roleName string, ids jenkins.GlobalPermissionIds, overwrite bool) (*jenkins.GlobalRole, error)

	DeleteProjectRoles(roleName ...string) error

	AddProjectRole(roleName string, pattern string, ids jenkins.ProjectPermissionIds, overwrite bool) (*jenkins.ProjectRole, error)

	DeleteUserInProject(username string) error

	CredentialOperator

	PipelineOperator

	AdapterOperator
}

type AdapterOperator interface {
	//SendPureRequest(path string, req *http.Request) ([]byte, error)

	//SendRequestWithHeaderResp(path string, req *http.Request) ([]byte, http.Header, error)

	ValidateJenkinsfile(jenkinsfile string) (*jenkins.ValidateJenkinsfileResponse, error)

	ValidatePipelineJson(json string) (*jenkins.ValidatePipelineJsonResponse, error)

	PipelineJsonToJenkinsfile(json string) (*jenkins.PipelineJsonToJenkinsfileResponse, error)

	JenkinsfileToPipelineJson(jenkinsfile string) (*jenkins.JenkinsfileToPipelineJsonResponse, error)
}
