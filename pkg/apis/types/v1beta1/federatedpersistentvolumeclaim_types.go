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
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedPersistentVolumeClaimList contains a list of federatedpersistentvolumeclaimlists
type FederatedPersistentVolumeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedPersistentVolumeClaim `json:"items"`
}
