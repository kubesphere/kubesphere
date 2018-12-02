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

package models

import (
	"time"

	v12 "k8s.io/client-go/listers/core/v1"

	"kubesphere.io/kubesphere/pkg/models/controllers"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/golang/glog"
)

// Namespaces need to watch
var SYSTEM_NAMESPACES = [...]string{"kubesphere-system", "openpitrix-system", "kube-system"}

type Component struct {
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	SelfLink        string      `json:"selfLink"`
	Label           interface{} `json:"label"`
	StartedAt       time.Time   `json:"startedAt"`
	TotalBackends   int         `json:"totalBackends"`
	HealthyBackends int         `json:"healthyBackends"`
}

func GetComponentStatus(namespace string, componentName string) (interface{}, error) {
	lister, err := controllers.GetLister(controllers.Services)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}

	serviceLister := lister.(v12.ServiceLister)
	service, err := serviceLister.Services(namespace).Get(componentName)

	if err != nil {
		glog.Error(err)
		return nil, err
	}

	lister, err = controllers.GetLister(controllers.Pods)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}

	podLister := lister.(v12.PodLister)

	set := labels.Set(service.Spec.Selector)

	pods, err := podLister.Pods(namespace).List(set.AsSelector())

	if err != nil {
		glog.Errorln(err)
		return nil, err
	} else {
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

}

func GetAllComponentsStatus() (map[string]interface{}, error) {

	status := make(map[string]interface{})
	var err error

	lister, err := controllers.GetLister(controllers.Services)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}

	serviceLister := lister.(v12.ServiceLister)

	lister, err = controllers.GetLister(controllers.Pods)

	if err != nil {
		glog.Errorln(err)
		return nil, err
	}

	podLister := lister.(v12.PodLister)

	for _, ns := range SYSTEM_NAMESPACES {

		nsStatus := make(map[string]interface{})

		services, err := serviceLister.Services(ns).List(labels.Everything())
		if err != nil {
			glog.Error(err)
			continue
		}

		for _, service := range services {

			set := labels.Set(service.Spec.Selector)

			if len(set) == 0 {
				continue
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

			pods, err := podLister.Pods(ns).List(set.AsSelector())
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
