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
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/gocraft/dbr"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"

	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/gojenkins/utils"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
)

type ProjectCredentialOperator interface {
	CreateProjectCredential(projectId, username string, credentialRequest *devops.Credential) (string, error)
	UpdateProjectCredential(projectId, credentialId string, credentialRequest *devops.Credential) (string, error)
	DeleteProjectCredential(projectId, credentialId string) (string, error)
	GetProjectCredential(projectId, credentialId, getContent string) (*devops.Credential, error)
	GetProjectCredentials(projectId string) ([]*devops.Credential, error)
}

type projectCredentialOperator struct {
	devopsClient devops.Interface
}

func newProjectCredentialOperator(client jenkins.Client) ProjectCredentialOperator {
	return &projectCredentialOperator{}
}

func (o *projectCredentialOperator) CreateProjectCredential(projectId, username string, credentialRequest *devops.Credential) (string, error) {
	switch credentialRequest.Type {
	case devops.CredentialTypeUsernamePassword:
		if credentialRequest.UsernamePasswordCredential == nil {
			err := fmt.Errorf("usename_password should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	case devops.CredentialTypeSsh:
		if credentialRequest.SshCredential == nil {
			err := fmt.Errorf("ssh should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	case devops.CredentialTypeSecretText:
		if credentialRequest.SecretTextCredential == nil {
			err := fmt.Errorf("secret_text should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	case devops.CredentialTypeKubeConfig:
		if credentialRequest.KubeconfigCredential == nil {
			err := fmt.Errorf("kubeconfig should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	default:
		err := fmt.Errorf("error unsupport credential type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())

	}
	credentialId, err := o.devopsClient.CreateCredentialInProject(projectId, credentialRequest)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = o.insertCredentialToDb(projectId, *credentialId, credentialRequest.Domain, username)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", err
	}
	return *credentialId, nil

}

func (o *projectCredentialOperator) UpdateProjectCredential(projectId, credentialId string, credentialRequest *devops.Credential) (string, error) {

	credential, err := o.devopsClient.GetCredentialInProject(projectId,
		credentialId, false)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	switch credential.Type {
	case devops.CredentialTypeUsernamePassword:
		if credentialRequest.UsernamePasswordCredential == nil {
			err := fmt.Errorf("usename_password should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	case devops.CredentialTypeSsh:
		if credentialRequest.SshCredential == nil {
			err := fmt.Errorf("ssh should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	case devops.CredentialTypeSecretText:
		if credentialRequest.SecretTextCredential == nil {
			err := fmt.Errorf("secret_text should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	case devops.CredentialTypeKubeConfig:
		if credentialRequest.KubeconfigCredential == nil {
			err := fmt.Errorf("kubeconfig should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	default:
		err := fmt.Errorf("error unsupport credential type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}
	credentialRequest.Id = credentialId
	_, err = o.devopsClient.UpdateCredentialInProject(projectId, credentialRequest)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())
	}

}

func (o *projectCredentialOperator) DeleteProjectCredential(projectId, credentialId string) (string, error) {

	dbClient, err := cs.ClientSets().MySQL()
	if err != nil {
		return "", restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	_, err = o.devopsClient.GetCredentialInProject(projectId,
		credentialId, false)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	id, err := o.devopsClient.DeleteCredentialInProject(projectId, credentialId)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	deleteConditions := append(make([]dbr.Builder, 0), db.Eq(ProjectCredentialProjectIdColumn, projectId))
	deleteConditions = append(deleteConditions, db.Eq(ProjectCredentialIdColumn, credentialId))
	deleteConditions = append(deleteConditions, db.Eq(ProjectCredentialDomainColumn, "_"))

	_, err = dbClient.DeleteFrom(ProjectCredentialTableName).
		Where(db.And(deleteConditions...)).Exec()
	if err != nil && err != db.ErrNotFound {
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return *id, nil

}

func (o *projectCredentialOperator) GetProjectCredential(projectId, credentialId, getContent string) (*devops.Credential, error) {

	dbClient, err := cs.ClientSets().MySQL()
	content := false
	if getContent != "" {
		content = true
	}
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	credential, err := o.devopsClient.GetCredentialInProject(projectId,
		credentialId,
		content)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	projectCredential := &ProjectCredential{}
	err = dbClient.Select(ProjectCredentialColumns...).
		From(ProjectCredentialTableName).Where(
		db.And(db.Eq(ProjectCredentialProjectIdColumn, projectId),
			db.Eq(ProjectCredentialIdColumn, credentialId),
			db.Eq(ProjectCredentialDomainColumn, credential.Domain))).LoadOne(projectCredential)

	if err != nil && err != db.ErrNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	response := formatCredentialResponse(credential, projectCredential)
	return response, nil

}

func (o *projectCredentialOperator) GetProjectCredentials(projectId string) ([]*devops.Credential, error) {

	dbClient, err := cs.ClientSets().MySQL()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	credentialResponses, err := o.devopsClient.GetCredentialsInProject(projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	selectCondition := db.Eq(ProjectCredentialProjectIdColumn, projectId)
	projectCredentials := make([]*ProjectCredential, 0)
	_, err = dbClient.Select(ProjectCredentialColumns...).
		From(ProjectCredentialTableName).Where(selectCondition).Load(&projectCredentials)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	response := formatCredentialsResponse(credentialResponses, projectCredentials)
	return response, nil
}

func (o *projectCredentialOperator) insertCredentialToDb(projectId, credentialId, domain, username string) error {
	dbClient, err := cs.ClientSets().MySQL()
	if err != nil {
		return err
	}

	projectCredential := NewProjectCredential(projectId, credentialId, domain, username)
	_, err = dbClient.InsertInto(ProjectCredentialTableName).Columns(ProjectCredentialColumns...).
		Record(projectCredential).Exec()
	if err != nil {
		klog.Errorf("%+v", err)
		return restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func formatCredentialResponse(
	credentialResponse *devops.Credential,
	dbCredentialResponse *ProjectCredential) *devops.Credential {
	response := &devops.Credential{}
	response.Id = credentialResponse.Id
	response.Description = credentialResponse.Description
	response.DisplayName = credentialResponse.DisplayName
	if credentialResponse.Fingerprint != nil && credentialResponse.Fingerprint.Hash != "" {
		response.Fingerprint = &struct {
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
		}{}
		response.Fingerprint.FileName = credentialResponse.Fingerprint.FileName
		response.Fingerprint.Hash = credentialResponse.Fingerprint.Hash
		for _, usage := range credentialResponse.Fingerprint.Usage {
			response.Fingerprint.Usage = append(response.Fingerprint.Usage, usage)
		}
	}
	response.Domain = credentialResponse.Domain

	if dbCredentialResponse != nil {
		response.CreateTime = &dbCredentialResponse.CreateTime
		response.Creator = dbCredentialResponse.Creator
	}

	credentialType, ok := devops.CredentialTypeMap[credentialResponse.Type]
	if ok {
		response.Type = credentialType
		return response
	}
	response.Type = credentialResponse.Type
	return response
}

func formatCredentialsResponse(credentialsResponse []*devops.Credential,
	projectCredentials []*ProjectCredential) []*devops.Credential {
	responseSlice := make([]*devops.Credential, 0)
	for _, credential := range credentialsResponse {
		var dbCredential *ProjectCredential = nil
		for _, projectCredential := range projectCredentials {
			if projectCredential.CredentialId == credential.Id &&
				projectCredential.Domain == credential.Domain {
				dbCredential = projectCredential
			}
		}
		responseSlice = append(responseSlice, formatCredentialResponse(credential, dbCredential))
	}
	return responseSlice
}
