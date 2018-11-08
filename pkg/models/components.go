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

	"k8s.io/apimachinery/pkg/labels"

	"github.com/golang/glog"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
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
	k8sClient := client.NewK8sClient()

	if service, err := k8sClient.CoreV1().Services(namespace).Get(componentName, meta_v1.GetOptions{}); err != nil {
		glog.Error(err)
		return nil, err
	} else {
		set := labels.Set(service.Spec.Selector)

		component := Component{
			Name:            service.Name,
			Namespace:       service.Namespace,
			SelfLink:        service.SelfLink,
			Label:           service.Spec.Selector,
			StartedAt:       service.CreationTimestamp.Time,
			HealthyBackends: 0,
			TotalBackends:   0,
		}

		if pods, err := k8sClient.CoreV1().Pods(namespace).List(meta_v1.ListOptions{LabelSelector: set.AsSelector().String()}); err != nil {
			glog.Error(err)
			return nil, err
		} else {
			for _, v := range pods.Items {
				component.TotalBackends++
				component.HealthyBackends++
				for _, c := range v.Status.ContainerStatuses {
					if !c.Ready {
						component.HealthyBackends--
						break
					}
				}
			}
		}

		return component, nil
	}

}

func GetAllComponentsStatus() (map[string]interface{}, error) {

	status := make(map[string]interface{})
	var err error

	k8sClient := client.NewK8sClient()

	for _, ns := range SYSTEM_NAMESPACES {

		nsStatus := make(map[string]interface{})

		services, err := k8sClient.CoreV1().Services(ns).List(meta_v1.ListOptions{})
		if err != nil {
			glog.Error(err)
			continue
		}

		for _, service := range services.Items {

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

			if pods, err := k8sClient.CoreV1().Pods(ns).List(meta_v1.ListOptions{LabelSelector: set.AsSelector().String()}); err != nil {
				glog.Error(err)
				continue
			} else {
				for _, v := range pods.Items {
					component.TotalBackends++
					component.HealthyBackends++
					for _, c := range v.Status.ContainerStatuses {
						if !c.Ready {
							component.HealthyBackends--
							break
						}
					}
				}
			}

			nsStatus[service.Name] = component
		}

		status[ns] = nsStatus
	}

	return status, err
}
