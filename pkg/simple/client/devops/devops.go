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

package devops

import (
	"fmt"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"sync"
)

const (
	jenkinsAllUserRoleName = "kubesphere-user"
)

type Client struct {
	jenkinsClient *gojenkins.Jenkins
}

func NewDevopsClient(options *Options) (*Client, error) {
	var d Client

	jenkins := gojenkins.CreateJenkins(nil, options.Host, options.MaxConnections, options.Username, options.Password)
	jenkins, err := jenkins.Init()
	if err != nil {
		klog.Errorf("failed to connecto to jenkins role, %+v", err)
		return nil, err
	}

	d.jenkinsClient = jenkins

	err = d.initializeJenkins()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &d, nil
}

func (c *Client) Jenkins() *gojenkins.Jenkins {
	return c.jenkinsClient
}

var mutex = sync.Mutex{}

func (c *Client) initializeJenkins() error {
	mutex.Lock()
	defer mutex.Unlock()

	if c.jenkinsClient == nil {
		return fmt.Errorf("jenkins intialization failed")
	}

	globalRole, err := c.jenkinsClient.GetGlobalRole(jenkinsAllUserRoleName)
	if err != nil {
		klog.Error(err)
		return err
	}

	// Jenkins uninitialized, create global role
	if globalRole == nil {
		_, err := c.jenkinsClient.AddGlobalRole(jenkinsAllUserRoleName, gojenkins.GlobalPermissionIds{GlobalRead: true}, true)
		if err != nil {
			klog.Error(err)
			return err
		}
	}

	_, err = c.jenkinsClient.AddProjectRole(jenkinsAllUserRoleName, "\\n\\s*\\r", gojenkins.ProjectPermissionIds{SCMTag: true}, true)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}
