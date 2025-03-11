package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkspace     = "Workspace"
	ResourceSingularWorkspace = "workspace"
	ResourcePluralWorkspace   = "workspaces"
	WorkspaceLabel            = "kubesphere.io/workspace"
)

// WorkspaceSpec defines the desired state of Workspace
// +k8s:openapi-gen=true
type WorkspaceSpec struct {
	Manager string `json:"manager,omitempty"`
}

// WorkspaceStatus defines the observed state of Workspace
// +k8s:openapi-gen=true
type WorkspaceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories="tenant",scope="Cluster"

// Workspace is the Schema for the workspaces API
// +k8s:openapi-gen=true
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceSpec   `json:"spec,omitempty"`
	Status            WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace
// +k8s:openapi-gen=true
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}
