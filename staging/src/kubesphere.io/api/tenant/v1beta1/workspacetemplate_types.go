package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkspaceTemplate     = "WorkspaceTemplate"
	ResourceSingularWorkspaceTemplate = "workspacetemplate"
	ResourcePluralWorkspaceTemplate   = "workspacetemplates"
)

// +k8s:openapi-gen=true
type WorkspaceTemplateSpec struct {
	Template  Template         `json:"template"`
	Placement GenericPlacement `json:"placement"`
}

type ObjectMeta struct {
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// +k8s:openapi-gen=true
type Template struct {
	ObjectMeta `json:"metadata,omitempty"`
	Spec       WorkspaceSpec `json:"spec,omitempty"`
}

type GenericClusterReference struct {
	Name string `json:"name"`
}

// +k8s:openapi-gen=true
type GenericPlacement struct {
	// +listType=map
	// +listMapKey=name
	Clusters        []GenericClusterReference `json:"clusters,omitempty"`
	ClusterSelector *metav1.LabelSelector     `json:"clusterSelector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories="tenant",scope="Cluster"

// WorkspaceTemplate is the Schema for the workspacetemplates API
// +k8s:openapi-gen=true
type WorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceTemplateList contains a list of WorkspaceTemplate
// +k8s:openapi-gen=true
type WorkspaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplate `json:"items"`
}
