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
	"flag"
	"fmt"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/models"
	"log"
	"net/http"
	"strings"
	"sync"
)

var (
	accountAPIServer string
	devopsAPIServer  string
	once             sync.Once
	c                client
)

type Interface interface {
	CreateGroup(group *models.Group) (*models.Group, error)
	UpdateGroup(group *models.Group) (*models.Group, error)
	DescribeGroup(name string) (*models.Group, error)
	DeleteGroup(name string) error
}

type client struct {
	client http.Client
}

func init() {
	flag.StringVar(&accountAPIServer, "ks-account-api-server", "http://ks-account.kubesphere-system.svc", "kubesphere account api server")
}

func Client() Interface {
	once.Do(func() {
		c = client{client: http.Client{}}
	})
	return c
}

type Error struct {
	status  int
	message string
}

func (e Error) Error() string {
	return fmt.Sprintf("status: %d,message: %s", e.status, e.message)
}

func (c client) CreateGroup(group *models.Group) (*models.Group, error) {
	data, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/groups", accountAPIServer), bytes.NewReader(data))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	log.Println(req.Method, req.URL, string(data))
	resp, err := c.client.Do(req)

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

	err = json.Unmarshal(data, group)

	if err != nil {
		return nil, err
	}

	return group, nil
}

func (c client) UpdateGroup(group *models.Group) (*models.Group, error) {
	data, err := json.Marshal(group)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/groups/%s", accountAPIServer, group.Name), bytes.NewReader(data))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	log.Println(req.Method, req.URL, string(data))
	resp, err := c.client.Do(req)

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

	err = json.Unmarshal(data, group)

	if err != nil {
		return nil, err
	}

	return group, nil
}

func (c client) DeleteGroup(name string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/groups/%s", accountAPIServer, name), nil)

	if err != nil {
		return err
	}

	log.Println(req.Method, req.URL)
	resp, err := c.client.Do(req)

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

func (c client) DescribeGroup(name string) (*models.Group, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/groups/%s", accountAPIServer, name), nil)

	if err != nil {
		return nil, err
	}
	log.Println(req.Method, req.URL)
	resp, err := c.client.Do(req)

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

	var group models.Group
	err = json.Unmarshal(data, &group)

	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (c client) ListUsers() (*models.PageableResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/kapis/iam.kubesphere.io/v1alpha2/users", accountAPIServer), nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", accountAPIServer)
	if err != nil {
		return nil, err
	}
	log.Println(req.Method, req.URL)
	resp, err := c.client.Do(req)

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

	var result models.PageableResponse
	err = json.Unmarshal(data, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
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
