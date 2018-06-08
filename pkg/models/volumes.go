package models

import (
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
)

type PodListByPvc struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Pods      []v12.Pod `json:"pods"`
}

// List pods of a specific persistent volume claims
func GetPodListByPvc(pvc string, ns string) (res []v12.Pod, err error) {
	cli := client.NewK8sClient()
	podList, err := cli.CoreV1().Pods(ns).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		if IsPvcInPod(pod, pvc) == true {
			res = append(res, pod)
		}
	}
	return res, nil
}

// Check if the persistent volume claim is related to the pod
func IsPvcInPod(pod v12.Pod, pvcname string) bool {
	for _, v := range pod.Spec.Volumes {
		if v.VolumeSource.PersistentVolumeClaim != nil &&
			v.VolumeSource.PersistentVolumeClaim.ClaimName == pvcname {
			return true
		}
	}
	return false
}
