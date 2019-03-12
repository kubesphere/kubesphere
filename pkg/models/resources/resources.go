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
package resources

import (
	"fmt"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/params"
	"strings"
)

func init() {
	namespacedResources[ConfigMaps] = &configMapSearcher{}
	namespacedResources[CronJobs] = &cronJobSearcher{}
	namespacedResources[DaemonSets] = &daemonSetSearcher{}
	namespacedResources[Deployments] = &deploymentSearcher{}
	namespacedResources[Ingresses] = &ingressSearcher{}
	namespacedResources[Jobs] = &jobSearcher{}
	namespacedResources[PersistentVolumeClaims] = &persistentVolumeClaimSearcher{}
	namespacedResources[Secrets] = &secretSearcher{}
	namespacedResources[Services] = &serviceSearcher{}
	namespacedResources[StatefulSets] = &statefulSetSearcher{}
	namespacedResources[Pods] = &podSearcher{}
	namespacedResources[Roles] = &roleSearcher{}

	clusterResources[Nodes] = &nodeSearcher{}
	clusterResources[Namespaces] = &namespaceSearcher{}
	clusterResources[ClusterRoles] = &clusterRoleSearcher{}
	clusterResources[StorageClasses] = &storageClassesSearcher{}
}

var namespacedResources = make(map[string]namespacedSearcherInterface)
var clusterResources = make(map[string]clusterSearcherInterface)

const (
	name                   = "name"
	label                  = "label"
	createTime             = "createTime"
	updateTime             = "updateTime"
	lastScheduleTime       = "lastScheduleTime"
	displayName            = "displayName"
	chart                  = "chart"
	release                = "release"
	annotation             = "annotation"
	keyword                = "keyword"
	status                 = "status"
	running                = "running"
	paused                 = "paused"
	updating               = "updating"
	stopped                = "stopped"
	failed                 = "failed"
	complete               = "complete"
	app                    = "app"
	Deployments            = "deployments"
	DaemonSets             = "daemonsets"
	Roles                  = "roles"
	CronJobs               = "cronjobs"
	ConfigMaps             = "configmaps"
	Ingresses              = "ingresses"
	Jobs                   = "jobs"
	PersistentVolumeClaims = "persistentvolumeclaims"
	Pods                   = "pods"
	Secrets                = "secrets"
	Services               = "services"
	StatefulSets           = "statefulsets"
	Nodes                  = "nodes"
	Namespaces             = "namespaces"
	StorageClasses         = "storageclasses"
	ClusterRoles           = "clusterroles"
)

type namespacedSearcherInterface interface {
	search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error)
}
type clusterSearcherInterface interface {
	search(conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error)
}

func ListNamespaceResource(namespace, resource string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	items := make([]interface{}, 0)
	total := 0
	var err error
	var result []interface{}

	if searcher, ok := namespacedResources[resource]; ok {
		result, err = searcher.search(namespace, conditions, orderBy, reverse)
	} else {
		return nil, fmt.Errorf("not support")
	}

	if err != nil {
		return nil, err
	}

	total = len(result)

	for i, d := range result {
		if i >= offset && (limit == -1 || len(items) < limit) {
			items = append(items, d)
		}
	}

	return &models.PageableResponse{TotalCount: total, Items: items}, nil
}

func ListClusterResource(resource string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	items := make([]interface{}, 0)
	total := 0
	var err error

	if err != nil {
		return nil, err
	}

	var result []interface{}

	if searcher, ok := clusterResources[resource]; ok {
		result, err = searcher.search(conditions, orderBy, reverse)
	} else {
		return nil, fmt.Errorf("not support")
	}

	if err != nil {
		return nil, err
	}

	total = len(result)

	for i, d := range result {
		if i >= offset && len(items) < limit {
			items = append(items, d)
		}
	}

	return &models.PageableResponse{TotalCount: total, Items: items}, nil
}

func searchFuzzy(m map[string]string, key, value string) bool {
	for k, v := range m {
		if key == "" {
			if strings.Contains(k, value) || strings.Contains(v, value) {
				return true
			}
		} else if k == key && strings.Contains(v, value) {
			return true
		}
	}
	return false
}
