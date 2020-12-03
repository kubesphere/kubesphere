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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourcePluralFederatedWorkspaceRoleBinding   = "federatedworkspacerolebindings"
	ResourceSingularFederatedWorkspaceRoleBinding = "federatedworkspacerolebinding"
	FederatedWorkspaceRoleBindingKind             = "FederatedWorkspaceRoleBinding"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type FederatedWorkspaceRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedWorkspaceRoleBindingSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedWorkspaceRoleBindingSpec struct {
	Template  WorkspaceRoleBindingTemplate `json:"template"`
	Placement GenericPlacementFields       `json:"placement"`
	Overrides []GenericOverrideItem        `json:"overrides,omitempty"`
}

type WorkspaceRoleBindingTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Subjects holds references to the objects the role applies to.
	// +optional
	Subjects []rbacv1.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef can only reference a WorkspaceRole.
	// If the RoleRef cannot be resolved, the Authorizer must return an error.
	RoleRef rbacv1.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedWorkspaceRoleBindingList contains a list of FederatedWorkspaceRoleBinding
type FederatedWorkspaceRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedWorkspaceRoleBinding `json:"items"`
}
