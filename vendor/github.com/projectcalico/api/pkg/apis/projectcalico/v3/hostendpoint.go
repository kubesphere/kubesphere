// Copyright (c) 2017,2020-2021 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectcalico/api/pkg/lib/numorstring"
)

const (
	KindHostEndpoint     = "HostEndpoint"
	KindHostEndpointList = "HostEndpointList"
)

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HostEndpointList is a list of HostEndpoint objects.
type HostEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []HostEndpoint `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HostEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec HostEndpointSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// HostEndpointSpec contains the specification for a HostEndpoint resource.
type HostEndpointSpec struct {
	// The node name identifying the Calico node instance.
	Node string `json:"node,omitempty" validate:"omitempty,name"`
	// Either "*", or the name of a specific Linux interface to apply policy to; or empty.  "*"
	// indicates that this HostEndpoint governs all traffic to, from or through the default
	// network namespace of the host named by the "Node" field; entering and leaving that
	// namespace via any interface, including those from/to non-host-networked local workloads.
	//
	// If InterfaceName is not "*", this HostEndpoint only governs traffic that enters or leaves
	// the host through the specific interface named by InterfaceName, or - when InterfaceName
	// is empty - through the specific interface that has one of the IPs in ExpectedIPs.
	// Therefore, when InterfaceName is empty, at least one expected IP must be specified.  Only
	// external interfaces (such as "eth0") are supported here; it isn't possible for a
	// HostEndpoint to protect traffic through a specific local workload interface.
	//
	// Note: Only some kinds of policy are implemented for "*" HostEndpoints; initially just
	// pre-DNAT policy.  Please check Calico documentation for the latest position.
	InterfaceName string `json:"interfaceName,omitempty" validate:"omitempty,interface"`
	// The expected IP addresses (IPv4 and IPv6) of the endpoint.
	// If "InterfaceName" is not present, Calico will look for an interface matching any
	// of the IPs in the list and apply policy to that.
	// Note:
	// 	When using the selector match criteria in an ingress or egress security Policy
	// 	or Profile, Calico converts the selector into a set of IP addresses. For host
	// 	endpoints, the ExpectedIPs field is used for that purpose. (If only the interface
	// 	name is specified, Calico does not learn the IPs of the interface for use in match
	// 	criteria.)
	ExpectedIPs []string `json:"expectedIPs,omitempty" validate:"omitempty,dive,ip"`
	// A list of identifiers of security Profile objects that apply to this endpoint. Each
	// profile is applied in the order that they appear in this list.  Profile rules are applied
	// after the selector-based security policy.
	Profiles []string `json:"profiles,omitempty" validate:"omitempty,dive,name"`
	// Ports contains the endpoint's named ports, which may be referenced in security policy rules.
	Ports []EndpointPort `json:"ports,omitempty" validate:"dive"`
}

type EndpointPort struct {
	Name     string               `json:"name" validate:"portName"`
	Protocol numorstring.Protocol `json:"protocol"`
	Port     uint16               `json:"port" validate:"gt=0"`
}

// NewHostEndpoint creates a new (zeroed) HostEndpoint struct with the TypeMetadata initialised to the current
// version.
func NewHostEndpoint() *HostEndpoint {
	return &HostEndpoint{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindHostEndpoint,
			APIVersion: GroupVersionCurrent,
		},
	}
}
