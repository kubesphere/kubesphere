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
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"kubesphere.io/kubesphere/pkg/models/metrics"
)

const inUse = "in_use_pods"

func (ctl *PodCtl) addAnnotationToPvc(item v1.Pod) {
	volumes := item.Spec.Volumes
	for _, volume := range volumes {
		pvc := volume.PersistentVolumeClaim
		if pvc != nil {
			name := pvc.ClaimName

			Pvc, _ := ctl.K8sClient.CoreV1().PersistentVolumeClaims(item.Namespace).Get(name, metaV1.GetOptions{})
			if Pvc.Annotations == nil {
				Pvc.Annotations = make(map[string]string)
			}
			annotation := Pvc.Annotations
			if len(annotation[inUse]) == 0 {
				pods := []string{item.Name}
				str, _ := json.Marshal(pods)
				annotation[inUse] = string(str)
			} else {
				var pods []string
				json.Unmarshal([]byte(annotation[inUse]), pods)
				for _, pod := range pods {
					if pod == item.Name {
						return
					}
					pods = append(pods, item.Name)
					str, _ := json.Marshal(pods)
					annotation[inUse] = string(str)
				}
			}
			ctl.K8sClient.CoreV1().PersistentVolumeClaims(item.Namespace).Update(Pvc)
		}
	}
}

func (ctl *PodCtl) delAnnotationFromPvc(item v1.Pod) {
	volumes := item.Spec.Volumes
	for _, volume := range volumes {
		pvc := volume.PersistentVolumeClaim
		if pvc != nil {
			name := pvc.ClaimName
			Pvc, _ := ctl.K8sClient.CoreV1().PersistentVolumeClaims(item.Namespace).Get(name, metaV1.GetOptions{})
			annotation := Pvc.Annotations
			var pods []string
			json.Unmarshal([]byte(annotation[inUse]), pods)

			for index, pod := range pods {
				if pod == item.Name {
					pods = append(pods[:index], pods[index+1:]...)
				}
			}

			str, _ := json.Marshal(pods)
			annotation[inUse] = string(str)
			ctl.K8sClient.CoreV1().PersistentVolumeClaims(item.Namespace).Update(Pvc)
		}
	}
}

func (ctl *PodCtl) generateObject(item v1.Pod) *Pod {
	name := item.Name
	namespace := item.Namespace
	podIp := item.Status.PodIP
	nodeName := item.Spec.NodeName
	nodeIp := item.Status.HostIP
	status := string(item.Status.Phase)
	createTime := item.CreationTimestamp.Time
	containerStatus := item.Status.ContainerStatuses
	containerSpecs := item.Spec.Containers

	var containers Containers

	for _, containerSpec := range containerSpecs {
		var container Container
		container.Name = containerSpec.Name
		container.Image = containerSpec.Image
		container.Ports = containerSpec.Ports
		container.Resources = containerSpec.Resources
		for _, status := range containerStatus {
			if container.Name == status.Name {
				container.Ready = status.Ready
			}
		}

		containers = append(containers, container)
	}

	object := &Pod{Namespace: namespace, Name: name, Node: nodeName, PodIp: podIp, Status: status, NodeIp: nodeIp,
		CreateTime: createTime, Annotation: Annotation{item.Annotations}, Containers: containers}

	return object
}

func (ctl *PodCtl) listAndWatch() {
	db := ctl.DB

	if db.HasTable(&Pod{}) {
		db.DropTable(&Pod{})

	}

	db = db.CreateTable(&Pod{})

	k8sClient := ctl.K8sClient
	kubeInformerFactory := informers.NewSharedInformerFactory(k8sClient, time.Second*resyncCircle)
	informer := kubeInformerFactory.Core().V1().Pods().Informer()
	lister := kubeInformerFactory.Core().V1().Pods().Lister()

	list, err := lister.List(labels.Everything())
	if err != nil {
		glog.Error(err)
		panic(err)
	}

	for _, item := range list {
		obj := ctl.generateObject(*item)
		db.Create(obj)
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			object := obj.(*v1.Pod)
			mysqlObject := ctl.generateObject(*object)
			db.Create(mysqlObject)

			ctl.addAnnotationToPvc(*object)
		},
		UpdateFunc: func(old, new interface{}) {
			object := new.(*v1.Pod)
			mysqlObject := ctl.generateObject(*object)

			db.Save(mysqlObject)

		},
		DeleteFunc: func(obj interface{}) {
			var item Pod
			object := obj.(*v1.Pod)
			db.Where("name=? And namespace=?", object.Name, object.Namespace).Find(&item)
			ctl.delAnnotationFromPvc(*object)
			db.Delete(item)
		},
	})

	informer.Run(ctl.stopChan)
}

func (ctl *PodCtl) CountWithConditions(conditions string) int {
	var object Pod

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *PodCtl) ListWithConditions(conditions string, paging *Paging) (int, interface{}, error) {
	var list []Pod
	var object Pod
	var total int

	order := "createTime desc"

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	ch := make(chan metrics.PodMetrics)

	for index, _ := range list {
		go metrics.GetSinglePodMetrics(list[index].Namespace, list[index].Name, ch)
	}

	var resultMetrics = make(map[string]metrics.PodMetrics)
	for range list {
		podMetric := <-ch
		resultMetrics[podMetric.PodName] = podMetric
	}

	for index, _ := range list {
		list[index].Metrics = resultMetrics[list[index].Name]
	}

	return total, list, nil
}

func (ctl *PodCtl) Count(namespace string) int {
	var count int
	db := ctl.DB
	if len(namespace) == 0 {
		db.Model(&Pod{}).Count(&count)
	} else {
		db.Model(&Pod{}).Where("namespace = ?", namespace).Count(&count)
	}
	return count
}
