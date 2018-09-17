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
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const inUse = "kubesphere.io/in_use_pods"

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
				json.Unmarshal([]byte(annotation[inUse]), &pods)
				for _, pod := range pods {
					if pod == item.Name {
						return
					}
				}
				pods = append(pods, item.Name)
				str, _ := json.Marshal(pods)
				annotation[inUse] = string(str)
			}
			Pvc.Annotations = annotation
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
			if Pvc.Annotations == nil {
				Pvc.Annotations = make(map[string]string)
			}
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

func getStatusAndRestartCount(pod v1.Pod) (string, int) {
	status := string(pod.Status.Phase)
	restarts := 0
	if pod.Status.Reason != "" {
		status = pod.Status.Reason
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					status = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					status = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				status = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			status = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			status = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				status = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				status = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					status = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					status = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if status == "Completed" && hasRunning {
			status = "Running"
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		status = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		status = "Terminating"
	}

	return status, restarts
}

func (ctl *PodCtl) generateObject(item v1.Pod) *Pod {
	name := item.Name
	namespace := item.Namespace
	podIp := item.Status.PodIP
	nodeName := item.Spec.NodeName
	nodeIp := item.Status.HostIP
	status, restartCount := getStatusAndRestartCount(item)
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

	object := &Pod{
		Namespace:    namespace,
		Name:         name,
		Node:         nodeName,
		PodIp:        podIp,
		Status:       status,
		NodeIp:       nodeIp,
		CreateTime:   createTime,
		Annotation:   MapString{item.Annotations},
		Containers:   containers,
		RestartCount: restartCount,
		Labels:       MapString{item.Labels},
	}

	return object
}

func (ctl *PodCtl) Name() string {
	return ctl.CommonAttribute.Name
}

func (ctl *PodCtl) sync(stopChan chan struct{}) {
	db := ctl.DB

	if db.HasTable(&Pod{}) {
		db.DropTable(&Pod{})
	}

	db = db.CreateTable(&Pod{})

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

func (ctl *PodCtl) total() int {
	list, err := ctl.lister.List(labels.Everything())
	if err != nil {
		glog.Errorf("count %s falied, reason:%s", err, ctl.Name())
		return 0
	}
	return len(list)
}

func (ctl *PodCtl) initListerAndInformer() {
	db := ctl.DB

	informerFactory := informers.NewSharedInformerFactory(ctl.K8sClient, time.Second*resyncCircle)

	ctl.lister = informerFactory.Core().V1().Pods().Lister()

	informer := informerFactory.Core().V1().Pods().Informer()
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
			ctl.addAnnotationToPvc(*object)
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

	ctl.informer = informer
}

func (ctl *PodCtl) CountWithConditions(conditions string) int {
	var object Pod

	return countWithConditions(ctl.DB, conditions, &object)
}

func (ctl *PodCtl) ListWithConditions(conditions string, paging *Paging, order string) (int, interface{}, error) {
	var list []Pod
	var object Pod
	var total int

	if len(order) == 0 {
		order = "createTime desc"
	}

	listWithConditions(ctl.DB, &total, &object, &list, conditions, paging, order)

	return total, list, nil
}

func (ctl *PodCtl) Lister() interface{} {

	return ctl.lister
}
