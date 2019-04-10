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
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	Unknown      = "-"
	DeploySuffix = "-Deployment"
	DaemonSuffix = "-DaemonSet"
	StateSuffix  = "-StatefulSet"
)

type Cluster struct {
	ClusterID       string        `json:"cluster_id"`
	Name            string        `json:"name"`
	AppID           string        `json:"app_id"`
	VersionID       string        `json:"version_id"`
	Status          string        `json:"status"`
	UpdateTime      time.Time     `json:"status_time"`
	CreateTime      time.Time     `json:"create_time"`
	RunTimeId       string        `json:"runtime_id"`
	Description     string        `json:"description"`
	ClusterRoleSets []ClusterRole `json:"cluster_role_set"`
}

type ClusterRole struct {
	ClusterID string `json:"cluster_id"`
	Role      string `json:"role"`
}

type ClusterList struct {
	Total    int       `json:"total_count"`
	Clusters []Cluster `json:"cluster_set"`
}

type VersionList struct {
	Total    int       `json:"total_count"`
	Versions []version `json:"app_version_set"`
}

type version struct {
	Name      string `json:"name"`
	VersionID string `json:"version_id"`
}

type runtime struct {
	RuntimeID string `json:"runtime_id"`
	Zone      string `json:"zone"`
}

type runtimeList struct {
	Total    int       `json:"total_count"`
	Runtimes []runtime `json:"runtime_set"`
}

type app struct {
	AppId     string `json:"app_id"`
	Name      string `json:"name"`
	ChartName string `json:"chart_name"`
	RepoId    string `json:"repo_id"`
}

type repo struct {
	RepoId string `json:"repo_id"`
	Name   string `json:"name"`
	Url    string `json:"url"`
}

type appList struct {
	Total int   `json:"total_count"`
	Apps  []app `json:"app_set"`
}

type repoList struct {
	Total int    `json:"total_count"`
	Repos []repo `json:"repo_set"`
}

func GetAppInfo(appId string) (string, string, string, error) {
	url := fmt.Sprintf("%s/v1/apps?app_id=%s", openpitrixAPIServer, appId)
	resp, err := makeHttpRequest("GET", url, "")
	if err != nil {
		glog.Error(err)
		return Unknown, Unknown, Unknown, err
	}

	var apps appList
	err = json.Unmarshal(resp, &apps)
	if err != nil {
		glog.Error(err)
		return Unknown, Unknown, Unknown, err
	}

	if len(apps.Apps) == 0 {
		return Unknown, Unknown, Unknown, err
	}

	return apps.Apps[0].ChartName, apps.Apps[0].RepoId, apps.Apps[0].AppId, nil
}

func GetCluster(clusterId string) (*Cluster, error) {
	if strings.HasSuffix(openpitrixAPIServer, "/") {
		openpitrixAPIServer = strings.TrimSuffix(openpitrixAPIServer, "/")
	}

	url := fmt.Sprintf("%s/v1/clusters?cluster_id=%s", openpitrixAPIServer, clusterId)

	resp, err := makeHttpRequest("GET", url, "")
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	var clusterList ClusterList
	err = json.Unmarshal(resp, &clusterList)

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	if len(clusterList.Clusters) == 0 {
		return nil, fmt.Errorf("NotFound, clusterId:%s", clusterId)
	}

	return &clusterList.Clusters[0], nil
}

func ListClusters(runtimeId, searchWord, status string, limit, offset int) (*ClusterList, error) {
	if strings.HasSuffix(openpitrixAPIServer, "/") {
		openpitrixAPIServer = strings.TrimSuffix(openpitrixAPIServer, "/")
	}

	defaultStatus := "status=active&status=stopped&status=pending&status=ceased"

	url := fmt.Sprintf("%s/v1/clusters?limit=%s&offset=%s", openpitrixAPIServer, strconv.Itoa(limit), strconv.Itoa(offset))

	if searchWord != "" {
		url = fmt.Sprintf("%s&search_word=%s", url, searchWord)
	}

	if status != "" {
		url = fmt.Sprintf("%s&status=%s", url, status)
	} else {
		url = fmt.Sprintf("%s&%s", url, defaultStatus)
	}

	if len(runtimeId) > 0 {
		url = fmt.Sprintf("%s&runtime_id=%s", url, runtimeId)
	}

	resp, err := makeHttpRequest("GET", url, "")
	if err != nil {
		glog.Errorf("request %s failed, reason: %s", url, err)
		return nil, err
	}

	var clusterList ClusterList
	err = json.Unmarshal(resp, &clusterList)

	if err != nil {
		return nil, err
	}

	return &clusterList, nil
}

func GetRepo(repoId string) (string, error) {
	url := fmt.Sprintf("%s/v1/repos?repo_id=%s", openpitrixAPIServer, repoId)
	resp, err := makeHttpRequest("GET", url, "")
	if err != nil {
		glog.Error(err)
		return Unknown, err
	}

	var repos repoList
	err = json.Unmarshal(resp, &repos)
	if err != nil {
		glog.Error(err)
		return Unknown, err
	}

	if len(repos.Repos) == 0 {
		return Unknown, err
	}

	return repos.Repos[0].Name, nil
}

func GetVersion(versionId string) (string, error) {
	versionUrl := fmt.Sprintf("%s/v1/app_versions?version_id=%s", openpitrixAPIServer, versionId)
	resp, err := makeHttpRequest("GET", versionUrl, "")
	if err != nil {
		glog.Error(err)
		return Unknown, err
	}

	var versions VersionList
	err = json.Unmarshal(resp, &versions)
	if err != nil {
		glog.Error(err)
		return Unknown, err
	}

	if len(versions.Versions) == 0 {
		return Unknown, nil
	}
	return versions.Versions[0].Name, nil
}

func GetRuntime(runtimeId string) (string, error) {

	versionUrl := fmt.Sprintf("%s/v1/runtimes?runtime_id=%s", openpitrixAPIServer, runtimeId)
	resp, err := makeHttpRequest("GET", versionUrl, "")
	if err != nil {
		glog.Error(err)
		return Unknown, err
	}

	var runtimes runtimeList
	err = json.Unmarshal(resp, &runtimes)
	if err != nil {
		glog.Error(err)
		return Unknown, err
	}

	if len(runtimes.Runtimes) == 0 {
		return Unknown, nil
	}

	return runtimes.Runtimes[0].Zone, nil
}

func makeHttpRequest(method, url, data string) ([]byte, error) {
	var req *http.Request

	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(data))
	}

	req.Header.Add("Authorization", openpitrixProxyToken)

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		err := fmt.Errorf("Request to %s failed, method: %s,token: %s, reason: %s ", url, method, openpitrixProxyToken, err)
		glog.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf(string(body))
	}
	return body, err
}
