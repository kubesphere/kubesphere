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
	Id          string `json:"id"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name,omitempty"`
	Fingerprint *struct {
		FileName string `json:"file_name,omitempty"`
		Hash     string `json:"hash,omitempty"`
		Usage    []*struct {
			Name   string `json:"name,omitempty"`
			Ranges struct {
				Ranges []*struct {
					Start int `json:"start,omitempty"`
					End   int `json:"end,omitempty"`
				} `json:"ranges,omitempty"`
			} `json:"ranges,omitempty"`
		} `json:"usage,omitempty"`
	} `json:"fingerprint,omitempty"`
	Description                string                      `json:"description,omitempty"`
	Domain                     string                      `json:"domain,omitempty"`
	CreateTime                 *time.Time                  `json:"create_time,omitempty"`
	Creator                    string                      `json:"creator,omitempty"`
	UsernamePasswordCredential *UsernamePasswordCredential `json:"username_password,omitempty"`
	SshCredential              *SshCredential              `json:"ssh,omitempty"`
	SecretTextCredential       *SecretTextCredential       `json:"secret_text,omitempty"`
	KubeconfigCredential       *KubeconfigCredential       `json:"kubeconfig,omitempty"`
}

type UsernamePasswordCredential struct {
	Id          string `json:"id"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Description string `json:"description,omitempty"`
}

type SshCredential struct {
	Id          string `json:"id"`
	Username    string `json:"username,omitempty"`
	Passphrase  string `json:"passphrase,omitempty"`
	PrivateKey  string `json:"private_key,omitempty" mapstructure:"private_key"`
	Description string `json:"description,omitempty"`
}

type SecretTextCredential struct {
	Id          string `json:"id"`
	Secret      string `json:"secret,omitempty"`
	Description string `json:"description,omitempty"`
}

type KubeconfigCredential struct {
	Id          string `json:"id"`
	Content     string `json:"content,omitempty"`
	Description string `json:"description,omitempty"`
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
