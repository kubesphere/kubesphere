/*
Copyright 2020 KubeSphere Authors

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
	v1 "k8s.io/api/core/v1"
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
	Description string `json:"description,omitempty" description:"Credential's description'"`
	Domain      string `json:"domain,omitempty" description:"Credential's domain,In ks we only use the default domain, default '_''"`
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

type CredentialOperator interface {
	CreateCredentialInProject(projectId string, credential *v1.Secret) (string, error)

	UpdateCredentialInProject(projectId string, credential *v1.Secret) (string, error)

	GetCredentialInProject(projectId, id string) (*Credential, error)

	DeleteCredentialInProject(projectId, id string) (string, error)
}
