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
	"kubesphere.io/kubesphere/pkg/apis/notification/v2alpha1"
)

const (
	ResourcePluralFederatedWebhookConfig   = "federatedwebhookconfigs"
	ResourceSingularFederatedWebhookConfig = "federatedwebhookconfig"
	FederatedWebhookConfigKind             = "FederatedWebhookConfig"
)

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
type FederatedWebhookConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FederatedWebhookConfigSpec `json:"spec"`

	Status *GenericFederatedStatus `json:"status,omitempty"`
}

type FederatedWebhookConfigSpec struct {
	Template  WebhookConfigTemplate  `json:"template"`
	Placement GenericPlacementFields `json:"placement"`
	Overrides []GenericOverrideItem  `json:"overrides,omitempty"`
}

type WebhookConfigTemplate struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              v2alpha1.WebhookConfigSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedWebhookConfigList contains a list of federatedwebhookconfiglists
type FederatedWebhookConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedWebhookConfig `json:"items"`
}
