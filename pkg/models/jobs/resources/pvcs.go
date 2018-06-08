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

type persistmentVolume struct {
	k8sClient *kubernetes.Clientset
}

func (pvc *persistmentVolume) list() (interface{}, error) {
	list, err := pvc.k8sClient.CoreV1().PersistentVolumeClaims("").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (pvc *persistmentVolume) getWatcher() (watch.Interface, error) {
	watcher, err := pvc.k8sClient.CoreV1().PersistentVolumeClaims("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (pvc *persistmentVolume) generateObject(item v1.PersistentVolumeClaim) OtherResourceObject {
	name := item.Name
	ns := item.Namespace

	object := OtherResourceObject{Namespace: ns, Name: name}

	return object

}

func (pvc *persistmentVolume) updateWithObject(status *ResourceStatus, item v1.PersistentVolumeClaim) {
	namespace := item.Namespace

	object := pvc.generateObject(item)
	status.ResourceList.update(namespace, object)
}

func (pvc *persistmentVolume) updateWithObjects(status *ResourceStatus, objects interface{}) {
	if status.ResourceList == nil {
		status.ResourceList = make(Resources)
	}

	items := objects.([]v1.PersistentVolumeClaim)

	for _, item := range items {
		pvc.updateWithObject(status, item)
	}

}

func (pvc *persistmentVolume) updateWithEvent(status *ResourceStatus, event watch.Event) {
	object := event.Object.(*v1.PersistentVolumeClaim)
	namespace := object.Namespace
	tmpObject := pvc.generateObject(*object)

	if event.Type == watch.Deleted {
		status.ResourceList.del(namespace, tmpObject)
		return
	}

	pvc.updateWithObject(status, *object)
}
