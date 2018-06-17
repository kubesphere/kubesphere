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
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"

	"kubesphere.io/kubesphere/pkg/client"
)

const (
	headlessSelector = "Headless(Selector)"
	headlessExternal = "Headless(ExternalName)"
	virtualIp        = "Virtual IP"
)

func (ctl *ServiceCtl) loadBalancerStatusStringer(item v1.Service) string {
	ingress := item.Status.LoadBalancer.Ingress
	result := sets.NewString()
	for i := range ingress {
		if ingress[i].IP != "" {
			result.Insert(ingress[i].IP)
		} else if ingress[i].Hostname != "" {
			result.Insert(ingress[i].Hostname)
		}
	}

	r := strings.Join(result.List(), ",")
	return r
}

func (ctl *ServiceCtl) getExternalIp(item v1.Service) string {
	switch item.Spec.Type {
	case "ClusterIP", "NodePort":
		if len(item.Spec.ExternalIPs) > 0 {
			return strings.Join(item.Spec.ExternalIPs, ",")
		}
	case "ExternalName":
		return item.Spec.ExternalName

	case "LoadBalancer":
		lbIps := ctl.loadBalancerStatusStringer(item)
		if len(item.Spec.ExternalIPs) > 0 {
			results := []string{}
			if len(lbIps) > 0 {
				results = append(results, strings.Split(lbIps, ",")...)
			}
			results = append(results, item.Spec.ExternalIPs...)
			return strings.Join(results, ",")
		}
		if len(lbIps) > 0 {
			return lbIps
		}
		return "<pending>"
	}
	return "-"
}

func (ctl *ServiceCtl) generateObject(item v1.Service) *Service {

	name := item.Name
	namespace := item.Namespace
	createTime := item.CreationTimestamp.Time
	externalIp := ctl.getExternalIp(item)
	serviceType := virtualIp
	vip := item.Spec.ClusterIP
	ports := ""

	if createTime.IsZero() {
		createTime = time.Now()
	}

	if item.Spec.ClusterIP == "None" {
		serviceType = headlessSelector
		vip = "-"
	}

	if len(item.Spec.ExternalName) > 0 {
		serviceType = headlessExternal
		vip = "-"
	}

	if len(item.Spec.ExternalIPs) > 0 {
		externalIp = strings.Join(item.Spec.ExternalIPs, ",")
	}

	for _, portItem := range item.Spec.Ports {
		port := portItem.Port
		targetPort := portItem.TargetPort.String()
		protocol := portItem.Protocol
		ports += fmt.Sprintf("%d:%s/%s,", port, targetPort, protocol)
	}
	if len(ports) == 0 {
		ports = "-"
	} else {
		ports = ports[0 : len(ports)-1]
	}

	annotation, _ := json.Marshal(item.Annotations)
	object := &Service{Namespace: namespace, Name: name, ServiceType: serviceType, ExternalIp: externalIp,
		VirtualIp: vip, CreateTime: createTime, Ports: ports, AnnotationStr: string(annotation)}

	return object
}

func (ctl *ServiceCtl) listAndWatch() {
	defer func() {
		defer close(ctl.aliveChan)
		if err := recover(); err != nil {
			glog.Error(err)
			return
		}
	}()

	db := ctl.DB

	if db.HasTable(&Service{}) {
		db.DropTable(&Service{})
	}

	db = db.CreateTable(&Service{})

	k8sClient := client.NewK8sClient()
	svcList, err := k8sClient.CoreV1().Services("").List(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for _, item := range svcList.Items {
		obj := ctl.generateObject(item)
		db.Create(obj)
	}

	watcher, err := k8sClient.CoreV1().Services("").Watch(metaV1.ListOptions{})
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-ctl.stopChan:
			return
		case event := <-watcher.ResultChan():
			var svc Service

			if event.Object == nil {
				panic("watch timeout, restart service controller")
			}
			object := event.Object.(*v1.Service)

			if event.Type == watch.Deleted {
				db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&svc)
				db.Delete(svc)
				break
			}
			obj := ctl.generateObject(*object)
			db.Save(obj)
		}
	}
}

func (ctl *ServiceCtl) CountWithConditions(conditions string) int {
	var object Service

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *ServiceCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Service
	var object Service
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

func (ctl *ServiceCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	if len(namespace) == 0 {
		db.Model(&Service{}).Count(&count)
	} else {
		db.Model(&Service{}).Where("namespace = ?", namespace).Count(&count)
	}
	return count
}
