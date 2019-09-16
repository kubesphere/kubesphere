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
	"io/ioutil"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"net/http"
	"strings"
)

type Interface interface {
	CreateGroup(group *models.Group) (*models.Group, error)
	UpdateGroup(group *models.Group) (*models.Group, error)
	DescribeGroup(name string) (*models.Group, error)
	DeleteGroup(name string) error
	ListUsers() (*models.PageableResponse, error)
	ListWorkspaceDevOpsProjects(workspace string) (*v1alpha2.PageableDevOpsProject, error)
	DeleteWorkspaceDevOpsProjects(workspace, devops string) error
}

type KubeSphereClient struct {
	client *http.Client

	apiServer     string
	accountServer string
}

func NewKubeSphereClient(options *KubeSphereOptions) *KubeSphereClient {
	return &KubeSphereClient{
		client:        &http.Client{},
		apiServer:     options.APIServer,
		accountServer: options.AccountServer,
	}
}

type Error struct {
	status  int
	message string
}

func (e Error) Error() string {
	return fmt.Sprintf("status: %d,message: %s", e.status, e.message)
}

func (c *KubeSphereClient) CreateGroup(group *models.Group) (*models.Group, error) {
	data, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/groups", c.accountServer), bytes.NewReader(data))

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if resp.StatusCode > http.StatusOK {
		return nil, Error{resp.StatusCode, string(data)}
	}

	err = json.Unmarshal(data, group)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return group, nil
}

func (c *KubeSphereClient) UpdateGroup(group *models.Group) (*models.Group, error) {
	data, err := json.Marshal(group)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/groups/%s", c.accountServer, group.Name), bytes.NewReader(data))

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if resp.StatusCode > http.StatusOK {
		return nil, Error{resp.StatusCode, string(data)}
	}

	err = json.Unmarshal(data, group)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return group, nil
}

func (c *KubeSphereClient) DeleteGroup(name string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/groups/%s", c.accountServer, name), nil)

	if err != nil {
		klog.Error(err)
		return err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		klog.Error(err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return err
	}

	if resp.StatusCode > http.StatusOK {
		return Error{resp.StatusCode, string(data)}
	}

	return nil
}

func (c *KubeSphereClient) DescribeGroup(name string) (*models.Group, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/groups/%s", c.accountServer, name), nil)

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	resp, err := c.client.Do(req)

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if resp.StatusCode > http.StatusOK {
		return nil, Error{resp.StatusCode, string(data)}
	}

	var group models.Group
	err = json.Unmarshal(data, &group)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &group, nil
}

func (c *KubeSphereClient) ListUsers() (*models.PageableResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/users", c.accountServer), nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", c.accountServer)
	resp, err := c.client.Do(req)

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if resp.StatusCode > http.StatusOK {
		return nil, Error{resp.StatusCode, string(data)}
	}

	var result models.PageableResponse
	err = json.Unmarshal(data, &result)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &result, nil
}

func (c *KubeSphereClient) ListWorkspaceDevOpsProjects(workspace string) (*v1alpha2.PageableDevOpsProject, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/kapis/tenant.kubesphere.io/v1alpha2/workspaces/%s/devops", c.apiServer, workspace), nil)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	req.Header.Add(constants.UserNameHeader, constants.AdminUserName)

	klog.Info(req.Method, req.URL)
	resp, err := c.client.Do(req)

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if resp.StatusCode > http.StatusOK {
		klog.Error(req.Method, req.URL, resp.StatusCode, string(data))
		return nil, Error{resp.StatusCode, string(data)}
	}

	var result v1alpha2.PageableDevOpsProject
	err = json.Unmarshal(data, &result)

	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &result, nil

}

func (c *KubeSphereClient) DeleteWorkspaceDevOpsProjects(workspace, devops string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/kapis/tenant.kubesphere.io/v1alpha2/workspaces/%s/devops/%s", c.apiServer, workspace, devops), nil)

	if err != nil {
		klog.Error(err)
		return err
	}
	req.Header.Add(constants.UserNameHeader, constants.AdminUserName)

	klog.Info(req.Method, req.URL)
	resp, err := c.client.Do(req)

	if err != nil {
		klog.Error(err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		klog.Error(err)
		return err
	}
	if resp.StatusCode > http.StatusOK {
		klog.Error(req.Method, req.URL, resp.StatusCode, string(data))
		return Error{resp.StatusCode, string(data)}
	}

	return nil
}

func IsNotFound(err error) bool {
	if e, ok := err.(Error); ok {
		if e.status == http.StatusNotFound {
			return true
		}
		if strings.Contains(e.message, "not exist") {
			return true
		}
		if strings.Contains(e.message, "not found") {
			return true
		}
	}
	return false
}

func IsExist(err error) bool {
	if e, ok := err.(Error); ok {
		if e.status == http.StatusConflict {
			return true
		}
		if strings.Contains(e.message, "Already Exists") {
			return true
		}
	}
	return false
}
