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
	"github.com/projectcalico/libcalico-go/lib/backend/encap"
	"github.com/projectcalico/libcalico-go/lib/net"
)

// IPPool contains the details of a Calico IP pool resource.
// A pool resource is used by Calico in two ways:
// 	- to provide a set of IP addresses from which Calico IPAM assigns addresses
// 	  for workloads.
// 	- to provide configuration specific to IP address range, such as configuration
// 	  for the BGP daemon (e.g. when to use a GRE tunnel to encapsulate packets
// 	  between compute hosts).
type IPPool struct {
	unversioned.TypeMetadata
	Metadata IPPoolMetadata `json:"metadata,omitempty"`
	Spec     IPPoolSpec     `json:"spec,omitempty"`
}

func (t IPPool) GetResourceMetadata() unversioned.ResourceMetadata {
	return t.Metadata
}

// String() returns the human-readable string representation of an IPPool instance
// which is defined by its CIDR.
func (t IPPool) String() string {
	return fmt.Sprintf("IPPool(CIDR=%s)", t.Metadata.CIDR.String())
}

// IPPoolMetadata contains the metadata for an IP pool resource.
type IPPoolMetadata struct {
	unversioned.ObjectMetadata
	CIDR net.IPNet `json:"cidr"`
}

// IPPoolSpec contains the specification for an IP pool resource.
type IPPoolSpec struct {
	// Contains configuration for ipip tunneling for this pool. If not specified,
	// then ipip tunneling is disabled for this pool.
	IPIP *IPIPConfiguration `json:"ipip,omitempty"`

	// When nat-outgoing is true, packets sent from Calico networked containers in
	// this pool to destinations outside of this pool will be masqueraded.
	NATOutgoing bool `json:"nat-outgoing,omitempty"`

	// When disabled is true, Calico IPAM will not assign addresses from this pool.
	Disabled bool `json:"disabled,omitempty"`
}

type IPIPConfiguration struct {
	// When enabled is true, ipip tunneling will be used to deliver packets to
	// destinations within this pool.
	Enabled bool `json:"enabled,omitempty"`

	// The IPIP mode.  This can be one of "always" or "cross-subnet".  A mode
	// of "always" will also use IPIP tunneling for routing to destination IP
	// addresses within this pool.  A mode of "cross-subnet" will only use IPIP
	// tunneling when the destination node is on a different subnet to the
	// originating node.  The default value (if not specified) is "always".
	Mode encap.Mode `json:"mode,omitempty" validate:"ipIpMode"`
}

// NewIPPool creates a new (zeroed) Pool struct with the TypeMetadata initialised to the current
// version.
func NewIPPool() *IPPool {
	return &IPPool{
		TypeMetadata: unversioned.TypeMetadata{
			Kind:       "ipPool",
			APIVersion: unversioned.VersionCurrent,
		},
	}
}

// IPPoolList contains a list of IP pool resources.  List types are returned from List()
// enumerations in the client interface.
type IPPoolList struct {
	unversioned.TypeMetadata
	Metadata unversioned.ListMetadata `json:"metadata,omitempty"`
	Items    []IPPool                 `json:"items" validate:"dive"`
}

// NewIPPool creates a new (zeroed) PoolList struct with the TypeMetadata initialised to the current
// version.
func NewIPPoolList() *IPPoolList {
	return &IPPoolList{
		TypeMetadata: unversioned.TypeMetadata{
			Kind:       "ipPoolList",
			APIVersion: unversioned.VersionCurrent,
		},
	}
}
