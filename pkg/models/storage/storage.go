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
	"kubesphere.io/kubesphere/pkg/simple/client"
	"strconv"

	"k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/informers"
)

const (
	IsDefaultStorageClassAnnotation     = "storageclass.kubernetes.io/is-default-class"
	betaIsDefaultStorageClassAnnotation = "storageclass.beta.kubernetes.io/is-default-class"
)

type ScMetrics struct {
	Capacity  string `json:"capacity,omitempty"`
	Usage     string `json:"usage,omitempty"`
	PvcNumber string `json:"pvcNumber"`
}

func GetPvcListBySc(scName string) ([]*v1.PersistentVolumeClaim, error) {
	persistentVolumeClaimLister := informers.SharedInformerFactory().Core().V1().PersistentVolumeClaims().Lister()
	all, err := persistentVolumeClaimLister.List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1.PersistentVolumeClaim, 0)

	for _, item := range all {
		if item.Spec.StorageClassName != nil {
			if *item.Spec.StorageClassName == scName {
				result = append(result, item.DeepCopy())
			}
		} else if item.GetAnnotations()[v1.BetaStorageClassAnnotation] == scName {
			result = append(result, item.DeepCopy())
		}
	}
	return result, nil
}

// Get info of metrics
func GetScMetrics(scName string) (*ScMetrics, error) {
	persistentVolumeLister := informers.SharedInformerFactory().Core().V1().PersistentVolumes().Lister()
	pvList, err := persistentVolumeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	// Get PVC
	pvcList, err := GetPvcListBySc(scName)

	if err != nil {
		return nil, err
	}

	// Get storage usage
	// Gathering usage of a specific StorageClass
	var total resource.Quantity
	for _, volume := range pvList {
		if volume.Spec.StorageClassName != scName {
			continue
		}
		total.Add(volume.Spec.Capacity[v1.ResourceStorage])
	}
	usage := total.String()

	// Get PVC number
	pvcNum := len(pvcList)

	return &ScMetrics{Usage: usage, PvcNumber: strconv.Itoa(pvcNum)}, nil
}

// Get SC item list
func GetScList() ([]*storageV1.StorageClass, error) {

	// Get StorageClass list
	scList, err := informers.SharedInformerFactory().Storage().V1().StorageClasses().Lister().List(labels.Everything())

	if err != nil {
		return nil, err
	}

	return scList, nil
}

func SetDefaultStorageClass(defaultScName string) (*storageV1.StorageClass, error) {
	scLister := informers.SharedInformerFactory().Storage().V1().StorageClasses().Lister()
	// 1. verify storage class name
	sc, err := scLister.Get(defaultScName)
	if sc == nil || err != nil {
		return sc, err
	}
	// 2. unset all default sc and then set default sc
	scList, err := scLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	k8sClient := client.ClientSets().K8s().Kubernetes()
	var defaultSc *storageV1.StorageClass
	for _, sc := range scList {
		_, hasDefault := sc.Annotations[IsDefaultStorageClassAnnotation]
		_, hasBeta := sc.Annotations[betaIsDefaultStorageClassAnnotation]
		if sc.Name == defaultScName || hasDefault || hasBeta {
			delete(sc.Annotations, IsDefaultStorageClassAnnotation)
			delete(sc.Annotations, betaIsDefaultStorageClassAnnotation)
			if sc.Name == defaultScName {
				sc.Annotations[IsDefaultStorageClassAnnotation] = "true"
				defaultSc = sc
			}
			_, err := k8sClient.StorageV1().StorageClasses().Update(sc)
			if err != nil {
				return nil, err
			}
		}
	}
	return defaultSc, nil
}
