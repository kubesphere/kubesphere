package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkspace     = "Workspace"
	ResourceSingularWorkspace = "workspace"
	ResourcePluralWorkspace   = "workspaces"
	WorkspaceLabel            = "kubesphere.io/workspace"
)

type WorkspaceSpec struct {
	Manager          string `json:"manager,omitempty"`
	NetworkIsolation *bool  `json:"networkIsolation,omitempty"`
}

type WorkspaceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion
// +kubebuilder:resource:categories="tenant",scope="Cluster"

type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceSpec   `json:"spec,omitempty"`
	Status            WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}
