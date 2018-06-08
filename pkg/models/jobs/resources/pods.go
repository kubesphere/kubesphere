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
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type pod struct {
	k8sClient *kubernetes.Clientset
}

func (po *pod) list() (interface{}, error) {
	list, err := po.k8sClient.CoreV1().Pods("").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (po *pod) getWatcher() (watch.Interface, error) {
	watcher, err := po.k8sClient.CoreV1().Pods("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (po *pod) generateObject(item v1.Pod) OtherResourceObject {
	name := item.Name
	ns := item.Namespace

	Object := OtherResourceObject{Namespace: ns, Name: name}

	return Object

}

func (po *pod) updateWithObject(status *ResourceStatus, item v1.Pod) {
	namespace := item.Namespace

	object := po.generateObject(item)
	status.ResourceList.update(namespace, object)
}

func (po *pod) updateWithObjects(status *ResourceStatus, objects interface{}) {
	if status.ResourceList == nil {
		status.ResourceList = make(Resources)
	}

	items := objects.([]v1.Pod)

	for _, item := range items {
		po.updateWithObject(status, item)
	}

}

func (po *pod) updateWithEvent(status *ResourceStatus, event watch.Event) {
	object := event.Object.(*v1.Pod)
	namespace := object.Namespace
	tmpObject := po.generateObject(*object)

	if event.Type == watch.Deleted {
		status.ResourceList.del(namespace, tmpObject)
		return
	}

	po.updateWithObject(status, *object)
}
