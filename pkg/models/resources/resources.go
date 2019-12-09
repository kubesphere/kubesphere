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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
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
	resources[HorizontalPodAutoscalers] = &hpaSearcher{}
	resources[Applications] = &appSearcher{}

	resources[Nodes] = &nodeSearcher{}
	resources[Namespaces] = &namespaceSearcher{}
	resources[ClusterRoles] = &clusterRoleSearcher{}
	resources[StorageClasses] = &storageClassesSearcher{}
	resources[S2iBuilderTemplates] = &s2iBuilderTemplateSearcher{}
	resources[Workspaces] = &workspaceSearcher{}
}

var (
	injector         = extraAnnotationInjector{}
	resources        = make(map[string]resourceSearchInterface)
	clusterResources = []string{Nodes, Workspaces, Namespaces, ClusterRoles, StorageClasses, S2iBuilderTemplates}
)

const (
	Name                     = "name"
	Label                    = "label"
	OwnerKind                = "ownerKind"
	OwnerName                = "ownerName"
	TargetKind               = "targetKind"
	TargetName               = "targetName"
	Role                     = "role"
	CreateTime               = "createTime"
	UpdateTime               = "updateTime"
	StartTime                = "startTime"
	LastScheduleTime         = "lastScheduleTime"
	chart                    = "chart"
	release                  = "release"
	annotation               = "annotation"
	Keyword                  = "keyword"
	UserFacing               = "userfacing"
	Status                   = "status"
	includeCronJob           = "includeCronJob"
	storageClassName         = "storageClassName"
	cronJobKind              = "CronJob"
	s2iRunKind               = "S2iRun"
	includeS2iRun            = "includeS2iRun"
	StatusRunning            = "running"
	StatusPaused             = "paused"
	StatusPending            = "pending"
	StatusUpdating           = "updating"
	StatusStopped            = "stopped"
	StatusFailed             = "failed"
	StatusBound              = "bound"
	StatusLost               = "lost"
	StatusComplete           = "complete"
	StatusWarning            = "warning"
	StatusUnschedulable      = "unschedulable"
	app                      = "app"
	Deployments              = "deployments"
	DaemonSets               = "daemonsets"
	Roles                    = "roles"
	Workspaces               = "workspaces"
	WorkspaceRoles           = "workspaceroles"
	CronJobs                 = "cronjobs"
	ConfigMaps               = "configmaps"
	Ingresses                = "ingresses"
	Jobs                     = "jobs"
	PersistentVolumeClaims   = "persistentvolumeclaims"
	Pods                     = "pods"
	Secrets                  = "secrets"
	Services                 = "services"
	StatefulSets             = "statefulsets"
	HorizontalPodAutoscalers = "horizontalpodautoscalers"
	Applications             = "applications"
	Nodes                    = "nodes"
	Namespaces               = "namespaces"
	StorageClasses           = "storageclasses"
	ClusterRoles             = "clusterroles"
	S2iBuilderTemplates      = "s2ibuildertemplates"
	S2iBuilders              = "s2ibuilders"
	S2iRuns                  = "s2iruns"
)

type resourceSearchInterface interface {
	get(namespace, name string) (interface{}, error)
	search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error)
}

func GetResource(namespace, resource, name string) (interface{}, error) {
	if searcher, ok := resources[resource]; ok {
		resource, err := searcher.get(namespace, name)
		if err != nil {
			klog.Errorf("resource %s.%s.%s not found: %s", namespace, resource, name, err)
			return nil, err
		}
		return resource, nil
	}
	return nil, fmt.Errorf("resource %s.%s.%s not found", namespace, resource, name)
}

func ListResources(namespace, resource string, conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	items := make([]interface{}, 0)
	var err error
	var result []interface{}

	// none namespace resource
	if namespace != "" && sliceutil.HasString(clusterResources, resource) {
		err = fmt.Errorf("namespaced resource %s not found", resource)
		klog.Errorln(err)
		return nil, err
	}

	if searcher, ok := resources[resource]; ok {
		result, err = searcher.search(namespace, conditions, orderBy, reverse)
	} else {
		err = fmt.Errorf("namespaced resource %s not found", resource)
		klog.Errorln(err)
		return nil, err
	}

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	if limit == -1 || limit+offset > len(result) {
		limit = len(result) - offset
	}

	result = result[offset : offset+limit]
	for _, item := range result {
		items = append(items, injector.addExtraAnnotations(item))
	}

	return &models.PageableResponse{TotalCount: len(result), Items: items}, nil
}

func searchFuzzy(m map[string]string, key, value string) bool {

	val, exist := m[key]

	if value == "" && (!exist || val == "") {
		return true
	} else if value != "" && strings.Contains(val, value) {
		return true
	}

	return false
}
