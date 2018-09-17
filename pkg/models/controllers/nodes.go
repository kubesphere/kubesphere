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

	"strings"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const nodeRole = "role"

func (ctl *NodeCtl) generateObject(item v1.Node) *Node {
	var status, ip, role, displayName, msgStr string
	var msg []string

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

	name := item.Name
	createTime := item.ObjectMeta.CreationTimestamp.Time
	annotation := item.Annotations

	if _, exist := item.Labels[nodeRole]; exist {
		role = item.Labels[nodeRole]
	}

	for _, condition := range item.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				status = Running
			} else {
				status = Error
			}

		} else {
			if condition.Status == "True" {
				msg = append(msg, condition.Reason)
			}
		}
	}

	if len(msg) > 0 {
		msgStr = strings.Join(msg, ",")
		if status == Running {
			status = Warning
		}
	}

	for _, address := range item.Status.Addresses {
		if address.Type == "InternalIP" {
			ip = address.Address
		}
	}

	object := &Node{
		Name:        name,
		DisplayName: displayName,
		Ip:          ip,
		Status:      status,
		CreateTime:  createTime,
		Annotation:  MapString{annotation},
		Taints:      Taints{item.Spec.Taints},
		Msg:         msgStr,
		Role:        role,
		Labels:      MapString{item.Labels}}

	return object
}

func (ctl *NodeCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *NodeCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&Node{}) {
		db.DropTable(&Node{})
	}

	db = db.CreateTable(&Node{})

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

func (ctl *NodeCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *NodeCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Core().V1().Nodes().Lister()

	informer := informerFactory.Core().V1().Nodes().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			object := obj.(*v1.Node)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.Node)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)

		},
		DeleteFunc: func(obj interface{}) {
			var item Node
			object := obj.(*v1.Node)
			db.Where("name=? ", object.Name, object.Namespace).Find(&item)
			db.Delete(item)
		},
	})

	ctl.informer = informer
}

func (ctl *NodeCtl) CountWithConditions(conditions string) int {
	var object Node

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *NodeCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []Node
	var object Node
	var total int

	if len(order) == 0 {
		order = "createTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *NodeCtl) Lister() interface{} {

	return ctl.lister
}
