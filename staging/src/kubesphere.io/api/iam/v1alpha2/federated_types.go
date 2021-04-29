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

package v1alpha2

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ResourcesSingularFedUser                 = "federateduser"
	ResourcesSingularFedGlobalRoleBinding    = "federatedglobalrolebinding"
	ResourcesSingularFedWorkspaceRoleBinding = "federatedworkspacerolebinding"
	ResourcesSingularFedGlobalRole           = "federatedglobalrole"
	ResourcesSingularFedWorkspaceRole        = "federatedworkspacerole"
	ResourcesPluralFedUser                   = "federatedusers"
	ResourcesPluralFedGlobalRoleBinding      = "federatedglobalrolebindings"
	ResourcesPluralFedWorkspaceRoleBinding   = "federatedworkspacerolebindings"
	ResourcesPluralFedGlobalRole             = "federatedglobalroles"
	ResourcesPluralFedWorkspaceRole          = "federatedworkspaceroles"
	FedClusterRoleBindingKind                = "FederatedClusterRoleBinding"
	FedClusterRoleKind                       = "FederatedClusterRole"
	FedGlobalRoleKind                        = "FederatedGlobalRole"
	FedWorkspaceRoleKind                     = "FederatedWorkspaceRole"
	FedGlobalRoleBindingKind                 = "FederatedGlobalRoleBinding"
	FedWorkspaceRoleBindingKind              = "FederatedWorkspaceRoleBinding"
	fedResourceGroup                         = "types.kubefed.io"
	fedResourceVersion                       = "v1beta1"
	FedUserKind                              = "FederatedUser"
)

var (
	FedUserResource = metav1.APIResource{
		Name:         ResourcesPluralFedUser,
		SingularName: ResourcesSingularFedUser,
		Namespaced:   false,
		Group:        fedResourceGroup,
		Version:      fedResourceVersion,
		Kind:         FedUserKind,
	}
	FedGlobalRoleBindingResource = metav1.APIResource{
		Name:         ResourcesPluralFedGlobalRoleBinding,
		SingularName: ResourcesSingularFedGlobalRoleBinding,
		Namespaced:   false,
		Group:        fedResourceGroup,
		Version:      fedResourceVersion,
		Kind:         FedGlobalRoleBindingKind,
	}
	FedWorkspaceRoleBindingResource = metav1.APIResource{
		Name:         ResourcesPluralFedWorkspaceRoleBinding,
		SingularName: ResourcesSingularFedWorkspaceRoleBinding,
		Namespaced:   false,
		Group:        fedResourceGroup,
		Version:      fedResourceVersion,
		Kind:         FedWorkspaceRoleBindingKind,
	}
	FedGlobalRoleResource = metav1.APIResource{
		Name:         ResourcesPluralFedGlobalRole,
		SingularName: ResourcesSingularFedGlobalRole,
		Namespaced:   false,
		Group:        fedResourceGroup,
		Version:      fedResourceVersion,
		Kind:         FedGlobalRoleKind,
	}

	FedWorkspaceRoleResource = metav1.APIResource{
		Name:         ResourcesPluralFedWorkspaceRole,
		SingularName: ResourcesSingularFedWorkspaceRole,
		Namespaced:   false,
		Group:        fedResourceGroup,
		Version:      fedResourceVersion,
		Kind:         FedWorkspaceRoleKind,
	}

	FederatedClusterRoleBindingResource = schema.GroupVersionResource{
		Group:    fedResourceGroup,
		Version:  fedResourceVersion,
		Resource: "federatedclusterrolebindings",
	}
)

type FederatedRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedRoleBindingSpec `json:"spec"`
}

type FederatedRoleBindingSpec struct {
	Template  RoleBindingTemplate `json:"template"`
	Placement Placement           `json:"placement"`
}
type RoleBindingTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Subjects          []rbacv1.Subject `json:"subjects,omitempty"`
	RoleRef           rbacv1.RoleRef   `json:"roleRef"`
}

type FederatedRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedRoleSpec `json:"spec"`
}

type FederatedRoleSpec struct {
	Template  RoleTemplate `json:"template"`
	Placement Placement    `json:"placement"`
}

type RoleTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Rules []rbacv1.PolicyRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
}

type FederatedUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedUserSpec `json:"spec"`
}

type FederatedUserSpec struct {
	Template  UserTemplate `json:"template"`
	Placement Placement    `json:"placement"`
}

type UserTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UserSpec `json:"spec"`
	// +optional
	Status UserStatus `json:"status,omitempty"`
}

type Placement struct {
	Clusters        []Cluster       `json:"clusters,omitempty"`
	ClusterSelector ClusterSelector `json:"clusterSelector,omitempty"`
}

type ClusterSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type Cluster struct {
	Name string `json:"name"`
}
