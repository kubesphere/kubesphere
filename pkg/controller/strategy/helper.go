package strategy

import (
	"fmt"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"k8s.io/api/core/v1"
	"kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
)

const (
	AppLabel = "app"
)

func getAppNameByStrategy(strategy *v1alpha2.Strategy) string {
	if len(strategy.Labels) > 0 && len(strategy.Labels[AppLabel]) > 0 {
		return strategy.Labels[AppLabel]
	}
	return ""
}

func fillDestinationPort(vs *v1alpha3.VirtualService, service *v1.Service) error {

	if len(service.Spec.Ports) == 0 {
		return fmt.Errorf("service %s/%s spec doesn't canotain any ports", service.Namespace, service.Name)
	}

	// fill http port
	for i := range vs.Spec.Http {
		for j := range vs.Spec.Http[i].Route {
			vs.Spec.Http[i].Route[j].Destination.Port.Number = uint32(service.Spec.Ports[0].Port)
		}
	}

	// fill tcp port
	for i := range vs.Spec.Tcp {
		for j := range vs.Spec.Tcp[i].Route {
			vs.Spec.Tcp[i].Route[j].Destination.Port.Number = uint32(service.Spec.Ports[0].Port)
		}
	}

	return nil
}
