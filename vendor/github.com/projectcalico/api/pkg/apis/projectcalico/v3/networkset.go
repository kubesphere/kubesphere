// Copyright (c) 2019,2021 Tigera, Inc. All rights reserved.

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
	KindNetworkSet     = "NetworkSet"
	KindNetworkSetList = "NetworkSetList"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkSetList is a list of NetworkSet objects.
type NetworkSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []NetworkSet `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NetworkSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec NetworkSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// NetworkSetSpec contains the specification for a NetworkSet resource.
type NetworkSetSpec struct {
	// The list of IP networks that belong to this set.
	Nets []string `json:"nets,omitempty" validate:"omitempty,dive,cidr"`
}

// NewNetworkSet creates a new (zeroed) NetworkSet struct with the TypeMetadata initialised to the current version.
func NewNetworkSet() *NetworkSet {
	return &NetworkSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindNetworkSet,
			APIVersion: GroupVersionCurrent,
		},
	}
}
