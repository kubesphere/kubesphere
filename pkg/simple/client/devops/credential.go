package devops

import (
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
			Name   string `json:"name,omitempty" description:"pipeline full name"`
			Ranges struct {
				Ranges []*struct {
					Start int `json:"start,omitempty" description:"Start build number"`
					End   int `json:"end,omitempty" description:"End build number"`
				} `json:"ranges,omitempty"`
			} `json:"ranges,omitempty" description:"The build number of all pipelines that use this credential"`
		} `json:"usage,omitempty" description:"all usage of Credential"`
	} `json:"fingerprint,omitempty" description:"usage of the Credential"`
	Description                string                      `json:"description,omitempty" description:"Credential's description'"`
	Domain                     string                      `json:"domain,omitempty" description:"Credential's domain,In ks we only use the default domain, default '_''"`
	CreateTime                 *time.Time                  `json:"create_time,omitempty" description:"Credential's create_time'"`
	Creator                    string                      `json:"creator,omitempty" description:"Creator's username"`
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

const (
	CredentialTypeUsernamePassword = "username_password"
	CredentialTypeSsh              = "ssh"
	CredentialTypeSecretText       = "secret_text"
	CredentialTypeKubeConfig       = "kubeconfig"
)

var CredentialTypeMap = map[string]string{
	"SSH Username with private key":         CredentialTypeSsh,
	"Username with password":                CredentialTypeUsernamePassword,
	"Secret text":                           CredentialTypeSecretText,
	"Kubernetes configuration (kubeconfig)": CredentialTypeKubeConfig,
}

type CredentialOperator interface {
	CreateCredentialInProject(projectId string, credential *Credential) (*string, error)

	UpdateCredentialInProject(projectId string, credential *Credential) (*string, error)

	GetCredentialInProject(projectId, id string, content bool) (*Credential, error)

	GetCredentialsInProject(projectId string) ([]*Credential, error)

	DeleteCredentialInProject(projectId, id string) (*string, error)
}
