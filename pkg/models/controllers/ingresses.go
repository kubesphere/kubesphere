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
	"encoding/json"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"kubesphere.io/kubesphere/pkg/client"
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

	annotation, _ := json.Marshal(item.Annotations)
	object := &Ingress{Namespace: namespace, Name: name, TlsTermination: tls, Ip: ip, CreateTime: createTime, AnnotationStr: string(annotation)}

	return object
}

func (ctl *IngressCtl) listAndWatch() {
	defer func() {
		defer close(ctl.aliveChan)
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

	k8sClient := client.NewK8sClient()
	list, err := k8sClient.ExtensionsV1beta1().Ingresses("").List(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range list.Items {
		obj := ctl.generateObject(item)
		db.Create(obj)
	}

	watcher, err := k8sClient.ExtensionsV1beta1().Ingresses("").Watch(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ctl.stopChan:
			return
		case event := <-watcher.ResultChan():
			var ing Ingress
			if event.Object == nil {
				panic("watch timeout, restart ingress controller")
			}
			object := event.Object.(*v1beta1.Ingress)
			if event.Type == watch.Deleted {
				db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&ing)
				db.Delete(ing)
				break
			}
			obj := ctl.generateObject(*object)
			db.Save(obj)
		}
	}
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

	for index, item := range list {
		annotation := make(map[string]string)
		json.Unmarshal([]byte(item.AnnotationStr), &annotation)
		list[index].Annotation = annotation
		list[index].AnnotationStr = ""
	}
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
