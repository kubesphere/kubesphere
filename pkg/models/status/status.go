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
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/params"

	"kubesphere.io/kubesphere/pkg/models/resources"
)

type workLoadStatus struct {
	Namespace string                 `json:"namespace"`
	Count     map[string]int         `json:"data"`
	Items     map[string]interface{} `json:"items,omitempty"`
}

func GetNamespacesResourceStatus(namespace string) (*workLoadStatus, error) {
	res := workLoadStatus{Count: make(map[string]int), Namespace: namespace, Items: make(map[string]interface{})}
	var notReadyList *models.PageableResponse
	var err error
	for _, resource := range []string{resources.Deployments, resources.StatefulSets, resources.DaemonSets, resources.PersistentVolumeClaims} {
		notReadyStatus := "updating"
		if resource == resources.PersistentVolumeClaims {
			notReadyStatus = "pending"
		}

		notReadyList, err = resources.ListNamespaceResource(namespace, resource, &params.Conditions{Match: map[string]string{"status": notReadyStatus}}, "", false, -1, 0)

		if err != nil {
			return nil, err
		}

		res.Count[resource] = notReadyList.TotalCount
	}

	return &res, nil
}

func GetClusterResourceStatus() (*workLoadStatus, error) {

	return GetNamespacesResourceStatus("")
}
