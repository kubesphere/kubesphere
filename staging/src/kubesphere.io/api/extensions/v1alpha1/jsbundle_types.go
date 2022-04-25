/*
Copyright 2022 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	// +optional
	FilenameOverride string `json:"filenameOverride,omitempty"`
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
	Link  string `json:"link,omitempty"`
	State string `json:"state,omitempty"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories="plugin",scope="Cluster"

// JSBundle declares a js bundle that needs to be injected into ks-console,
// the endpoint can be provided by a service or a static file.
type JSBundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JSBundleSpec   `json:"spec,omitempty"`
	Status JSBundleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type JSBundleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JSBundle `json:"items"`
}
