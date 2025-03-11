package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Watching scope, when both are empty, watching all namespaces,
// when neither is empty, namespaces is preferred.
type Scope struct {
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`
	// +optional
	NamespaceSelector string `json:"namespaceSelector,omitempty"`
}
type IngressClass struct {
	Name    string `json:"name,omitempty"`
	Default bool   `json:"default,omitempty"`
}

type IngressClassScopeSpec struct {
	Scope        Scope        `json:"scope,omitempty"`
	IngressClass IngressClass `json:"ingressClass,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

// IngressClassScope is a special resource used to
// connect other gateways to the KubeSphere platform.
type IngressClassScope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IngressClassScopeSpec `json:"spec,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	Status runtime.RawExtension `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type IngressClassScopeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IngressClassScope `json:"items"`
}
