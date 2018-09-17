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

package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/kubesphere/pkg/client"
)

const (
	unknown       = "-"
	deploySurffix = "-Deployment"
	daemonSurffix = "-DaemonSet"
	stateSurffix  = "-StatefulSet"
)

type ApplicationCtl struct {
	OpenpitrixAddr string
}

type Application struct {
	Name        string     `json:"name"`
	RepoName    string     `json:"repoName"`
	Runtime     string     `json:"namespace"`
	RuntimeId   string     `json:"runtime_id"`
	Version     string     `json:"version"`
	VersionId   string     `json:"version_id"`
	Status      string     `json:"status"`
	UpdateTime  time.Time  `json:"updateTime"`
	CreateTime  time.Time  `json:"createTime"`
	App         string     `json:"app"`
	AppId       string     `json:"app_id"`
	Description string     `json:"description,omitempty"`
	WorkLoads   *workLoads `json:"workloads,omitempty"`
	Services    *[]Service `json:"services,omitempty"`
	Ingresses   *[]ing     `json:"ingresses,omitempty"`
	ClusterID   string     `json:"cluster_id"`
}

type ing struct {
	Name  string        `json:"name"`
	Rules []ingressRule `json:"rules"`
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
	Deployments  []Deployment  `json:"deployments,omitempty"`
	Statefulsets []Statefulset `json:"statefulsets,omitempty"`
	Daemonsets   []Daemonset   `json:"daemonsets,omitempty"`
}

//type description struct {
//	Creator string `json:"creator"`
//}

type appList struct {
	Total int   `json:"total_count"`
	Apps  []app `json:"app_set"`
}

type repoList struct {
	Total int    `json:"total_count"`
	Repos []repo `json:"repo_set"`
}

