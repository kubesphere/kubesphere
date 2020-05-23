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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models"
	"math/rand"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	namespace        = constants.KubeSphereControlNamespace
	deployNameFormat = "kubectl-%s"
)

type Interface interface {
	GetKubectlPod(username string) (models.PodInfo, error)
	CreateKubectlDeploy(username string) error
}

type operator struct {
	k8sClient   kubernetes.Interface
	k8sInformer k8sinformers.SharedInformerFactory
	ksInformer  ksinformers.SharedInformerFactory
}

func NewOperator(k8sClient kubernetes.Interface, k8sInformer k8sinformers.SharedInformerFactory, ksInformer ksinformers.SharedInformerFactory) Interface {
	return &operator{k8sClient: k8sClient, k8sInformer: k8sInformer, ksInformer: ksInformer}
}

var DefaultImage = "kubesphere/kubectl:advanced-1.0.0"

func init() {
	if env := os.Getenv("KUBECTL_IMAGE"); env != "" {
		DefaultImage = env
	}
}

func (o *operator) GetKubectlPod(username string) (models.PodInfo, error) {
	deployName := fmt.Sprintf(deployNameFormat, username)
	deploy, err := o.k8sInformer.Apps().V1().Deployments().Lister().Deployments(namespace).Get(deployName)
	if err != nil {
		klog.Errorln(err)
		return models.PodInfo{}, err
	}

	selectors := deploy.Spec.Selector.MatchLabels
	labelSelector := labels.Set(selectors).AsSelector()
	pods, err := o.k8sInformer.Core().V1().Pods().Lister().Pods(namespace).List(labelSelector)
	if err != nil {
		klog.Errorln(err)
		return models.PodInfo{}, err
	}

	pod, err := selectCorrectPod(namespace, pods)
	if err != nil {
		klog.Errorln(err)
		return models.PodInfo{}, err
	}

	info := models.PodInfo{Namespace: pod.Namespace, Pod: pod.Name, Container: pod.Status.ContainerStatuses[0].Name}

	return info, nil

}

func selectCorrectPod(namespace string, pods []*v1.Pod) (kubectlPod *v1.Pod, err error) {

	var kubectlPodList []*v1.Pod
	for _, pod := range pods {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				kubectlPodList = append(kubectlPodList, pod)
			}
		}
	}
	if len(kubectlPodList) < 1 {
		err = fmt.Errorf("cannot find valid kubectl pod in namespace:%s", namespace)
		return &v1.Pod{}, err
	}

	random := rand.Intn(len(kubectlPodList))

	return kubectlPodList[random], nil
}

func (o *operator) CreateKubectlDeploy(username string) error {
	deployName := fmt.Sprintf(deployNameFormat, username)

	user, err := o.ksInformer.Iam().V1alpha2().Users().Lister().Get(username)

	if err != nil {
		klog.Error(err)
		// ignore if user not exist
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	replica := int32(1)
	selector := metav1.LabelSelector{MatchLabels: map[string]string{constants.UsernameLabelKey: username}}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &selector,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						constants.UsernameLabelKey: username,
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

	err = controllerutil.SetControllerReference(user, deployment, scheme.Scheme)

	if err != nil {
		klog.Errorln(err)
		return err
	}

	_, err = o.k8sClient.AppsV1().Deployments(namespace).Create(deployment)

	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	return nil
}
