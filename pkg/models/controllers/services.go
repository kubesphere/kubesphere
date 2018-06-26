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
	"fmt"
	"strings"
	"time"

	"strconv"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
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
	return ""
}

func (ctl *ServiceCtl) generateObject(item v1.Service) *Service {

	name := item.Name
	namespace := item.Namespace
	createTime := item.CreationTimestamp.Time
	externalIp := ctl.getExternalIp(item)
	serviceType := item.Spec.Type
	vip := item.Spec.ClusterIP
	ports := ""
	var nodePorts []string

	if createTime.IsZero() {
		createTime = time.Now()
	}

	if len(item.Spec.ExternalIPs) > 0 {
		externalIp = strings.Join(item.Spec.ExternalIPs, ",")
	}

	for _, portItem := range item.Spec.Ports {
		port := portItem.Port
		targetPort := portItem.TargetPort.String()
		protocol := portItem.Protocol
		ports += fmt.Sprintf("%d:%s/%s,", port, targetPort, protocol)

		if portItem.NodePort != 0 {
			nodePorts = append(nodePorts, strconv.FormatInt(int64(portItem.NodePort), 10))
		}
	}

	if len(ports) == 0 {
		ports = "-"
	} else {
		ports = ports[0 : len(ports)-1]
	}

	object := &Service{
		Namespace:   namespace,
		Name:        name,
		ServiceType: string(serviceType),
		ExternalIp:  externalIp,
		VirtualIp:   vip,
		CreateTime:  createTime,
		Ports:       ports,
		NodePorts:   strings.Join(nodePorts, ","),
		Annotation:  Annotation{item.Annotations},
	}

	return object
}

func (ctl *ServiceCtl) listAndWatch() {
	defer func() {
		close(ctl.aliveChan)
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

	k8sClient := ctl.K8sClient
	kubeInformerFactory := informers.NewSharedInformerFactory(k8sClient, time.Second*resyncCircle)
	informer := kubeInformerFactory.Core().V1().Services().Informer()
	lister := kubeInformerFactory.Core().V1().Services().Lister()

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

			object := obj.(*v1.Service)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.Service)
			mysqlObject := ctl.generateObject(*object)
			db.Save(mysqlObject)
		},
		DeleteFunc: func(obj interface{}) {
			var item Service
			object := obj.(*v1.Service)
			db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&item)
			db.Delete(item)

		},
	})

	informer.Run(ctl.stopChan)
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
