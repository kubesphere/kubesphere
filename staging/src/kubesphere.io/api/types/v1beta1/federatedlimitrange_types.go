/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourcePluralFederatedLimitRange   = "federatedlimitranges"
	ResourceSingularFederatedLimitRange = "federatedlimitrange"
	FederatedLimitRangeKind             = "FederatedLimitRange"
)

// +genclient
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
type FederatedLimitRange struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedLimitRangeSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedLimitRangeSpec struct {
	Template  LimitRangeTemplate     `json:"template"`
	Placement GenericPlacementFields `json:"placement"`
	Overrides []GenericOverrideItem  `json:"overrides,omitempty"`
}

type LimitRangeTemplate struct {
	Spec corev1.LimitRangeSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// FederatedLimitRangeList contains a list of federatedlimitrangelists
type FederatedLimitRangeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedLimitRange `json:"items"`
}
