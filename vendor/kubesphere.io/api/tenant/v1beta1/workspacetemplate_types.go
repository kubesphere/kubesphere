/*
Copyright 2019 The KubeSphere Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkspaceTemplate     = "WorkspaceTemplate"
	ResourceSingularWorkspaceTemplate = "workspacetemplate"
	ResourcePluralWorkspaceTemplate   = "workspacetemplates"
)

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

type Template struct {
	ObjectMeta `json:"metadata,omitempty"`
	Spec       WorkspaceSpec `json:"spec,omitempty"`
}

type GenericClusterReference struct {
	Name string `json:"name"`
}

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
type WorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceTemplateList contains a list of WorkspaceTemplate
type WorkspaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplate `json:"items"`
}
