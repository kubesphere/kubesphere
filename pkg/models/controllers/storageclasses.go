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
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/storage/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (ctl *StorageClassCtl) generateObject(item v1.StorageClass) *StorageClass {

	name := item.Name
	createTime := item.CreationTimestamp.Time
	isDefault := false
	if item.Annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true" {
		isDefault = true
	}

	if createTime.IsZero() {
		createTime = time.Now()
	}

	object := &StorageClass{Name: name, CreateTime: createTime, IsDefault: isDefault, Annotation: Annotation{item.Annotations}}

	return object
}

func (ctl *StorageClassCtl) listAndWatch() {
	defer func() {
		close(ctl.aliveChan)
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

	k8sClient := ctl.K8sClient
	kubeInformerFactory := informers.NewSharedInformerFactory(k8sClient, time.Second*resyncCircle)
	informer := kubeInformerFactory.Storage().V1().StorageClasses().Informer()
	lister := kubeInformerFactory.Storage().V1().StorageClasses().Lister()

	list, err := lister.List(labels.Everything())
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list {
		obj := ctl.generateObject(*item)
		db.Create(obj)

	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			object := obj.(*v1.StorageClass)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.StorageClass)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
		},
		DeleteFunc: func(obj interface{}) {
			var item StorageClass
			object := obj.(*v1.StorageClass)
			db.Where("name=?", object.Name).Find(&item)
			db.Delete(item)

		},
	})

	informer.Run(ctl.stopChan)

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
