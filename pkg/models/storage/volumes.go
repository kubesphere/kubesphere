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

package storage

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/informers"
)

// List pods of a specific persistent volume claims
func GetPodListByPvc(pvc string, ns string) (res []*v1.Pod, err error) {
	podLister := informers.SharedInformerFactory().Core().V1().Pods().Lister()
	podList, err := podLister.Pods(ns).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, pod := range podList {
		if IsPvcInPod(pod, pvc) == true {
			res = append(res, pod.DeepCopy())
		}
	}
	return res, nil
}

// Check if the persistent volume claim is related to the pod
func IsPvcInPod(pod *v1.Pod, pvcName string) bool {
	for _, v := range pod.Spec.Volumes {
		if v.VolumeSource.PersistentVolumeClaim != nil &&
			v.VolumeSource.PersistentVolumeClaim.ClaimName == pvcName {
			return true
		}
	}
	return false
}
