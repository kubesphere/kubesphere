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
	"time"

	"github.com/golang/glog"

	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (ctl *StatefulsetCtl) generateObject(item v1.StatefulSet) *Statefulset {
	var app string
	var status string
	var displayName string

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}
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
	}

	if item.Annotations["state"] == "stop" {
		status = Stopped
	} else {
		if availablePodNum >= desirePodNum {
			status = Running
		} else {
			status = Updating
		}
	}

	statefulSetObject := &Statefulset{
		Namespace:   namespace,
		Name:        name,
		DisplayName: displayName,
		Available:   availablePodNum,
		Desire:      desirePodNum,
		App:         app,
		CreateTime:  createTime,
		Status:      status,
		Annotation:  MapString{item.Annotations},
		Labels:      MapString{item.Spec.Selector.MatchLabels},
	}

	return statefulSetObject
}

func (ctl *StatefulsetCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *StatefulsetCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&Statefulset{}) {
		db.DropTable(&Statefulset{})
	}

	db = db.CreateTable(&Statefulset{})

	ctl.initListerAndInformer()
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list {
		obj := ctl.generateObject(*item)
		db.Create(obj)
	}

	ctl.informer.Run(stopChan)
}

func (ctl *StatefulsetCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *StatefulsetCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Apps().V1().StatefulSets().Lister()

	informer := informerFactory.Apps().V1().StatefulSets().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			object := obj.(*v1.StatefulSet)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.StatefulSet)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
		},
		DeleteFunc: func(obj interface{}) {
			var item Statefulset
			object := obj.(*v1.StatefulSet)
			db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&item)
			db.Delete(item)

		},
	})

	ctl.informer = informer
}

func (ctl *StatefulsetCtl) CountWithConditions(conditions string) int {
	var object Statefulset

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *StatefulsetCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []Statefulset
	var object Statefulset
	var total int

	if len(order) == 0 {
		order = "createTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *StatefulsetCtl) Lister() interface{} {

	return ctl.lister
}
