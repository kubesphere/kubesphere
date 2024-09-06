/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package components

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api/resource/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
)

type Getter interface {
	GetComponentStatus(name string) (v1alpha2.ComponentStatus, error)
	GetSystemHealthStatus() (v1alpha2.HealthStatus, error)
	GetAllComponentsStatus() ([]v1alpha2.ComponentStatus, error)
}

type componentsGetter struct {
	cache runtimeclient.Reader
}

func NewComponentsGetter(cache runtimeclient.Reader) Getter {
	return &componentsGetter{cache: cache}
}

func (c *componentsGetter) GetComponentStatus(name string) (v1alpha2.ComponentStatus, error) {

	service := &corev1.Service{}
	var err error
	for _, ns := range constants.SystemNamespaces {
		if err := c.cache.Get(context.Background(), types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, service); err == nil {
			break
		}
	}
	if err != nil {
		return v1alpha2.ComponentStatus{}, err
	}

	if len(service.Spec.Selector) == 0 {
		return v1alpha2.ComponentStatus{}, fmt.Errorf("component %s has no selector", name)
	}

	pods := &corev1.PodList{}
	if err := c.cache.List(context.Background(), pods, &runtimeclient.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(service.Spec.Selector),
		Namespace:     service.Namespace,
	}); err != nil {
		return v1alpha2.ComponentStatus{}, err
	}

	component := v1alpha2.ComponentStatus{
		Name:            service.Name,
		Namespace:       service.Namespace,
		Label:           service.Spec.Selector,
		StartedAt:       service.CreationTimestamp.Time,
		HealthyBackends: 0,
		TotalBackends:   0,
	}
	for _, pod := range pods.Items {
		component.TotalBackends++
		if pod.Status.Phase == corev1.PodRunning && isAllContainersReady(&pod) {
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

	nodes := &corev1.NodeList{}
	// get node status
	if err := c.cache.List(context.Background(), nodes); err != nil {
		klog.Errorln(err)
		return status, nil
	}

	totalNodes := 0
	healthyNodes := 0
	for _, nodes := range nodes.Items {
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

		services := &corev1.ServiceList{}
		if err := c.cache.List(context.Background(), services, &runtimeclient.ListOptions{
			Namespace: ns,
		}); err != nil {
			klog.Error(err)
			continue
		}

		for _, service := range services.Items {
			// skip services without a selector
			if len(service.Spec.Selector) == 0 {
				continue
			}

			component := v1alpha2.ComponentStatus{
				Name:            service.Name,
				Namespace:       service.Namespace,
				Label:           service.Spec.Selector,
				StartedAt:       service.CreationTimestamp.Time,
				HealthyBackends: 0,
				TotalBackends:   0,
			}

			pods := &corev1.PodList{}
			if err := c.cache.List(context.Background(), pods, &runtimeclient.ListOptions{
				LabelSelector: labels.SelectorFromValidatedSet(service.Spec.Selector),
				Namespace:     ns,
			}); err != nil {
				klog.Error(err)
				continue
			}

			if err != nil {
				klog.Errorln(err)
				continue
			}

			for _, pod := range pods.Items {
				component.TotalBackends++
				if pod.Status.Phase == corev1.PodRunning && isAllContainersReady(&pod) {
					component.HealthyBackends++
				}
			}

			components = append(components, component)
		}
	}

	return components, err
}
