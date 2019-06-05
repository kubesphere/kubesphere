/*
Copyright 2019 The KubeSphere Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	Id          string `json:"id" description:"id of credential"`
	Type        string `json:"type" description:"type of credential,such as ssh/kubeconfig"`
	DisplayName string `json:"display_name,omitempty" description:"credential's display name'"`
	Fingerprint *struct {
		FileName string `json:"file_name,omitempty" description:"credential's display name and description"`
		Hash     string `json:"hash,omitempty" description:"credential's hash'"`
		Usage    []*struct {
			Name   string `json:"name,omitempty" description:"jenkins pipeline full name"`
			Ranges struct {
				Ranges []*struct {
					Start int `json:"start,omitempty" description:"start build number"`
					End   int `json:"end,omitempty" description:"end build number"`
				} `json:"ranges,omitempty"`
			} `json:"ranges,omitempty" description:"all build num using credential"`
		} `json:"usage,omitempty" description:"all usage of credential"`
	} `json:"fingerprint,omitempty" description:"usage of credential"`
	Description                string                      `json:"description,omitempty" description:"credential's description'"`
	Domain                     string                      `json:"domain,omitempty" description:"credential's domain, default '_''"`
	CreateTime                 *time.Time                  `json:"create_time,omitempty" description:"credential's create_time'"`
	Creator                    string                      `json:"creator,omitempty" description:"creator's username"`
	UsernamePasswordCredential *UsernamePasswordCredential `json:"username_password,omitempty" description:"username password credential struct"`
	SshCredential              *SshCredential              `json:"ssh,omitempty" description:"ssh credential struct"`
	SecretTextCredential       *SecretTextCredential       `json:"secret_text,omitempty" description:"secret_text credential struct"`
	KubeconfigCredential       *KubeconfigCredential       `json:"kubeconfig,omitempty" description:"kubeconfig credential struct"`
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
