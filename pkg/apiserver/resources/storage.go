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
	"github.com/emicklei/go-restful"
	"k8s.io/api/core/v1"
	"net/http"

	"kubesphere.io/kubesphere/pkg/models/storage"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

type pvcList struct {
	Name  string                      `json:"name"`
	Items []*v1.PersistentVolumeClaim `json:"items"`
}

type podListByPvc struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Pods      []*v1.Pod `json:"pods"`
}

// List all pods of a specific PVC
// Extended API URL: "GET /api/v1alpha2/namespaces/{namespace}/persistentvolumeclaims/{name}/pods"
func GetPodListByPvc(request *restful.Request, response *restful.Response) {

	pvcName := request.PathParameter("pvc")
	nsName := request.PathParameter("namespace")
	pods, err := storage.GetPodListByPvc(pvcName, nsName)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}
	result := podListByPvc{Name: pvcName, Namespace: nsName, Pods: pods}
	response.WriteAsJson(result)
}

// List all PersistentVolumeClaims of a specific StorageClass
// Extended API URL: "GET /api/v1alpha2/storageclasses/{storageclass}/persistentvolumeclaims"
func GetPvcListBySc(request *restful.Request, response *restful.Response) {
	scName := request.PathParameter("storageclass")
	claims, err := storage.GetPvcListBySc(scName)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	result := pvcList{
		Name: scName, Items: claims,
	}

	response.WriteAsJson(result)
}
