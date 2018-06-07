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
	"k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type role struct {
	k8sClient *kubernetes.Clientset
}

func (r *role) list() (interface{}, error) {
	list, err := r.k8sClient.RbacV1().Roles("").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (r *role) getWatcher() (watch.Interface, error) {
	watcher, err := r.k8sClient.RbacV1().Roles("").Watch(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (r *role) generateObject(item v1.Role) OtherResourceObject {
	name := item.Name
	ns := item.Namespace

	object := OtherResourceObject{Namespace: ns, Name: name}

	return object

}

func (r *role) updateWithObject(status *ResourceStatus, item v1.Role) {
	namespace := item.Namespace

	object := r.generateObject(item)
	status.ResourceList.update(namespace, object)
}

func (r *role) updateWithObjects(status *ResourceStatus, objects interface{}) {
	if status.ResourceList == nil {
		status.ResourceList = make(Resources)
	}

	items := objects.([]v1.Role)

	for _, item := range items {
		r.updateWithObject(status, item)
	}

}

func (r *role) updateWithEvent(status *ResourceStatus, event watch.Event) {
	object := event.Object.(*v1.Role)
	namespace := object.Namespace
	tmpObject := r.generateObject(*object)

	if event.Type == watch.Deleted {
		status.ResourceList.del(namespace, tmpObject)
		return
	}

	r.updateWithObject(status, *object)
}
