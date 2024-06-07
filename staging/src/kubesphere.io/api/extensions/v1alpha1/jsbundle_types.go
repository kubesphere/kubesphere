package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JSBundleSpec struct {
	// +optional
	Raw []byte `json:"raw,omitempty"`
	// +optional
	RawFrom RawFrom `json:"rawFrom,omitempty"`
}

type RawFrom struct {
	Endpoint `json:",inline"`
	// Selects a key of a ConfigMap.
	ConfigMapKeyRef *ConfigMapKeyRef `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	SecretKeyRef *SecretKeyRef `json:"secretKeyRef,omitempty"`
}

type ConfigMapKeyRef struct {
	corev1.ConfigMapKeySelector `json:",inline"`
	Namespace                   string `json:"namespace"`
}

type SecretKeyRef struct {
	corev1.SecretKeySelector `json:",inline"`
	Namespace                string `json:"namespace"`
}

type JSBundleStatus struct {
	// Link is the path for downloading JS file, default to "/dist/{jsBundleName}/index.js".
	// +optional
	Link  string `json:"link,omitempty"`
	State string `json:"state,omitempty"`
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

// JSBundle declares a js bundle that needs to be injected into ks-console,
// the endpoint can be provided by a service or a static file.
type JSBundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JSBundleSpec   `json:"spec,omitempty"`
	Status JSBundleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type JSBundleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JSBundle `json:"items"`
}
