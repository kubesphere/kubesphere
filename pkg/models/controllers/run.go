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
	"github.com/jinzhu/gorm"
	"k8s.io/client-go/kubernetes"

	"kubesphere.io/kubesphere/pkg/client"
)

type resourceControllers struct {
	Controllers map[string]Controller
	k8sClient   *kubernetes.Clientset
}

var stopChan chan struct{}
var ResourceControllers resourceControllers

func (rec *resourceControllers) runContoller(name string) {
	var ctl Controller
	attr := CommonAttribute{DB: client.NewDBClient(), K8sClient: rec.k8sClient, stopChan: stopChan,
		aliveChan: make(chan struct{}), Name: name}
	switch name {
	case Deployments:
		ctl = &DeploymentCtl{CommonAttribute: attr}
	case Statefulsets:
		ctl = &StatefulsetCtl{CommonAttribute: attr}
	case Daemonsets:
		ctl = &DaemonsetCtl{CommonAttribute: attr}
	case Ingresses:
		ctl = &IngressCtl{CommonAttribute: attr}
	case PersistentVolumeClaim:
		ctl = &PvcCtl{CommonAttribute: attr}
	case Roles:
		ctl = &RoleCtl{CommonAttribute: attr}
	case ClusterRoles:
		ctl = &ClusterRoleCtl{CommonAttribute: attr}
	case Services:
		ctl = &ServiceCtl{CommonAttribute: attr}
	case Pods:
		ctl = &PodCtl{CommonAttribute: attr}
	case Namespaces:
		ctl = &NamespaceCtl{CommonAttribute: attr}
	case StorageClasses:
		ctl = &StorageClassCtl{CommonAttribute: attr}
	default:
		return
	}

	rec.Controllers[name] = ctl
	go listAndWatch(ctl)

}

func dbHealthCheck(db *gorm.DB) {
	for {
		count := 0
		var err error
		for k := 0; k < 5; k++ {
			err = db.DB().Ping()
			if err != nil {
				count++
			}
			time.Sleep(1 * time.Second)
		}

		if count > 3 {
			panic(err)
		}
	}

}

func Run() {

	stopChan := make(chan struct{})
	defer close(stopChan)

	k8sClient := client.NewK8sClient()
	ResourceControllers = resourceControllers{k8sClient: k8sClient, Controllers: make(map[string]Controller)}

	for _, item := range []string{Deployments, Statefulsets, Daemonsets, PersistentVolumeClaim, Pods, Services,
		Ingresses, Roles, ClusterRoles, Namespaces, StorageClasses} {
		ResourceControllers.runContoller(item)
	}

	go dbHealthCheck(client.NewDBClient())

	for {
		for ctlName, controller := range ResourceControllers.Controllers {
			select {
			case _, isClose := <-controller.chanAlive():
				if !isClose {
					glog.Errorf("controller %s have stopped, restart it", ctlName)
					ResourceControllers.runContoller(ctlName)
				}
			default:
				time.Sleep(5 * time.Second)
			}
		}
	}
}
