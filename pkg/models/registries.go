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

package models

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeclient "kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
)

const TYPE = "kubernetes.io/dockerconfigjson"

const SECRET = "Secret"

const APIVERSION = "v1"

type AuthInfo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ServerHost string `json:"serverhost"`
}

func NewAuthInfo(para Registries) *AuthInfo {

	return &AuthInfo{
		Username:   para.RegUsername,
		Password:   para.RegPassword,
		ServerHost: para.RegServerHost,
	}

}

func convert2DockerJson(authinfo AuthInfo) []byte {

	datastr := []byte(authinfo.Username + ":" + authinfo.Password)
	auth := base64.StdEncoding.EncodeToString(datastr)

	dockercfg := fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"auth\":\"%s\"}}}",
		authinfo.ServerHost, authinfo.Username, authinfo.Password, auth)

	return []byte(dockercfg)

}

type Registries struct {
	DisplayName   string      `json:"display_name,omitempty"`
	AuthProject   string      `json:"auth_project,omitempty"`
	RegServerHost string      `json:"reg_server_host,omitempty"`
	RegUsername   string      `json:"reg_username,omitempty"`
	RegPassword   string      `json:"reg_password,omitempty"`
	Annotations   interface{} `json:"annotations"`
}

type ValidationMsg struct {
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

const DOCKERCLIENTERROR = "Docker client error"

func RegistryLoginAuth(authinfo AuthInfo) ValidationMsg {

	var result ValidationMsg

	datastr := []byte(authinfo.Username + ":" + authinfo.Password)
	auth := base64.StdEncoding.EncodeToString(datastr)
	ctx := context.Background()
	cli, err := client.NewClientWithOpts()

	if err != nil {

		glog.Error(err)

	}

	authcfg := types.AuthConfig{

		Username:      authinfo.Username,
		Password:      authinfo.Password,
		Auth:          auth,
		ServerAddress: authinfo.ServerHost,
	}

	authmsg, err := cli.RegistryLogin(ctx, authcfg)

	cli.Close()

	if err != nil {

		glog.Error(err)

	}

	if authmsg.Status == "Login Succeeded" {

		result.Message = "Verified"

	} else {

		result.Message = "Unverified"
		result.Reason = "Username or password is incorrect "

	}

	return result

}

//create registries
func CreateRegistries(registries Registries) (msg constants.MessageResponse, err error) {

	projects := strings.Split(registries.AuthProject, ",")

	var secret v1.Secret

	secret.Kind = SECRET
	secret.APIVersion = APIVERSION
	secret.Type = TYPE
	secret.Name = registries.DisplayName

	authinfo := NewAuthInfo(registries)
	data := make(map[string][]byte)
	data[".dockerconfigjson"] = convert2DockerJson(*authinfo)

	secret.Data = data
	k8sclient := kubeclient.NewK8sClient()

	labels := make(map[string]string)

	labels["app"] = "dockerhubkey"
	secret.Labels = labels

	annotations := make(map[string]string)

	for key, value := range registries.Annotations.(map[string]interface{}) {

		annotations[key] = value.(string)

	}
	secret.Annotations = annotations

	for _, pro := range projects {

		glog.Infof("create secret %s in %s ", registries.DisplayName, pro)
		_, err := k8sclient.CoreV1().Secrets(pro).Create(&secret)

		if err != nil {
			glog.Error(err)
			return msg, err
		}

	} //end for

	msg.Message = "success"

	return msg, nil

}

//query registries host
func QueryRegistries(project string) ([]Registries, error) {

	regList := make([]Registries, 0)

	k8sclient := kubeclient.NewK8sClient()

	var options meta_v1.ListOptions
	options.LabelSelector = "app=dockerhubkey"

	var reg Registries

	secrets, err := k8sclient.CoreV1().Secrets(project).List(options)

	if err != nil {

		glog.Errorln(err)
		return regList, err

	}

	if len(secrets.Items) > 0 {

		for _, secret := range secrets.Items {

			reg.DisplayName = secret.Name
			reg.AuthProject = project

			if err != nil {

				glog.Errorln(err)
				return regList, err

			}

			var data map[string]interface{}
			err := json.Unmarshal(secret.Data[".dockerconfigjson"], &data)

			if err != nil {

				glog.Errorln(err)
				return regList, err

			}

			hostMap := data["auths"].(map[string]interface{})

			for key, _ := range hostMap {

				reg.RegServerHost = key

			}

			regList = append(regList, reg)

		}

	}

	return regList, nil

}
