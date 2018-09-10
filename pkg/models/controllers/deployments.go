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

func (ctl *DeploymentCtl) generateObject(item v1.Deployment) *Deployment {
	var app, status, displayName string
	var updateTime time.Time

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

	name := item.Name
	namespace := item.Namespace
	availablePodNum := item.Status.AvailableReplicas
	desirePodNum := *item.Spec.Replicas
	release := item.ObjectMeta.Labels["release"]
	chart := item.ObjectMeta.Labels["chart"]

	if len(release) > 0 && len(chart) > 0 {
		app = release + "/" + chart
	}

	for _, condition := range item.Status.Conditions {
		if updateTime.IsZero() {
			updateTime = condition.LastUpdateTime.Time
		} else {
			if updateTime.Before(condition.LastUpdateTime.Time) {
				updateTime = condition.LastUpdateTime.Time
			}
		}
	}
	if updateTime.IsZero() {
		updateTime = time.Now()
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

	return &Deployment{
		Namespace:   namespace,
		Name:        name,
		Available:   availablePodNum,
		Desire:      desirePodNum,
		App:         app,
		UpdateTime:  updateTime,
		Status:      status,
		Annotation:  MapString{item.Annotations},
		Labels:      MapString{item.Spec.Selector.MatchLabels},
		DisplayName: displayName,
	}
}

func (ctl *DeploymentCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *DeploymentCtl) sync(stopChan chan struct{}) {
	db := ctl.DB
	if db.HasTable(&Deployment{}) {
		db.DropTable(&Deployment{})
	}

	db = db.CreateTable(&Deployment{})

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

func (ctl *DeploymentCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", ctl.Name(), err)
		return 0
	}

	return len(list)
}

func (ctl *DeploymentCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Apps().V1().Deployments().Lister()

	informer := informerFactory.Apps().V1().Deployments().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			object := obj.(*v1.Deployment)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.Deployment)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
		},
		DeleteFunc: func(obj interface{}) {
			var deploy Deployment
			object := obj.(*v1.Deployment)
			db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&deploy)
			db.Delete(deploy)

		},
	})
	ctl.informer = informer
}

func (ctl *DeploymentCtl) CountWithConditions(conditions string) int {
	var object Deployment

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *DeploymentCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []Deployment
	var object Deployment
	var total int

	if len(order) == 0 {
		order = "updateTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *DeploymentCtl) Lister() interface{} {

	return ctl.lister
}
