// Copyright (c) 2019 Tigera, Inc. All rights reserved.

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
)

const (
	KindIPAMBlock     = "IPAMBlock"
	KindIPAMBlockList = "IPAMBlockList"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAMBlock contains information about a block for IP address assignment.
type IPAMBlock struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the IPAMBlock.
	Spec IPAMBlockSpec `json:"spec,omitempty"`
}

// IPAMBlockSpec contains the specification for a IPAMBlock resource.
type IPAMBlockSpec struct {
	CIDR           string                `json:"cidr"`
	Affinity       *string               `json:"affinity"`
	StrictAffinity bool                  `json:"strictAffinity"`
	Allocations    []*int                `json:"allocations"`
	Unallocated    []int                 `json:"unallocated"`
	Attributes     []AllocationAttribute `json:"attributes"`
	Deleted        bool                  `json:"deleted`
}

type AllocationAttribute struct {
	AttrPrimary   *string           `json:"handle_id"`
	AttrSecondary map[string]string `json:"secondary"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAMBlockList contains a list of IPAMBlock resources.
type IPAMBlockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IPAMBlock `json:"items"`
}

// NewIPAMBlock creates a new (zeroed) IPAMBlock struct with the TypeMetadata initialised to the current
// version.
func NewIPAMBlock() *IPAMBlock {
	return &IPAMBlock{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPAMBlock,
			APIVersion: GroupVersionCurrent,
		},
	}
}

// NewIPAMBlockList creates a new (zeroed) IPAMBlockList struct with the TypeMetadata initialised to the current
// version.
func NewIPAMBlockList() *IPAMBlockList {
	return &IPAMBlockList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPAMBlockList,
			APIVersion: GroupVersionCurrent,
		},
	}
}
