package kubernetes

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// FilterPodsForService returns a subpart of pod list filtered according service selector
func FilterPodsForService(s *v1.Service, allPods []v1.Pod) []v1.Pod {
	if s == nil || allPods == nil {
		return nil
	}
	serviceSelector := labels.Set(s.Spec.Selector).AsSelector()
	pods := FilterPodsForSelector(serviceSelector, allPods)

	return pods
}

func FilterPodsForSelector(selector labels.Selector, allPods []v1.Pod) []v1.Pod {
	var pods []v1.Pod
	for _, pod := range allPods {
		if selector.Matches(labels.Set(pod.ObjectMeta.Labels)) {
			pods = append(pods, pod)
		}
	}
	return pods
}

// FilterPodsForEndpoints performs a second pass was selector may return too many data
// This case happens when a "nil" selector (such as one of default/kubernetes service) is used
func FilterPodsForEndpoints(endpoints *v1.Endpoints, unfiltered []v1.Pod) []v1.Pod {
	endpointPods := make(map[string]bool)
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			if address.TargetRef != nil && address.TargetRef.Kind == "Pod" {
				endpointPods[address.TargetRef.Name] = true
			}
		}
	}
	var pods []v1.Pod
	for _, pod := range unfiltered {
		if _, ok := endpointPods[pod.Name]; ok {
			pods = append(pods, pod)
		}
	}
	return pods
}

func FilterPodsForController(controllerName string, controllerType string, allPods []v1.Pod) []v1.Pod {
	var pods []v1.Pod
	for _, pod := range allPods {
		for _, ref := range pod.OwnerReferences {
			if ref.Controller != nil && *ref.Controller && ref.Name == controllerName && ref.Kind == controllerType {
				pods = append(pods, pod)
				break
			}
		}
	}
	return pods
}

func FilterServicesForSelector(selector labels.Selector, allServices []v1.Service) []v1.Service {
	var services []v1.Service
	for _, svc := range allServices {
		if selector.Matches(labels.Set(svc.Spec.Selector)) {
			services = append(services, svc)
		}
	}
	return services
}
