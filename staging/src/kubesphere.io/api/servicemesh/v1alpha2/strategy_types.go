/*
Copyright 2019 The KubeSphere authors.

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
	"istio.io/api/networking/v1alpha3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ResourceKindStrategy     = "Strategy"
	ResourceSingularStrategy = "strategy"
	ResourcePluralStrategy   = "strategies"
)

type strategyType string

const (
	// Canary strategy type
	CanaryType strategyType = "Canary"

	// BlueGreen strategy type
	BlueGreenType strategyType = "BlueGreen"

	// Mirror strategy type
	Mirror strategyType = "Mirror"
)

type StrategyPolicy string

const (
	// apply strategy only until workload is ready
	PolicyWaitForWorkloadReady StrategyPolicy = "WaitForWorkloadReady"

	// apply strategy immediately no matter workload status is
	PolicyImmediately StrategyPolicy = "Immediately"

	// pause strategy
	PolicyPause StrategyPolicy = "Paused"
)

// StrategySpec defines the desired state of Strategy
type StrategySpec struct {
	// Strategy type
	Type strategyType `json:"type,omitempty"`

	// Principal version, the one as reference version
	// label version value
	// +optional
	PrincipalVersion string `json:"principal,omitempty"`

	// Governor version, the version takes control of all incoming traffic
	// label version value
	// +optional
	GovernorVersion string `json:"governor,omitempty"`

	// Label selector for virtual services.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Template describes the virtual service that will be created.
	Template VirtualServiceTemplateSpec `json:"template,omitempty"`

	// strategy policy, how the strategy will be applied
	// by the strategy controller
	StrategyPolicy StrategyPolicy `json:"strategyPolicy,omitempty"`
}

// VirtualServiceTemplateSpec
type VirtualServiceTemplateSpec struct {

	// Metadata of the virtual services created from this template
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec indicates the behavior of a virtual service.
	// +optional
	Spec v1alpha3.VirtualService `json:"spec,omitempty"`
}

// StrategyStatus defines the observed state of Strategy
type StrategyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The latest available observations of an object's current state.
	// +optional
	Conditions []StrategyCondition `json:"conditions,omitempty"`

	// Represents time when the strategy was acknowledged by the controller.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the strategy was completed.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

type StrategyConditionType string

// These are valid conditions of a strategy.
const (
	// StrategyComplete means the strategy has been delivered to istio.
	StrategyComplete StrategyConditionType = "Complete"

	// StrategyFailed means the strategy has failed its delivery to istio.
	StrategyFailed StrategyConditionType = "Failed"
)

// StrategyCondition describes current state of a strategy.
type StrategyCondition struct {
	// Type of strategy condition, Complete or Failed.
	Type StrategyConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown
	Status apiextensions.ConditionStatus `json:"status,omitempty"`

	// Last time the condition was checked.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`

	// Last time the condition transit from one status to another
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// reason for the condition's last transition
	Reason string `json:"reason,omitempty"`

	// Human readable message indicating details about last transition.
	// +optinal
	Message string `json:"message,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true

// Strategy is the Schema for the strategies API
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="type of strategy"
// +kubebuilder:printcolumn:name="Hosts",type="string",JSONPath=".spec.template.spec.hosts",description="destination hosts"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
// +k8s:openapi-gen=true
type Strategy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StrategySpec   `json:"spec,omitempty"`
	Status StrategyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StrategyList contains a list of Strategy
type StrategyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Strategy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Strategy{}, &StrategyList{})
}
