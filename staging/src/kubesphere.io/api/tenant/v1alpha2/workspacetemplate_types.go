package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1beta1"
)

const (
	ResourceKindWorkspaceTemplate     = "WorkspaceTemplate"
	ResourceSingularWorkspaceTemplate = "workspacetemplate"
	ResourcePluralWorkspaceTemplate   = "workspacetemplates"
)

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion
// +kubebuilder:resource:categories="tenant",scope="Cluster"

type WorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedWorkspaceSpec `json:"spec,omitempty"`
}

type FederatedWorkspaceSpec struct {
	Template  WorkspaceTemplateSpec  `json:"template"`
	Placement GenericPlacementFields `json:"placement"`
	Overrides []GenericOverrideItem  `json:"overrides,omitempty"`
}

type WorkspaceTemplateSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              tenantv1alpha1.WorkspaceSpec `json:"spec,omitempty"`
}

type GenericClusterReference struct {
	Name string `json:"name"`
}

type GenericPlacementFields struct {
	// +listType=map
	// +listMapKey=name
	Clusters        []GenericClusterReference `json:"clusters,omitempty"`
	ClusterSelector *metav1.LabelSelector     `json:"clusterSelector,omitempty"`
}

type ClusterOverride struct {
	Op   string `json:"op,omitempty"`
	Path string `json:"path"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Value runtime.RawExtension `json:"value,omitempty"`
}

type GenericOverrideItem struct {
	ClusterName      string            `json:"clusterName"`
	ClusterOverrides []ClusterOverride `json:"clusterOverrides,omitempty"`
}

// +kubebuilder:object:root=true

type WorkspaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplate `json:"items"`
}
