package devops

import (
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
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
}

const (
	ProjectOwner      = "owner"
	ProjectMaintainer = "maintainer"
	ProjectDeveloper  = "developer"
	ProjectReporter   = "reporter"
)

const (
	JenkinsAllUserRoleName = "kubesphere-user"
)

type Role struct {
	Name        string `json:"name" description:"role's name e.g. owner'"`
	Description string `json:"description" description:"role 's description'"`
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

var JenkinsOwnerProjectPermissionIds = &jenkins.ProjectPermissionIds{
	CredentialCreate:        true,
	CredentialDelete:        true,
	CredentialManageDomains: true,
	CredentialUpdate:        true,
	CredentialView:          true,
	ItemBuild:               true,
	ItemCancel:              true,
	ItemConfigure:           true,
	ItemCreate:              true,
	ItemDelete:              true,
	ItemDiscover:            true,
	ItemMove:                true,
	ItemRead:                true,
	ItemWorkspace:           true,
	RunDelete:               true,
	RunReplay:               true,
	RunUpdate:               true,
	SCMTag:                  true,
}

var JenkinsProjectPermissionMap = map[string]jenkins.ProjectPermissionIds{
	ProjectOwner: {
		CredentialCreate:        true,
		CredentialDelete:        true,
		CredentialManageDomains: true,
		CredentialUpdate:        true,
		CredentialView:          true,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           true,
		ItemCreate:              true,
		ItemDelete:              true,
		ItemDiscover:            true,
		ItemMove:                true,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  true,
	},
	ProjectMaintainer: {
		CredentialCreate:        true,
		CredentialDelete:        true,
		CredentialManageDomains: true,
		CredentialUpdate:        true,
		CredentialView:          true,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           false,
		ItemCreate:              true,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  true,
	},
	ProjectDeveloper: {
		CredentialCreate:        false,
		CredentialDelete:        false,
		CredentialManageDomains: false,
		CredentialUpdate:        false,
		CredentialView:          false,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           false,
		ItemCreate:              false,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  false,
	},
	ProjectReporter: {
		CredentialCreate:        false,
		CredentialDelete:        false,
		CredentialManageDomains: false,
		CredentialUpdate:        false,
		CredentialView:          false,
		ItemBuild:               false,
		ItemCancel:              false,
		ItemConfigure:           false,
		ItemCreate:              false,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           false,
		RunDelete:               false,
		RunReplay:               false,
		RunUpdate:               false,
		SCMTag:                  false,
	},
}

var JenkinsPipelinePermissionMap = map[string]jenkins.ProjectPermissionIds{
	ProjectOwner: {
		CredentialCreate:        true,
		CredentialDelete:        true,
		CredentialManageDomains: true,
		CredentialUpdate:        true,
		CredentialView:          true,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           true,
		ItemCreate:              true,
		ItemDelete:              true,
		ItemDiscover:            true,
		ItemMove:                true,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  true,
	},
	ProjectMaintainer: {
		CredentialCreate:        true,
		CredentialDelete:        true,
		CredentialManageDomains: true,
		CredentialUpdate:        true,
		CredentialView:          true,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           true,
		ItemCreate:              true,
		ItemDelete:              true,
		ItemDiscover:            true,
		ItemMove:                true,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  true,
	},
	ProjectDeveloper: {
		CredentialCreate:        false,
		CredentialDelete:        false,
		CredentialManageDomains: false,
		CredentialUpdate:        false,
		CredentialView:          false,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           false,
		ItemCreate:              false,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  false,
	},
	ProjectReporter: {
		CredentialCreate:        false,
		CredentialDelete:        false,
		CredentialManageDomains: false,
		CredentialUpdate:        false,
		CredentialView:          false,
		ItemBuild:               false,
		ItemCancel:              false,
		ItemConfigure:           false,
		ItemCreate:              false,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           false,
		RunDelete:               false,
		RunReplay:               false,
		RunUpdate:               false,
		SCMTag:                  false,
	},
}
