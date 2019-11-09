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

package kubectl

import (
	"fmt"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"math/rand"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/api/errors"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	namespace = constants.KubeSphereControlNamespace
)

var DefaultImage = "kubesphere/kubectl:advanced-1.0.0"

func init() {
	if env := os.Getenv("KUBECTL_IMAGE"); env != "" {
		DefaultImage = env
	}
}

func GetKubectlPod(username string) (models.PodInfo, error) {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	deployName := fmt.Sprintf("kubectl-%s", username)
	deploy, err := k8sClient.AppsV1().Deployments(namespace).Get(deployName, metav1.GetOptions{})
	if err != nil {
		klog.Errorln(err)
		return models.PodInfo{}, err
	}

	selectors := deploy.Spec.Selector.MatchLabels
	labelSelector := labels.Set(selectors).AsSelector().String()
	podList, err := k8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		klog.Errorln(err)
		return models.PodInfo{}, err
	}

	pod, err := selectCorrectPod(namespace, podList.Items)
	if err != nil {
		klog.Errorln(err)
		return models.PodInfo{}, err
	}

	info := models.PodInfo{Namespace: pod.Namespace, Pod: pod.Name, Container: pod.Status.ContainerStatuses[0].Name}

	return info, nil

}

func selectCorrectPod(namespace string, pods []v1.Pod) (kubectlPod v1.Pod, err error) {

	var kubectlPodList []v1.Pod
	for _, pod := range pods {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				kubectlPodList = append(kubectlPodList, pod)
			}
		}
	}
	if len(kubectlPodList) < 1 {
		err = fmt.Errorf("cannot find valid kubectl pod in namespace:%s", namespace)
		return v1.Pod{}, err
	}

	random := rand.Intn(len(kubectlPodList))
	return kubectlPodList[random], nil
}

func CreateKubectlDeploy(username string) error {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	deployName := fmt.Sprintf("kubectl-%s", username)
	_, err := k8sClient.AppsV1().Deployments(namespace).Get(deployName, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	replica := int32(1)
	selector := metav1.LabelSelector{MatchLabels: map[string]string{"username": username}}
	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &selector,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"username": username,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "kubectl",
							Image: DefaultImage,
						},
					},
					ServiceAccountName: "kubesphere-cluster-admin",
				},
			},
		},
	}

	_, err = k8sClient.AppsV1().Deployments(namespace).Create(&deployment)

	return err
}

func DelKubectlDeploy(username string) error {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	deployName := fmt.Sprintf("kubectl-%s", username)
	_, err := k8sClient.AppsV1().Deployments(namespace).Get(deployName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		err := fmt.Errorf("delete username %s failed, reason:%v", username, err)
		return err
	}

	deletePolicy := metav1.DeletePropagationBackground

	err = k8sClient.AppsV1().Deployments(namespace).Delete(deployName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	if err != nil {
		err := fmt.Errorf("delete username %s failed, reason:%v", username, err)
		return err
	}

	return nil
}
