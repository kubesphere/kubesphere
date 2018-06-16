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

func RegistryLoginAuth(authinfo AuthInfo) ValidationMsg {

	var result ValidationMsg

	datastr := []byte(authinfo.Username + ":" + authinfo.Password)
	auth := base64.StdEncoding.EncodeToString(datastr)
	ctx := context.Background()
	cli, err := client.NewEnvClient()

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

//list all registries
func ListAllRegistries() ([]Registries, error) {

	result := make([]Registries, 0)
	k8sclient := kubeclient.NewK8sClient()
	var registries Registries
	var options meta_v1.ListOptions
	options.LabelSelector = "app=dockerhubkey"

	secrets, err := k8sclient.CoreV1().Secrets("").List(options)

	if err != nil {
		return result, err
	}

	if len(secrets.Items) > 0 {

		for _, secret := range secrets.Items {

			registries.DisplayName = secret.Name
			registries.AuthProject = secret.Namespace
			var data map[string]interface{}
			err := json.Unmarshal(secret.Data[".dockerconfigjson"], &data)

			if err != nil {

				glog.Errorln(err)
				return result, err

			}

			hostMap := data["auths"].(map[string]interface{})

			for key, val := range hostMap {

				registries.RegServerHost = key
				info := val.(map[string]interface{})
				registries.RegUsername = info["username"].(string)
				registries.RegPassword = info["password"].(string)

			}

			registries.Annotations = secret.Annotations

			if len(result) == 0 {

				result = append(result, registries)

			} else {

				if !containSame(registries, result) {

					result = append(result, registries)

				}

			}

		}

	}

	return result, nil

}

func containSame(registries Registries, list []Registries) bool {

	flag := false

	for ind, reg := range list {

		if reg.DisplayName == registries.DisplayName {
			list[ind].AuthProject = reg.AuthProject + "," + registries.AuthProject
			flag = true
		}

	}

	return flag

}

//delete registries

func DeleteRegistries(name string) (constants.MessageResponse, error) {

	var msg constants.MessageResponse
	k8sclient := kubeclient.NewK8sClient()

	var options meta_v1.ListOptions
	options.FieldSelector = "metadata.name=" + name
	secretList, err := k8sclient.CoreV1().Secrets("").List(options)

	if err != nil {
		return msg, err
	}

	if len(secretList.Items) > 0 {
		for _, secret := range secretList.Items {
			err := k8sclient.CoreV1().Secrets(secret.Namespace).Delete(secret.Name, &meta_v1.DeleteOptions{})
			if err != nil {
				return msg, err
			}
		}
	}
	msg.Message = "success"
	return msg, nil

}

//update registries

func UpdateRegistries(name string, registries Registries) (Registries, error) {

	DeleteRegistries(name)

	k8sclient := kubeclient.NewK8sClient()
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
	labels := make(map[string]string)
	labels["app"] = "dockerhubkey"
	secret.Labels = labels
	annotations := make(map[string]string)
	for key, value := range registries.Annotations.(map[string]interface{}) {

		annotations[key] = value.(string)

	}
	secret.Annotations = annotations

	for _, pro := range projects {

		glog.Infof("alter secret %s in %s ", registries.DisplayName, pro)
		_, err := k8sclient.CoreV1().Secrets(pro).Create(&secret)

		if err != nil {
			glog.Error(err)
			return registries, err
		}

	} //end for

	return registries, nil

}

// Get registries detail
func GetReisgtries(name string) (Registries, error) {

	var reg Registries

	k8sclient := kubeclient.NewK8sClient()

	var options meta_v1.ListOptions
	options.FieldSelector = "metadata.name=" + name
	secretList, err := k8sclient.CoreV1().Secrets("").List(options)

	if err != nil {

		return reg, err
	}

	if len(secretList.Items) > 0 {

		for _, secret := range secretList.Items {

			secret, err := k8sclient.CoreV1().Secrets(secret.Namespace).Get(secret.Name, meta_v1.GetOptions{})

			if err == nil {
				reg.DisplayName = secret.Name
				var data map[string]interface{}
				json.Unmarshal(secret.Data[".dockerconfigjson"], &data)
				if len(reg.AuthProject) == 0 {
					reg.AuthProject = secret.Namespace
				} else {
					reg.AuthProject = reg.AuthProject + "," + secret.Namespace
				}
				hostMap := data["auths"].(map[string]interface{})
				for key, val := range hostMap {
					reg.RegServerHost = key
					info := val.(map[string]interface{})
					reg.RegUsername = info["username"].(string)
					reg.RegPassword = info["password"].(string)
				}
				reg.Annotations = secret.Annotations
			}

		}

	}

	return reg, nil

}
