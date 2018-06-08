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

type statefulset struct {
	k8sClient *kubernetes.Clientset
}

func (ss *statefulset) list() (interface{}, error) {
	daemonsetList, err := ss.k8sClient.AppsV1beta2().StatefulSets("").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return daemonsetList.Items, nil
}

func (ss *statefulset) getWatcher() (watch.Interface, error) {
	watcher, err := ss.k8sClient.AppsV1beta2().StatefulSets("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (ss *statefulset) generateObject(item v1beta2.StatefulSet) WorkLoadObject {
	var app string
	var ready bool
	name := item.Name
	namespace := item.Namespace
	availablePodNum := item.Status.ReadyReplicas
	desirePodNum := *item.Spec.Replicas
	createTime := item.CreationTimestamp
	release := item.ObjectMeta.Labels["release"]
	chart := item.ObjectMeta.Labels["chart"]

	if len(release) > 0 && len(chart) > 0 {
		app = release + "/" + chart
	} else {
		app = "-"
	}
	if availablePodNum >= desirePodNum {
		ready = true
	} else {
		ready = false
	}

	statefulSetObject := WorkLoadObject{Namespace: namespace, Name: name, Available: availablePodNum, Desire: desirePodNum,
		App: app, CreateTime: createTime, Ready: ready}

	return statefulSetObject

}

func (ss *statefulset) updateWithObject(status *ResourceStatus, item v1beta2.StatefulSet) {
	namespace := item.Namespace
	ssObject := ss.generateObject(item)

	status.ResourceList.update(namespace, ssObject)

}

func (ss *statefulset) updateWithObjects(status *ResourceStatus, objects interface{}) {
	if status.ResourceList == nil {
		status.ResourceList = make(Resources)
	}

	items := objects.([]v1beta2.StatefulSet)

	for _, item := range items {
		ss.updateWithObject(status, item)
	}

}

func (ss *statefulset) updateWithEvent(status *ResourceStatus, event watch.Event) {
	object := event.Object.(*v1beta2.StatefulSet)
	namespace := object.Namespace
	ssObject := ss.generateObject(*object)

	if event.Type == watch.Deleted {
		status.ResourceList.del(namespace, ssObject)
		return
	}

	ss.updateWithObject(status, *object)
}