func (ctl *ApplicationCtl) GetAppInfo(appId string) (string, string, string, error) {
	url := fmt.Sprintf("%s/v1/apps?app_id=%s", ctl.OpenpitrixAddr, appId)
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

func (ctl *ApplicationCtl) GetRepo(repoId string) (string, error) {
	url := fmt.Sprintf("%s/v1/repos?repo_id=%s", ctl.OpenpitrixAddr, repoId)
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

func (ctl *ApplicationCtl) GetVersion(versionId string) (string, error) {
	versionUrl := fmt.Sprintf("%s/v1/app_versions?version_id=%s", ctl.OpenpitrixAddr, versionId)
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

func (ctl *ApplicationCtl) GetRuntime(runtimeId string) (string, error) {

	versionUrl := fmt.Sprintf("%s/v1/runtimes?runtime_id=%s", ctl.OpenpitrixAddr, runtimeId)
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

func (ctl *ApplicationCtl) GetWorkLoads(namespace string, clusterRoles []clusterRole) *workLoads {

	var works workLoads
	for _, clusterRole := range clusterRoles {
		workLoadName := clusterRole.Role
		if len(workLoadName) > 0 {
			if strings.HasSuffix(workLoadName, deploySurffix) {
				name := strings.Split(workLoadName, deploySurffix)[0]
				ctl := ResourceControllers.Controllers[Deployments]
				_, items, _ := ctl.ListWithConditions(fmt.Sprintf("namespace='%s' and name = '%s'", namespace, name), nil, "")
				works.Deployments = append(works.Deployments, items.([]Deployment)...)
				continue
			}

			if strings.HasSuffix(workLoadName, daemonSurffix) {
				name := strings.Split(workLoadName, daemonSurffix)[0]
				ctl := ResourceControllers.Controllers[Daemonsets]
				_, items, _ := ctl.ListWithConditions(fmt.Sprintf("namespace='%s' and name = '%s'", namespace, name), nil, "")
				works.Daemonsets = append(works.Daemonsets, items.([]Daemonset)...)
				continue
			}

			if strings.HasSuffix(workLoadName, stateSurffix) {
				name := strings.Split(workLoadName, stateSurffix)[0]
				ctl := ResourceControllers.Controllers[Statefulsets]
				_, items, _ := ctl.ListWithConditions(fmt.Sprintf("namespace='%s' and name = '%s'", namespace, name), nil, "")
				works.Statefulsets = append(works.Statefulsets, items.([]Statefulset)...)
				continue
			}
		}
	}
	return &works
}

func (ctl *ApplicationCtl) getLabels(namespace string, workloads *workLoads) *[]map[string]string {
	k8sClient := client.NewK8sClient()

	var workloadLables []map[string]string
	if workloads == nil {
		return nil
	}

	for _, workload := range workloads.Deployments {
		deploy, err := k8sClient.AppsV1().Deployments(namespace).Get(workload.Name, metaV1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, deploy.Labels)
	}

	for _, workload := range workloads.Daemonsets {
		daemonset, err := k8sClient.AppsV1().DaemonSets(namespace).Get(workload.Name, metaV1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, daemonset.Labels)
	}

	for _, workload := range workloads.Statefulsets {
		statefulset, err := k8sClient.AppsV1().StatefulSets(namespace).Get(workload.Name, metaV1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		workloadLables = append(workloadLables, statefulset.Labels)
	}

	return &workloadLables
}

func isExist(svcs []Service, svc v1.Service) bool {
	for _, item := range svcs {
		if item.Name == svc.Name && item.Namespace == svc.Namespace {
			return true
		}
	}
	return false
}

func (ctl *ApplicationCtl) getSvcs(namespace string, workLoadLabels *[]map[string]string) *[]Service {
	if len(*workLoadLabels) == 0 {
		return nil
	}
	k8sClient := client.NewK8sClient()
	var services []Service
	for _, label := range *workLoadLabels {
		labelSelector := labels.Set(label).AsSelector().String()
		svcs, err := k8sClient.CoreV1().Services(namespace).List(metaV1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			glog.Errorf("get app's svc failed, reason: %v", err)
		}
		for _, item := range svcs.Items {
			if !isExist(services, item) {
				services = append(services, *generateSvcObject(item))
			}
		}
	}

	return &services
}

func (ctl *ApplicationCtl) getIng(namespace string, services *[]Service) *[]ing {
	if services == nil {
		return nil
	}

	ingCtl := ResourceControllers.Controllers[Ingresses]
	var ings []ing
	for _, svc := range *services {
		_, items, err := ingCtl.ListWithConditions(fmt.Sprintf("namespace = '%s' and rules like '%%%s%%' ", namespace, svc.Name), nil, "")
		if err != nil {
			glog.Error(err)
			return nil
		}

		glog.Error(items)
		for _, ingress := range items.([]Ingress) {
			var rules []ingressRule
			err := json.Unmarshal([]byte(ingress.Rules), &rules)
			if err != nil {
				return nil
			}

			exist := false
			var tmpRules []ingressRule
			for _, rule := range rules {
				if rule.Service == svc.Name {
					exist = true
					tmpRules = append(tmpRules, rule)
				}
			}

			if exist {
				ings = append(ings, ing{Name: ingress.Name, Rules: tmpRules})
			}
		}
	}

	return &ings
}

func (ctl *ApplicationCtl) ListApplication(runtimeId string, match, fuzzy map[string]string, paging *Paging) (int, interface{}, error) {
	limit := paging.Limit
	offset := paging.Offset
	if strings.HasSuffix(ctl.OpenpitrixAddr, "/") {
		ctl.OpenpitrixAddr = strings.TrimSuffix(ctl.OpenpitrixAddr, "/")
	}

	defaultStatus := "status=active&status=stopped&status=pending&status=ceased"

	url := fmt.Sprintf("%s/v1/clusters?limit=%s&offset=%s", ctl.OpenpitrixAddr, strconv.Itoa(limit), strconv.Itoa(offset))

	if len(fuzzy["name"]) > 0 {
		url = fmt.Sprintf("%s&search_word=%s", url, fuzzy["name"])
	}

	if len(match["status"]) > 0 {
		url = fmt.Sprintf("%s&status=%s", url, match["status"])
	} else {
		url = fmt.Sprintf("%s&%s", url, defaultStatus)
	}

	if len(runtimeId) > 0 {
		url = fmt.Sprintf("%s&runtime_id=%s", url, runtimeId)
	}

	resp, err := makeHttpRequest("GET", url, "")
	if err != nil {
		glog.Errorf("request %s failed, reason: %s", url, err)
		return 0, nil, err
	}

	var clusterList clusters
	err = json.Unmarshal(resp, &clusterList)

	if err != nil {
		return 0, nil, err
	}

	var apps []Application

	for _, item := range clusterList.Clusters {
		var app Application

		app.Name = item.Name
		app.ClusterID = item.ClusterID
		app.UpdateTime = item.UpdateTime
		app.Status = item.Status
		versionInfo, _ := ctl.GetVersion(item.VersionID)
		app.Version = versionInfo
		app.VersionId = item.VersionID
		runtimeInfo, _ := ctl.GetRuntime(item.RunTimeId)
		app.Runtime = runtimeInfo
		app.RuntimeId = item.RunTimeId
		appInfo, _, appId, _ := ctl.GetAppInfo(item.AppID)
		app.App = appInfo
		app.AppId = appId

		apps = append(apps, app)
	}

	return clusterList.Total, apps, nil
}

func (ctl *ApplicationCtl) GetApp(clusterId string) (*Application, error) {
	if strings.HasSuffix(ctl.OpenpitrixAddr, "/") {
		ctl.OpenpitrixAddr = strings.TrimSuffix(ctl.OpenpitrixAddr, "/")
	}

	url := fmt.Sprintf("%s/v1/clusters?cluster_id=%s", ctl.OpenpitrixAddr, clusterId)

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
	versionInfo, _ := ctl.GetVersion(item.VersionID)
	app.Version = versionInfo
	app.VersionId = item.VersionID

	runtimeInfo, _ := ctl.GetRuntime(item.RunTimeId)
	app.Runtime = runtimeInfo
	app.RuntimeId = item.RunTimeId
	appInfo, repoId, appId, _ := ctl.GetAppInfo(item.AppID)
	app.App = appInfo
	app.AppId = appId
	app.Description = item.Description

	app.RepoName, _ = ctl.GetRepo(repoId)
	app.WorkLoads = ctl.GetWorkLoads(app.Runtime, item.ClusterRoleSets)
	workloadLabels := ctl.getLabels(app.Runtime, app.WorkLoads)
	app.Services = ctl.getSvcs(app.Runtime, workloadLabels)
	app.Ingresses = ctl.getIng(app.Runtime, app.Services)

	return &app, nil
}
