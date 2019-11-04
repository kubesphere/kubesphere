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
	"fmt"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/kubesphere/pkg/constants"
)

func GetComponentStatus(name string) (interface{}, error) {

	var service *corev1.Service
	var err error

	serviceLister := informers.SharedInformerFactory().Core().V1().Services().Lister()

	for _, ns := range constants.SystemNamespaces {
		service, err = serviceLister.Services(ns).Get(name)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	if len(service.Spec.Selector) == 0 {
		return nil, fmt.Errorf("component %s has no selector", name)
	}

	podLister := informers.SharedInformerFactory().Core().V1().Pods().Lister()

	pods, err := podLister.Pods(service.Namespace).List(labels.SelectorFromValidatedSet(service.Spec.Selector))

	if err != nil {
		return nil, err
	}

	component := models.ComponentStatus{
		Name:            service.Name,
		Namespace:       service.Namespace,
		SelfLink:        service.SelfLink,
		Label:           service.Spec.Selector,
		StartedAt:       service.CreationTimestamp.Time,
		HealthyBackends: 0,
		TotalBackends:   0,
	}
	for _, pod := range pods {
		component.TotalBackends++
		if pod.Status.Phase == corev1.PodRunning && isAllContainersReady(pod) {
			component.HealthyBackends++
		}
	}
	return component, nil
}

func isAllContainersReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.ContainerStatuses {
		if !c.Ready {
			return false
		}
	}
	return true
}

func GetSystemHealthStatus() (*models.HealthStatus, error) {

	status := &models.HealthStatus{}

	// get kubesphere-system components
	components, err := GetAllComponentsStatus()
	if err != nil {
		klog.Errorln(err)
	}

	status.KubeSphereComponents = components

	nodeLister := informers.SharedInformerFactory().Core().V1().Nodes().Lister()
	// get node status
	nodes, err := nodeLister.List(labels.Everything())
	if err != nil {
		klog.Errorln(err)
		return status, nil
	}

	totalNodes := 0
	healthyNodes := 0
	for _, nodes := range nodes {
		totalNodes++
		for _, condition := range nodes.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				healthyNodes++
			}
		}
	}
	nodeStatus := models.NodeStatus{TotalNodes: totalNodes, HealthyNodes: healthyNodes}

	status.NodeStatus = nodeStatus

	return status, nil

}

func GetAllComponentsStatus() ([]models.ComponentStatus, error) {
	serviceLister := informers.SharedInformerFactory().Core().V1().Services().Lister()
	podLister := informers.SharedInformerFactory().Core().V1().Pods().Lister()

	components := make([]models.ComponentStatus, 0)

	var err error
	for _, ns := range constants.SystemNamespaces {

		services, err := serviceLister.Services(ns).List(labels.Everything())

		if err != nil {
			klog.Error(err)
			continue
		}

		for _, service := range services {

			// skip services without a selector
			if len(service.Spec.Selector) == 0 {
				continue
			}

			component := models.ComponentStatus{
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
				klog.Errorln(err)
				continue
			}

			for _, pod := range pods {
				component.TotalBackends++
				if pod.Status.Phase == corev1.PodRunning && isAllContainersReady(pod) {
					component.HealthyBackends++
				}
			}

			components = append(components, component)
		}
	}

	return components, err
}
