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

package quotas

import (
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/params"
)

const (
	podsKey                   = "count/pods"
	daemonsetsKey             = "count/daemonsets.apps"
	deploymentsKey            = "count/deployments.apps"
	ingressKey                = "count/ingresses.extensions"
	servicesKey               = "count/services"
	statefulsetsKey           = "count/statefulsets.apps"
	persistentvolumeclaimsKey = "persistentvolumeclaims"
	storageClassesKey         = "count/storageClass"
	namespaceKey              = "count/namespace"
	jobsKey                   = "count/jobs.batch"
	cronJobsKey               = "count/cronjobs.batch"
)

var (
	resourceMap = map[string]string{daemonsetsKey: resources.DaemonSets, deploymentsKey: resources.Deployments,
		ingressKey: resources.Ingresses, servicesKey: resources.Services,
		statefulsetsKey: resources.StatefulSets, persistentvolumeclaimsKey: resources.PersistentVolumeClaims, podsKey: resources.Pods,
		namespaceKey: resources.Namespaces, storageClassesKey: resources.StorageClasses,
		jobsKey: resources.Jobs, cronJobsKey: resources.CronJobs}
)

func getUsage(namespace, resource string) (int, error) {
	var result *models.PageableResponse
	var err error
	if resource == resources.Namespaces || resource == resources.StorageClasses {
		result, err = resources.ListResources("", resource, &params.Conditions{}, "", false, 1, 0)
	} else {
		result, err = resources.ListResources(namespace, resource, &params.Conditions{}, "", false, 1, 0)
	}

	if err != nil {
		return 0, err
	}

	return result.TotalCount, nil
}

func GetClusterQuotas() (*models.ResourceQuota, error) {

	quota := v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}

	for k, v := range resourceMap {
		used, err := getUsage("", v)
		if err != nil {
			return nil, err
		}
		var quantity resource.Quantity
		quantity.Set(int64(used))
		quota.Used[v1.ResourceName(k)] = quantity
	}

	return &models.ResourceQuota{Namespace: "\"\"", Data: quota}, nil

}

func GetNamespaceQuotas(namespace string) (*models.ResourceQuota, error) {
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

			used, err := getUsage(namespace, v)
			if err != nil {
				return nil, err
			}
			var quantity resource.Quantity
			quantity.Set(int64(used))
			quota.Used[v1.ResourceName(k)] = quantity
		}
	}

	return &models.ResourceQuota{Namespace: namespace, Data: *quota}, nil
}

func updateNamespaceQuota(tmpResourceList, resourceList v1.ResourceList) {
	if tmpResourceList == nil {
		tmpResourceList = resourceList
	}
	for res, usage := range resourceList {
		tmpUsage, exist := tmpResourceList[res]
		if !exist {
			tmpResourceList[res] = usage
		}
		if tmpUsage.Cmp(usage) == 1 {
			tmpResourceList[res] = usage
		}
	}

}

func getNamespaceResourceQuota(namespace string) (*v1.ResourceQuotaStatus, error) {
	resourceQuotaLister := informers.SharedInformerFactory().Core().V1().ResourceQuotas().Lister()
	quotaList, err := resourceQuotaLister.ResourceQuotas(namespace).List(labels.Everything())
	if err != nil || len(quotaList) == 0 {
		return nil, err
	}

	quotaStatus := v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}

	for _, quota := range quotaList {
		updateNamespaceQuota(quotaStatus.Hard, quota.Status.Hard)
		updateNamespaceQuota(quotaStatus.Used, quota.Status.Used)
	}

	return &quotaStatus, nil
}
