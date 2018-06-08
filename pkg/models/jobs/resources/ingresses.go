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
	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type ingress struct {
	k8sClient *kubernetes.Clientset
}

func (ing *ingress) list() (interface{}, error) {
	list, err := ing.k8sClient.ExtensionsV1beta1().Ingresses("").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (ing *ingress) getWatcher() (watch.Interface, error) {
	watcher, err := ing.k8sClient.ExtensionsV1beta1().Ingresses("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (ing *ingress) generateObject(item v1beta1.Ingress) OtherResourceObject {
	name := item.Name
	ns := item.Namespace

	object := OtherResourceObject{Namespace: ns, Name: name}

	return object

}

func (ing *ingress) updateWithObject(status *ResourceStatus, item v1beta1.Ingress) {
	namespace := item.Namespace

	object := ing.generateObject(item)
	status.ResourceList.update(namespace, object)
}

func (ing *ingress) updateWithObjects(status *ResourceStatus, objects interface{}) {
	if status.ResourceList == nil {
		status.ResourceList = make(Resources)
	}

	items := objects.([]v1beta1.Ingress)

	for _, item := range items {
		ing.updateWithObject(status, item)
	}

}

func (ing *ingress) updateWithEvent(status *ResourceStatus, event watch.Event) {
	object := event.Object.(*v1beta1.Ingress)
	namespace := object.Namespace
	tmpObject := ing.generateObject(*object)

	if event.Type == watch.Deleted {
		status.ResourceList.del(namespace, tmpObject)
		return
	}

	ing.updateWithObject(status, *object)
}
