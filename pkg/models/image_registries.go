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

	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"

	kubeclient "kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	TYPE               = "kubernetes.io/dockerconfigjson"
	SECRET             = "Secret"
	APIVERSION         = "v1"
	TYPEHARBOR         = "harbor"
	TYPEDOCKERHUB      = "dockerhub"
	TYPEDOCKERREGISTRY = "docker-registry"
)

type AuthInfo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ServerHost string `json:"serverhost"`
}

type DockerConfigEntry struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

type RegistryInfo struct {
	user, password, registryType, url string
}

type dockerConfig map[string]map[string]DockerConfigEntry

type harborRepo struct {
	RepoName string `json:"repository_name"`
}

type harborRepos struct {
	Repos []harborRepo `json:"repository"`
}

type registryRepos struct {
	Repositories []string
}

type registryTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type dockerhubRepo struct {
	RepoName string `json:"repo_name"`
}
type dockerhubRepos struct {
	Repositories []dockerhubRepo `json:"results"`
}

type dockerhubTag struct {
	TagName string `json:"name"`
}

type dockerhubTags struct {
	Tags []dockerhubTag `json:"results"`
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

// by image secret to get registry'info, like username, password, registry url ...
func getRegistryInfo(namespace, registryName string) *RegistryInfo {

	var registry RegistryInfo
	k8sClient := kubeclient.NewK8sClient()
	secret, err := k8sClient.CoreV1().Secrets(namespace).Get(registryName, meta_v1.GetOptions{})
	if err != nil {
		glog.Error(err)
		return nil
	}

	registry.registryType = secret.Annotations["type"]

	data := secret.Data[v1.DockerConfigJsonKey]

	authsMap := make(dockerConfig)
	err = json.Unmarshal(data, &authsMap)
	if err != nil {
		glog.Error(err)
		return nil
	}

	for url, config := range authsMap["auths"] {
		registry.url = url
		registry.user = config.Username
		registry.password = config.Password
		break
	}

	return &registry
}

func ImageSearch(namespace, registryName, searchWord string) []string {
	registry := getRegistryInfo(namespace, registryName)
	if registry == nil {
		return nil
	}

	switch registry.registryType {
	case TYPEDOCKERHUB:
		return searchDockerHub(registry.url, searchWord)
	case TYPEDOCKERREGISTRY:
		return searchDockerRegistry(registry.url, searchWord)
	case TYPEHARBOR:
		return searchHarbor(registry.url, registry.user, registry.password, searchWord)
	}

	return nil
}

func GetImageTags(namespace, registryName, imageName string) []string {
	registry := getRegistryInfo(namespace, registryName)
	if registry == nil {
		return nil
	}

	switch registry.registryType {
	case TYPEDOCKERHUB:
		return getTagInDockerHub(registry.url, imageName)
	case TYPEDOCKERREGISTRY:
		return getTagInDockerRegistry(registry.url, imageName)
	case TYPEHARBOR:
		return getTagInHarbor(registry.url, registry.user, registry.password, imageName)
	}

	return nil
}

func httpGet(url, username, password string, insecure bool) ([]byte, error) {
	var httpClient *http.Client
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if insecure {
		httpClient = &http.Client{}
	} else {
		req.SetBasicAuth(username, password)
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		httpClient = &http.Client{Timeout: 20 * time.Second, Transport: tr}
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		err := fmt.Errorf("Request to %s failed reason: %s ", url, err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest || err != nil {
		return nil, err
	}

	return body, nil
}

func searchHarbor(url, username, password, searchWord string) []string {
	url = strings.TrimSuffix(url, "/") + fmt.Sprintf("/api/search?q=%s", searchWord)

	body, err := httpGet(url, username, password, false)
	if err != nil || len(body) == 0 {
		glog.Error(err)
		return nil
	}

	var repos harborRepos
	repoList := make([]string, 0, 100)
	err = json.Unmarshal(body, &repos)
	if err != nil {
		glog.Error(err)
		return nil
	}

	for _, repo := range repos.Repos {
		repoList = append(repoList, repo.RepoName)
	}

	return repoList
}

func searchDockerRegistry(url, searchword string) []string {
	url = strings.TrimSuffix(url, "/") + "/v2/_catalog"
	body, err := httpGet(url, "", "", true)
	if err != nil || len(body) == 0 {
		glog.Error(err)
		return nil
	}

	var repos registryRepos
	err = json.Unmarshal(body, &repos)
	if err != nil {
		glog.Error(err)
		return nil
	}

	repoList := make([]string, 0, 100)
	for _, repo := range repos.Repositories {
		if strings.HasPrefix(repo, searchword) {
			repoList = append(repoList, repo)
		}
	}

	return repoList
}

func searchDockerHub(url, searchWord string) []string {
	url = fmt.Sprintf("https://hub.docker.com/v2/search/repositories/?page=1&query=%s&page_size=50", searchWord)
	body, err := httpGet(url, "", "", true)
	if err != nil || len(body) == 0 {
		glog.Error(err)
		return nil
	}

	var repos dockerhubRepos
	err = json.Unmarshal(body, &repos)
	if err != nil {
		glog.Error(err)
		return nil
	}

	repoList := make([]string, 0, 50)
	for _, repo := range repos.Repositories {
		repoList = append(repoList, repo.RepoName)
	}

	return repoList
}

func getTagInHarbor(url, username, password, imageName string) []string {
	url = strings.TrimSuffix(url, "/") + fmt.Sprintf("/api/repositories/%s/tags", imageName)
	body, err := httpGet(url, username, password, false)
	if err != nil || len(body) == 0 {
		glog.Error(err)
		return nil
	}

	var tagList []string
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		glog.Error(err)
		return nil
	}

	return tagList
}

func getTagInDockerRegistry(url, imageName string) []string {
	url = strings.TrimSuffix(url, "/") + fmt.Sprintf("/v2/%s/tags/list", imageName)
	body, err := httpGet(url, "", "", true)
	if err != nil || len(body) == 0 {
		glog.Error(err)
		return nil
	}

	var tags registryTags
	err = json.Unmarshal(body, &tags)
	if err != nil {
		glog.Error(err)
		return nil
	}

	return tags.Tags
}

func getTagInDockerHub(url, imageName string) []string {
	if !strings.Contains(imageName, "/") {
		imageName = fmt.Sprintf("library/%s", imageName)
	}
	url = fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/?page=1&page_size=200", imageName)

	body, err := httpGet(url, "", "", true)
	if err != nil || len(body) == 0 {
		glog.Error(err)
		return nil
	}

	var tags dockerhubTags
	err = json.Unmarshal(body, &tags)
	if err != nil {
		glog.Error(err)
		return nil
	}

	tagList := make([]string, 0, 200)
	for _, tag := range tags.Tags {
		tagList = append(tagList, tag.TagName)
	}

	return tagList
}
