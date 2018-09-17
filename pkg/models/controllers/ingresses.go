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

	"encoding/json"

	"github.com/golang/glog"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (ctl *IngressCtl) generateObject(item v1beta1.Ingress) *Ingress {

	var ip, tls, displayName string

	name := item.Name
	namespace := item.Namespace

	if item.Annotations != nil && len(item.Annotations[DisplayName]) > 0 {
		displayName = item.Annotations[DisplayName]
	}

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

	var ingRules []ingressRule
	for _, rule := range item.Spec.Rules {
		host := rule.Host
		for _, path := range rule.HTTP.Paths {
			var ingRule ingressRule
			ingRule.Host = host
			ingRule.Service = path.Backend.ServiceName
			ingRule.Port = path.Backend.ServicePort.IntVal
			ingRule.Path = path.Path
			ingRules = append(ingRules, ingRule)
		}
	}

	ruleStr, _ := json.Marshal(ingRules)

	object := &Ingress{
		Namespace:      namespace,
		Name:           name,
		DisplayName:    displayName,
		TlsTermination: tls,
		Ip:             ip,
		CreateTime:     createTime,
		Annotation:     MapString{item.Annotations},
		Rules:          string(ruleStr),
		Labels:         MapString{item.Labels},
	}

	return object
}

func (ctl *IngressCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *IngressCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&Ingress{}) {
		db.DropTable(&Ingress{})
	}

	db = db.CreateTable(&Ingress{})

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

func (ctl *IngressCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *IngressCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Extensions().V1beta1().Ingresses().Lister()

	informer := informerFactory.Extensions().V1beta1().Ingresses().Informer()
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

	ctl.informer = informer
}

func (ctl *IngressCtl) CountWithConditions(conditions string) int {
	var object Ingress

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *IngressCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []Ingress
	var object Ingress
	var total int

	if len(order) == 0 {
		order = "createTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *IngressCtl) Lister() interface{} {

	return ctl.lister
}
