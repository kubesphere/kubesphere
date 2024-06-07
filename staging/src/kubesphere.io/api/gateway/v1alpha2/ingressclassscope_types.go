/*
Copyright 2023 The KubeSphere Authors.

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
