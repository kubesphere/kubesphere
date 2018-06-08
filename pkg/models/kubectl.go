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

package models

import (
	"fmt"
	"math/rand"

	"github.com/golang/glog"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/options"
)

const namespace = constants.NameSpace

type kubectlPodInfo struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

func GetKubectlPod(user string) (kubectlPodInfo, error) {
	k8sClient := client.NewK8sClient()
	deploy, err := k8sClient.AppsV1beta2().Deployments(namespace).Get(user, meta_v1.GetOptions{})
	if err != nil {
		glog.Errorln(err)
		return kubectlPodInfo{}, err
	}

	selectors := deploy.Spec.Selector.MatchLabels
	labelSelector := labels.Set(selectors).AsSelector().String()
	podList, err := k8sClient.CoreV1().Pods(namespace).List(meta_v1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		glog.Errorln(err)
		return kubectlPodInfo{}, err
	}

	pod, err := selectCorrectPod(namespace, podList.Items)
	if err != nil {
		glog.Errorln(err)
		return kubectlPodInfo{}, err
	}

	info := kubectlPodInfo{Namespace: pod.Namespace, Pod: pod.Name, Container: pod.Status.ContainerStatuses[0].Name}

	return info, nil

}

func selectCorrectPod(namespace string, pods []v1.Pod) (kubectlPod v1.Pod, err error) {

	var kubectPodList []v1.Pod
	for _, pod := range pods {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				kubectPodList = append(kubectPodList, pod)
			}
		}
	}
	if len(kubectPodList) < 1 {
		err = fmt.Errorf("cannot find valid kubectl pod in namespace:%s", namespace)
		return v1.Pod{}, err
	}

	random := rand.Intn(len(kubectPodList))
	return kubectPodList[random], nil
}

func CreateKubectlPod(user string) error {

	replica := int32(1)
	selector := meta_v1.LabelSelector{MatchLabels: map[string]string{"user": user}}
	config := v1.ConfigMapVolumeSource{Items: []v1.KeyToPath{{Key: "config", Path: "config"}}, LocalObjectReference: v1.LocalObjectReference{Name: user}}
	deployment := v1beta2.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: user,
		},
		Spec: v1beta2.DeploymentSpec{
			Replicas: &replica,
			Selector: &selector,
			Template: v1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Labels: map[string]string{
						"user": user,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "kubectl",
							Image:        options.ServerOptions.GetKubectlImage(),
							VolumeMounts: []v1.VolumeMount{{Name: "kubeconfig", MountPath: "/root/.kube"}},
						},
					},
					Volumes: []v1.Volume{{Name: "kubeconfig", VolumeSource: v1.VolumeSource{ConfigMap: &config}}},
				},
			},
		},
	}

	k8sClient := client.NewK8sClient()
	_, err := k8sClient.AppsV1beta2().Deployments(namespace).Create(&deployment)

	return err
}

func DelKubectlPod(user string) error {
	k8sClient := client.NewK8sClient()
	deploy, err := k8sClient.AppsV1beta2().Deployments(namespace).Get(user, meta_v1.GetOptions{})
	if err != nil {
		return err
	}
	replicas := int32(0)
	deploy.Spec.Replicas = &replicas
	k8sClient.AppsV1beta2().Deployments(namespace).Update(deploy)
	err = k8sClient.AppsV1beta2().Deployments(namespace).Delete(user, &meta_v1.DeleteOptions{})
	return err
}
