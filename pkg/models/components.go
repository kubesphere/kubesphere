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
	"kubesphere.io/kubesphere/pkg/constants"
)

type ComponentsCount struct {
	KubernetesCount int `json:"kubernetesCount"`
	OpenpitrixCount int `json:"openpitrixCount"`
	KubesphereCount int `json:"kubesphereCount"`
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

	label := ""
	namespaces := []string{constants.KubeSystemNamespace, constants.OpenPitrixNamespace, constants.KubeSphereNamespace}
	for _, ns := range namespaces {

		if ns == constants.KubeSystemNamespace {
			label = "kubernetes.io/cluster-service=true"
		} else if ns == constants.OpenPitrixNamespace {
			label = "app=openpitrix"
		} else {
			label = "app=kubesphere"
		}
		option := meta_v1.ListOptions{

			LabelSelector: label,
		}
		servicelists, err := k8sClient.CoreV1().Services(ns).List(option)

		if err != nil {

			glog.Error(err)

			return result, err
		}

		if len(servicelists.Items) > 0 {

			for _, service := range servicelists.Items {

				switch ns {

				case constants.KubeSystemNamespace:
					count.KubernetesCount++
				case constants.OpenPitrixNamespace:
					count.OpenpitrixCount++
				default:
					count.KubesphereCount++
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
					var health bool
					for _, pod := range podsList.Items {

						for _, status := range pod.Status.ContainerStatuses {

							if status.Ready == false {
								health = status.Ready
								break
							} else {
								health = status.Ready
							}

						}

						if health == false {
							components.HealthStatus = "unhealth"
							break
						}

					}

					if health == true {
						components.HealthStatus = "health"
					}

				} else {
					components.HealthStatus = "unhealth"
				}

				componentsList = append(componentsList, components)

			}

		}

	}
	result["count"] = count
	result["item"] = componentsList
	return result, nil

}
