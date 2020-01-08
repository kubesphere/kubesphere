package devops

import (
	"net/http"
)

type Interface interface {
	SendJenkinsRequest(baseUrl string, req *http.Request) ([]byte, error)

	SendJenkinsRequestWithHeaderResp(baseUrl string, req *http.Request) ([]byte, http.Header, error)

	//ValidateJenkinsfile(jenkinsfile string) (*jenkins.ValidateJenkinsfileResponse, error)
	//
	//ValidatePipelineJson(json string) (*jenkins.ValidatePipelineJsonResponse, error)
	//
	//PipelineJsonToJenkinsfile(json string) (*jenkins.PipelineJsonToJenkinsfileResponse, error)
	//
	//JenkinsfileToPipelineJson(jenkinsfile string) (*jenkins.JenkinsfileToPipelineJsonResponse, error)
	//
	//CreateFolder(name, description string, parents ...string) (*jenkins.Folder, error)
	//
	//CreateJobInFolder(config string, jobName string, parentIDs ...string) (*jenkins.Job, error)
	//
	//DeleteJob(name string, parentIDs ...string) (bool, error)
	//
	//BuildJob(name string, options ...interface{}) (int64, error)
	//
	//GetBuild(jobName string, number int64) (*jenkins.Build, error)
	//
	//GetJob(id string, parentIDs ...string) (*jenkins.Job, error)
	//
	//GetFolder(id string, parents ...string) (*jenkins.Folder, error)
	//
	//GetGlobalRole(roleName string) (*jenkins.GlobalRole, error)
	//
	//GetProjectRole(roleName string) (*jenkins.ProjectRole, error)
	//
	//AddGlobalRole(roleName string, ids jenkins.GlobalPermissionIds, overwrite bool) (*jenkins.GlobalRole, error)
	//
	//DeleteProjectRoles(roleName ...string) error
	//
	//AddProjectRole(roleName string, pattern string, ids jenkins.ProjectPermissionIds, overwrite bool) (*jenkins.ProjectRole, error)

	DeleteUserInProject(username string) error

	CredentialOperator

	BuildGetter

	PipelineOperator

	ProjectMemberOperator
}

const (
	ProjectOwner      = "owner"
	ProjectMaintainer = "maintainer"
	ProjectDeveloper  = "developer"
	ProjectReporter   = "reporter"
)

type Role struct {
	Name        string `json:"name" description:"role's name e.g. owner'"`
	Description string `json:"description" description:"role 's description'"`
}

	CredentialOperator

	PipelineOperator

	//AdapterOperator
}

type AdapterOperator interface {
	//SendPureRequest(path string, req *http.Request) ([]byte, error)

	//SendRequestWithHeaderResp(path string, req *http.Request) ([]byte, http.Header, error)

// 	ValidateJenkinsfile(jenkinsfile string) (*jenkins.ValidateJenkinsfileResponse, error)

// 	ValidatePipelineJson(json string) (*jenkins.ValidatePipelineJsonResponse, error)

// 	PipelineJsonToJenkinsfile(json string) (*jenkins.PipelineJsonToJenkinsfileResponse, error)

// 	JenkinsfileToPipelineJson(jenkinsfile string) (*jenkins.JenkinsfileToPipelineJsonResponse, error)
}
var DefaultRoles = []*Role{
	{
		Name:        ProjectOwner,
		Description: "Owner have access to do all the operations of a DevOps project and own the highest permissions as well.",
	},
	{
		Name:        ProjectMaintainer,
		Description: "Maintainer have access to manage pipeline and credential configuration in a DevOps project.",
	},
	{
		Name:        ProjectDeveloper,
		Description: "Developer is able to view and trigger the pipeline.",
	},
	{
		Name:        ProjectReporter,
		Description: "Reporter is only allowed to view the status of the pipeline.",
	},
}

var AllRoleSlice = []string{ProjectDeveloper, ProjectReporter, ProjectMaintainer, ProjectOwner}
