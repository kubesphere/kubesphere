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

type service struct {
	k8sClient *kubernetes.Clientset
}

func (svc *service) list() (interface{}, error) {
	list, err := svc.k8sClient.CoreV1().Services("").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (svc *service) getWatcher() (watch.Interface, error) {
	watcher, err := svc.k8sClient.CoreV1().Services("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (svc *service) generateObject(item v1.Service) OtherResourceObject {
	name := item.Name
	ns := item.Namespace

	object := OtherResourceObject{Namespace: ns, Name: name}

	return object

}

func (svc *service) updateWithObject(status *ResourceStatus, item v1.Service) {
	namespace := item.Namespace

	object := svc.generateObject(item)
	status.ResourceList.update(namespace, object)
}

func (svc *service) updateWithObjects(status *ResourceStatus, objects interface{}) {
	if status.ResourceList == nil {
		status.ResourceList = make(Resources)
	}

	items := objects.([]v1.Service)

	for _, item := range items {
		svc.updateWithObject(status, item)
	}

}

func (svc *service) updateWithEvent(status *ResourceStatus, event watch.Event) {
	object := event.Object.(*v1.Service)
	namespace := object.Namespace
	tmpObject := svc.generateObject(*object)

	if event.Type == watch.Deleted {
		status.ResourceList.del(namespace, tmpObject)
		return
	}

	svc.updateWithObject(status, *object)
}
