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
	"time"

	"github.com/golang/glog"
	"k8s.io/api/storage/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"kubesphere.io/kubesphere/pkg/client"
)

func (ctl *StorageClassCtl) generateObject(item v1beta1.StorageClass) *StorageClass {

	name := item.Name
	createTime := item.CreationTimestamp.Time
	isDefault := false
	if item.Annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true" {
		isDefault = true
	}

	if createTime.IsZero() {
		createTime = time.Now()
	}

	annotation, _ := json.Marshal(item.Annotations)
	object := &StorageClass{Name: name, CreateTime: createTime, IsDefault: isDefault, AnnotationStr: string(annotation)}

	return object
}

func (ctl *StorageClassCtl) listAndWatch() {
	defer func() {
		defer close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&StorageClass{}) {
		db.DropTable(&StorageClass{})
	}

	db = db.CreateTable(&StorageClass{})

	k8sClient := client.NewK8sClient()
	list, err := k8sClient.StorageV1beta1().StorageClasses().List(meta_v1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list.Items {
		obj := ctl.generateObject(item)
		db.Create(obj)
	}

	watcher, err := k8sClient.StorageV1beta1().StorageClasses().Watch(meta_v1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ctl.stopChan:
			return
		case event := <-watcher.ResultChan():
			var sc StorageClass
			if event.Object == nil {
				panic("watch timeout, restart storageClass controller")
			}
			object := event.Object.(*v1beta1.StorageClass)
			if event.Type == watch.Deleted {
				db.Where("name=?", object.Name).Find(&sc)
				db.Delete(sc)
				break
			}
			obj := ctl.generateObject(*object)
			db.Save(obj)
		}
	}
}

func (ctl *StorageClassCtl) CountWithConditions(conditions string) int {
	var object StorageClass

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *StorageClassCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []StorageClass
	var object StorageClass
	var total int

	order := "createTime desc"

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	for index, storageClass := range list {
		name := storageClass.Name
		annotation := make(map[string]string)
		json.Unmarshal([]byte(storageClass.AnnotationStr), &annotation)
		list[index].Annotation = annotation
		list[index].AnnotationStr = ""
		pvcCtl := PvcCtl{CommonAttribute{K8sClient: ctl.K8sClient, DB: ctl.DB}}

		list[index].Count = pvcCtl.CountWithConditions(fmt.Sprintf("storage_class=\"%s\"", name))
	}

	return total, list, nil
}

func (ctl *StorageClassCtl) Count(name string) int {
	var count int
	db := ctl.DB
	db.Model(&StorageClass{}).Count(&count)
	return count
}
