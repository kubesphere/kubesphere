/*

 Copyright 2020 The KubeSphere Authors.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourcesPluralFedNamespace   = "federatednamespaces"
	ResourcesSingularFedNamespace = "federatednamespace"
	FedNamespaceKind              = "FederatedNamespace"
)

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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FederatedNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedNamespaceSpec `json:"spec"`
}

type FederatedNamespaceSpec struct {
	Template  NamespaceTemplate `json:"template"`
	Placement Placement         `json:"placement"`
	Overrides []Override        `json:"overrides,omitempty"`
}

type NamespaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              corev1.NamespaceSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedNamespaceList contains a list of federatednamespacelists
type FederatedNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedNamespace `json:"items"`
}
