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

package expansion

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func getPodMountPVC(pods []*corev1.Pod, pvc string) []*corev1.Pod {
	var res []*corev1.Pod
	for _, pod := range pods {
		klog.V(4).Infof("check pod %s is mount pvc %s", pod.Name, pvc)
		curPod := pod
		if isMounted(pod, pvc) {
			res = append(res, curPod)
		}
	}
	return res
}

func isMounted(pod *corev1.Pod, pvc string) bool {
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == pvc {
			return true
		}
	}
	return false
}

// claimToClaimKey return namespace/name string for pvc
func claimToClaimKey(claim *corev1.PersistentVolumeClaim) string {
	return fmt.Sprintf("%s/%s", claim.Namespace, claim.Name)
}

// GetPersistentVolumeClaimClass returns StorageClassName. If no storage class was
// requested, it returns "".
func getPersistentVolumeClaimClass(claim *corev1.PersistentVolumeClaim) string {
	// Use beta annotation first
	if class, found := claim.Annotations[corev1.BetaStorageClassAnnotation]; found {
		return class
	}

	if claim.Spec.StorageClassName != nil {
		return *claim.Spec.StorageClassName
	}

	return ""
}
