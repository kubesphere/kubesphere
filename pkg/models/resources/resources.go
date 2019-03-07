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
	"regexp"
	"strings"

	"kubesphere.io/kubesphere/pkg/informers"

	"kubesphere.io/kubesphere/pkg/errors"
)

func init() {
	namespacedResources[ConfigMaps] = &configMapSearcher{
		configMapLister: informers.SharedInformerFactory().Core().V1().ConfigMaps().Lister(),
	}
	namespacedResources[CronJobs] = &cronJobSearcher{
		cronJobLister: informers.SharedInformerFactory().Batch().V2alpha1().CronJobs().Lister(),
	}
	namespacedResources[DaemonSets] = &daemonSetSearcher{
		daemonSetLister: informers.SharedInformerFactory().Apps().V1().DaemonSets().Lister(),
	}
	namespacedResources[Deployments] = &deploymentSearcher{
		deploymentLister: informers.SharedInformerFactory().Apps().V1().Deployments().Lister(),
	}
	namespacedResources[Ingresses] = &ingressSearcher{
		ingressLister: informers.SharedInformerFactory().Extensions().V1beta1().Ingresses().Lister(),
	}
	namespacedResources[Jobs] = &jobSearcher{
		jobLister: informers.SharedInformerFactory().Batch().V1().Jobs().Lister(),
	}
	namespacedResources[PersistentVolumeClaims] = &persistentVolumeClaimSearcher{
		persistentVolumeClaimLister: informers.SharedInformerFactory().Core().V1().PersistentVolumeClaims().Lister(),
	}
	namespacedResources[Secrets] = &secretSearcher{
		secretLister: informers.SharedInformerFactory().Core().V1().Secrets().Lister(),
	}
	namespacedResources[Services] = &serviceSearcher{
		serviceLister: informers.SharedInformerFactory().Core().V1().Services().Lister(),
	}
	namespacedResources[StatefulSets] = &statefulSetSearcher{
		statefulSetLister: informers.SharedInformerFactory().Apps().V1().StatefulSets().Lister(),
	}
	namespacedResources[Pods] = &podSearcher{
		podLister: informers.SharedInformerFactory().Core().V1().Pods().Lister(),
	}
	namespacedResources[Roles] = &roleSearcher{
		roleLister: informers.SharedInformerFactory().Rbac().V1().Roles().Lister(),
	}

	clusterResources[Nodes] = &nodeSearcher{
		nodeLister: informers.SharedInformerFactory().Core().V1().Nodes().Lister(),
	}
	clusterResources[Namespaces] = &namespaceSearcher{
		namespaceLister: informers.SharedInformerFactory().Core().V1().Namespaces().Lister(),
	}
	clusterResources[ClusterRoles] = &clusterRoleSearcher{
		clusterRoleLister: informers.SharedInformerFactory().Rbac().V1().ClusterRoles().Lister(),
	}
	clusterResources[StorageClasses] = &storageClassesSearcher{
		storageClassesLister: informers.SharedInformerFactory().Storage().V1().StorageClasses().Lister(),
	}
}

var namespacedResources = make(map[string]namespacedSearcherInterface)
var clusterResources = make(map[string]clusterSearcherInterface)

type conditions struct {
	match map[string]string
	fuzzy map[string]string
}

const (
	name                   = "name"
	label                  = "label"
	createTime             = "createTime"
	updateTime             = "updateTime"
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
	search(namespace string, conditions *conditions, orderBy string, reverse bool) ([]interface{}, error)
}
type clusterSearcherInterface interface {
	search(conditions *conditions, orderBy string, reverse bool) ([]interface{}, error)
}

func ListNamespaceResource(namespace, resource, conditionStr, orderBy string, reverse bool, limit, offset int) (*ResourceList, error) {
	items := make([]interface{}, 0)
	total := 0
	var err error

	conditions, err := parseToConditions(conditionStr)

	if err != nil {
		return nil, err
	}

	var result []interface{}

	if searcher, ok := namespacedResources[resource]; ok {
		result, err = searcher.search(namespace, conditions, orderBy, reverse)
	} else {
		return nil, errors.New(errors.NotImplement, "not support")
	}

	if err != nil {
		return nil, errors.New(errors.Internal, err.Error())
	}

	total = len(result)

	for i, d := range result {
		if i >= offset && (limit == -1 || len(items) < limit) {
			items = append(items, d)
		}
	}

	return &ResourceList{TotalCount: total, Items: items}, nil
}

func ListClusterResource(resource, conditionStr, orderBy string, reverse bool, limit, offset int) (*ResourceList, error) {
	items := make([]interface{}, 0)
	total := 0
	var err error

	conditions, err := parseToConditions(conditionStr)

	if err != nil {
		return nil, err
	}

	var result []interface{}

	if searcher, ok := clusterResources[resource]; ok {
		result, err = searcher.search(conditions, orderBy, reverse)
	} else {
		return nil, errors.New(errors.NotImplement, "not support")
	}

	if err != nil {
		return nil, errors.New(errors.Internal, err.Error())
	}

	total = len(result)

	for i, d := range result {
		if i >= offset && len(items) < limit {
			items = append(items, d)
		}
	}

	return &ResourceList{TotalCount: total, Items: items}, nil
}

func parseToConditions(str string) (*conditions, error) {
	conditions := &conditions{match: make(map[string]string, 0), fuzzy: make(map[string]string, 0)}

	if str == "" {
		return conditions, nil
	}

	for _, item := range strings.Split(str, ",") {
		if strings.Count(item, "=") > 1 || strings.Count(item, "~") > 1 {
			return nil, errors.New(errors.InvalidArgument, "invalid condition")
		}
		if groups := regexp.MustCompile(`(\S+)([=~])(\S+)`).FindStringSubmatch(item); len(groups) == 4 {
			if groups[2] == "=" {
				conditions.match[groups[1]] = groups[3]
			} else {
				conditions.fuzzy[groups[1]] = groups[3]
			}
		} else {
			return nil, errors.New(errors.InvalidArgument, "invalid condition")
		}
	}
	return conditions, nil
}

type ResourceList struct {
	TotalCount int           `json:"total_count"`
	Items      []interface{} `json:"items"`
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
