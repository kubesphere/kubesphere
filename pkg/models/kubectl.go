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
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/client"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)


const (
	deploymentName = "kubectl"
)

type kubectlPodInfo struct {
	Namespace string `json: "namespace"`
	Pod string `json: "podname"`
	Container string `json: "container"`
}

func GetKubectlPod(namespace string) (kubectlPodInfo, error) {
	k8sClient := client.NewK8sClient()
	deploy, err := k8sClient.AppsV1beta2().Deployments(namespace).Get(deploymentName, meta_v1.GetOptions{})
	if err != nil {
		glog.Errorln(err)
		return kubectlPodInfo{}, err
	}

	selectors := deploy.Spec.Selector.MatchLabels
	labelSelector := labels.Set(selectors).AsSelector().String()
	podList, err := k8sClient.CoreV1().Pods(namespace).List(meta_v1.ListOptions{LabelSelector:labelSelector})
	if err != nil {
		glog.Errorln(err)
		return kubectlPodInfo{}, err
	}

	pod, err := selectCorrectPod(namespace, podList.Items)
	if err != nil{
		glog.Errorln(err)
		return kubectlPodInfo{}, err
	}

	info := kubectlPodInfo{Namespace:pod.Namespace, Pod:pod.Name, Container:pod.Status.ContainerStatuses[0].Name}

	return info, nil

}


func selectCorrectPod(namespace string, pods []v1.Pod) (kubectlPod v1.Pod, err error) {

	var kubectPodList []v1.Pod
	for _, pod := range pods{
		for _, condition := range pod.Status.Conditions{
			if condition.Type == "Ready" && condition.Status == "True"{
				kubectPodList = append(kubectPodList, pod)
			}
		}
	}
	if len(kubectPodList) < 1{
		err = fmt.Errorf("cannot find valid kubectl pod in namespace:%s", namespace)
		return v1.Pod{}, err
	}

	random := rand.Intn(len(kubectPodList))
	return kubectPodList[random], nil
}
