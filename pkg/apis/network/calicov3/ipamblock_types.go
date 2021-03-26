// Copyright (c) 2019-2020 Tigera, Inc. All rights reserved.

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

package calicov3

import (
	"strings"

	v3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Cluster
type IPAMBlock struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              v3.IPAMBlockSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAMBlockList contains a list of IPAMBlock resources.
type IPAMBlockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IPAMBlock `json:"items"`
}

func (b *IPAMBlock) NumReservedAddresses() int {
	sum := 0
	for _, attrIdx := range b.Spec.Allocations {
		if attrIdx == nil {
			continue
		}
		attrs := b.Spec.Attributes[*attrIdx]
		if attrs.AttrPrimary == nil || strings.ToLower(*attrs.AttrPrimary) == WindowsReservedHandle {
			sum += 1
		}
	}
	return sum
}

// Get number of addresses covered by the block
func (b *IPAMBlock) NumAddresses() int {
	_, cidr, _ := cnet.ParseCIDR(b.Spec.CIDR)
	ones, size := cidr.Mask.Size()
	numAddresses := 1 << uint(size-ones)
	return numAddresses
}

func (b *IPAMBlock) NumFreeAddresses() int {
	return len(b.Spec.Unallocated)
}

// windwowsReservedHandle is the handle used to reserve addresses required for Windows
// networking so that workloads do not get assigned these addresses.
const WindowsReservedHandle = "windows-reserved-ipam-handle"

func (b *IPAMBlock) Empty() bool {
	for _, attrIdx := range b.Spec.Allocations {
		if attrIdx == nil {
			continue
		}
		attrs := b.Spec.Attributes[*attrIdx]
		if attrs.AttrPrimary == nil || strings.ToLower(*attrs.AttrPrimary) != WindowsReservedHandle {
			return false
		}
	}
	return true
}
