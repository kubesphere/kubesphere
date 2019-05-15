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
package openpitrix

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"sync"
)

var (
	openpitrixAPIServer  string
	openpitrixProxyToken string
	once                 sync.Once
	c                    client
)

type RunTime struct {
	RuntimeId         string `json:"runtime_id"`
	RuntimeUrl        string `json:"runtime_url"`
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	Zone              string `json:"zone"`
	RuntimeCredential string `json:"runtime_credential"`
}

type Interface interface {
	CreateRuntime(runtime *RunTime) error
	DeleteRuntime(runtimeId string) error
}

type Error struct {
	status  int
	message string
}

func (e Error) Error() string {
	return fmt.Sprintf("status: %d,message: %s", e.status, e.message)
}

type client struct {
	client http.Client
}

func init() {
	flag.StringVar(&openpitrixAPIServer, "openpitrix-api-server", "http://openpitrix-api-gateway.openpitrix-system.svc:9100", "openpitrix api server")
	flag.StringVar(&openpitrixProxyToken, "openpitrix-proxy-token", "", "openpitrix proxy token")
}

func Client() Interface {
	once.Do(func() {
		c = client{client: http.Client{}}
	})
	return c
}

func (c client) CreateRuntime(runtime *RunTime) error {

	data, err := json.Marshal(runtime)
	if err != nil {
		glog.Error(err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/runtimes", openpitrixAPIServer), bytes.NewReader(data))

	if err != nil {
		glog.Error(err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", openpitrixProxyToken)

	resp, err := c.client.Do(req)

	if err != nil {
		glog.Error(err)
		return err
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		glog.Error(err)
		return err
	}

	if resp.StatusCode > http.StatusOK {
		return Error{resp.StatusCode, string(data)}
	}

	return nil
}

func (c client) DeleteRuntime(runtimeId string) error {
	data := []byte(fmt.Sprintf(`{"runtime_id":"%s"}`, runtimeId))
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/runtimes", openpitrixAPIServer), bytes.NewReader(data))

	if err != nil {
		glog.Error(err)

		return err
	}

	req.Header.Add("Authorization", openpitrixProxyToken)

	resp, err := c.client.Do(req)

	if err != nil {
		glog.Error(err)
		return err
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		glog.Error(err)
		return err
	}

	if resp.StatusCode > http.StatusOK {
		return Error{resp.StatusCode, string(data)}
	}

	return nil
}
