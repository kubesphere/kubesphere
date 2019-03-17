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
package applications

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	v12 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	OpenPitrixProxyToken string
	OpenPitrixServer     string
)

const (
	unknown      = "-"
	deploySuffix = "-Deployment"
	daemonSuffix = "-DaemonSet"
	stateSuffix  = "-StatefulSet"
)

type Application struct {
	Name        string            `json:"name"`
	RepoName    string            `json:"repoName"`
	Runtime     string            `json:"namespace"`
	RuntimeId   string            `json:"runtime_id"`
	Version     string            `json:"version"`
	VersionId   string            `json:"version_id"`
	Status      string            `json:"status"`
	UpdateTime  time.Time         `json:"updateTime"`
	CreateTime  time.Time         `json:"createTime"`
	App         string            `json:"app"`
	AppId       string            `json:"app_id"`
	Description string            `json:"description,omitempty"`
	WorkLoads   *workLoads        `json:"workloads,omitempty"`
	Services    []v1.Service      `json:"services,omitempty"`
	Ingresses   []v1beta1.Ingress `json:"ingresses,omitempty"`
	ClusterID   string            `json:"cluster_id"`
}

type clusterRole struct {
	ClusterID string `json:"cluster_id"`
	Role      string `json:"role"`
}

type cluster struct {
	ClusterID       string        `json:"cluster_id"`
	Name            string        `json:"name"`
	AppID           string        `json:"app_id"`
	VersionID       string        `json:"version_id"`
	Status          string        `json:"status"`
	UpdateTime      time.Time     `json:"status_time"`
	CreateTime      time.Time     `json:"create_time"`
	RunTimeId       string        `json:"runtime_id"`
	Description     string        `json:"description"`
	ClusterRoleSets []clusterRole `json:"cluster_role_set"`
}

type clusters struct {
	Total    int       `json:"total_count"`
	Clusters []cluster `json:"cluster_set"`
}

