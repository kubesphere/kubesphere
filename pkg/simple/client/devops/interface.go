package devops

import "net/http"

type Interface interface {

	SendJenkinsRequest(baseUrl string, req *http.Request) ([]byte, error)

	SendJenkinsRequestWithHeaderResp(baseUrl string, req *http.Request) ([]byte, http.Header, error)

	ValidateJenkinsfile(jenkinsfile string) (*ValidateJenkinsfileResponse, error)

	ValidatePipelineJson(json string) (*ValidatePipelineJsonResponse, error)

	PipelineJsonToJenkinsfile(json string) (*PipelineJsonToJenkinsfileResponse, error)

	JenkinsfileToPipelineJson(jenkinsfile string) (*JenkinsfileToPipelineJsonResponse, error)

	StepsJsonToJenkinsfile(json string) (*StepJsonToJenkinsfileResponse, error)

	StepsJenkinsfileToJson(jenkinsfile string) (*StepsJenkinsfileToJsonResponse, error)

	Init() (*Jenkins, error)

	Info() (*ExecutorResponse, error)

	CreateNode(name string, numExecutors int, description string, remoteFS string, label string, options ...interface{}) (*Node, error)

	DeleteNode(name string) (bool, error)

	CreateFolder(name, description string, parents ...string) (*Folder, error)

	CreateJobInFolder(config string, jobName string, parentIDs ...string) (*Job, error)

	CreateJob(config string, options ...interface{}) (*Job, error)

	RenameJob(job string, name string) *Job

	CopyJob(copyFrom string, newName string) (*Job, error)

	DeleteJob(name string, parentIDs ...string) (bool, error)

	BuildJob(name string, options ...interface{}) (int64, error)

	GetNode(name string) (*Node, error)

	GetLabel(name string) (*Label, error)

	GetBuild(jobName string, number int64) (*Build, error)

	GetJob(id string, parentIDs ...string) (*Job, error)

	GetSubJob(parentId string, childId string) (*Job, error)

	GetFolder(id string, parents ...string) (*Folder, error)

	GetAllNodes() ([]*Node, error)

	GetAllBuildIds(job string) ([]JobBuild, error)

	GetAllBuildStatus(jobId string) ([]JobBuildStatus, error)

	GetAllJobNames() ([]InnerJob, error)

	GetAllJobs() ([]*Job, error)

	GetQueue() (*Queue, error)

	GetQueueUrl() string

	GetArtifactData(id string) (*FingerPrintResponse, error)

	GetPlugins(depth int) (*Plugins, error)

	HasPlugin(name string) (*Plugin, error)

	ValidateFingerPrint(id string) (bool, error)

	GetView(name string) (*View, error)

	GetAllViews() ([]*View, error)

	CreateView(name string, viewType string) (*View, error)

	Poll() (int, error)

	CreateSshCredential(id, username, passphrase, privateKey, description string) (*string, error)

	CreateUsernamePasswordCredential(id, username, password, description string) (*string, error)

	CreateSshCredentialInFolder(domain, id, username, passphrase, privateKey, description string, folders ...string) (*string, error)

	CreateUsernamePasswordCredentialInFolder(domain, id, username, password, description string, folders ...string) (*string, error)

	CreateSecretTextCredentialInFolder(domain, id, secret, description string, folders ...string) (*string, error)

	CreateKubeconfigCredentialInFolder(domain, id, content, description string, folders ...string) (*string, error)

	UpdateSshCredentialInFolder(domain, id, username, passphrase, privateKey, description string, folders ...string) (*string, error)

	UpdateUsernamePasswordCredentialInFolder(domain, id, username, password, description string, folders ...string) (*string, error)

	UpdateSecretTextCredentialInFolder(domain, id, secret, description string, folders ...string) (*string, error)

	UpdateKubeconfigCredentialInFolder(domain, id, content, description string, folders ...string) (*string, error)

	GetCredentialInFolder(domain, id string, folders ...string) (*CredentialResponse, error)

	GetCredentialContentInFolder(domain, id string, folders ...string) (string, error)

	GetCredentialsInFolder(domain string, folders ...string) ([]*CredentialResponse, error)

	DeleteCredentialInFolder(domain, id string, folders ...string) (*string, error)

	GetGlobalRole(roleName string) (*GlobalRole, error)

	GetProjectRole(roleName string) (*ProjectRole, error)

	AddGlobalRole(roleName string, ids GlobalPermissionIds, overwrite bool) (*GlobalRole, error)

	DeleteProjectRoles(roleName ...string) error

	AddProjectRole(roleName string, pattern string, ids ProjectPermissionIds, overwrite bool) (*ProjectRole, error)

	DeleteUserInProject(username string) error

	GetQueueItem(number int64) (*QueueItemResponse, error)
}
