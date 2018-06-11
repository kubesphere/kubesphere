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
	"strings"
	"time"

	"github.com/golang/glog"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
)

const KUBESYSTEM = "kube-system"
const OPENPITRIX = "openpitrix-system"
const ISTIO = "istio-system"
const KUBESPHERE = "kubesphere-system"

type Components struct {
	Name         string      `json:"name"`
	Version      string      `json:"version"`
	Kind         string      `json:"kind"`
	Namespace    string      `json:"namespace"`
	Label        interface{} `json:"label"`
	Replicas     int         `json:"replicas"`
	HealthStatus string      `json:"healthStatus"`
	SelfLink     string      `json:"selfLink"`
	UpdateTime   time.Time   `json:"updateTime"`
}

/***
* get all components from k8s and kubesphere system,
* there are master component, node component,addons component , kubesphere component
*
 */
func GetComponents() ([]Components, error) {

	result := make([]Components, 0)
	k8sClient := client.NewK8sClient()
	label := "tier=control-plane"
	option := meta_v1.ListOptions{

		LabelSelector: label,
	}

	podlists, err := k8sClient.CoreV1().Pods(KUBESYSTEM).List(option)

	if err != nil {

		glog.Error(err)

		return result, err
	}

	var components Components

	templates := []string{"kube-apiserver", "etcd", "kube-scheduler", "kube-controller-manager", "cloud-controller-manager"}

	if len(podlists.Items) > 0 {

		for _, pod := range podlists.Items {

			for _, template := range templates {

				if strings.Contains(pod.Name, template) {

					components.Name = template
					components.Kind = "Pod"
					components.SelfLink = pod.SelfLink
					components.Label = pod.Labels
					components.Namespace = pod.Namespace
					version := strings.Split(pod.Spec.Containers[0].Image, ":")

					if len(version) < 2 {

						components.Version = "latest"

					} else {

						components.Version = version[1]

					}
					components.Replicas = 1

					if pod.Status.Phase == "Running" {

						components.HealthStatus = "health"

					} else {

						components.HealthStatus = "unhealth"

					}
					components.UpdateTime = pod.Status.Conditions[0].LastTransitionTime.Time

					result = append(result, components)

				}

			}

		}

	}

	label = "component=kube-addon-manager"
	option.LabelSelector = label

	kubeaddon, err := k8sClient.CoreV1().Pods(KUBESYSTEM).List(option)

	if err != nil {

		glog.Error(err)

		return result, err
	}

	if len(kubeaddon.Items) > 0 {

		for _, pod := range kubeaddon.Items {

			components.Name = "kube-addon-manager"
			components.Kind = "Pod"
			components.SelfLink = pod.SelfLink
			components.Label = pod.Labels
			components.Namespace = pod.Namespace
			version := strings.Split(pod.Spec.Containers[0].Image, ":")

			if len(version) < 2 {

				components.Version = "latest"

			} else {

				components.Version = version[1]

			}
			components.Replicas = 1

			if pod.Status.Phase == "Running" {

				components.HealthStatus = "health"

			} else {

				components.HealthStatus = "unhealth"

			}
			components.UpdateTime = pod.Status.Conditions[0].LastTransitionTime.Time

			result = append(result, components)

		}

	}

	option.LabelSelector = ""
	dsList, err := k8sClient.AppsV1beta2().DaemonSets(KUBESYSTEM).List(option)

	if err != nil {

		glog.Error(err)

		return result, err
	}

	templates = []string{"flannel", "kube-proxy", "calico"}

	if len(dsList.Items) > 0 {

		for _, ds := range dsList.Items {

			for _, template := range templates {

				if strings.Contains(ds.Name, template) {

					components.Name = ds.Name
					components.Kind = "Daemonset"
					components.SelfLink = ds.SelfLink
					components.Label = ds.Spec.Selector.MatchLabels
					components.Namespace = ds.Namespace
					version := strings.Split(ds.Spec.Template.Spec.Containers[0].Image, ":")

					if len(version) < 2 {

						components.Version = "latest"

					} else {

						components.Version = version[1]

					}

					components.UpdateTime = ds.CreationTimestamp.Time
					components.Replicas = int(ds.Status.DesiredNumberScheduled)

					if ds.Status.NumberAvailable == ds.Status.DesiredNumberScheduled {

						components.HealthStatus = "health"

					} else {

						components.HealthStatus = "unhealth"

					}
					result = append(result, components)
				}

			}

		}

	}

	templates = []string{"kube-dns", "heapster", "monitoring-influxdb", "iam", "openpitrix", "istio", "kubesphere"}
	namespaces := []string{KUBESYSTEM, OPENPITRIX, ISTIO, KUBESPHERE}

	for _, ns := range namespaces {

		deployList, err := k8sClient.AppsV1beta1().Deployments(ns).List(option)

		if err != nil {

			glog.Error(err)

			return result, err
		}

		if len(deployList.Items) > 0 {

			for _, dm := range deployList.Items {

				for _, template := range templates {

					if strings.Contains(dm.Name, template) {

						components.Name = dm.Name
						components.Kind = "Deployment"
						components.SelfLink = dm.SelfLink
						components.Label = dm.Spec.Selector.MatchLabels
						components.Namespace = dm.Namespace
						components.Replicas = int(dm.Status.Replicas)
						version := strings.Split(dm.Spec.Template.Spec.Containers[0].Image, ":")
						if len(version) < 2 {

							components.Version = "latest"

						} else {

							components.Version = version[1]

						}

						components.UpdateTime = dm.Status.Conditions[0].LastUpdateTime.Time

						if dm.Status.AvailableReplicas == dm.Status.Replicas {

							components.HealthStatus = "health"

						} else {

							components.HealthStatus = "unhealth"

						}

						result = append(result, components)

					}

				}

			}

		}

	}

	return result, nil

}
