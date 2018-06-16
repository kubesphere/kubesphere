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
	"time"

	"github.com/golang/glog"
	"k8s.io/api/apps/v1beta2"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"kubesphere.io/kubesphere/pkg/client"
)

func (ctl *StatefulsetCtl) generateObject(item v1beta2.StatefulSet) *Statefulset {
	var app string
	var status string
	name := item.Name
	namespace := item.Namespace
	availablePodNum := item.Status.ReadyReplicas
	desirePodNum := *item.Spec.Replicas
	createTime := item.CreationTimestamp.Time
	release := item.ObjectMeta.Labels["release"]
	chart := item.ObjectMeta.Labels["chart"]

	if createTime.IsZero() {
		createTime = time.Now()
	}

	if len(release) > 0 && len(chart) > 0 {
		app = release + "/" + chart
	} else {
		app = "-"
	}

	if item.Annotations["state"] == "stop" {
		status = stopping
	} else {
		if availablePodNum >= desirePodNum {
			status = running
		} else {
			status = updating
		}
	}

	annotation, _ := json.Marshal(item.Annotations)

	statefulSetObject := &Statefulset{Namespace: namespace, Name: name, Available: availablePodNum, Desire: desirePodNum,
		App: app, CreateTime: createTime, Status: status, AnnotationStr: string(annotation)}

	return statefulSetObject
}

func (ctl *StatefulsetCtl) listAndWatch() {
	defer func() {
		defer close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB
	if db.HasTable(&Statefulset{}) {
		db.DropTable(&Statefulset{})
	}

	db = db.CreateTable(&Statefulset{})
	k8sClient := client.NewK8sClient()
	deoloyList, err := k8sClient.AppsV1beta2().StatefulSets("").List(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
	}

	for _, item := range deoloyList.Items {
		obj := ctl.generateObject(item)
		db.Create(obj)
	}

	watcher, err := k8sClient.AppsV1beta2().StatefulSets("").Watch(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ctl.stopChan:
			return
		case event := <-watcher.ResultChan():
			var tmp Statefulset
			if event.Object == nil {
				break
			}
			object := event.Object.(*v1beta2.StatefulSet)
			if event.Type == watch.Deleted {
				db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&tmp)
				db.Delete(tmp)
				break
			}
			obj := ctl.generateObject(*object)
			db.Save(obj)
		}
	}
}

func (ctl *StatefulsetCtl) CountWithConditions(conditions string) int {
	var object Statefulset

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *StatefulsetCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Statefulset
	var object Statefulset
	var total int

	order := "createTime desc"

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	for index, item := range list {
		annotation := make(map[string]string)
		json.Unmarshal([]byte(item.AnnotationStr), &annotation)
		list[index].Annotation = annotation
		list[index].AnnotationStr = ""
	}
	return total, list, nil
}

func (ctl *StatefulsetCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	if len(namespace) == 0 {
		db.Model(&Statefulset{}).Count(&count)
	} else {
		db.Model(&Statefulset{}).Where("namespace = ?", namespace).Count(&count)
	}
	return count
}
