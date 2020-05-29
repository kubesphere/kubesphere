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
	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"strconv"
)

const SSHCrenditalStaplerClass = "com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey"
const DirectSSHCrenditalStaplerClass = "com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey$DirectEntryPrivateKeySource"
const UsernamePassswordCredentialStaplerClass = "com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl"
const SecretTextCredentialStaplerClass = "org.jenkinsci.plugins.plaincredentials.impl.StringCredentialsImpl"
const KubeconfigCredentialStaplerClass = "com.microsoft.jenkins.kubernetes.credentials.KubeconfigCredentials"
const DirectKubeconfigCredentialStaperClass = "com.microsoft.jenkins.kubernetes.credentials.KubeconfigCredentials$DirectEntryKubeconfigSource"
const GLOBALScope = "GLOBAL"

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

func NewSshCredential(secret *v1.Secret) *SshCredential {
	id := secret.Name
	username := string(secret.Data[devopsv1alpha3.SSHAuthUsernameKey])
	passphrase := string(secret.Data[devopsv1alpha3.SSHAuthPassphraseKey])
	privatekey := string(secret.Data[devopsv1alpha3.SSHAuthPrivateKey])

	keySource := PrivateKeySource{
		StaplerClass: DirectSSHCrenditalStaplerClass,
		PrivateKey:   privatekey,
	}

	return &SshCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Username:     username,
		Passphrase:   passphrase,
		KeySource:    keySource,
		StaplerClass: SSHCrenditalStaplerClass,
	}
}

func NewUsernamePasswordCredential(secret *v1.Secret) *UsernamePasswordCredential {
	id := secret.Name
	username := string(secret.Data[devopsv1alpha3.BasicAuthUsernameKey])
	password := string(secret.Data[devopsv1alpha3.BasicAuthPasswordKey])
	return &UsernamePasswordCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Username:     username,
		Password:     password,
		StaplerClass: UsernamePassswordCredentialStaplerClass,
	}
}

func NewSecretTextCredential(secret *v1.Secret) *SecretTextCredential {
	id := secret.Name
	secretContent := string(secret.Data[devopsv1alpha3.SecretTextSecretKey])
	return &SecretTextCredential{
		Scope:        GLOBALScope,
		Id:           id,
		Secret:       secretContent,
		StaplerClass: SecretTextCredentialStaplerClass,
	}
}

func NewKubeconfigCredential(secret *v1.Secret) *KubeconfigCredential {
	id := secret.Name
	secretContent := string(secret.Data[devopsv1alpha3.KubeConfigSecretKey])

	credentialSource := KubeconfigSource{
		StaplerClass: DirectKubeconfigCredentialStaperClass,
		Content:      secretContent,
	}

	return &KubeconfigCredential{
		Scope:            GLOBALScope,
		Id:               id,
		KubeconfigSource: credentialSource,
		StaplerClass:     KubeconfigCredentialStaplerClass,
	}
}

func (j *Jenkins) GetCredentialInProject(projectId, id string) (*devops.Credential, error) {
	responseStruct := &devops.Credential{}

	domain := "_"

	response, err := j.Requester.GetJSON(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/_/credential/%s", projectId, id),
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
	return responseStruct, nil
}

func (j *Jenkins) GetCredentialsInProject(projectId string) ([]*devops.Credential, error) {
	domain := "_"
	var responseStruct = &struct {
		Credentials []*devops.Credential `json:"credentials"`
	}{}
	response, err := j.Requester.GetJSON(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/_", projectId),
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

func (j *Jenkins) CreateCredentialInProject(projectId string, credential *v1.Secret) (string, error) {

	var request interface{}
	responseString := ""
	switch credential.Type {
	case devopsv1alpha3.SecretTypeBasicAuth:
		request = NewUsernamePasswordCredential(credential)
	case devopsv1alpha3.SecretTypeSSHAuth:
		request = NewSshCredential(credential)
	case devopsv1alpha3.SecretTypeSecretText:
		request = NewSecretTextCredential(credential)
	case devopsv1alpha3.SecretTypeKubeConfig:
		request = NewKubeconfigCredential(credential)
	default:
		err := fmt.Errorf("error unsupport credential type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}

	response, err := j.Requester.Post(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/_/createCredentials", projectId),
		nil, &responseString, map[string]string{
			"json": makeJson(map[string]interface{}{
				"credentials": request,
			}),
		})
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.New(strconv.Itoa(response.StatusCode))
	}
	return credential.Name, nil
}

func (j *Jenkins) UpdateCredentialInProject(projectId string, credential *v1.Secret) (string, error) {

	requestContent := ""
	switch credential.Type {
	case devopsv1alpha3.SecretTypeBasicAuth:
		requestStruct := NewUsernamePasswordCredential(credential)
		requestContent = makeJson(requestStruct)
	case devopsv1alpha3.SecretTypeSSHAuth:
		requestStruct := NewSshCredential(credential)
		requestContent = makeJson(requestStruct)
	case devopsv1alpha3.SecretTypeSecretText:
		requestStruct := NewSecretTextCredential(credential)
		requestContent = makeJson(requestStruct)
	case devopsv1alpha3.SecretTypeKubeConfig:
		requestStruct := NewKubeconfigCredential(credential)
		requestContent = makeJson(requestStruct)
	default:
		err := fmt.Errorf("error unsupport credential type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
	response, err := j.Requester.Post(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/_/credential/%s/updateSubmit", projectId, credential.Name),
		nil, nil, map[string]string{
			"json": requestContent,
		})
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", errors.New(strconv.Itoa(response.StatusCode))
	}
	return credential.Name, nil
}

func (j *Jenkins) DeleteCredentialInProject(projectId, id string) (string, error) {
	response, err := j.Requester.Post(
		fmt.Sprintf("/job/%s/credentials/store/folder/domain/_/credential/%s/doDelete", projectId, id),
		nil, nil, nil)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return "", errors.New(strconv.Itoa(response.StatusCode))
	}
	return id, nil
}
