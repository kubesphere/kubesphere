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

	apiv3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
)

const (
	KindIPAMConfig       = "IPAMConfig"
	KindIPAMConfigList   = "IPAMConfigList"
	GlobalIPAMConfigName = "default"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAMConfig contains information about a block for IP address assignment.
type IPAMConfig struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the IPAMConfig.
	Spec IPAMConfigSpec `json:"spec,omitempty"`
}

// IPAMConfigSpec contains the specification for an IPAMConfig resource.
type IPAMConfigSpec struct {
	StrictAffinity     bool `json:"strictAffinity"`
	AutoAllocateBlocks bool `json:"autoAllocateBlocks"`

	// MaxBlocksPerHost, if non-zero, is the max number of blocks that can be
	// affine to each host.
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:validation:Maximum:=2147483647
	// +optional
	MaxBlocksPerHost int `json:"maxBlocksPerHost,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAMConfigList contains a list of IPAMConfig resources.
type IPAMConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IPAMConfig `json:"items"`
}

// NewIPAMConfig creates a new (zeroed) IPAMConfig struct with the TypeMetadata initialised to the current
// version.
func NewIPAMConfig() *IPAMConfig {
	return &IPAMConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPAMConfig,
			APIVersion: apiv3.GroupVersionCurrent,
		},
	}
}

// NewIPAMConfigList creates a new (zeroed) IPAMConfigList struct with the TypeMetadata initialised to the current
// version.
func NewIPAMConfigList() *IPAMConfigList {
	return &IPAMConfigList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPAMConfigList,
			APIVersion: apiv3.GroupVersionCurrent,
		},
	}
}
