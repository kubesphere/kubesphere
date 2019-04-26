package devops

import (
	"github.com/asaskevich/govalidator"
	"time"
)

const (
	CredentialTypeUsernamePassword = "username_password"
	CredentialTypeSsh              = "ssh"
	CredentialTypeSecretText       = "secret_text"
	CredentialTypeKubeConfig       = "kubeconfig"
)

type JenkinsCredential struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	Fingerprint *struct {
		FileName string `json:"file_name,omitempty"`
		Hash     string `json:"hash,omitempty"`
		Usage    []*struct {
			Name   string `json:"name,omitempty"`
			Ranges struct {
				Ranges []*struct {
					Start int `json:"start"`
					End   int `json:"end"`
				} `json:"ranges"`
			} `json:"ranges"`
		} `json:"usage,omitempty"`
	} `json:"fingerprint,omitempty"`
	Description                string                      `json:"description"`
	Domain                     string                      `json:"domain"`
	CreateTime                 *time.Time                  `json:"create_time,omitempty"`
	Creator                    string                      `json:"creator,omitempty"`
	UsernamePasswordCredential *UsernamePasswordCredential `json:"username_password"`
	SshCredential              *SshCredential              `json:"ssh"`
	SecretTextCredential       *SecretTextCredential       `json:"secret_text"`
	KubeconfigCredential       *KubeconfigCredential       `json:"kubeconfig"`
}

type UsernamePasswordCredential struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"`
	Description string `json:"description"`
}

type SshCredential struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	Passphrase  string `json:"passphrase"`
	PrivateKey  string `json:"private_key" mapstructure:"private_key"`
	Description string `json:"description"`
}

type SecretTextCredential struct {
	Id          string `json:"id"`
	Secret      string `json:"secret"`
	Description string `json:"description"`
}

type KubeconfigCredential struct {
	Id          string `json:"id"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

type DeleteCredentialRequest struct {
	Domain string `json:"domain"`
}

type CopySshCredentialRequest struct {
	Id string `json:"id"`
}

const (
	ProjectCredentialTableName       = "project_credential"
	ProjectCredentialIdColumn        = "credential_id"
	ProjectCredentialDomainColumn    = "domain"
	ProjectCredentialProjectIdColumn = "project_id"
)

var CredentialTypeMap = map[string]string{
	"SSH Username with private key":         CredentialTypeSsh,
	"Username with password":                CredentialTypeUsernamePassword,
	"Secret text":                           CredentialTypeSecretText,
	"Kubernetes configuration (kubeconfig)": CredentialTypeKubeConfig,
}

type ProjectCredential struct {
	ProjectId    string    `json:"project_id"`
	CredentialId string    `json:"credential_id"`
	Domain       string    `json:"domain"`
	Creator      string    `json:"creator"`
	CreateTime   time.Time `json:"create_time"`
}

var ProjectCredentialColumns = GetColumnsFromStruct(&ProjectCredential{})

func NewProjectCredential(projectId, credentialId, domain, creator string) *ProjectCredential {
	if govalidator.IsNull(domain) {
		domain = "_"
	}
	return &ProjectCredential{
		ProjectId:    projectId,
		CredentialId: credentialId,
		Domain:       domain,
		Creator:      creator,
		CreateTime:   time.Now(),
	}
}
