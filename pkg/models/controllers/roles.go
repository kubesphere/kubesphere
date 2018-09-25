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
	var displayName string

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

	name := item.Name
	if strings.HasPrefix(name, systemPrefix) {
		return nil
	}
	namespace := item.Namespace
	createTime := item.CreationTimestamp.Time
	if createTime.IsZero() {
		createTime = time.Now()
	}

	object := &Role{
		Namespace:   namespace,
		Name:        name,
		DisplayName: displayName,
		CreateTime:  createTime,
		Annotation:  MapString{item.Annotations},
	}

	return object
}

func (ctl *RoleCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *RoleCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&Role{}) {
		db.DropTable(&Role{})
	}

	db = db.CreateTable(&Role{})

	ctl.initListerAndInformer()
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list {
		obj := ctl.generateObject(*item)
		if obj != nil {
			db.Create(obj)
		}
	}

	ctl.informer.Run(stopChan)
}

func (ctl *RoleCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}

	count := 0
	for _, item := range list {
		if !strings.HasPrefix(item.Name, systemPrefix) {
			count++
		}
	}

	return count
}

func (ctl *RoleCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Rbac().V1().Roles().Lister()

	informer := informerFactory.Rbac().V1().Roles().Informer()
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

	ctl.informer = informer
}

func (ctl *RoleCtl) CountWithConditions(conditions string) int {
	var object Role

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *RoleCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []Role
	var object Role
	var total int

	if len(order) == 0 {
		order = "createTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *RoleCtl) Lister() interface{} {

	return ctl.lister
}
