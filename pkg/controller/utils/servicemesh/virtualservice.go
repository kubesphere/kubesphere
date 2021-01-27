/*
Copyright 2020 KubeSphere Authors

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

package servicemesh

import (
	apinetworkingv1alpha3 "istio.io/api/networking/v1alpha3"
	clientgonetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
)

func GenerateVirtualServiceSpec(strategy *servicemeshv1alpha2.Strategy, service *v1.Service) *clientgonetworkingv1alpha3.VirtualService {

	// Define VirtualService to be created
	vs := &clientgonetworkingv1alpha3.VirtualService{
		Spec: strategy.Spec.Template.Spec,
	}

	// one version rules them all
	if len(strategy.Spec.GovernorVersion) > 0 {
		governorDestinationWeight := apinetworkingv1alpha3.HTTPRouteDestination{
			Destination: &apinetworkingv1alpha3.Destination{
				Host:   service.Name,
				Subset: strategy.Spec.GovernorVersion,
			},
			Weight: 100,
		}

		if len(strategy.Spec.Template.Spec.Http) > 0 {
			governorRoute := apinetworkingv1alpha3.HTTPRoute{
				Route: []*apinetworkingv1alpha3.HTTPRouteDestination{&governorDestinationWeight},
			}

			vs.Spec.Http = []*apinetworkingv1alpha3.HTTPRoute{&governorRoute}
		} else if len(strategy.Spec.Template.Spec.Tcp) > 0 {
			tcpRoute := apinetworkingv1alpha3.TCPRoute{
				Route: []*apinetworkingv1alpha3.RouteDestination{
					{
						Destination: &apinetworkingv1alpha3.Destination{
							Host:   governorDestinationWeight.Destination.Host,
							Subset: governorDestinationWeight.Destination.Subset,
						},
						Weight: governorDestinationWeight.Weight,
					},
				},
			}

			//governorRoute := v1alpha3.TCPRoute{tcpRoute}
			vs.Spec.Tcp = []*apinetworkingv1alpha3.TCPRoute{&tcpRoute}
		}
	}

	FillDestinationPort(vs, service)
	return vs
}

// if destinationrule not specified with port number, then fill with service first port
func FillDestinationPort(vs *clientgonetworkingv1alpha3.VirtualService, service *v1.Service) {
	// fill http port
	for i := range vs.Spec.Http {
		for j := range vs.Spec.Http[i].Route {
			port := vs.Spec.Http[i].Route[j].Destination.Port
			if port == nil || port.Number == 0 {
				vs.Spec.Http[i].Route[j].Destination.Port = &apinetworkingv1alpha3.PortSelector{
					Number: uint32(service.Spec.Ports[0].Port),
				}
			}
		}

		if vs.Spec.Http[i].Mirror != nil && (vs.Spec.Http[i].Mirror.Port == nil || vs.Spec.Http[i].Mirror.Port.Number == 0) {
			vs.Spec.Http[i].Mirror.Port = &apinetworkingv1alpha3.PortSelector{
				Number: uint32(service.Spec.Ports[0].Port),
			}
		}
	}

	// fill tcp port
	for i := range vs.Spec.Tcp {
		for j := range vs.Spec.Tcp[i].Route {
			if vs.Spec.Tcp[i].Route[j].Destination.Port == nil || vs.Spec.Tcp[i].Route[j].Destination.Port.Number == 0 {
				vs.Spec.Tcp[i].Route[j].Destination.Port = &apinetworkingv1alpha3.PortSelector{
					Number: uint32(service.Spec.Ports[0].Port),
				}
			}
		}
	}
}

// Get subsets from strategy
func GetStrategySubsets(strategy *servicemeshv1alpha2.Strategy) sets.String {
	set := sets.String{}
	for _, httpRoute := range strategy.Spec.Template.Spec.Http {
		for _, dw := range httpRoute.Route {
			set.Insert(dw.Destination.Subset)
		}

		if httpRoute.Mirror != nil {
			set.Insert(httpRoute.Mirror.Subset)
		}
	}

	for _, tcpRoute := range strategy.Spec.Template.Spec.Tcp {
		for _, dw := range tcpRoute.Route {
			set.Insert(dw.Destination.Subset)
		}
	}

	for _, tlsRoute := range strategy.Spec.Template.Spec.Tls {
		for _, dw := range tlsRoute.Route {
			set.Insert(dw.Destination.Subset)
		}
	}
	return set
}
