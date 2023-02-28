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
	KindIPAMHandle     = "IPAMHandle"
	KindIPAMHandleList = "IPAMHandleList"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAMHandle contains information about an IPAMHandle resource.
type IPAMHandle struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the IPAMHandle.
	Spec IPAMHandleSpec `json:"spec,omitempty"`
}

// IPAMHandleSpec contains the specification for an IPAMHandle resource.
type IPAMHandleSpec struct {
	HandleID string         `json:"handleID"`
	Block    map[string]int `json:"block"`

	// +optional
	Deleted bool `json:"deleted"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPAMHandleList contains a list of IPAMHandle resources.
type IPAMHandleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IPAMHandle `json:"items"`
}

// NewIPAMHandle creates a new (zeroed) IPAMHandle struct with the TypeMetadata initialised to the current
// version.
func NewIPAMHandle() *IPAMHandle {
	return &IPAMHandle{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPAMHandle,
			APIVersion: apiv3.GroupVersionCurrent,
		},
	}
}

// NewIPAMHandleList creates a new (zeroed) IPAMHandleList struct with the TypeMetadata initialised to the current
// version.
func NewIPAMHandleList() *IPAMHandleList {
	return &IPAMHandleList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPAMHandleList,
			APIVersion: apiv3.GroupVersionCurrent,
		},
	}
}
