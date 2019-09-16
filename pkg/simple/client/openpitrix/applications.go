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
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"strconv"
	"strings"
)

const (
	Unknown      = "-"
	DeploySuffix = "-Deployment"
	DaemonSuffix = "-DaemonSet"
	StateSuffix  = "-StatefulSet"
)

func (c *OpenPitrixClient) GetAppInfo(appId string) (string, string, string, error) {
	url := fmt.Sprintf("%s/v1/apps?app_id=%s", c.apiServer, appId)
	resp, err := c.makeHttpRequest("GET", url, "")
	if err != nil {
		klog.Error(err)
		return Unknown, Unknown, Unknown, err
	}

	var apps appList
	err = json.Unmarshal(resp, &apps)
	if err != nil {
		klog.Error(err)
		return Unknown, Unknown, Unknown, err
	}

	if len(apps.Apps) == 0 {
		return Unknown, Unknown, Unknown, err
	}

	return apps.Apps[0].ChartName, apps.Apps[0].RepoId, apps.Apps[0].AppId, nil
}

func (c *OpenPitrixClient) GetCluster(clusterId string) (*Cluster, error) {
	url := fmt.Sprintf("%s/v1/clusters?cluster_id=%s", c.apiServer, clusterId)

	resp, err := c.makeHttpRequest("GET", url, "")
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var clusterList ClusterList
	err = json.Unmarshal(resp, &clusterList)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(clusterList.Clusters) == 0 {
		return nil, fmt.Errorf("NotFound, clusterId:%s", clusterId)
	}

	return &clusterList.Clusters[0], nil
}

func (c *OpenPitrixClient) ListClusters(runtimeId, searchWord, status string, limit, offset int) (*ClusterList, error) {

	defaultStatus := "status=active&status=stopped&status=pending&status=ceased"

	url := fmt.Sprintf("%s/v1/clusters?limit=%s&offset=%s", c.apiServer, strconv.Itoa(limit), strconv.Itoa(offset))

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

	resp, err := c.makeHttpRequest("GET", url, "")
	if err != nil {
		klog.Errorf("request %s failed, reason: %s", url, err)
		return nil, err
	}

	var clusterList ClusterList
	err = json.Unmarshal(resp, &clusterList)

	if err != nil {
		return nil, err
	}

	return &clusterList, nil
}

func (c *OpenPitrixClient) GetRepo(repoId string) (string, error) {
	url := fmt.Sprintf("%s/v1/repos?repo_id=%s", c.apiServer, repoId)
	resp, err := c.makeHttpRequest("GET", url, "")
	if err != nil {
		klog.Error(err)
		return Unknown, err
	}

	var repos repoList
	err = json.Unmarshal(resp, &repos)
	if err != nil {
		klog.Error(err)
		return Unknown, err
	}

	if len(repos.Repos) == 0 {
		return Unknown, err
	}

	return repos.Repos[0].Name, nil
}

func (c *OpenPitrixClient) GetVersion(versionId string) (string, error) {
	versionUrl := fmt.Sprintf("%s/v1/app_versions?version_id=%s", c.apiServer, versionId)
	resp, err := c.makeHttpRequest("GET", versionUrl, "")
	if err != nil {
		klog.Error(err)
		return Unknown, err
	}

	var versions VersionList
	err = json.Unmarshal(resp, &versions)
	if err != nil {
		klog.Error(err)
		return Unknown, err
	}

	if len(versions.Versions) == 0 {
		return Unknown, nil
	}
	return versions.Versions[0].Name, nil
}

func (c *OpenPitrixClient) GetRuntime(runtimeId string) (string, error) {

	versionUrl := fmt.Sprintf("%s/v1/runtimes?runtime_id=%s", c.apiServer, runtimeId)
	resp, err := c.makeHttpRequest("GET", versionUrl, "")
	if err != nil {
		klog.Error(err)
		return Unknown, err
	}

	var runtimes runtimeList
	err = json.Unmarshal(resp, &runtimes)
	if err != nil {
		klog.Error(err)
		return Unknown, err
	}

	if len(runtimes.Runtimes) == 0 {
		return Unknown, nil
	}

	return runtimes.Runtimes[0].Zone, nil
}

func (c *OpenPitrixClient) CreateCluster(request CreateClusterRequest) error {

	versionUrl := fmt.Sprintf("%s/v1/clusters/create", c.apiServer)

	data, err := json.Marshal(request)

	if err != nil {
		klog.Error(err)
		return err
	}

	data, err = c.makeHttpRequest("POST", versionUrl, string(data))

	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func (c *OpenPitrixClient) DeleteCluster(request DeleteClusterRequest) error {

	versionUrl := fmt.Sprintf("%s/v1/clusters/delete", c.apiServer)

	data, err := json.Marshal(request)

	if err != nil {
		klog.Error(err)
		return err
	}

	data, err = c.makeHttpRequest("POST", versionUrl, string(data))

	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func (c *OpenPitrixClient) makeHttpRequest(method, url, data string) ([]byte, error) {
	var req *http.Request

	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(data))
	}

	req.Header.Add("Authorization", c.token)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		err := fmt.Errorf("Request to %s failed, method: %s,token: %s, reason: %s ", url, method, c.apiServer, err)
		klog.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf(string(body))
	}
	return body, err
}
