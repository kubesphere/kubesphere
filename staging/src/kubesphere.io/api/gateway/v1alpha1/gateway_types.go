/*
Copyright 2021 The KubeSphere Authors.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// GatewaySpec defines the desired state of Gateway
type GatewaySpec struct {
	Conroller  ControllerSpec `json:"controller,omitempty"`
	Service    ServiceSpec    `json:"service,omitempty"`
	Deployment DeploymentSpec `json:"deployment,omitempty"`
}

type ControllerSpec struct {
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +optional
	Config map[string]string `json:"config,omitempty"`
	// +optional
	Scope Scope `json:"scope,omitempty"`
}

type ServiceSpec struct {
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +optional
	Type corev1.ServiceType `json:"type,omitempty"`
}

type DeploymentSpec struct {
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Scope struct {
	Enabled   bool   `json:"enabled,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+genclient

// Gateway is the Schema for the gateways API
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GatewaySpec `json:"spec,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	Status runtime.RawExtension `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GatewayList contains a list of Gateway
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
}
