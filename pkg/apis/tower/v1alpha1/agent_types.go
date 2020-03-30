/*

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AgentSpec defines the desired state of Agent
type AgentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Token used by agents to connect to proxy.
	// +optional
	Token string `json:"token,omitempty"`

	// Proxy address
	// +optional
	Proxy string `json:"proxy,omitempty"`

	// KubeAPIServerPort is the port which listens for forwarding kube-apiserver traffic
	// +optional
	KubernetesAPIServerPort uint16 `json:"kubernetesAPIServerPort,omitempty"`

	// KubeSphereAPIServerPort is the port which listens for forwarding kubesphere apigateway traffic
	// +optional
	KubeSphereAPIServerPort uint16 `json:"kubesphereAPIServerPort,omitempty"`

	// Indicates that the agent is paused.
	// +optional
	Paused bool
}

type AgentConditionType string

const (
	// Agent is initialized, and waiting for establishing to a proxy server
	AgentInitialized AgentConditionType = "Initialized"

	// Agent has successfully connected to proxy server
	AgentConnected AgentConditionType = "Connected"
)

type AgentCondition struct {
	// Type of AgentCondition
	Type AgentConditionType `json:"type,omitempty"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// AgentStatus defines the observed state of Agent
type AgentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Represents the latest available observations of a agent's current state.
	Conditions []AgentCondition `json:"conditions,omitempty"`

	// Represents the connection quality, in ms
	Ping uint64 `json:"ping,omitempty"`

	// Issued new kubeconfig by proxy server
	KubeConfig []byte
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// Agent is the Schema for the agents API
type Agent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSpec   `json:"spec,omitempty"`
	Status AgentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AgentList contains a list of Agent
type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Agent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Agent{}, &AgentList{})
}
