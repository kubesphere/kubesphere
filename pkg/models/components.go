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

	"github.com/golang/glog"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
)

const KUBESYSTEM = "kube-system"
const OPENPITRIX = "openpitrix-system"
const ISTIO = "istio-system"
const KUBESPHERE = "kubesphere-system"

type ComponentsCount struct {
	KubernetesCount int `json:"kubernetesCount"`
	OpenpitrixCount int `json:"openpitrixCount"`
	KubesphereCount int `json:"kubesphereCount"`
	IstioCount      int `json:"istioCount"`
}

type Components struct {
	Name         string      `json:"name"`
	Namespace    string      `json:"namespace"`
	SelfLink     string      `json:"selfLink"`
	Label        interface{} `json:"label"`
	HealthStatus string      `json:"healthStatus"`
	CreateTime   time.Time   `json:"createTime"`
}

/***
* get all components from k8s and kubesphere system
*
 */
func GetComponents() (map[string]interface{}, error) {

	result := make(map[string]interface{})
	componentsList := make([]Components, 0)
	k8sClient := client.NewK8sClient()
	var count ComponentsCount
	var components Components
	label := "kubernetes.io/cluster-service=true"
	option := meta_v1.ListOptions{

		LabelSelector: label,
	}

	namespaces := []string{KUBESYSTEM, OPENPITRIX, ISTIO, KUBESPHERE}
	for _, ns := range namespaces {

		if ns != KUBESYSTEM {
			option.LabelSelector = ""
		}
		servicelists, err := k8sClient.CoreV1().Services(ns).List(option)

		if err != nil {

			glog.Error(err)

			return result, err
		}

		if len(servicelists.Items) > 0 {

			for _, service := range servicelists.Items {

				switch ns {

				case KUBESYSTEM:
					count.KubernetesCount++
				case OPENPITRIX:
					count.OpenpitrixCount++
				case KUBESPHERE:
					count.KubesphereCount++

				default:
					count.IstioCount++
				}

				components.Name = service.Name
				components.Namespace = service.Namespace
				components.CreateTime = service.CreationTimestamp.Time
				components.Label = service.Spec.Selector
				components.SelfLink = service.SelfLink
				label := service.Spec.Selector
				combination := ""
				for key, val := range label {

					labelstr := key + "=" + val

					if combination == "" {

						combination = labelstr

					} else {

						combination = combination + "," + labelstr

					}

				}
				option := meta_v1.ListOptions{
					LabelSelector: combination,
				}
				podsList, err := k8sClient.CoreV1().Pods(ns).List(option)

				if err != nil {

					glog.Error(err)
					return result, err
				}

				if len(podsList.Items) > 0 {

					for _, pod := range podsList.Items {

						if pod.Status.Phase == "Running" {

							components.HealthStatus = "health"

						} else {

							components.HealthStatus = "unhealth"

						}

					}

				}

				componentsList = append(componentsList, components)

			}

		}

	}
	result["count"] = count
	result["item"] = componentsList
	return result, nil

}

func GetComponentsByNamespace(ns string) ([]Components, error) {

	result := make([]Components, 0)
	k8sClient := client.NewK8sClient()
	var components Components

	label := "kubernetes.io/cluster-service=true"
	option := meta_v1.ListOptions{

		LabelSelector: label,
	}
	if ns != KUBESYSTEM {
		option.LabelSelector = ""
	}
	servicelists, err := k8sClient.CoreV1().Services(ns).List(option)

	if err != nil {

		glog.Error(err)

		return result, err
	}

	if len(servicelists.Items) > 0 {

		for _, service := range servicelists.Items {

			components.Name = service.Name
			components.Namespace = service.Namespace
			components.CreateTime = service.CreationTimestamp.Time
			components.SelfLink = service.SelfLink
			components.Label = service.Spec.Selector
			label := service.Spec.Selector
			combination := ""
			for key, val := range label {

				labelstr := key + "=" + val

				if combination == "" {

					combination = labelstr

				} else {

					combination = combination + "," + labelstr

				}

			}
			option := meta_v1.ListOptions{
				LabelSelector: combination,
			}
			podsList, err := k8sClient.CoreV1().Pods(ns).List(option)

			if err != nil {

				glog.Error(err)
				return result, err
			}

			if len(podsList.Items) > 0 {

				for _, pod := range podsList.Items {

					if pod.Status.Phase == "Running" {

						components.HealthStatus = "health"

					} else {

						components.HealthStatus = "unhealth"

					}

				}

			}

			result = append(result, components)

		}

	}

	return result, nil

}
