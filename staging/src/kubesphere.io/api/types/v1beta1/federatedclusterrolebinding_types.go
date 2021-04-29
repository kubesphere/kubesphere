/*
Copyright 2020 KubeSphere Authors

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
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourcePluralFederatedClusterRoleBindingBinding = "federatedclusterrolebindings"
	ResourceSingularFederatedClusterRoleBinding      = "federatedclusterrolebinding"
	FederatedClusterRoleBindingKind                  = "FederatedClusterRoleBinding"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type FederatedClusterRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedClusterRoleBindingSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedClusterRoleBindingSpec struct {
	Template  ClusterRoleBindingTemplate `json:"template"`
	Placement GenericPlacementFields     `json:"placement"`
	Overrides []GenericOverrideItem      `json:"overrides,omitempty"`
}

type ClusterRoleBindingTemplate struct {
	// +optional
	Subjects []v1.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef can only reference a ClusterRole in the global namespace.
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	RoleRef v1.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedClusterRoleBindingList contains a list of federatedclusterrolebindinglists
type FederatedClusterRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedClusterRoleBinding `json:"items"`
}
