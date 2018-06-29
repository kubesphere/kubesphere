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

func (ctl *ClusterRoleCtl) generateObject(item v1.ClusterRole) *ClusterRole {
	name := item.Name
	if strings.HasPrefix(name, "system:") {
		return nil
	}

	createTime := item.CreationTimestamp.Time
	if createTime.IsZero() {
		createTime = time.Now()
	}

	object := &ClusterRole{Name: name, CreateTime: createTime, Annotation: Annotation{item.Annotations}}

	return object
}

func (ctl *ClusterRoleCtl) listAndWatch() {
	defer func() {
		close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&ClusterRole{}) {
		db.DropTable(&ClusterRole{})

	}

	db = db.CreateTable(&ClusterRole{})

	k8sClient := ctl.K8sClient
	kubeInformerFactory := informers.NewSharedInformerFactory(k8sClient, time.Second*resyncCircle)
	informer := kubeInformerFactory.Rbac().V1().ClusterRoles().Informer()
	lister := kubeInformerFactory.Rbac().V1().ClusterRoles().Lister()

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

			object := obj.(*v1.ClusterRole)
			mysqlObject := ctl.generateObject(*object)
			if mysqlObject != nil {
				db.Create(mysqlObject)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.ClusterRole)
			mysqlObject := ctl.generateObject(*object)
			if mysqlObject != nil {
				db.Save(mysqlObject)
			}
		},
		DeleteFunc: func(obj interface{}) {
			var item ClusterRole
			object := obj.(*v1.ClusterRole)
			db.Where("name=?", object.Name).Find(&item)
			db.Delete(item)

		},
	})

	informer.Run(ctl.stopChan)
}

func (ctl *ClusterRoleCtl) CountWithConditions(conditions string) int {
	var object ClusterRole

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *ClusterRoleCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var object ClusterRole
	var list []ClusterRole
	var total int

	order := "createTime desc"
	db := ctl.DB

	listWithConditions(db, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *ClusterRoleCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	db.Model(&ClusterRole{}).Count(&count)
	return count
}
