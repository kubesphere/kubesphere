// Copyright (c) 2017,2020 Tigera, Inc. All rights reserved.

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

	apiv3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"

	"github.com/projectcalico/api/pkg/lib/numorstring"
)

const (
	KindNode     = "Node"
	KindNodeList = "NodeList"
	CalicoNodeIP = "CalicoNodeIP"
	InternalIP   = "InternalIP"
	ExternalIP   = "ExternalIP"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Node contains information about a Node resource.
type Node struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the Node.
	Spec NodeSpec `json:"spec,omitempty"`
	// Status of the Node.
	Status NodeStatus `json:"status,omitempty"`
}

// NodeSpec contains the specification for a Node resource.
type NodeSpec struct {
	// BGP configuration for this node.
	BGP *NodeBGPSpec `json:"bgp,omitempty" validate:"omitempty"`

	// IPv4VXLANTunnelAddr is the IPv4 address of the VXLAN tunnel.
	IPv4VXLANTunnelAddr string `json:"ipv4VXLANTunnelAddr,omitempty" validate:"omitempty,ipv4"`

	// VXLANTunnelMACAddr is the MAC address of the VXLAN tunnel.
	VXLANTunnelMACAddr string `json:"vxlanTunnelMACAddr,omitempty" validate:"omitempty,mac"`

	// IPv6VXLANTunnelAddr is the address of the IPv6 VXLAN tunnel.
	IPv6VXLANTunnelAddr string `json:"ipv6VXLANTunnelAddr,omitempty" validate:"omitempty,ipv6"`

	// VXLANTunnelMACAddrV6 is the MAC address of the IPv6 VXLAN tunnel.
	VXLANTunnelMACAddrV6 string `json:"vxlanTunnelMACAddrV6,omitempty" validate:"omitempty,mac"`

	// OrchRefs for this node.
	OrchRefs []OrchRef `json:"orchRefs,omitempty" validate:"omitempty"`

	// Wireguard configuration for this node.
	Wireguard *NodeWireguardSpec `json:"wireguard,omitempty" validate:"omitempty"`

	// Addresses list address that a client can reach the node at.
	Addresses []NodeAddress `json:"addresses,omitempty" validate:"omitempty"`
}

// NodeAddress represents an address assigned to a node.
type NodeAddress struct {
	// Address is a string representation of the actual address.
	Address string `json:"address" validate:"net"`

	// Type is the node IP type
	Type string `json:"type,omitempty" validate:"omitempty,ipType"`
}

type NodeStatus struct {
	// WireguardPublicKey is the IPv4 Wireguard public-key for this node.
	// wireguardPublicKey validates if the string is a valid base64 encoded key.
	WireguardPublicKey string `json:"wireguardPublicKey,omitempty" validate:"omitempty,wireguardPublicKey"`

	// WireguardPublicKeyV6 is the IPv6 Wireguard public-key for this node.
	// wireguardPublicKey validates if the string is a valid base64 encoded key.
	WireguardPublicKeyV6 string `json:"wireguardPublicKeyV6,omitempty" validate:"omitempty,wireguardPublicKey"`

	// PodCIDR is a reflection of the Kubernetes node's spec.PodCIDRs field.
	PodCIDRs []string `json:"podCIDRs,omitempty" validate:"omitempty"`
}

// OrchRef is used to correlate a Calico node to its corresponding representation in a given orchestrator
type OrchRef struct {
	// NodeName represents the name for this node according to the orchestrator.
	NodeName string `json:"nodeName,omitempty" validate:"omitempty"`
	// Orchestrator represents the orchestrator using this node.
	Orchestrator string `json:"orchestrator"`
}

// NodeBGPSpec contains the specification for the Node BGP configuration.
type NodeBGPSpec struct {
	// The AS Number of the node.  If this is not specified, the global
	// default value will be used.
	ASNumber *numorstring.ASNumber `json:"asNumber,omitempty"`
	// IPv4Address is the IPv4 address and network of this node.  The IPv4 address
	// should always be specified if you are using BGP.
	IPv4Address string `json:"ipv4Address,omitempty" validate:"omitempty,cidrv4"`
	// IPv6Address is the IPv6 address and network of this node.  Not required if you
	// are not using BGP or you do not require IPv6 routing.
	IPv6Address string `json:"ipv6Address,omitempty" validate:"omitempty,cidrv6"`
	// IPv4IPIPTunnelAddr is the IPv4 address of the IP in IP tunnel.
	IPv4IPIPTunnelAddr string `json:"ipv4IPIPTunnelAddr,omitempty" validate:"omitempty,ipv4"`
	// RouteReflectorClusterID enables this node as a route reflector within the given
	// cluster.
	RouteReflectorClusterID string `json:"routeReflectorClusterID,omitempty" validate:"omitempty,ipv4"`
}

// NodeWireguardSpec contains the specification for the Node wireguard configuration.
type NodeWireguardSpec struct {
	// InterfaceIPv4Address is the IP address for the IPv4 Wireguard interface.
	InterfaceIPv4Address string `json:"interfaceIPv4Address,omitempty" validate:"omitempty,ipv4"`

	// InterfaceIPv6Address is the IP address for the IPv6 Wireguard interface.
	InterfaceIPv6Address string `json:"interfaceIPv6Address,omitempty" validate:"omitempty,ipv6"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeList contains a list of Node resources.
type NodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Node `json:"items"`
}

// NewNode creates a new (zeroed) Node struct with the TypeMetadata initialised to the current
// version.
func NewNode() *Node {
	return &Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindNode,
			APIVersion: apiv3.GroupVersionCurrent,
		},
	}
}

// NewNodeList creates a new (zeroed) NodeList struct with the TypeMetadata initialised to the current
// version.
func NewNodeList() *NodeList {
	return &NodeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindNodeList,
			APIVersion: apiv3.GroupVersionCurrent,
		},
	}
}
