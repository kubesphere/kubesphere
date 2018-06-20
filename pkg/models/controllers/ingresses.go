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
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (ctl *IngressCtl) generateObject(item v1beta1.Ingress) *Ingress {
	name := item.Name
	namespace := item.Namespace
	ip := "-"
	tls := "-"
	createTime := item.CreationTimestamp.Time
	if createTime.IsZero() {
		createTime = time.Now()
	}

	var ipList []string
	for _, lb := range item.Status.LoadBalancer.Ingress {
		if len(lb.IP) > 0 {
			ipList = append(ipList, lb.IP)
		}
	}
	if len(ipList) > 0 {
		ip = strings.Join(ipList, ",")
	}

	object := &Ingress{Namespace: namespace, Name: name, TlsTermination: tls, Ip: ip, CreateTime: createTime, Annotation: Annotation{item.Annotations}}

	return object
}

func (ctl *IngressCtl) listAndWatch() {
	defer func() {
		close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&Ingress{}) {
		db.DropTable(&Ingress{})

	}

	db = db.CreateTable(&Ingress{})

	k8sClient := ctl.K8sClient
	kubeInformerFactory := informers.NewSharedInformerFactory(k8sClient, time.Second*resyncCircle)
	informer := kubeInformerFactory.Extensions().V1beta1().Ingresses().Informer()
	lister := kubeInformerFactory.Extensions().V1beta1().Ingresses().Lister()

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

			object := obj.(*v1beta1.Ingress)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1beta1.Ingress)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
		},
		DeleteFunc: func(obj interface{}) {
			var item Ingress
			object := obj.(*v1beta1.Ingress)
			db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&item)
			db.Delete(item)

		},
	})

	informer.Run(ctl.stopChan)
}

func (ctl *IngressCtl) CountWithConditions(conditions string) int {
	var object Ingress

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *IngressCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Ingress
	var object Ingress
	var total int

	order := "createTime desc"

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *IngressCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	if len(namespace) == 0 {
		db.Model(&Ingress{}).Count(&count)
	} else {
		db.Model(&Ingress{}).Where("namespace = ?", namespace).Count(&count)
	}
	return count
}