type versionList struct {
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

type workLoads struct {
	Deployments  []v12.Deployment  `json:"deployments,omitempty"`
	Statefulsets []v12.StatefulSet `json:"statefulsets,omitempty"`
	Daemonsets   []v12.DaemonSet   `json:"daemonsets,omitempty"`
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
	url := fmt.Sprintf("%s/v1/apps?app_id=%s", OpenPitrixServer, appId)
	resp, err := makeHttpRequest("GET", url, "")
	if err != nil {
		glog.Error(err)
		return unknown, unknown, unknown, err
	}

	var apps appList
	err = json.Unmarshal(resp, &apps)
	if err != nil {
		glog.Error(err)
		return unknown, unknown, unknown, err
	}

	if len(apps.Apps) == 0 {
		return unknown, unknown, unknown, err
	}

	return apps.Apps[0].ChartName, apps.Apps[0].RepoId, apps.Apps[0].AppId, nil
}

func GetRepo(repoId string) (string, error) {
	url := fmt.Sprintf("%s/v1/repos?repo_id=%s", OpenPitrixServer, repoId)
	resp, err := makeHttpRequest("GET", url, "")
	if err != nil {
		glog.Error(err)
		return unknown, err
	}

	var repos repoList
	err = json.Unmarshal(resp, &repos)
	if err != nil {
		glog.Error(err)
		return unknown, err
	}

	if len(repos.Repos) == 0 {
		return unknown, err
	}

	return repos.Repos[0].Name, nil
}

func GetVersion(versionId string) (string, error) {
	versionUrl := fmt.Sprintf("%s/v1/app_versions?version_id=%s", OpenPitrixServer, versionId)
	resp, err := makeHttpRequest("GET", versionUrl, "")
	if err != nil {
		glog.Error(err)
		return unknown, err
	}

	var versions versionList
	err = json.Unmarshal(resp, &versions)
	if err != nil {
		glog.Error(err)
		return unknown, err
	}

	if len(versions.Versions) == 0 {
		return unknown, nil
	}
	return versions.Versions[0].Name, nil
}

func GetRuntime(runtimeId string) (string, error) {

	versionUrl := fmt.Sprintf("%s/v1/runtimes?runtime_id=%s", OpenPitrixServer, runtimeId)
	resp, err := makeHttpRequest("GET", versionUrl, "")
	if err != nil {
		glog.Error(err)
		return unknown, err
	}

	var runtimes runtimeList
	err = json.Unmarshal(resp, &runtimes)
	if err != nil {
		glog.Error(err)
		return unknown, err
	}

	if len(runtimes.Runtimes) == 0 {
		return unknown, nil
	}

	return runtimes.Runtimes[0].Zone, nil
}

func GetWorkLoads(namespace string, clusterRoles []clusterRole) (*workLoads, error) {

	var works workLoads
	for _, clusterRole := range clusterRoles {
		workLoadName := clusterRole.Role
		if len(workLoadName) > 0 {
			if strings.HasSuffix(workLoadName, deploySuffix) {
				name := strings.Split(workLoadName, deploySuffix)[0]

				item, err := informers.SharedInformerFactory().Apps().V1().Deployments().Lister().Deployments(namespace).Get(name)

				if err != nil {
					return nil, err
				}

				works.Deployments = append(works.Deployments, *item)
				continue
			}

			if strings.HasSuffix(workLoadName, daemonSuffix) {
				name := strings.Split(workLoadName, daemonSuffix)[0]
				item, err := informers.SharedInformerFactory().Apps().V1().DaemonSets().Lister().DaemonSets(namespace).Get(name)
				if err != nil {
					return nil, err
				}
				works.Daemonsets = append(works.Daemonsets, *item)
				continue
			}

			if strings.HasSuffix(workLoadName, stateSuffix) {
				name := strings.Split(workLoadName, stateSuffix)[0]
				item, err := informers.SharedInformerFactory().Apps().V1().StatefulSets().Lister().StatefulSets(namespace).Get(name)
				if err != nil {
					return nil, err
				}
				works.Statefulsets = append(works.Statefulsets, *item)
				continue
			}
		}
	}
	return &works, nil
}

func getLabels(namespace string, workloads *workLoads) *[]map[string]string {
	k8sClient := k8s.Client()

	var workloadLables []map[string]string
	if workloads == nil {
		return nil
	}

	for _, workload := range workloads.Deployments {
		deploy, err := k8sClient.AppsV1().Deployments(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, deploy.Labels)
	}

	for _, workload := range workloads.Daemonsets {
		daemonset, err := k8sClient.AppsV1().DaemonSets(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, daemonset.Labels)
	}

	for _, workload := range workloads.Statefulsets {
		statefulset, err := k8sClient.AppsV1().StatefulSets(namespace).Get(workload.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, statefulset.Labels)
	}

	return &workloadLables
}

func isExist(svcs []v1.Service, svc v1.Service) bool {
	for _, item := range svcs {
		if item.Name == svc.Name && item.Namespace == svc.Namespace {
			return true
		}
	}
	return false
}

func getSvcs(namespace string, workLoadLabels *[]map[string]string) []v1.Service {
	if len(*workLoadLabels) == 0 {
		return nil
	}
	k8sClient := k8s.Client()
	var services []v1.Service
	for _, label := range *workLoadLabels {
		labelSelector := labels.Set(label).AsSelector().String()
		svcs, err := k8sClient.CoreV1().Services(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			glog.Errorf("get app's svc failed, reason: %v", err)
		}
		for _, item := range svcs.Items {
			if !isExist(services, item) {
				services = append(services, item)
			}
		}
	}

	return services
}

func getIng(namespace string, services []v1.Service) []v1beta1.Ingress {
	if services == nil {
		return nil
	}

	var ings []v1beta1.Ingress
	for _, svc := range services {
		result, err := resources.ListNamespaceResource(namespace, "ingress", &params.Conditions{Fuzzy: map[string]string{"serviceName": svc.Name}}, "", false, -1, 0)
		if err != nil {
			glog.Error(err)
			return nil
		}

		glog.Error(result)
		for _, i := range result.Items {
			ingress := i.(*v1beta1.Ingress)

			exist := false
			var tmpRules []v1beta1.IngressRule
			for _, rule := range ingress.Spec.Rules {
				for _, p := range rule.HTTP.Paths {
					if p.Backend.ServiceName == svc.Name {
						exist = true
						tmpRules = append(tmpRules, rule)
					}
				}

			}

			if exist {
				ing := v1beta1.Ingress{}
				ing.Name = ingress.Name
				ing.Spec.Rules = tmpRules
				ings = append(ings, ing)
			}
		}
	}

	return ings
}

func ListApplication(runtimeId string, conditions *params.Conditions, limit, offset int) (*models.PageableResponse, error) {
	if strings.HasSuffix(OpenPitrixServer, "/") {
		OpenPitrixServer = strings.TrimSuffix(OpenPitrixServer, "/")
	}

	defaultStatus := "status=active&status=stopped&status=pending&status=ceased"

	url := fmt.Sprintf("%s/v1/clusters?limit=%s&offset=%s", OpenPitrixServer, strconv.Itoa(limit), strconv.Itoa(offset))

	if len(conditions.Fuzzy["name"]) > 0 {
		url = fmt.Sprintf("%s&search_word=%s", url, conditions.Fuzzy["name"])
	}

	if len(conditions.Match["status"]) > 0 {
		url = fmt.Sprintf("%s&status=%s", url, conditions.Match["status"])
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

	var clusterList clusters
	err = json.Unmarshal(resp, &clusterList)

	if err != nil {
		return nil, err
	}

	result := models.PageableResponse{TotalCount: clusterList.Total}
	result.Items = make([]interface{}, 0)
	for _, item := range clusterList.Clusters {
		var app Application

		app.Name = item.Name
		app.ClusterID = item.ClusterID
		app.UpdateTime = item.UpdateTime
		app.Status = item.Status
		versionInfo, _ := GetVersion(item.VersionID)
		app.Version = versionInfo
		app.VersionId = item.VersionID
		runtimeInfo, _ := GetRuntime(item.RunTimeId)
		app.Runtime = runtimeInfo
		app.RuntimeId = item.RunTimeId
		appInfo, _, appId, _ := GetAppInfo(item.AppID)
		app.App = appInfo
		app.AppId = appId
		app.Description = item.Description

		result.Items = append(result.Items, app)
	}

	return &result, nil
}

func GetApp(clusterId string) (*Application, error) {
	if strings.HasSuffix(OpenPitrixServer, "/") {
		OpenPitrixServer = strings.TrimSuffix(OpenPitrixServer, "/")
	}

	url := fmt.Sprintf("%s/v1/clusters?cluster_id=%s", OpenPitrixServer, clusterId)

	resp, err := makeHttpRequest("GET", url, "")
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	var clusterList clusters
	err = json.Unmarshal(resp, &clusterList)

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	if len(clusterList.Clusters) == 0 {
		return nil, fmt.Errorf("NotFound, clusterId:%s", clusterId)
	}

	item := clusterList.Clusters[0]
	var app Application

	app.Name = item.Name
	app.ClusterID = item.ClusterID
	app.UpdateTime = item.UpdateTime
	app.CreateTime = item.CreateTime
	app.Status = item.Status
	versionInfo, _ := GetVersion(item.VersionID)
	app.Version = versionInfo
	app.VersionId = item.VersionID

	runtimeInfo, _ := GetRuntime(item.RunTimeId)
	app.Runtime = runtimeInfo
	app.RuntimeId = item.RunTimeId
	appInfo, repoId, appId, _ := GetAppInfo(item.AppID)
	app.App = appInfo
	app.AppId = appId
	app.Description = item.Description

	app.RepoName, _ = GetRepo(repoId)

	workloads, err := GetWorkLoads(app.Runtime, item.ClusterRoleSets)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	app.WorkLoads = workloads
	workloadLabels := getLabels(app.Runtime, app.WorkLoads)
	app.Services = getSvcs(app.Runtime, workloadLabels)
	app.Ingresses = getIng(app.Runtime, app.Services)

	return &app, nil
}

func makeHttpRequest(method, url, data string) ([]byte, error) {
	var req *http.Request

	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(data))
	}

	req.Header.Add("Authorization", OpenPitrixProxyToken)

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	if err != nil {
		err := fmt.Errorf("Request to %s failed, method: %s, reason: %s ", url, method, err)
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
