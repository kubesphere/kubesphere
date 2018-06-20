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
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (ctl *RoleCtl) generateObject(item v1.Role) *Role {
	name := item.Name
	if strings.HasPrefix(name, "system:") {
		return nil
	}
	namespace := item.Namespace
	createTime := item.CreationTimestamp.Time
	if createTime.IsZero() {
		createTime = time.Now()
	}

	object := &Role{Namespace: namespace, Name: name, CreateTime: createTime, Annotation: Annotation{item.Annotations}}

	return object
}

func (ctl *RoleCtl) listAndWatch() {
	defer func() {
		close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&Role{}) {
		db.DropTable(&Role{})

	}

	db = db.CreateTable(&Role{})

	k8sClient := ctl.K8sClient
	kubeInformerFactory := informers.NewSharedInformerFactory(k8sClient, time.Second*resyncCircle)
	informer := kubeInformerFactory.Rbac().V1().Roles().Informer()
	lister := kubeInformerFactory.Rbac().V1().Roles().Lister()

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

			object := obj.(*v1.Role)
			mysqlObject := ctl.generateObject(*object)
			if mysqlObject != nil {
				db.Create(mysqlObject)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.Role)
			mysqlObject := ctl.generateObject(*object)
			if mysqlObject != nil {
				db.Save(mysqlObject)
			}
		},
		DeleteFunc: func(obj interface{}) {
			var item Role
			object := obj.(*v1.Role)
			db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&item)
			db.Delete(item)

		},
	})

	informer.Run(ctl.stopChan)
}

func (ctl *RoleCtl) CountWithConditions(conditions string) int {
	var object Role

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *RoleCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Role
	var object Role
	var total int

	order := "createTime desc"

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *RoleCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	db.Model(&Role{}).Where("namespace = ?", namespace).Count(&count)
	return count
}
