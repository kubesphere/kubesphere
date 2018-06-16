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

package controllers

import (
	"encoding/json"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"kubesphere.io/kubesphere/pkg/client"
)

func (ctl *PodCtl) generateObject(item v1.Pod) *Pod {
	name := item.Name
	namespace := item.Namespace
	podIp := item.Status.PodIP
	nodeName := item.Spec.NodeName
	nodeIp := item.Status.HostIP
	status := string(item.Status.Phase)
	createTime := item.CreationTimestamp.Time
	containerStatus := item.Status.ContainerStatuses
	containerSpecs := item.Spec.Containers
	var containers []Container

	for _, containerSpec := range containerSpecs {
		var container Container
		container.Name = containerSpec.Name
		container.Image = containerSpec.Image
		container.Ports = containerSpec.Ports
		for _, status := range containerStatus {
			if container.Name == status.Name {
				container.Ready = status.Ready
			}
		}

		containers = append(containers, container)
	}

	containerStr, _ := json.Marshal(containers)

	annotation, _ := json.Marshal(item.Annotations)

	object := &Pod{Namespace: namespace, Name: name, Node: nodeName, PodIp: podIp, Status: status, NodeIp: nodeIp,
		CreateTime: createTime, ContainerStr: string(containerStr), AnnotationStr: string(annotation)}

	return object
}

func (ctl *PodCtl) listAndWatch() {
	defer func() {
		defer close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&Pod{}) {
		db.DropTable(&Pod{})

	}

	db = db.CreateTable(&Pod{})

	k8sClient := client.NewK8sClient()
	list, err := k8sClient.CoreV1().Pods("").List(meta_v1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list.Items {
		obj := ctl.generateObject(item)
		db.Create(obj)
	}

	watcher, err := k8sClient.CoreV1().Pods("").Watch(meta_v1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ctl.stopChan:
			return
		case event := <-watcher.ResultChan():
			var po Pod
			if event.Object == nil {
				break
			}
			object := event.Object.(*v1.Pod)
			if event.Type == watch.Deleted {
				db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&po)
				db.Delete(po)
				break
			}
			obj := ctl.generateObject(*object)
			db.Save(obj)
		}
	}
}

func (ctl *PodCtl) CountWithConditions(conditions string) int {
	var object Pod

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *PodCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Pod
	var object Pod
	var total int

	order := "createTime desc"

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	for index, item := range list {
		var containers []Container
		json.Unmarshal([]byte(item.ContainerStr), &containers)
		list[index].Containers = containers
		list[index].ContainerStr = ""

		annotation := make(Annotation)
		json.Unmarshal([]byte(item.AnnotationStr), &annotation)
		list[index].Annotation = annotation
		list[index].AnnotationStr = ""
	}
	return total, list, nil
}

func (ctl *PodCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	if len(namespace) == 0 {
		db.Model(&Pod{}).Count(&count)
	} else {
		db.Model(&Pod{}).Where("namespace = ?", namespace).Count(&count)
	}
	return count
}
