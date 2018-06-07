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

package cronjobs

import (
	"encoding/json"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/jobs/resources"
	"time"
)

const (
	pods                   = "count/pods"
	daemonsets             = "count/daemonsets.apps"
	deployments            = "count/deployments.apps"
	ingress                = "count/ingresses.extensions"
	roles                  = "count/roles.rbac.authorization.k8s.io"
	services               = "count/services"
	statefulsets           = "count/statefulsets.apps"
	persistentvolumeclaims = "persistentvolumeclaims"
)

type resourceUsage struct {
	NameSpace       string
	Data            v1.ResourceQuotaStatus
	UpdateTimeStamp int64
}

type resourceQuotaWorker struct {
	k8sClient *kubernetes.Clientset
	resChan   chan dataType
	stopChan  chan struct{}
}

func (ru resourceUsage) namespace() string {
	return ru.NameSpace
}

type workloadList map[string][]resources.WorkLoadObject

type otherResourceList map[string][]resources.OtherResourceObject

type workload struct {
	ResourceType    string       `json:"type"`
	ResourceList    workloadList `json:"lists"`
	UpdateTimeStamp int64        `json:"updateTimestamp"`
}

type otherResource struct {
	ResourceType    string            `json:"type"`
	ResourceList    otherResourceList `json:"lists"`
	UpdateTimeStamp int64             `json:"updateTimestamp"`
}

var workLoads = []string{"deployments", "daemonsets", "statefulsets"}

var resourceMap = map[string]string{daemonsets: "daemonsets", deployments: "deployments", ingress: "ingresses",
	roles: "roles", services: "services", statefulsets: "statefulsets", persistentvolumeclaims: "persistent-volume-claim", pods:"pods"}

func contain(items []string, item string) bool {
	for _, v := range items {
		if v == item {
			return false
		}
	}
	return true
}

func (rw *resourceQuotaWorker) getResourceusage(namespace, resourceName string) (int, error) {

	etcdcli, err := client.NewEtcdClient()
	if err != nil {
		glog.Error(err)
		return 0, err
	}

	defer etcdcli.Close()
	key := constants.Root + "/" + resourceName
	value, err := etcdcli.Get(key)
	if err != nil {
		glog.Error(err)
	}

	if contain(workLoads, resourceName) {
		resourceStatus := workload{ResourceList: make(workloadList)}

		err := json.Unmarshal(value, &resourceStatus)
		if err != nil {
			glog.Error(err)
			return 0, nil
		}

		return len(resourceStatus.ResourceList[namespace]), nil
	} else {
		resourceStatus := otherResource{ResourceList: make(otherResourceList)}

		err := json.Unmarshal(value, &resourceStatus)
		if err != nil {
			glog.Error(err)
			return 0, err
		}

		return len(resourceStatus.ResourceList[namespace]), nil
	}

	return 0, nil
}

func (rw *resourceQuotaWorker) updateNamespaceQuota(tmpResourceList, resourceList v1.ResourceList) {
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

func (rw *resourceQuotaWorker) getNamespaceResourceUsageByQuota(namespace string) (*v1.ResourceQuotaStatus, error) {
	quotaList, err := rw.k8sClient.CoreV1().ResourceQuotas(namespace).List(meta_v1.ListOptions{})
	if err != nil || len(quotaList.Items) == 0 {
		return nil, err
	}

	quotaStatus := v1.ResourceQuotaStatus{Hard: make(v1.ResourceList), Used: make(v1.ResourceList)}

	for _, quota := range quotaList.Items {
		rw.updateNamespaceQuota(quotaStatus.Hard, quota.Status.Hard)
		rw.updateNamespaceQuota(quotaStatus.Used, quota.Status.Used)
	}

	return &quotaStatus, nil
}

func (rw *resourceQuotaWorker) getNamespaceQuota(namespace string) (v1.ResourceQuotaStatus, error) {
	quota, err := rw.getNamespaceResourceUsageByQuota(namespace)
	if err != nil {
		return v1.ResourceQuotaStatus{}, err
	}

	if quota == nil {
		quota = new(v1.ResourceQuotaStatus)
		quota.Used = make(v1.ResourceList)
	}

	for k, v := range resourceMap {
		if _, exist := quota.Used[v1.ResourceName(k)]; !exist {
			used, err := rw.getResourceusage(namespace, v)
			if err != nil {
				continue
			}

			var quantity resource.Quantity
			quantity.Set(int64(used))
			quota.Used[v1.ResourceName(k)] = quantity
		}
	}

	return *quota, nil
}

func (rw *resourceQuotaWorker) workOnce() {
	clusterQuota := new(v1.ResourceQuotaStatus)
	clusterQuota.Used = make(v1.ResourceList)
	namespaces, err := rw.k8sClient.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, ns := range namespaces.Items {
		namespace := ns.Name
		nsquota, err := rw.getNamespaceQuota(namespace)
		if err != nil {
			glog.Error(err)
			return
		}
		res := resourceUsage{NameSpace: namespace, Data: nsquota, UpdateTimeStamp: time.Now().Unix()}
		rw.resChan <- res

		for k, v := range nsquota.Used {
			tmp := clusterQuota.Used[k]
			tmp.Add(v)
			clusterQuota.Used[k] = tmp
		}
	}

	var quantity resource.Quantity
	quantity.Set(int64(len(namespaces.Items)))

	clusterQuota.Used["count/namespaces"] = quantity
	res := resourceUsage{NameSpace: "\"\"", Data: *clusterQuota, UpdateTimeStamp: time.Now().Unix()}
	rw.resChan <- res

}

func (rw *resourceQuotaWorker) chanStop() chan struct{} {
	return rw.stopChan
}

func (rw *resourceQuotaWorker) chanRes() chan dataType {
	return rw.resChan
}
