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

	"os"
	"sync"
	"syscall"

	"kubesphere.io/kubesphere/pkg/client"
)

type resourceControllers struct {
	Controllers map[string]Controller
	k8sClient   *kubernetes.Clientset
}

var ResourceControllers resourceControllers

func (rec *resourceControllers) runContoller(name string, stopChan chan struct{}, wg *sync.WaitGroup) {
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
	case Jobs:
		ctl = &JobCtl{CommonAttribute: attr}
	case Cronjobs:
		ctl = &CronJobCtl{CommonAttribute: attr}
	case Nodes:
		ctl = &NodeCtl{CommonAttribute: attr}
	case Replicasets:
		ctl = &ReplicaSetCtl{CommonAttribute: attr}
	case ControllerRevisions:
		ctl = &ControllerRevisionCtl{CommonAttribute: attr}
	case ConfigMaps:
		ctl = &ConfigMapCtl{CommonAttribute: attr}
	case Secrets:
		ctl = &SecretCtl{CommonAttribute: attr}
	default:
		return
	}

	rec.Controllers[name] = ctl
	wg.Add(1)
	go listAndWatch(ctl, wg)

}

func dbHealthCheck(db *gorm.DB) {
	defer db.Close()

	for {
		count := 0
		var err error
		for k := 0; k < 5; k++ {
			err = db.DB().Ping()
			if err != nil {
				count++
			}
			time.Sleep(5 * time.Second)
		}

		if count > 3 {
			syscall.Kill(os.Getpid(), syscall.SIGTERM)
		}
	}

}

func Run(stopChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	k8sClient := client.NewK8sClient()
	ResourceControllers = resourceControllers{k8sClient: k8sClient, Controllers: make(map[string]Controller)}

	for _, item := range []string{Deployments, Statefulsets, Daemonsets, PersistentVolumeClaim, Pods, Services,
		Ingresses, Roles, ClusterRoles, Namespaces, StorageClasses, Jobs, Cronjobs, Nodes, Replicasets,
		ControllerRevisions, ConfigMaps, Secrets} {
		ResourceControllers.runContoller(item, stopChan, wg)
	}

	go dbHealthCheck(client.NewDBClient())

	for {
		for ctlName, controller := range ResourceControllers.Controllers {
			select {
			case <-stopChan:
				return
			case _, isClose := <-controller.chanAlive():
				if !isClose {
					glog.Errorf("controller %s have stopped, restart it", ctlName)
					ResourceControllers.runContoller(ctlName, stopChan, wg)
				}
			default:
				time.Sleep(3 * time.Second)
			}
		}
	}
}
