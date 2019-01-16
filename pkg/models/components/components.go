/*

 Copyright 2019 The KubeSphere Authors.

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
package components

import (
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	lister "k8s.io/client-go/listers/core/v1"

	"kubesphere.io/kubesphere/pkg/client"

	"kubesphere.io/kubesphere/pkg/informers"

	"github.com/golang/glog"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/kubesphere/pkg/constants"
)

type Component struct {
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	SelfLink        string      `json:"selfLink"`
	Label           interface{} `json:"label"`
	StartedAt       time.Time   `json:"startedAt"`
	TotalBackends   int         `json:"totalBackends"`
	HealthyBackends int         `json:"healthyBackends"`
}

var (
	//componentStatusLister lister.ComponentStatusLister
	serviceLister lister.ServiceLister
	podLister     lister.PodLister
	nodeLister    lister.NodeLister
)

func init() {
	//componentStatusLister = informers.SharedInformerFactory().Core().V1().ComponentStatuses().Lister()
	serviceLister = informers.SharedInformerFactory().Core().V1().Services().Lister()
	podLister = informers.SharedInformerFactory().Core().V1().Pods().Lister()
	nodeLister = informers.SharedInformerFactory().Core().V1().Nodes().Lister()
}

func GetComponentStatus(name string) (interface{}, error) {

	var service *coreV1.Service
	var err error
	for _, ns := range constants.SystemNamespaces {
		service, err = serviceLister.Services(ns).Get(name)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	pods, err := podLister.Pods(service.Namespace).List(labels.SelectorFromValidatedSet(service.Spec.Selector))

	if err != nil {
		return nil, err
	}

	component := Component{
		Name:            service.Name,
		Namespace:       service.Namespace,
		SelfLink:        service.SelfLink,
		Label:           service.Spec.Selector,
		StartedAt:       service.CreationTimestamp.Time,
		HealthyBackends: 0,
		TotalBackends:   0,
	}
	for _, v := range pods {
		component.TotalBackends++
		component.HealthyBackends++
		for _, c := range v.Status.ContainerStatuses {
			if !c.Ready {
				component.HealthyBackends--
				break
			}
		}
	}
	return component, nil
}

func GetSystemHealthStatus() (map[string]interface{}, error) {

	status := make(map[string]interface{})

	componentStatuses, err := client.K8sClient().CoreV1().ComponentStatuses().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, cs := range componentStatuses.Items {
		status[cs.Name] = cs.Conditions[0]
	}

	// get kubesphere-system components
	systemComponentStatus, err := GetAllComponentsStatus()
	if err != nil {
		glog.Errorln(err)
	}

	for k, v := range systemComponentStatus {
		status[k] = v
	}
	// get node status
	nodes, err := nodeLister.List(labels.Everything())
	if err != nil {
		glog.Errorln(err)
		return status, nil
	}

	nodeStatus := make(map[string]int)
	totalNodes := 0
	healthyNodes := 0
	for _, nodes := range nodes {
		totalNodes++
		for _, condition := range nodes.Status.Conditions {
			if condition.Type == coreV1.NodeReady && condition.Status == coreV1.ConditionTrue {
				healthyNodes++
			}
		}
	}
	nodeStatus["totalNodes"] = totalNodes
	nodeStatus["healthyNodes"] = healthyNodes

	status["nodes"] = nodeStatus

	return status, nil

}

func GetAllComponentsStatus() (map[string]interface{}, error) {

	status := make(map[string]interface{})

	var err error

	for _, ns := range constants.SystemNamespaces {

		nsStatus := make(map[string]interface{})

		services, err := serviceLister.Services(ns).List(labels.Everything())

		if err != nil {
			glog.Error(err)
			continue
		}

		for _, service := range services {
			component := Component{
				Name:            service.Name,
				Namespace:       service.Namespace,
				SelfLink:        service.SelfLink,
				Label:           service.Spec.Selector,
				StartedAt:       service.CreationTimestamp.Time,
				HealthyBackends: 0,
				TotalBackends:   0,
			}

			pods, err := podLister.Pods(ns).List(labels.SelectorFromValidatedSet(service.Spec.Selector))

			if err != nil {
				glog.Errorln(err)
				continue
			}

			for _, pod := range pods {
				component.TotalBackends++
				component.HealthyBackends++
				for _, c := range pod.Status.ContainerStatuses {
					if !c.Ready {
						component.HealthyBackends--
						break
					}
				}
			}

			nsStatus[service.Name] = component
		}

		if len(nsStatus) > 0 {
			status[ns] = nsStatus
		}
	}

	return status, err
}
