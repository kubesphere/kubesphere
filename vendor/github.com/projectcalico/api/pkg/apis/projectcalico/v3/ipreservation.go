// Copyright (c) 2017, 2021 Tigera, Inc. All rights reserved.

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
	KindIPReservation     = "IPReservation"
	KindIPReservationList = "IPReservationList"
)

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPReservationList contains a list of IPReservation resources.
type IPReservationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []IPReservation `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPReservation allows certain IP addresses to be reserved (i.e. prevented from being allocated) by Calico
// IPAM.  Reservations only block new allocations, they do not cause existing IP allocations to be released.
// The current implementation is only suitable for reserving small numbers of IP addresses relative to the
// size of the IP pool.  If large portions of an IP pool are reserved, Calico IPAM may hunt for a long time
// to find a non-reserved IP.
type IPReservation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec IPReservationSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// IPReservationSpec contains the specification for an IPReservation resource.
type IPReservationSpec struct {
	// ReservedCIDRs is a list of CIDRs and/or IP addresses that Calico IPAM will exclude from new allocations.
	ReservedCIDRs []string `json:"reservedCIDRs,omitempty" validate:"cidrs,omitempty"`
}

// NewIPReservation creates a new (zeroed) IPReservation struct with the TypeMetadata initialised to the current
// version.
func NewIPReservation() *IPReservation {
	return &IPReservation{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindIPReservation,
			APIVersion: GroupVersionCurrent,
		},
	}
}
