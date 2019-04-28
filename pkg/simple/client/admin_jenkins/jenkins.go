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

package admin_jenkins

import (
	"flag"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"sync"
)

var (
	jenkinsClientOnce    sync.Once
	jenkinsClient        *gojenkins.Jenkins
	jenkinsAdminAddress  string
	jenkinsAdminUsername string
	jenkinsAdminPassword string
	jenkinsMaxConn       int
)

const (
	JenkinsAllUserRoleName = "kubesphere-user"
)

func init() {
	flag.StringVar(&jenkinsAdminAddress, "jenkins-address", "http://ks-jenkins.kubesphere-devops-system.svc/", "data source name")
	flag.StringVar(&jenkinsAdminUsername, "jenkins-username", "admin", "username of jenkins")
	flag.StringVar(&jenkinsAdminPassword, "jenkins-password", "passw0rd", "password of jenkins")
	flag.IntVar(&jenkinsMaxConn, "jenkins-max-conn", 20, "max conn to jenkins")
}

func Client() *gojenkins.Jenkins {
	jenkinsClientOnce.Do(func() {
		jenkins := gojenkins.CreateJenkins(nil, jenkinsAdminAddress, jenkinsMaxConn, jenkinsAdminUsername, jenkinsAdminPassword)
		jenkins, err := jenkins.Init()
		if err != nil {
			glog.Error("failed to connect jenkins")
			return
		}
		jenkinsClient = jenkins
		globalRole, err := jenkins.GetGlobalRole(JenkinsAllUserRoleName)
		if err != nil {
			glog.Error("failed to get jenkins role")
		}
		if globalRole == nil {
			_, err := jenkins.AddGlobalRole(JenkinsAllUserRoleName, gojenkins.GlobalPermissionIds{
				GlobalRead: true,
			}, true)
			if err != nil {
				glog.Error("failed to create jenkins global role")
				return
			}
		}
		_, err = jenkins.AddProjectRole(JenkinsAllUserRoleName, "\\n\\s*\\r", gojenkins.ProjectPermissionIds{
			SCMTag: true,
		}, true)
		if err != nil {
			glog.Error("failed to create jenkins project role")
			return
		}
	})

	return jenkinsClient

}
