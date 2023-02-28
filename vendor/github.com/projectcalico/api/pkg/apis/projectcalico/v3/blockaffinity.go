// Copyright (c) 2022 Tigera, Inc. All rights reserved.

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
	KindBlockAffinity     = "BlockAffinity"
	KindBlockAffinityList = "BlockAffinityList"
)

type BlockAffinityState string

const (
	StateConfirmed       BlockAffinityState = "confirmed"
	StatePending         BlockAffinityState = "pending"
	StatePendingDeletion BlockAffinityState = "pendingDeletion"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BlockAffinity maintains a block affinity's state
type BlockAffinity struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the BlockAffinity.
	Spec BlockAffinitySpec `json:"spec,omitempty"`
}

// BlockAffinitySpec contains the specification for a BlockAffinity resource.
type BlockAffinitySpec struct {
	// The state of the block affinity with regard to any referenced IPAM blocks.
	State BlockAffinityState `json:"state"`

	// The node that this block affinity is assigned to.
	Node string `json:"node"`

	// The CIDR range this block affinity references.
	CIDR string `json:"cidr"`

	// Deleted indicates whether or not this block affinity is disabled and is
	// used as part of race-condition prevention. When set to true, clients
	// should treat this block as if it does not exist.
	Deleted bool `json:"deleted,omitempty"`
}

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BlockAffinityList contains a list of BlockAffinity resources.
type BlockAffinityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []BlockAffinity `json:"items"`
}

// NewBlockAffinity creates a new (zeroed) BlockAffinity struct with the TypeMetadata initialised to the current
// version.
func NewBlockAffinity() *BlockAffinity {
	return &BlockAffinity{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindBlockAffinity,
			APIVersion: GroupVersionCurrent,
		},
	}
}

// NewBlockAffinityList creates a new (zeroed) BlockAffinityList struct with the TypeMetadata initialised to the current
// version.
func NewBlockAffinityList() *BlockAffinityList {
	return &BlockAffinityList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindBlockAffinityList,
			APIVersion: GroupVersionCurrent,
		},
	}
}
