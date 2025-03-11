package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CategorySpec defines the desired state of HelmRepo
type CategorySpec struct {
	Icon string `json:"icon,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=appctg
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="total",type=string,JSONPath=`.status.total`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Category is the Schema for the categories API
type Category struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CategorySpec   `json:"spec,omitempty"`
	Status CategoryStatus `json:"status,omitempty"`
}

type CategoryStatus struct {
	Total int `json:"total"`
}

// +kubebuilder:object:root=true

// CategoryList contains a list of Category
type CategoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Category `json:"items"`
}
