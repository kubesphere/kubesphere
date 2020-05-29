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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api/resource/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
)

type ComponentsGetter interface {
	GetComponentStatus(name string) (v1alpha2.ComponentStatus, error)
	GetSystemHealthStatus() (v1alpha2.HealthStatus, error)
	GetAllComponentsStatus() ([]v1alpha2.ComponentStatus, error)
}

type componentsGetter struct {
	informers informers.SharedInformerFactory
}

func NewComponentsGetter(informers informers.SharedInformerFactory) ComponentsGetter {
	return &componentsGetter{informers: informers}
}

func (c *componentsGetter) GetComponentStatus(name string) (v1alpha2.ComponentStatus, error) {

	var service *corev1.Service
	var err error

	for _, ns := range constants.SystemNamespaces {
		service, err = c.informers.Core().V1().Services().Lister().Services(ns).Get(name)
		if err == nil {
			break
		}
	}

	if err != nil {
		return v1alpha2.ComponentStatus{}, err
	}

	if len(service.Spec.Selector) == 0 {
		return v1alpha2.ComponentStatus{}, fmt.Errorf("component %s has no selector", name)
	}

	pods, err := c.informers.Core().V1().Pods().Lister().Pods(service.Namespace).List(labels.SelectorFromValidatedSet(service.Spec.Selector))

	if err != nil {
		return v1alpha2.ComponentStatus{}, err
	}

	component := v1alpha2.ComponentStatus{
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

func (c *componentsGetter) GetSystemHealthStatus() (v1alpha2.HealthStatus, error) {

	status := v1alpha2.HealthStatus{}

	// get kubesphere-system components
	components, err := c.GetAllComponentsStatus()
	if err != nil {
		klog.Errorln(err)
	}

	status.KubeSphereComponents = components

	// get node status
	nodes, err := c.informers.Core().V1().Nodes().Lister().List(labels.Everything())
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
	nodeStatus := v1alpha2.NodeStatus{TotalNodes: totalNodes, HealthyNodes: healthyNodes}

	status.NodeStatus = nodeStatus

	return status, nil

}

func (c *componentsGetter) GetAllComponentsStatus() ([]v1alpha2.ComponentStatus, error) {

	components := make([]v1alpha2.ComponentStatus, 0)

	var err error
	for _, ns := range constants.SystemNamespaces {

		services, err := c.informers.Core().V1().Services().Lister().Services(ns).List(labels.Everything())

		if err != nil {
			klog.Error(err)
			continue
		}

		for _, service := range services {

			// skip services without a selector
			if len(service.Spec.Selector) == 0 {
				continue
			}

			component := v1alpha2.ComponentStatus{
				Name:            service.Name,
				Namespace:       service.Namespace,
				SelfLink:        service.SelfLink,
				Label:           service.Spec.Selector,
				StartedAt:       service.CreationTimestamp.Time,
				HealthyBackends: 0,
				TotalBackends:   0,
			}

			pods, err := c.informers.Core().V1().Pods().Lister().Pods(ns).List(labels.SelectorFromValidatedSet(service.Spec.Selector))

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
