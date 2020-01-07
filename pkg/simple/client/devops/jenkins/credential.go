/*
Copyright 2018 The KubeSphere Authors.
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

package jenkins

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"strconv"
	"strings"
)

const SSHCrenditalStaplerClass = "com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey"
const DirectSSHCrenditalStaplerClass = "com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey$DirectEntryPrivateKeySource"
const UsernamePassswordCredentialStaplerClass = "com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl"
const SecretTextCredentialStaplerClass = "org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl"
const KubeconfigCredentialStaplerClass = "com.microsoft.jenkins.kubernetes.credentials.KubeconfigCredentials"
const DirectKubeconfigCredentialStaperClass = "com.microsoft.jenkins.kubernetes.credentials.KubeconfigCredentials$DirectEntryKubeconfigSource"
const GLOBALScope = "GLOBAL"

type CreateSshCredentialRequest struct {
	Credentials SshCredential `json:"credentials"`
}

type CreateUsernamePasswordCredentialRequest struct {
	Credentials UsernamePasswordCredential `json:"credentials"`
}

type CreateSecretTextCredentialRequest struct {
	Credentials SecretTextCredential `json:"credentials"`
}

type CreateKubeconfigCredentialRequest struct {
	Credentials KubeconfigCredential `json:"credentials"`
}

type UsernamePasswordCredential struct {
	Scope        string `json:"scope"`
	Id           string `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Description  string `json:"description"`
	StaplerClass string `json:"stapler-class"`
}

type SshCredential struct {
	Scope        string           `json:"scope"`
	Id           string           `json:"id"`
	Username     string           `json:"username"`
	Passphrase   string           `json:"passphrase"`
	KeySource    PrivateKeySource `json:"privateKeySource"`
	Description  string           `json:"description"`
	StaplerClass string           `json:"stapler-class"`
}

type SecretTextCredential struct {
	Scope        string `json:"scope"`
	Id           string `json:"id"`
	Secret       string `json:"secret"`
	Description  string `json:"description"`
	StaplerClass string `json:"stapler-class"`
}

type KubeconfigCredential struct {
	Scope            string           `json:"scope"`
	Id               string           `json:"id"`
	Description      string           `json:"description"`
	KubeconfigSource KubeconfigSource `json:"kubeconfigSource"`
	StaplerClass     string           `json:"stapler-class"`
}

type PrivateKeySource struct {
	StaplerClass string `json:"stapler-class"`
	PrivateKey   string `json:"privateKey"`
}

type KubeconfigSource struct {
	StaplerClass string `json:"stapler-class"`
	Content      string `json:"content"`
}

type CredentialResponse struct {
	Id          string `json:"id"`
	TypeName    string `json:"typeName"`
	DisplayName string `json:"displayName"`
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
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain"`
}

func NewCreateSshCredentialRequest(id, username, passphrase, privateKey, description string) *CreateSshCredentialRequest {

	keySource := PrivateKeySource{
		StaplerClass: DirectSSHCrenditalStaplerClass,
		PrivateKey:   privateKey,
	}

	sshCredential := SshCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Username:     username,
		Passphrase:   passphrase,
		KeySource:    keySource,
		Description:  description,
		StaplerClass: SSHCrenditalStaplerClass,
	}
	return &CreateSshCredentialRequest{
		Credentials: sshCredential,
	}

}

func NewCreateUsernamePasswordRequest(id, username, password, description string) *CreateUsernamePasswordCredentialRequest {
	credential := UsernamePasswordCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Username:     username,
		Password:     password,
		Description:  description,
		StaplerClass: UsernamePassswordCredentialStaplerClass,
	}
	return &CreateUsernamePasswordCredentialRequest{
		Credentials: credential,
	}
}

func NewCreateSecretTextCredentialRequest(id, secret, description string) *CreateSecretTextCredentialRequest {
	credential := SecretTextCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Secret:       secret,
		Description:  description,
		StaplerClass: SecretTextCredentialStaplerClass,
	}
	return &CreateSecretTextCredentialRequest{
		Credentials: credential,
	}
}

func NewCreateKubeconfigCredentialRequest(id, content, description string) *CreateKubeconfigCredentialRequest {

	credentialSource := KubeconfigSource{
		StaplerClass: DirectKubeconfigCredentialStaperClass,
		Content:      content,
	}

	credential := KubeconfigCredential{
		Scope:            GLOBALScope,
		Id:               id,
		Description:      description,
		KubeconfigSource: credentialSource,
		StaplerClass:     KubeconfigCredentialStaplerClass,
	}
	return &CreateKubeconfigCredentialRequest{
		credential,
	}
}

func NewSshCredential(id, username, passphrase, privateKey, description string) *SshCredential {
	keySource := PrivateKeySource{
		StaplerClass: DirectSSHCrenditalStaplerClass,
		PrivateKey:   privateKey,
	}

	return &SshCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Username:     username,
		Passphrase:   passphrase,
		KeySource:    keySource,
		Description:  description,
		StaplerClass: SSHCrenditalStaplerClass,
	}
}

func NewUsernamePasswordCredential(id, username, password, description string) *UsernamePasswordCredential {
	return &UsernamePasswordCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Username:     username,
		Password:     password,
		Description:  description,
		StaplerClass: UsernamePassswordCredentialStaplerClass,
	}
}

func NewSecretTextCredential(id, secret, description string) *SecretTextCredential {
	return &SecretTextCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Secret:       secret,
		Description:  description,
		StaplerClass: SecretTextCredentialStaplerClass,
	}
}

func NewKubeconfigCredential(id, content, description string) *KubeconfigCredential {
	credentialSource := KubeconfigSource{
		StaplerClass: DirectKubeconfigCredentialStaperClass,
		Content:      content,
	}

	return &KubeconfigCredential{
		Scope:            GLOBALScope,
		Id:               id,
		Description:      description,
		KubeconfigSource: credentialSource,
		StaplerClass:     KubeconfigCredentialStaplerClass,
	}
}


func (j *Jenkins) GetCredentialInProject(projectId, id string, content bool) (*devops.Credential, error) {
	responseStruct := &devops.Credential{}

	domain := "_"

	response, err := j.Requester.GetJSON(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/%s/credential/%s",projectId, domain, id),
		responseStruct, map[string]string{
			"depth": "2",
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	responseStruct.Domain = domain
	if content {

	}
	contentString := ""
	response, err = j.Requester.GetHtml(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/%s/credential/%s/update", projectId, domain, id),
		&contentString, nil)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	stringReader := strings.NewReader(contentString)
	doc, err := goquery.NewDocumentFromReader(stringReader)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	switch responseStruct.Type {
	case devops.CredentialTypeKubeConfig:
		content := &devops.KubeconfigCredential{}
		doc.Find("textarea[name*=content]").Each(func(i int, selection *goquery.Selection) {
			value := selection.Text()
			content.Content = value
		})
		responseStruct.KubeconfigCredential = content
	case devops.CredentialTypeUsernamePassword:
		content := &devops.UsernamePasswordCredential{}
		doc.Find("input[name*=username]").Each(func(i int, selection *goquery.Selection) {
			value, _ := selection.Attr("value")
			content.Username = value
		})

		responseStruct.UsernamePasswordCredential = content
	case devops.CredentialTypeSsh:
		content := &devops.SshCredential{}
		doc.Find("input[name*=username]").Each(func(i int, selection *goquery.Selection) {
			value, _ := selection.Attr("value")
			content.Username = value
		})

		doc.Find("textarea[name*=privateKey]").Each(func(i int, selection *goquery.Selection) {
			value := selection.Text()
			content.PrivateKey = value
		})
		responseStruct.SshCredential = content
	}
	return responseStruct, nil
}


func (j *Jenkins) GetCredentialsInProject(projectId string) ([]* devops.Credential, error) {
	domain := "_"
	var responseStruct = &struct {
		Credentials []*devops.Credential `json:"credentials"`
	}{}
	response, err := j.Requester.GetJSON(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/%s", projectId, domain),
		responseStruct, map[string]string{
			"depth": "2",
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	for _, credential := range responseStruct.Credentials {
		credential.Domain = domain
	}
	return responseStruct.Credentials, nil

}
