package devops

import (
	"net/http"
	"time"
)

type Credential struct {
	Id          string `json:"id" description:"Id of Credential, e.g. dockerhub-id"`
	Type        string `json:"type" description:"Type of Credential, e.g. ssh/kubeconfig"`
	DisplayName string `json:"display_name,omitempty" description:"Credential's display name"`
	Fingerprint *struct {
		FileName string `json:"file_name,omitempty" description:"Credential's display name and description"`
		Hash     string `json:"hash,omitempty" description:"Credential's hash"`
		Usage    []*struct {
			Name   string `json:"name,omitempty" description:"Jenkins pipeline full name"`
			Ranges struct {
				Ranges []*struct {
					Start int `json:"start,omitempty" description:"Start build number"`
					End   int `json:"end,omitempty" description:"End build number"`
				} `json:"ranges,omitempty"`
			} `json:"ranges,omitempty" description:"The build number of all pipelines that use this credential"`
		} `json:"usage,omitempty" description:"all usage of Credential"`
	} `json:"fingerprint,omitempty" description:"usage of the Credential"`
	Description                string                              `json:"description,omitempty" description:"Credential's description'"`
	Domain                     string                              `json:"domain,omitempty" description:"Credential's domain,In ks we only use the default domain, default '_''"`
	CreateTime                 *time.Time                          `json:"create_time,omitempty" description:"Credential's create_time'"`
	Creator                    string                              `json:"creator,omitempty" description:"Creator's username"`
	UsernamePasswordCredential *UsernamePasswordCredential `json:"username_password,omitempty" description:"username password Credential struct"`
	SshCredential              *SshCredential              `json:"ssh,omitempty" description:"ssh Credential struct"`
	SecretTextCredential       *SecretTextCredential       `json:"secret_text,omitempty" description:"secret_text Credential struct"`
	KubeconfigCredential       *KubeconfigCredential       `json:"kubeconfig,omitempty" description:"kubeconfig Credential struct"`
}

type UsernamePasswordCredential struct {
	Username string `json:"username,omitempty" description:"username of username_password credential"`
	Password string `json:"password,omitempty" description:"password of username_password credential"`
}

type SshCredential struct {
	Username   string `json:"username,omitempty" description:"username of ssh credential"`
	Passphrase string `json:"passphrase,omitempty" description:"passphrase of ssh credential, password of ssh credential"`
	PrivateKey string `json:"private_key,omitempty" mapstructure:"private_key" description:"private key of ssh credential"`
}

type SecretTextCredential struct {
	Secret string `json:"secret,omitempty" description:"secret content of credential"`
}

type KubeconfigCredential struct {
	Content string `json:"content,omitempty" description:"content of kubeconfig"`
}

type Interface interface {
	SendJenkinsRequest(baseUrl string, req *http.Request) ([]byte, error)

	SendJenkinsRequestWithHeaderResp(baseUrl string, req *http.Request) ([]byte, http.Header, error)

	ValidateJenkinsfile(jenkinsfile string) (*ValidateJenkinsfileResponse, error)

	ValidatePipelineJson(json string) (*ValidatePipelineJsonResponse, error)

	PipelineJsonToJenkinsfile(json string) (*PipelineJsonToJenkinsfileResponse, error)

	JenkinsfileToPipelineJson(jenkinsfile string) (*JenkinsfileToPipelineJsonResponse, error)

	CreateFolder(name, description string, parents ...string) (*Folder, error)

	CreateJobInFolder(config string, jobName string, parentIDs ...string) (*Job, error)

	DeleteJob(name string, parentIDs ...string) (bool, error)

	BuildJob(name string, options ...interface{}) (int64, error)

	GetBuild(jobName string, number int64) (*Build, error)

	GetJob(id string, parentIDs ...string) (*Job, error)

	GetFolder(id string, parents ...string) (*Folder, error)

	CreateCredentialInProject(projectId string, credential *Credential) (*string, error)

	UpdateCredentialInProject(projectId string, credential *Credential) (*string, error)

	GetCredentialInProject(projectId, id string) (*Credential, error)

	GetCredentialContentInProject(projectId, id string) (string, error)

	GetCredentialsInProject(projectId string) ([]Credential, error)

	DeleteCredentialInFolder(projectId, id string) (*string, error)

	GetGlobalRole(roleName string) (*GlobalRole, error)

	GetProjectRole(roleName string) (*ProjectRole, error)

	AddGlobalRole(roleName string, ids GlobalPermissionIds, overwrite bool) (*GlobalRole, error)

	DeleteProjectRoles(roleName ...string) error

	AddProjectRole(roleName string, pattern string, ids ProjectPermissionIds, overwrite bool) (*ProjectRole, error)

	DeleteUserInProject(username string) error
}
