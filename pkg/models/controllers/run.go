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
	controllers map[string]Controller
	db          *gorm.DB
	k8sClient   *kubernetes.Clientset
}

var stopChan chan struct{}

func (rec *resourceControllers) runContoller(name string) {
	var ctl Controller
	attr := CommonAttribute{DB: rec.db, K8sClient: rec.k8sClient, stopChan: stopChan, aliveChan: make(chan struct{})}
	switch name {
	case Deployments:
		ctl = &DeploymentCtl{attr}
	case Statefulsets:
		ctl = &StatefulsetCtl{attr}
	case Daemonsets:
		ctl = &DaemonsetCtl{attr}
	case Ingresses:
		ctl = &IngressCtl{attr}
	case PersistentVolumeClaim:
		ctl = &PvcCtl{attr}
	case Roles:
		ctl = &RoleCtl{attr}
	case ClusterRoles:
		ctl = &ClusterRoleCtl{attr}
	case Services:
		ctl = &ServiceCtl{attr}
	case Pods:
		ctl = &PodCtl{attr}
	case Namespaces:
		ctl = &NamespaceCtl{attr}
	case StorageClasses:
		ctl = &StorageClassCtl{attr}
	default:
		return
	}

	rec.controllers[name] = ctl
	go ctl.listAndWatch()

}

func Run() {
	db := client.NewDBClient()

	stopChan := make(chan struct{})
	defer db.Commit()
	defer db.Close()
	defer close(stopChan)

	rec := resourceControllers{k8sClient: client.NewK8sClient(), db: db, controllers: make(map[string]Controller)}

	for _, item := range []string{Deployments, Statefulsets, Daemonsets, PersistentVolumeClaim, Pods, Services,
		Ingresses, Roles, ClusterRoles, Namespaces, StorageClasses} {
		rec.runContoller(item)
	}

	for {
		for ctlName, controller := range rec.controllers {
			select {
			case _, isClose := <-controller.chanAlive():
				if !isClose {
					glog.Errorf("controller %s have stopped, restart it", ctlName)
					rec.runContoller(ctlName)
				}
			default:
				time.Sleep(5 * time.Second)
			}
		}
	}
}
