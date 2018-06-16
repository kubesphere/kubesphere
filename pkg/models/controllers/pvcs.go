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
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"kubesphere.io/kubesphere/pkg/client"
)

const creator = "creator"

func (ctl *PvcCtl) generateObject(item *v1.PersistentVolumeClaim) *Pvc {
	name := item.Name
	namespace := item.Namespace
	status := fmt.Sprintf("%s", item.Status.Phase)
	createTime := item.CreationTimestamp.Time
	capacity := "-"

	if createTime.IsZero() {
		createTime = time.Now()
	}

	if storage, exist := item.Status.Capacity["storage"]; exist {
		capacity = storage.String()
	}

	storageClass := "-"
	if item.Spec.StorageClassName != nil {
		storageClass = *item.Spec.StorageClassName
	}

	accessModeStr := "-"

	var accessModeList []string
	for _, accessMode := range item.Status.AccessModes {
		accessModeList = append(accessModeList, string(accessMode))
	}

	accessModeStr = strings.Join(accessModeList, ",")
	annotation, _ := json.Marshal(item.Annotations)

	object := &Pvc{Namespace: namespace, Name: name, Status: status, Capacity: capacity,
		AccessMode: accessModeStr, StorageClassName: storageClass, CreateTime: createTime, AnnotationStr: string(annotation)}

	return object
}

func (ctl *PvcCtl) listAndWatch() {
	defer func() {
		defer close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&Pvc{}) {
		db.DropTable(&Pvc{})

	}

	db = db.CreateTable(&Pvc{})

	k8sClient := client.NewK8sClient()
	pvcList, err := k8sClient.CoreV1().PersistentVolumeClaims("").List(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range pvcList.Items {
		obj := ctl.generateObject(&item)
		db.Create(obj)
	}

	watcher, err := k8sClient.CoreV1().PersistentVolumeClaims("").Watch(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ctl.stopChan:
			return
		case event := <-watcher.ResultChan():
			var pvc Pvc
			if event.Object == nil {
				break
			}
			object := event.Object.(*v1.PersistentVolumeClaim)
			if event.Type == watch.Deleted {
				db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&pvc)
				db.Delete(pvc)
				break
			}
			obj := ctl.generateObject(object)
			db.Save(obj)
		}
	}
}

func (ctl *PvcCtl) CountWithConditions(conditions string) int {
	var object Pvc

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *PvcCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Pvc
	var object Pvc
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

func (ctl *PvcCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	if len(namespace) == 0 {
		db.Model(&Pvc{}).Count(&count)
	} else {
		db.Model(&Pvc{}).Where("namespace = ?", namespace).Count(&count)
	}
	return count
}
