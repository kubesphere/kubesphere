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
	"k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (ctl *CronJobCtl) generateObject(item v1beta1.CronJob) *CronJob {
	var status, displayName string
	var lastScheduleTime *time.Time

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

	name := item.Name
	namespace := item.Namespace

	status = Running
	if *item.Spec.Suspend {
		status = Pause
	}

	schedule := item.Spec.Schedule
	if item.Status.LastScheduleTime != nil {
		lastScheduleTime = &item.Status.LastScheduleTime.Time
	}

	active := len(item.Status.Active)

	object := &CronJob{
		Namespace:        namespace,
		Name:             name,
		DisplayName:      displayName,
		LastScheduleTime: lastScheduleTime,
		Active:           active,
		Schedule:         schedule,
		Status:           status,
		Annotation:       MapString{item.Annotations},
		Labels:           MapString{item.ObjectMeta.Labels},
	}

	return object
}

func (ctl *CronJobCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *CronJobCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&CronJob{}) {
		db.DropTable(&CronJob{})
	}

	db = db.CreateTable(&CronJob{})

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

func (ctl *CronJobCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *CronJobCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)
	ctl.lister = informerFactory.Batch().V1beta1().CronJobs().Lister()

	informer := informerFactory.Batch().V1beta1().CronJobs().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			object := obj.(*v1beta1.CronJob)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1beta1.CronJob)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
		},
		DeleteFunc: func(obj interface{}) {
			var item CronJob
			object := obj.(*v1beta1.CronJob)
			db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&item)
			db.Delete(item)

		},
	})

	ctl.informer = informer
}

func (ctl *CronJobCtl) CountWithConditions(conditions string) int {
	var object CronJob

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *CronJobCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []CronJob
	var object CronJob
	var total int

	if len(order) == 0 {
		order = "lastScheduleTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *CronJobCtl) Lister() interface{} {

	return ctl.lister
}
