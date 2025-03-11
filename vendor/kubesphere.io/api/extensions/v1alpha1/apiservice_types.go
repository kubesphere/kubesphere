package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIServiceSpec struct {
	Group    string `json:"group,omitempty"`
	Version  string `json:"version,omitempty"`
	Endpoint `json:",inline"`
}

type APIServiceStatus struct {
	State      string             `json:"state,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

// APIService is a special resource used in Ks-apiserver
// declares a directional proxy path for a resource type API，
// it's similar to Kubernetes API Aggregation Layer.
type APIService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIServiceSpec   `json:"spec,omitempty"`
	Status APIServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type APIServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIService `json:"items"`
}
