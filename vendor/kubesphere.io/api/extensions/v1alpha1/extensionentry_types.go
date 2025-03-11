package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ExtensionEntrySpec struct {
	Entries []runtime.RawExtension `json:"entries,omitempty"`
}

type ExtensionEntryStatus struct {
	State string `json:"state,omitempty"`
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

// ExtensionEntry declares an entry endpoint that needs to be injected into ks-console.
type ExtensionEntry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExtensionEntrySpec   `json:"spec,omitempty"`
	Status ExtensionEntryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ExtensionEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExtensionEntry `json:"items"`
}
