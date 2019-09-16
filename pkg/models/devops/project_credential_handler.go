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
	"github.com/PuerkitoBio/goquery"
	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful"
	"github.com/gocraft/dbr"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"kubesphere.io/kubesphere/pkg/gojenkins/utils"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
	"strings"
)

func CreateProjectCredential(projectId, username string, credentialRequest *JenkinsCredential) (string, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return "", restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	jenkinsClient := devops.Jenkins()

	err = checkJenkinsCredentialExists(projectId, credentialRequest.Domain, credentialRequest.Id)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", err
	}

	switch credentialRequest.Type {
	case CredentialTypeUsernamePassword:
		if credentialRequest.UsernamePasswordCredential == nil {
			err := fmt.Errorf("usename_password should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
		credentialId, err := jenkinsClient.CreateUsernamePasswordCredentialInFolder(credentialRequest.Domain,
			credentialRequest.Id,
			credentialRequest.UsernamePasswordCredential.Username,
			credentialRequest.UsernamePasswordCredential.Password,
			credentialRequest.Description,
			projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		err = insertCredentialToDb(projectId, *credentialId, credentialRequest.Domain, username)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", err
		}
		return *credentialId, nil
	case CredentialTypeSsh:
		if credentialRequest.SshCredential == nil {
			err := fmt.Errorf("ssh should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
		credentialId, err := jenkinsClient.CreateSshCredentialInFolder(credentialRequest.Domain,
			credentialRequest.Id,
			credentialRequest.SshCredential.Username,
			credentialRequest.SshCredential.Passphrase,
			credentialRequest.SshCredential.PrivateKey,
			credentialRequest.Description,
			projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		err = insertCredentialToDb(projectId, *credentialId, credentialRequest.Domain, username)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		return *credentialId, nil
	case CredentialTypeSecretText:
		if credentialRequest.SecretTextCredential == nil {
			err := fmt.Errorf("secret_text should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}

		credentialId, err := jenkinsClient.CreateSecretTextCredentialInFolder(credentialRequest.Domain,
			credentialRequest.Id,
			credentialRequest.SecretTextCredential.Secret,
			credentialRequest.Description,
			projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		err = insertCredentialToDb(projectId, *credentialId, credentialRequest.Domain, username)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		return *credentialId, nil
	case CredentialTypeKubeConfig:
		if credentialRequest.KubeconfigCredential == nil {
			err := fmt.Errorf("kubeconfig should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
		credentialId, err := jenkinsClient.CreateKubeconfigCredentialInFolder(credentialRequest.Domain,
			credentialRequest.Id,
			credentialRequest.KubeconfigCredential.Content,
			credentialRequest.Description,
			projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		err = insertCredentialToDb(projectId, *credentialId, credentialRequest.Domain, username)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		return *credentialId, nil
	default:
		err := fmt.Errorf("error unsupport credential type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())

	}

}

func UpdateProjectCredential(projectId, credentialId string, credentialRequest *JenkinsCredential) (string, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return "", restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	jenkinsCredential, err := jenkinsClient.GetCredentialInFolder(credentialRequest.Domain,
		credentialId,
		projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	credentialType := CredentialTypeMap[jenkinsCredential.TypeName]
	switch credentialType {
	case CredentialTypeUsernamePassword:
		if credentialRequest.UsernamePasswordCredential == nil {
			err := fmt.Errorf("usename_password should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
		credentialId, err := jenkinsClient.UpdateUsernamePasswordCredentialInFolder(credentialRequest.Domain,
			credentialId,
			credentialRequest.UsernamePasswordCredential.Username,
			credentialRequest.UsernamePasswordCredential.Password,
			credentialRequest.Description,
			projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		return *credentialId, nil
	case CredentialTypeSsh:
		if credentialRequest.SshCredential == nil {
			err := fmt.Errorf("ssh should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
		credentialId, err := jenkinsClient.UpdateSshCredentialInFolder(credentialRequest.Domain,
			credentialId,
			credentialRequest.SshCredential.Username,
			credentialRequest.SshCredential.Passphrase,
			credentialRequest.SshCredential.PrivateKey,
			credentialRequest.Description,
			projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}

		return *credentialId, nil
	case CredentialTypeSecretText:
		if credentialRequest.SecretTextCredential == nil {
			err := fmt.Errorf("secret_text should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
		credentialId, err := jenkinsClient.UpdateSecretTextCredentialInFolder(credentialRequest.Domain,
			credentialId,
			credentialRequest.SecretTextCredential.Secret,
			credentialRequest.Description,
			projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		return *credentialId, nil
	case CredentialTypeKubeConfig:
		if credentialRequest.KubeconfigCredential == nil {
			err := fmt.Errorf("kubeconfig should not be nil")
			klog.Error(err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
		credentialId, err := jenkinsClient.UpdateKubeconfigCredentialInFolder(credentialRequest.Domain,
			credentialId,
			credentialRequest.KubeconfigCredential.Content,
			credentialRequest.Description,
			projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		return *credentialId, nil
	default:
		err := fmt.Errorf("error unsupport credential type")
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusBadRequest, err.Error())

	}

}

func DeleteProjectCredential(projectId, credentialId string, credentialRequest *JenkinsCredential) (string, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return "", restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	dbClient, err := cs.ClientSets().MySQL()
	if err != nil {
		return "", restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	_, err = jenkinsClient.GetCredentialInFolder(credentialRequest.Domain,
		credentialId,
		projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	id, err := jenkinsClient.DeleteCredentialInFolder(credentialRequest.Domain, credentialId, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	deleteConditions := append(make([]dbr.Builder, 0), db.Eq(ProjectCredentialProjectIdColumn, projectId))
	deleteConditions = append(deleteConditions, db.Eq(ProjectCredentialIdColumn, credentialId))
	if !govalidator.IsNull(credentialRequest.Domain) {
		deleteConditions = append(deleteConditions, db.Eq(ProjectCredentialDomainColumn, credentialRequest.Domain))
	} else {
		deleteConditions = append(deleteConditions, db.Eq(ProjectCredentialDomainColumn, "_"))
	}

	_, err = dbClient.DeleteFrom(ProjectCredentialTableName).
		Where(db.And(deleteConditions...)).Exec()
	if err != nil && err != db.ErrNotFound {
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return *id, nil

}

func GetProjectCredential(projectId, credentialId, domain, getContent string) (*JenkinsCredential, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	dbClient, err := cs.ClientSets().MySQL()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsResponse, err := jenkinsClient.GetCredentialInFolder(domain,
		credentialId,
		projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	projectCredential := &ProjectCredential{}
	err = dbClient.Select(ProjectCredentialColumns...).
		From(ProjectCredentialTableName).Where(
		db.And(db.Eq(ProjectCredentialProjectIdColumn, projectId),
			db.Eq(ProjectCredentialIdColumn, credentialId),
			db.Eq(ProjectCredentialDomainColumn, jenkinsResponse.Domain))).LoadOne(projectCredential)

	if err != nil && err != db.ErrNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	response := formatCredentialResponse(jenkinsResponse, projectCredential)
	if getContent != "" {
		stringBody, err := jenkinsClient.GetCredentialContentInFolder(jenkinsResponse.Domain, credentialId, projectId)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		stringReader := strings.NewReader(stringBody)
		doc, err := goquery.NewDocumentFromReader(stringReader)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusInternalServerError, err.Error())
		}
		switch response.Type {
		case CredentialTypeKubeConfig:
			content := &KubeconfigCredential{}
			doc.Find("textarea[name*=content]").Each(func(i int, selection *goquery.Selection) {
				value := selection.Text()
				content.Content = value
			})
			response.KubeconfigCredential = content
		case CredentialTypeUsernamePassword:
			content := &UsernamePasswordCredential{}
			doc.Find("input[name*=username]").Each(func(i int, selection *goquery.Selection) {
				value, _ := selection.Attr("value")
				content.Username = value
			})

			response.UsernamePasswordCredential = content
		case CredentialTypeSsh:
			content := &SshCredential{}
			doc.Find("input[name*=username]").Each(func(i int, selection *goquery.Selection) {
				value, _ := selection.Attr("value")
				content.Username = value
			})

			doc.Find("textarea[name*=privateKey]").Each(func(i int, selection *goquery.Selection) {
				value := selection.Text()
				content.PrivateKey = value
			})
			response.SshCredential = content
		}
	}
	return response, nil

}

func GetProjectCredentials(projectId, domain string) ([]*JenkinsCredential, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	dbClient, err := cs.ClientSets().MySQL()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsCredentialResponses, err := jenkinsClient.GetCredentialsInFolder(domain, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	selectCondition := db.Eq(ProjectCredentialProjectIdColumn, projectId)
	if !govalidator.IsNull(domain) {
		selectCondition = db.And(selectCondition, db.Eq(ProjectCredentialDomainColumn, domain))
	}
	projectCredentials := make([]*ProjectCredential, 0)
	_, err = dbClient.Select(ProjectCredentialColumns...).
		From(ProjectCredentialTableName).Where(selectCondition).Load(&projectCredentials)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	response := formatCredentialsResponse(jenkinsCredentialResponses, projectCredentials)
	return response, nil
}

func insertCredentialToDb(projectId, credentialId, domain, username string) error {
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

func checkJenkinsCredentialExists(projectId, domain, credentialId string) error {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return restful.NewError(http.StatusServiceUnavailable, err.Error())
	}

	jenkinsClient := devops.Jenkins()

	credential, err := jenkinsClient.GetCredentialInFolder(domain, credentialId, projectId)
	if credential != nil {
		err := fmt.Errorf("credential id [%s] has been used", credential.Id)
		klog.Warning(err.Error())
		return restful.NewError(http.StatusConflict, err.Error())
	}
	if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)

		return restful.NewError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func formatCredentialResponse(
	jenkinsCredentialResponse *gojenkins.CredentialResponse,
	dbCredentialResponse *ProjectCredential) *JenkinsCredential {
	response := &JenkinsCredential{}
	response.Id = jenkinsCredentialResponse.Id
	response.Description = jenkinsCredentialResponse.Description
	response.DisplayName = jenkinsCredentialResponse.DisplayName
	if jenkinsCredentialResponse.Fingerprint != nil && jenkinsCredentialResponse.Fingerprint.Hash != "" {
		response.Fingerprint = &struct {
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
		}{}
		response.Fingerprint.FileName = jenkinsCredentialResponse.Fingerprint.FileName
		response.Fingerprint.Hash = jenkinsCredentialResponse.Fingerprint.Hash
		for _, usage := range jenkinsCredentialResponse.Fingerprint.Usage {
			response.Fingerprint.Usage = append(response.Fingerprint.Usage, usage)
		}
	}
	response.Domain = jenkinsCredentialResponse.Domain

	if dbCredentialResponse != nil {
		response.CreateTime = &dbCredentialResponse.CreateTime
		response.Creator = dbCredentialResponse.Creator
	}

	credentialType, ok := CredentialTypeMap[jenkinsCredentialResponse.TypeName]
	if ok {
		response.Type = credentialType
		return response
	}
	response.Type = jenkinsCredentialResponse.TypeName
	return response
}

func formatCredentialsResponse(jenkinsCredentialsResponse []*gojenkins.CredentialResponse,
	projectCredentials []*ProjectCredential) []*JenkinsCredential {
	responseSlice := make([]*JenkinsCredential, 0)
	for _, jenkinsCredential := range jenkinsCredentialsResponse {
		var dbCredential *ProjectCredential = nil
		for _, projectCredential := range projectCredentials {
			if projectCredential.CredentialId == jenkinsCredential.Id &&
				projectCredential.Domain == jenkinsCredential.Domain {
				dbCredential = projectCredential
			}
		}
		responseSlice = append(responseSlice, formatCredentialResponse(jenkinsCredential, dbCredential))
	}
	return responseSlice
}
