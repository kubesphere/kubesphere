// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.

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

package v1

import (
	"fmt"

	"github.com/projectcalico/libcalico-go/lib/apis/v1/unversioned"
	"github.com/projectcalico/libcalico-go/lib/net"
)

// HostEndpoint contains information about a “bare-metal” interfaces attached to the host that is
// running Calico’s agent, Felix. By default, Calico doesn’t apply any policy to such interfaces.
type HostEndpoint struct {
	unversioned.TypeMetadata
	Metadata HostEndpointMetadata `json:"metadata,omitempty"`
	Spec     HostEndpointSpec     `json:"spec,omitempty"`
}

func (t HostEndpoint) GetResourceMetadata() unversioned.ResourceMetadata {
	return t.Metadata
}

// String() returns the human-readable string representation of a HostEndpoint instance
// which is defined by its Node and Name.
func (t HostEndpoint) String() string {
	return fmt.Sprintf("HostEndpoint(Node=%s, Name=%s)", t.Metadata.Node, t.Metadata.Name)
}

// HostEndpointMetadata contains the Metadata for a HostEndpoint resource.
type HostEndpointMetadata struct {
	unversioned.ObjectMetadata

	// The name of the endpoint.
	Name string `json:"name,omitempty" validate:"omitempty,namespacedName"`

	// The node name identifying the Calico node instance.
	Node string `json:"node,omitempty" validate:"omitempty,name"`

	// The labels applied to the host endpoint.  It is expected that many endpoints share
	// the same labels. For example, they could be used to label all “production” workloads
	// with “deployment=prod” so that security policy can be applied to production workloads.
	Labels map[string]string `json:"labels,omitempty" validate:"omitempty,labels"`
}

// HostEndpointSpec contains the specification for a HostEndpoint resource.
type HostEndpointSpec struct {
	// The name of the linux interface to apply policy to; for example “eth0”.
	// If "InterfaceName" is not present then at least one expected IP must be specified.
	InterfaceName string `json:"interfaceName,omitempty" validate:"omitempty,interface"`

	// The expected IP addresses (IPv4 and IPv6) of the endpoint.
	// If "InterfaceName" is not present, Calico will look for an interface matching any
	// of the IPs in the list and apply policy to that.
	//
	// Note:
	// 	When using the selector|tag match criteria in an ingress or egress security Policy
	// 	or Profile, Calico converts the selector into a set of IP addresses. For host
	// 	endpoints, the ExpectedIPs field is used for that purpose. (If only the interface
	// 	name is specified, Calico does not learn the IPs of the interface for use in match
	// 	criteria.)
	ExpectedIPs []net.IP `json:"expectedIPs,omitempty" validate:"omitempty"`

	// A list of identifiers of security Profile objects that apply to this endpoint. Each
	// profile is applied in the order that they appear in this list.  Profile rules are applied
	// after the selector-based security policy.
	Profiles []string `json:"profiles,omitempty" validate:"omitempty,dive,namespacedName"`

	// Ports contains the endpoint's named ports, which may be referenced in security policy rules.
	Ports []EndpointPort `json:"ports,omitempty" validate:"omitempty,dive"`
}

// NewHostEndpoint creates a new (zeroed) HostEndpoint struct with the TypeMetadata initialised to the current
// version.
func NewHostEndpoint() *HostEndpoint {
	return &HostEndpoint{
		TypeMetadata: unversioned.TypeMetadata{
			Kind:       "hostEndpoint",
			APIVersion: unversioned.VersionCurrent,
		},
	}
}

// HostEndpointList contains a list of Host Endpoint resources.  List types are returned from List()
// enumerations in the client interface.
type HostEndpointList struct {
	unversioned.TypeMetadata
	Metadata unversioned.ListMetadata `json:"metadata,omitempty"`
	Items    []HostEndpoint           `json:"items" validate:"dive"`
}

// NewHostEndpoint creates a new (zeroed) HostEndpointList struct with the TypeMetadata initialised to the current
// version.
func NewHostEndpointList() *HostEndpointList {
	return &HostEndpointList{
		TypeMetadata: unversioned.TypeMetadata{
			Kind:       "hostEndpointList",
			APIVersion: unversioned.VersionCurrent,
		},
	}
}
