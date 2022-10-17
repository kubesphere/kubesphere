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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourcePluralFederatedDeployment   = "federateddeployments"
	ResourceSingularFederatedDeployment = "federateddeployment"
	FederatedDeploymentKind             = "FederatedDeployment"
)

// +genclient
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
type FederatedDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedDeploymentSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedDeploymentSpec struct {
	Template  DeploymentTemplate     `json:"template"`
	Placement GenericPlacementFields `json:"placement"`
	Overrides []GenericOverrideItem  `json:"overrides,omitempty"`
}

type DeploymentTemplate struct {
	Spec appsv1.DeploymentSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// FederatedDeploymentList contains a list of federateddeploymentlists
type FederatedDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedDeployment `json:"items"`
}
