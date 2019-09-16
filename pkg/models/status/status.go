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
package status

import (
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"strings"

	"kubesphere.io/kubesphere/pkg/models/resources"
)

type WorkLoadStatus struct {
	Namespace string                 `json:"namespace" description:"the name of the namespace"`
	Count     map[string]int         `json:"data" description:"the number of unhealthy workloads"`
	Items     map[string]interface{} `json:"items,omitempty" description:"unhealthy workloads"`
}

func GetNamespacesResourceStatus(namespace string) (*WorkLoadStatus, error) {
	res := WorkLoadStatus{Count: make(map[string]int), Namespace: namespace, Items: make(map[string]interface{})}
	var notReadyList *models.PageableResponse
	var err error
	for _, resource := range []string{resources.Deployments, resources.StatefulSets, resources.DaemonSets, resources.PersistentVolumeClaims, resources.Jobs} {
		var notReadyStatus string

		switch resource {
		case resources.PersistentVolumeClaims:
			notReadyStatus = strings.Join([]string{resources.StatusPending, resources.StatusLost}, "|")
		case resources.Jobs:
			notReadyStatus = resources.StatusFailed
		default:
			notReadyStatus = resources.StatusUpdating
		}

		notReadyList, err = resources.ListResources(namespace, resource, &params.Conditions{Match: map[string]string{resources.Status: notReadyStatus}}, "", false, -1, 0)

		if err != nil {
			klog.Errorf("list resources failed: %+v", err)
			return nil, err
		}

		res.Count[resource] = notReadyList.TotalCount
	}

	return &res, nil
}

func GetClusterResourceStatus() (*WorkLoadStatus, error) {

	return GetNamespacesResourceStatus("")
}
