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
	appsv1informers "k8s.io/client-go/informers/apps/v1"
	coreinfomers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	namespace        = constants.KubeSphereControlNamespace
	deployNameFormat = "kubectl-%s"
)

type Interface interface {
	GetKubectlPod(username string) (models.PodInfo, error)
	CreateKubectlDeploy(username string, owner metav1.Object) error
}

type operator struct {
	k8sClient          kubernetes.Interface
	deploymentInformer appsv1informers.DeploymentInformer
	podInformer        coreinfomers.PodInformer
	userInformer       iamv1alpha2informers.UserInformer
	kubectlImage       string
}

func NewOperator(k8sClient kubernetes.Interface, deploymentInformer appsv1informers.DeploymentInformer,
	podInformer coreinfomers.PodInformer, userInformer iamv1alpha2informers.UserInformer, kubectlImage string) Interface {
	return &operator{k8sClient: k8sClient, deploymentInformer: deploymentInformer, podInformer: podInformer,
		userInformer: userInformer, kubectlImage: kubectlImage}
}

func (o *operator) GetKubectlPod(username string) (models.PodInfo, error) {
	deployName := fmt.Sprintf(deployNameFormat, username)
	deploy, err := o.deploymentInformer.Lister().Deployments(namespace).Get(deployName)
	if err != nil {
		klog.Errorln(err)
		return models.PodInfo{}, err
	}

	selectors := deploy.Spec.Selector.MatchLabels
	labelSelector := labels.Set(selectors).AsSelector()
	pods, err := o.podInformer.Lister().Pods(namespace).List(labelSelector)
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

func (o *operator) CreateKubectlDeploy(username string, owner metav1.Object) error {
	deployName := fmt.Sprintf(deployNameFormat, username)

	_, err := o.userInformer.Lister().Get(username)
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
						{
							Name:  "kubectl",
							Image: o.kubectlImage,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "host-time",
									MountPath: "/etc/localtime",
								},
							},
						},
					},
					ServiceAccountName: "kubesphere-cluster-admin",
					Volumes: []v1.Volume{
						{
							Name: "host-time",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/etc/localtime",
								},
							},
						},
					},
				},
			},
		},
	}

	// bind the lifecycle of role binding
	err = controllerutil.SetControllerReference(owner, deployment, scheme.Scheme)
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
