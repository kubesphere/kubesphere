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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
)

const (
	ResourceKindWorkspaceTemplate     = "WorkspaceTemplate"
	ResourceSingularWorkspaceTemplate = "workspacetemplate"
	ResourcePluralWorkspaceTemplate   = "workspacetemplates"
	ResourcesPluralFedWorkspace       = "federatedworkspaces"
	ResourcesSingularFedWorkspace     = "federatedworkspace"
	FedWorkspaceKind                  = "FederatedWorkspace"
	fedResourceGroup                  = "types.kubefed.io"
	fedResourceVersion                = "v1beta1"
)

var (
	FedWorkspaceResource = metav1.APIResource{
		Name:         ResourcesPluralFedWorkspace,
		SingularName: ResourcesSingularFedWorkspace,
		Namespaced:   false,
		Group:        fedResourceGroup,
		Version:      fedResourceVersion,
		Kind:         FedWorkspaceKind,
	}
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// WorkspaceTemplate is the Schema for the workspacetemplates API
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="tenant",scope="Cluster"
type WorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceTemplateSpec `json:"spec,omitempty"`
}

type WorkspaceTemplateSpec struct {
	Template  Template   `json:"template"`
	Placement Placement  `json:"placement"`
	Overrides []Override `json:"overrides,omitempty"`
}

type Template struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              v1alpha1.WorkspaceSpec `json:"spec"`
}

type Placement struct {
	Clusters        []Cluster        `json:"clusters,omitempty"`
	ClusterSelector *ClusterSelector `json:"clusterSelector,omitempty"`
}

type ClusterSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type Cluster struct {
	Name string `json:"name"`
}

type Override struct {
	ClusterName      string            `json:"clusterName"`
	ClusterOverrides []ClusterOverride `json:"clusterOverrides"`
}

type ClusterOverride struct {
	Path  string               `json:"path"`
	Op    string               `json:"op,omitempty"`
	Value runtime.RawExtension `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// WorkspaceTemplateList contains a list of WorkspaceTemplate
type WorkspaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceTemplate{}, &WorkspaceTemplateList{})
}

type FederatedWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceTemplateSpec `json:"spec"`
}
