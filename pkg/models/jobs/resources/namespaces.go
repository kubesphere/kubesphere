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

type namespace struct {
	k8sClient *kubernetes.Clientset
}

func (ns *namespace) list() (interface{}, error) {
	nsList, err := ns.k8sClient.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nsList.Items, nil
}

func (ns *namespace) getWatcher() (watch.Interface, error) {
	watcher, err := ns.k8sClient.CoreV1().Namespaces().Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (ns *namespace) generateObject(item v1.Namespace) OtherResourceObject {
	name := item.Name
	nsp := item.Namespace

	object := OtherResourceObject{Namespace: nsp, Name: name}

	return object

}

func (ns *namespace) updateWithObject(status *ResourceStatus, item v1.Namespace) {
	namespace := item.Namespace

	object := ns.generateObject(item)
	status.ResourceList.update(namespace, object)
}

func (ns *namespace) updateWithObjects(status *ResourceStatus, objects interface{}) {
	if status.ResourceList == nil {
		status.ResourceList = make(Resources)
	}

	items := objects.([]v1.Namespace)

	for _, item := range items {
		ns.updateWithObject(status, item)
	}

}

func (ns *namespace) updateWithEvent(status *ResourceStatus, event watch.Event) {
	object := event.Object.(*v1.Namespace)
	namespace := object.Namespace
	tmpObject := ns.generateObject(*object)

	if event.Type == watch.Deleted {
		status.ResourceList.del(namespace, tmpObject)
		return
	}

	ns.updateWithObject(status, *object)
}
