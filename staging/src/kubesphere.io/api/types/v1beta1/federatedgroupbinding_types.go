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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/iam/v1alpha2"
)

const (
	ResourcePluralFederatedGroupBinding   = "federatedgroupbindings"
	ResourceSingularFederatedGroupBinding = "federatedgroupbinding"
	FederatedGroupBindingKind             = "FederatedGroupBinding"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
type FederatedGroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedGroupBindingSpec `json:"spec"`
	Status            *GenericFederatedStatus   `json:"status,omitempty"`
}

type FederatedGroupBindingSpec struct {
	Template  GroupBindingTemplate   `json:"template"`
	Placement GenericPlacementFields `json:"placement"`
	Overrides []GenericOverrideItem  `json:"overrides,omitempty"`
}

type GroupBindingTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	GroupRef v1alpha2.GroupRef `json:"groupRef,omitempty"`

	// +optional
	Users []string `json:"users,omitempty"`
}

// +kubebuilder:object:root=true

// FederatedGroupBindingList contains a list of federateduserlists
type FederatedGroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedGroupBinding `json:"items"`
}
