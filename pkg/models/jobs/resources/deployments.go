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

package resources

import (
	"k8s.io/api/apps/v1beta2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type deployment struct {
	k8sClient *kubernetes.Clientset
}

func (deploy *deployment) list() (interface{}, error) {
	deoloyList, err := deploy.k8sClient.AppsV1beta2().Deployments("").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deoloyList.Items, nil
}

func (deploy *deployment) getWatcher() (watch.Interface, error) {
	watcher, err := deploy.k8sClient.AppsV1beta2().Deployments("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (deploy *deployment) generateObject(item v1beta2.Deployment) WorkLoadObject {
	var app string
	var ready bool
	var updateTime meta_v1.Time
	name := item.Name
	namespace := item.Namespace
	availablePodNum := item.Status.AvailableReplicas
	desirePodNum := *item.Spec.Replicas
	release := item.ObjectMeta.Labels["release"]
	chart := item.ObjectMeta.Labels["chart"]

	if len(release) > 0 && len(chart) > 0 {
		app = release + "/" + chart
	} else {
		app = "-"
	}

	for _, conditon := range item.Status.Conditions {
		if conditon.Type == "Progressing" {
			updateTime = conditon.LastUpdateTime
		}
	}

	if availablePodNum >= desirePodNum {
		ready = true
	} else {
		ready = false
	}

	deployObject := WorkLoadObject{Namespace: namespace, Name: name, Available: availablePodNum, Desire: desirePodNum,
		App: app, UpdateTime: updateTime, Ready: ready}

	return deployObject

}

func (deploy *deployment) updateWithObject(status *ResourceStatus, item v1beta2.Deployment) {
	namespace := item.Namespace

	deployObject := deploy.generateObject(item)
	status.ResourceList.update(namespace, deployObject)
}

func (deploy *deployment) updateWithObjects(status *ResourceStatus, objects interface{}) {
	if status.ResourceList == nil {
		status.ResourceList = make(Resources)
	}

	items := objects.([]v1beta2.Deployment)

	for _, item := range items {
		deploy.updateWithObject(status, item)
	}

}

func (deploy *deployment) updateWithEvent(status *ResourceStatus, event watch.Event) {
	object := event.Object.(*v1beta2.Deployment)
	namespace := object.Namespace
	deployObject := deploy.generateObject(*object)

	if event.Type == watch.Deleted {
		status.ResourceList.del(namespace, deployObject)
		return
	}

	deploy.updateWithObject(status, *object)
}
