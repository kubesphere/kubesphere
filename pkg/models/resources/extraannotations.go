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
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/informers"
	"strconv"
)

type extraAnnotationInjector struct {
}

func (i extraAnnotationInjector) addExtraAnnotations(item interface{}) interface{} {

	switch item.(type) {
	case *corev1.PersistentVolumeClaim:
		return i.injectPersistentVolumeClaim(item.(*corev1.PersistentVolumeClaim))
	case *storagev1.StorageClass:
		return i.injectStorageClass(item.(*storagev1.StorageClass))
	}

	return item
}

func (i extraAnnotationInjector) injectStorageClass(item *storagev1.StorageClass) *storagev1.StorageClass {

	count, err := countPvcByStorageClass(item.Name)

	if err != nil {
		klog.Errorf("inject annotation failed %+v", err)
		return item
	}

	item = item.DeepCopy()

	if item.Annotations == nil {
		item.Annotations = make(map[string]string, 0)
	}

	item.Annotations["kubesphere.io/pvc-count"] = strconv.Itoa(count)

	return item
}

func (i extraAnnotationInjector) injectPersistentVolumeClaim(item *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	podLister := informers.SharedInformerFactory().Core().V1().Pods().Lister()
	pods, err := podLister.Pods(item.Namespace).List(labels.Everything())
	if err != nil {
		klog.Errorf("inject annotation failed %+v", err)
		return item
	}

	item = item.DeepCopy()

	if item.Annotations == nil {
		item.Annotations = make(map[string]string, 0)
	}

	if isPvcInUse(pods, item.Name) {
		item.Annotations["kubesphere.io/in-use"] = "true"
	} else {
		item.Annotations["kubesphere.io/in-use"] = "false"
	}

	return item
}

func isPvcInUse(pods []*corev1.Pod, pvcName string) bool {
	for _, pod := range pods {
		volumes := pod.Spec.Volumes
		for _, volume := range volumes {
			pvc := volume.PersistentVolumeClaim
			if pvc != nil && pvc.ClaimName == pvcName {
				return true
			}
		}
	}
	return false
}

func countPvcByStorageClass(scName string) (int, error) {
	persistentVolumeClaimLister := informers.SharedInformerFactory().Core().V1().PersistentVolumeClaims().Lister()
	all, err := persistentVolumeClaimLister.List(labels.Everything())

	if err != nil {
		return 0, err
	}

	count := 0

	for _, item := range all {
		if item.Spec.StorageClassName != nil {
			if *item.Spec.StorageClassName == scName {
				count++
			}
		} else if item.GetAnnotations()[corev1.BetaStorageClassAnnotation] == scName {
			count++
		}
	}
	return count, nil
}
