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
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/models/controllers"
)

const (
	podsKey                   = "count/pods"
	daemonsetsKey             = "count/daemonsets.apps"
	deploymentsKey            = "count/deployments.apps"
	ingressKey                = "count/ingresses.extensions"
	rolesKey                  = "count/roles.rbac.authorization.k8s.io"
	clusterRolesKey           = "count/cluster-role"
	servicesKey               = "count/services"
	statefulsetsKey           = "count/statefulsets.apps"
	persistentvolumeclaimsKey = "persistentvolumeclaims"
	storageClassesKey         = "count/storageClass"
	namespaceKey              = "count/namespace"
	jobsKey                   = "count/jobs.batch"
	cronJobsKey               = "count/cronjobs.batch"
)

var resourceMap = map[string]string{daemonsetsKey: controllers.Daemonsets, deploymentsKey: controllers.Deployments,
	ingressKey: controllers.Ingresses, rolesKey: controllers.Roles, servicesKey: controllers.Services,
	statefulsetsKey: controllers.Statefulsets, persistentvolumeclaimsKey: controllers.PersistentVolumeClaim, podsKey: controllers.Pods,
	namespaceKey: controllers.Namespaces, storageClassesKey: controllers.StorageClasses, clusterRolesKey: controllers.ClusterRoles,
	jobsKey: controllers.Jobs, cronJobsKey: controllers.Cronjobs}

type ResourceQuota struct {
	NameSpace string                 `json:"namespace"`
	Data      v1.ResourceQuotaStatus `json:"data"`
}

func getUsage(namespace, resource string) int {
	ctl, err := getController(resource)
	if err != nil {
		return 0
	}

	if len(namespace) == 0 {
		return ctl.CountWithConditions("")
	}

	return ctl.CountWithConditions(fmt.Sprintf("namespace = '%s' ", namespace))
}

func GetClusterQuota() (*ResourceQuota, error) {

	quota := v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}
	for k, v := range resourceMap {
		used := getUsage("", v)
		var quantity resource.Quantity
		quantity.Set(int64(used))
		quota.Used[v1.ResourceName(k)] = quantity
	}

	return &ResourceQuota{NameSpace: "\"\"", Data: quota}, nil

}

func GetNamespaceQuota(namespace string) (*ResourceQuota, error) {
	quota, err := getNamespaceResourceQuota(namespace)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	if quota == nil {
		quota = &v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}
	}

	for k, v := range resourceMap {
		if _, exist := quota.Used[v1.ResourceName(k)]; !exist {
			if k == namespaceKey || k == storageClassesKey {
				continue
			}

			used := getUsage(namespace, v)
			var quantity resource.Quantity
			quantity.Set(int64(used))
			quota.Used[v1.ResourceName(k)] = quantity
		}
	}

	return &ResourceQuota{NameSpace: namespace, Data: *quota}, nil
}

func updateNamespaceQuota(tmpResourceList, resourceList v1.ResourceList) {
	if tmpResourceList == nil {
		tmpResourceList = resourceList
	}
	for resource, usage := range resourceList {
		tmpUsage, exist := tmpResourceList[resource]
		if !exist {
			tmpResourceList[resource] = usage
		}
		if tmpUsage.Cmp(usage) == 1 {
			tmpResourceList[resource] = usage
		}
	}

}

func getNamespaceResourceQuota(namespace string) (*v1.ResourceQuotaStatus, error) {
	quotaList, err := client.NewK8sClient().CoreV1().ResourceQuotas(namespace).List(metaV1.ListOptions{})
	if err != nil || len(quotaList.Items) == 0 {
		return nil, err
	}

	quotaStatus := v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}

	for _, quota := range quotaList.Items {
		updateNamespaceQuota(quotaStatus.Hard, quota.Status.Hard)
		updateNamespaceQuota(quotaStatus.Used, quota.Status.Used)
	}

	return &quotaStatus, nil
}
