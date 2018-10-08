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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (ctl *ConfigMapCtl) generateObject(item v1.ConfigMap) *ConfigMap {
	var displayName string
	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

	createTime := item.CreationTimestamp.Time
	if createTime.IsZero() {
		createTime = time.Now()
	}

	var entries []string

	for entry := range item.Data {
		entries = append(entries, entry)
	}

	object := &ConfigMap{
		Name:        item.Name,
		Namespace:   item.Namespace,
		CreateTime:  createTime,
		Annotation:  MapString{item.Annotations},
		DisplayName: displayName,
		Entries:     strings.Join(entries, ","),
	}

	return object
}

func (ctl *ConfigMapCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *ConfigMapCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&ConfigMap{}) {
		db.DropTable(&ConfigMap{})
	}

	db = db.CreateTable(&ConfigMap{})

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

func (ctl *ConfigMapCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}

	return len(list)
}

func (ctl *ConfigMapCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)
	ctl.lister = informerFactory.Core().V1().ConfigMaps().Lister()

	informer := informerFactory.Core().V1().ConfigMaps().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			object := obj.(*v1.ConfigMap)
			mysqlObject := ctl.generateObject(*object)
			if mysqlObject != nil {
				db.Create(mysqlObject)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.ConfigMap)
			mysqlObject := ctl.generateObject(*object)
			if mysqlObject != nil {
				db.Save(mysqlObject)
			}
		},
		DeleteFunc: func(obj interface{}) {
			var item ConfigMap
			object := obj.(*v1.ConfigMap)
			db.Where("name=?", object.Name).Find(&item)
			db.Delete(item)

		},
	})
	ctl.informer = informer
}

func (ctl *ConfigMapCtl) CountWithConditions(conditions string) int {
	var object ConfigMap

	if strings.Contains(conditions, "namespace") {
		conditions = ""
	}
	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *ConfigMapCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var object ConfigMap
	var list []ConfigMap
	var total int

	if len(order) == 0 {
		order = "createTime desc"
	}

	db := ctl.DB

	listWithConditions(db, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *ConfigMapCtl) Lister() interface{} {

	return ctl.lister
}
