// Copyright (c) 2017 Tigera, Inc. All rights reserved.

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

	apiv1 "github.com/projectcalico/libcalico-go/lib/apis/v1"
	"github.com/projectcalico/libcalico-go/lib/selector"
)

const (
	KindIPPool     = "IPPool"
	KindIPPoolList = "IPPoolList"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPPool contains information about a IPPool resource.
type IPPool struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the IPPool.
	Spec IPPoolSpec `json:"spec,omitempty"`
}

// IPPoolSpec contains the specification for an IPPool resource.
type IPPoolSpec struct {
	// The pool CIDR.
	CIDR string `json:"cidr" validate:"net"`

	// Contains configuration for VXLAN tunneling for this pool. If not specified,
	// then this is defaulted to "Never" (i.e. VXLAN tunelling is disabled).
	VXLANMode VXLANMode `json:"vxlanMode,omitempty" validate:"omitempty,vxlanMode"`

	// Contains configuration for IPIP tunneling for this pool. If not specified,
	// then this is defaulted to "Never" (i.e. IPIP tunelling is disabled).
	IPIPMode IPIPMode `json:"ipipMode,omitempty" validate:"omitempty,ipIpMode"`

	// When nat-outgoing is true, packets sent from Calico networked containers in
	// this pool to destinations outside of this pool will be masqueraded.
	NATOutgoing bool `json:"natOutgoing,omitempty"`

	// When disabled is true, Calico IPAM will not assign addresses from this pool.
	Disabled bool `json:"disabled,omitempty"`

	// The block size to use for IP address assignments from this pool. Defaults to 26 for IPv4 and 112 for IPv6.
	BlockSize int `json:"blockSize,omitempty"`

	// Allows IPPool to allocate for a specific node by label selector.
	NodeSelector string `json:"nodeSelector,omitempty" validate:"omitempty,selector"`

	// Deprecated: this field is only used for APIv1 backwards compatibility.
	// Setting this field is not allowed, this field is for internal use only.
	IPIP *apiv1.IPIPConfiguration `json:"ipip,omitempty" validate:"omitempty,mustBeNil"`

	// Deprecated: this field is only used for APIv1 backwards compatibility.
	// Setting this field is not allowed, this field is for internal use only.
	NATOutgoingV1 bool `json:"nat-outgoing,omitempty" validate:"omitempty,mustBeFalse"`
}

// SelectsNode determines whether or not the IPPool's nodeSelector
// matches the labels on the given node.
func (pool IPPool) SelectsNode(n Node) (bool, error) {
	// No node selector means that the pool matches the node.
	if len(pool.Spec.NodeSelector) == 0 {
		return true, nil
	}
	// Check for valid selector syntax.
	sel, err := selector.Parse(pool.Spec.NodeSelector)
	if err != nil {
		return false, err
	}
	// Return whether or not the selector matches.
	return sel.Evaluate(n.Labels), nil
}

type VXLANMode string

const (
	VXLANModeNever       VXLANMode = "Never"
	VXLANModeAlways                = "Always"
	VXLANModeCrossSubnet           = "CrossSubnet"
)

type IPIPMode string

const (
	IPIPModeNever       IPIPMode = "Never"
	IPIPModeAlways               = "Always"
	IPIPModeCrossSubnet          = "CrossSubnet"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPPoolList contains a list of IPPool resources.
type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IPPool `json:"items"`
}

// NewIPPool creates a new (zeroed) IPPool struct with the TypeMetadata initialised to the current
// version.
func NewIPPool() *IPPool {
	return &IPPool{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPPool,
			APIVersion: GroupVersionCurrent,
		},
	}
}

// NewIPPoolList creates a new (zeroed) IPPoolList struct with the TypeMetadata initialised to the current
// version.
func NewIPPoolList() *IPPoolList {
	return &IPPoolList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPPoolList,
			APIVersion: GroupVersionCurrent,
		},
	}
}
