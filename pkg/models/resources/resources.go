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
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"strings"
)

func init() {
	resources[ConfigMaps] = &configMapSearcher{}
	resources[CronJobs] = &cronJobSearcher{}
	resources[DaemonSets] = &daemonSetSearcher{}
	resources[Deployments] = &deploymentSearcher{}
	resources[Ingresses] = &ingressSearcher{}
	resources[Jobs] = &jobSearcher{}
	resources[PersistentVolumeClaims] = &persistentVolumeClaimSearcher{}
	resources[Secrets] = &secretSearcher{}
	resources[Services] = &serviceSearcher{}
	resources[StatefulSets] = &statefulSetSearcher{}
	resources[Pods] = &podSearcher{}
	resources[Roles] = &roleSearcher{}
	resources[S2iBuilders] = &s2iBuilderSearcher{}
	resources[S2iRuns] = &s2iRunSearcher{}

	resources[Nodes] = &nodeSearcher{}
	resources[Namespaces] = &namespaceSearcher{}
	resources[ClusterRoles] = &clusterRoleSearcher{}
	resources[StorageClasses] = &storageClassesSearcher{}
	resources[S2iBuilderTemplates] = &s2iBuilderTemplateSearcher{}
	resources[Workspaces] = &workspaceSearcher{}
}

var (
	resources        = make(map[string]resourceSearchInterface)
	clusterResources = []string{Nodes, Workspaces, Namespaces, ClusterRoles, StorageClasses, S2iBuilderTemplates}
)

const (
	Name                   = "name"
	Label                  = "label"
	OwnerKind              = "ownerKind"
	OwnerName              = "ownerName"
	CreateTime             = "createTime"
	UpdateTime             = "updateTime"
	LastScheduleTime       = "lastScheduleTime"
	chart                  = "chart"
	release                = "release"
	annotation             = "annotation"
	Keyword                = "keyword"
	status                 = "status"
	includeCronJob         = "includeCronJob"
	cronJobKind            = "CronJob"
	s2iRunKind             = "S2iRun"
	includeS2iRun          = "includeS2iRun"
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
	Workspaces             = "workspaces"
	WorkspaceRoles         = "workspaceroles"
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
	S2iBuilderTemplates    = "s2ibuildertemplates"
	S2iBuilders            = "s2ibuilders"
	S2iRuns                = "s2iruns"
)

type resourceSearchInterface interface {
	get(namespace, name string) (interface{}, error)
	search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error)
}

func GetResource(namespace, resource, name string) (interface{}, error) {
	if searcher, ok := resources[resource]; ok {
		resource, err := searcher.get(namespace, name)
		if err != nil {
			glog.Errorln("get resource", namespace, resource, name, err)
			return nil, err
		}
		return resource, nil
	}
	return nil, fmt.Errorf("resource %s not found", resource)
}

func ListResources(namespace, resource string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	items := make([]interface{}, 0)
	var err error
	var result []interface{}

	// none namespace resource
	if namespace != "" && sliceutil.HasString(clusterResources, resource) {
		glog.Errorln("resources not found", resource)
		return nil, fmt.Errorf("not found")
	}

	if searcher, ok := resources[resource]; ok {
		result, err = searcher.search(namespace, conditions, orderBy, reverse)
	} else {
		glog.Errorln("resources not found", resource)
		return nil, fmt.Errorf("not found")
	}

	if err != nil {
		glog.Errorln("resources search", err)
		return nil, err
	}

	for i, d := range result {
		if i >= offset && (limit == -1 || len(items) < limit) {
			items = append(items, d)
		}
	}

	return &models.PageableResponse{TotalCount: len(result), Items: items}, nil
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
