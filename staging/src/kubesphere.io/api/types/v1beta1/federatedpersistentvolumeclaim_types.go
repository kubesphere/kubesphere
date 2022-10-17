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
	ResourcePluralFederatedPersistentVolumeClaim   = "federatedpersistentvolumeclaims"
	ResourceSingularFederatedPersistentVolumeClaim = "federatedpersistentvolumeclaim"
	FederatedPersistentVolumeClaimKind             = "FederatedPersistentVolumeClaim"
)

// +genclient
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
type FederatedPersistentVolumeClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedPersistentVolumeClaimSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedPersistentVolumeClaimSpec struct {
	Template  PersistentVolumeClaimTemplate `json:"template"`
	Placement GenericPlacementFields        `json:"placement"`
	Overrides []GenericOverrideItem         `json:"overrides,omitempty"`
}

type PersistentVolumeClaimTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              corev1.PersistentVolumeClaimSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// FederatedPersistentVolumeClaimList contains a list of federatedpersistentvolumeclaimlists
type FederatedPersistentVolumeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedPersistentVolumeClaim `json:"items"`
}
