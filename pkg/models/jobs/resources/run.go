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

package resources

import (
	"encoding/json"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"time"
)

var etcdClient *client.EtcdClient

var stopChan = make(chan struct{})

const (
	pods                  = "pods"
	deployments           = "deployments"
	daemonsets            = "daemonsets"
	statefulsets          = "statefulsets"
	namespaces            = "namespaces"
	ingresses             = "ingresses"
	persistentVolumeClaim = "persistent-volume-claim"
	roles                 = "roles"
	services              = "services"
)

func registerResource(resourceChans map[string]ResourceChan, resourceType string) {
	resourceChan := ResourceChan{Type: resourceType, StatusChan: make(chan *ResourceStatus), StopChan: stopChan}
	resourceChans[resourceType] = resourceChan
}

func updateStatus(resource Resource, resourceChan ResourceChan) {

	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			close(resourceChan.StatusChan)
		}
	}()

	var clusterStatus ResourceStatus
	clusterStatus.UpdateTimeStamp = time.Now().Unix()
	clusterStatus.ResourceType = resourceChan.Type

	items, err := resource.list()
	if err != nil {
		glog.Errorln(err)
		return
	}
	resource.updateWithObjects(&clusterStatus, items)

	watcher, err := resource.getWatcher()
	if err != nil {
		glog.Error(err)
		return
	}

	for {
		select {
		case <-resourceChan.StopChan:
			return
		case event := <-watcher.ResultChan():
			resource.updateWithEvent(&clusterStatus, event)
			break

		default:
			break
		}

		if time.Now().Unix()-clusterStatus.UpdateTimeStamp > constants.UpdateCircle {
			clusterStatus.UpdateTimeStamp = time.Now().Unix()
			resourceChan.StatusChan <- &clusterStatus

		}

	}
}

func updateResourceStatus(resourceChan ResourceChan) {
	glog.Infof("updateResourceStatus:%s", resourceChan.Type)
	client := client.NewK8sClient()
	switch resourceChan.Type {
	case deployments:
		deploy := deployment{k8sClient: client}
		go updateStatus(&deploy, resourceChan)
	case daemonsets:
		ds := daemonset{k8sClient: client}
		go updateStatus(&ds, resourceChan)
	case statefulsets:
		ss := statefulset{k8sClient: client}
		go updateStatus(&ss, resourceChan)
	case namespaces:
		ns := namespace{k8sClient: client}
		go updateStatus(&ns, resourceChan)
	case ingresses:
		ing := ingress{k8sClient: client}
		go updateStatus(&ing, resourceChan)
	case persistentVolumeClaim:
		pvc := persistmentVolume{k8sClient: client}
		go updateStatus(&pvc, resourceChan)
	case roles:
		r := role{k8sClient: client}
		go updateStatus(&r, resourceChan)
	case services:
		svc := service{k8sClient: client}
		go updateStatus(&svc, resourceChan)
	case pods:
		po := pod{k8sClient: client}
		go updateStatus(&po, resourceChan)
	}

}

func updateAllResourceStatus(resourceChans map[string]ResourceChan) {
	for _, resourceChan := range resourceChans {
		updateResourceStatus(resourceChan)
	}
}

func receiveResourceStatus(resourceChans map[string]ResourceChan) {
	defer func() {
		close(stopChan)
	}()

	for {
		for _, resourceChan := range resourceChans {
			select {
			case res, ok := <-resourceChan.StatusChan:
				if !ok {
					glog.Errorf("job:calculate %s' status have stopped", resourceChan.Type)
					registerResource(resourceChans, resourceChan.Type)
					updateResourceStatus(resourceChans[resourceChan.Type])
				} else {
					value, _ := json.Marshal(res)
					key := constants.Root + "/" + res.ResourceType
					etcdClient.Put(key, string(value))
				}
			default:
				continue
			}
		}
	}
}

func Run() {
	glog.Info("Begin to submit resource status")
	var err error
	etcdClient, err = client.NewEtcdClient()
	defer etcdClient.Close()
	if err != nil {
		glog.Error(err)
	}
	resourceChans := make(map[string]ResourceChan)
	resourceList := []string{statefulsets, deployments, daemonsets, namespaces, ingresses, services, roles, persistentVolumeClaim, pods}
	for _, resource := range resourceList {
		registerResource(resourceChans, resource)
	}
	updateAllResourceStatus(resourceChans)
	receiveResourceStatus(resourceChans)
}
