package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	workspacev1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
)

const (
	ResourcePluralFederatedWorkspace   = "federatedworkspaces"
	ResourceSingularFederatedWorkspace = "federatedworkspace"
	FederatedWorkspaceKind             = "FederatedWorkspace"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type FederatedWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedWorkspaceSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedWorkspaceSpec struct {
	Template  WorkspaceTemplate      `json:"template"`
	Placement GenericPlacementFields `json:"placement"`
	Overrides []GenericOverrideItem  `json:"overrides,omitempty"`
}

type WorkspaceTemplate struct {
	Spec workspacev1alpha1.WorkspaceSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedWorkspaceList contains a list of federatedworkspacelists
type FederatedWorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedWorkspace `json:"items"`
}
