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
package kubesphere

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"net/http"
)

func (c client) DeleteDevopsProject(username string, projectId string) error {
	request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1alpha/projects/%s", devopsAPIServer, projectId), nil)
	if username == "" {
		username = constants.AdminUserName
	}
	request.Header.Add("X-Token-Username", username)

	resp, err := c.client.Do(request)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > http.StatusOK {
		return Error{resp.StatusCode, string(data)}
	}
	return nil
}

func (c client) GetUserDevopsRole(username string, projectId string) (string, error) {

	if username == "admin" {
		return "owner", nil
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1alpha/projects/%s/members", devopsAPIServer, projectId), nil)

	if err != nil {
		return "", err
	}
	req.Header.Set(constants.UserNameHeader, username)
	resp, err := c.client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if resp.StatusCode > http.StatusOK {
		return "", Error{resp.StatusCode, string(data)}
	}

	var result []map[string]string

	err = json.Unmarshal(data, &result)

	if err != nil {
		return "", err
	}

	for _, item := range result {
		if item["username"] == username {
			return item["role"], nil
		}
	}

	return "", nil
}

func (c client) CreateDevopsProject(username string, project *models.DevopsProject) (*models.DevopsProject, error) {
	data, err := json.Marshal(project)

	if err != nil {
		return nil, err
	}

	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1alpha/projects", devopsAPIServer), bytes.NewReader(data))
	request.Header.Add("X-Token-Username", username)
	request.Header.Add("Content-Type", "application/json")
	resp, err := c.client.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode > http.StatusOK {
		return nil, Error{resp.StatusCode, string(data)}
	}

	var created models.DevopsProject

	err = json.Unmarshal(data, &created)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (c client) CreateDevopsRoleBinding(projectId string, user string, role string) {

	projects := make([]string, 0)
	projects = append(projects, projectId)

	for _, project := range projects {
		data := []byte(fmt.Sprintf(`{"username":"%s","role":"%s"}`, user, role))
		request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1alpha/projects/%s/members", devopsAPIServer, project), bytes.NewReader(data))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("X-Token-Username", "admin")
		resp, err := c.client.Do(request)
		if err != nil || resp.StatusCode > 200 {
			glog.Warning(fmt.Sprintf("create devops role binding failed %s,%s,%s", project, user, role))
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func (c client) ListDevopsProjects(username string) ([]models.DevopsProject, error) {
	projects := make([]models.DevopsProject, 0)

	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1alpha/projects", devopsAPIServer), nil)
	request.Header.Add(constants.UserNameHeader, username)

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode > http.StatusOK {
		return nil, Error{resp.StatusCode, string(data)}
	}

	err = json.Unmarshal(data, &projects)

	if err != nil {
		return nil, err
	}

	return projects, nil
}
