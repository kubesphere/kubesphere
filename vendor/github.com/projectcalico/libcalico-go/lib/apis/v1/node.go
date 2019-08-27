// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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
	"github.com/projectcalico/libcalico-go/lib/numorstring"
)

// Node contains the details of a node resource which contains the configuration
// for a Calico node instance running on a compute host.
//
// In addition to creating a Node resource through calicoctl or the Calico API,
// the Calico node instance must also be running on the specific host and should be
// provided the same Name as that configured on the Node resource.  Note that, by
// default, the Calico node instance uses the hostname of the compute host when it
// is not explicitly specified - in this case, the equivalent Node resource should
// be created using the same hostname as the Name of the Node resource.
//
// Operations on the Node resources is expected to be required when adding a new
// host into a Calico network, and when removing a host from a Calico network, and
// occasionally to modify certain configuration.  Care should be taken when operating
// on Node resources: deleting a Node resource will remove all Node specific data.
type Node struct {
	unversioned.TypeMetadata
	Metadata NodeMetadata `json:"metadata,omitempty"`
	Spec     NodeSpec     `json:"spec,omitempty"`
}

func (t Node) GetResourceMetadata() unversioned.ResourceMetadata {
	return t.Metadata
}

// String() returns the human-readable string representation of a Node instance
// which is defined by its Name.
func (t Node) String() string {
	return fmt.Sprintf("Node(Name=%s)", t.Metadata.Name)
}

// NodeMetadata contains the metadata for a Calico Node resource.
type NodeMetadata struct {
	unversioned.ObjectMetadata

	// The name of the node.
	Name string `json:"name,omitempty" validate:"omitempty,name"`
}

// NodeSpec contains the specification for a Calico Node resource.
type NodeSpec struct {
	// BGP configuration for this node.  If this omitted, the Calico node
	// will be run in policy-only mode.
	BGP *NodeBGPSpec `json:"bgp,omitempty" validate:"omitempty"`

	// OrchRefs for this node.
	OrchRefs []OrchRef `json:"orchRefs,omitempty" validate:"omitempty"`
}

// OrchRef is used to correlate a Calico node to its corresponding representation in a given orchestrator
type OrchRef struct {
	// NodeName represents the name for this node according to the orchestrator.
	NodeName string `json:"nodeName,omitempty" validate:"omitempty"`
	// Orchestrator represents the orchestrator using this node.
	Orchestrator string `json:"orchestrator"`
}

// NodeSpec contains the specification for a Calico Node resource.
type NodeBGPSpec struct {
	// The AS Number of the node.  If this is not specified, the global
	// default value will be used.
	ASNumber *numorstring.ASNumber `json:"asNumber,omitempty"`

	// IPv4Address is the IPv4 address and network of this node.  At least
	// one of the IPv4 and IPv6 addresses should be specified.
	IPv4Address *net.IPNet `json:"ipv4Address,omitempty" validate:"omitempty"`

	// IPv6Address is the IPv6 address and network of this node.  At least
	// one of the IPv4 and IPv6 addresses should be specified.
	IPv6Address *net.IPNet `json:"ipv6Address,omitempty" validate:"omitempty"`
}

// NewNode creates a new (zeroed) NodeList struct with the TypeMetadata initialised to the current
// version.
func NewNode() *Node {
	return &Node{
		TypeMetadata: unversioned.TypeMetadata{
			Kind:       "node",
			APIVersion: unversioned.VersionCurrent,
		},
	}
}

// A NodeList contains a list of Node resources.  List types are returned from List()
// enumerations on the client interface.
type NodeList struct {
	unversioned.TypeMetadata
	Metadata unversioned.ListMetadata `json:"metadata,omitempty"`
	Items    []Node                   `json:"items" validate:"dive,omitempty"`
}

// NewNodeList creates a new (zeroed) NodeList struct with the TypeMetadata initialised to the current
// version.
func NewNodeList() *NodeList {
	return &NodeList{
		TypeMetadata: unversioned.TypeMetadata{
			Kind:       "nodeList",
			APIVersion: unversioned.VersionCurrent,
		},
	}
}
