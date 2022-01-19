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

const (
	ResourceKindServicePolicy     = "ServicePolicy"
	ResourceSingularServicePolicy = "servicepolicy"
	ResourcePluralServicePolicy   = "servicepolicies"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServicePolicySpec defines the desired state of ServicePolicy
type ServicePolicySpec struct {

	// Label selector for destination rules.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Template used to create a destination rule
	// +optional
	Template DestinationRuleSpecTemplate `json:"template,omitempty"`
}

type DestinationRuleSpecTemplate struct {

	// Metadata of the virtual services created from this template
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec indicates the behavior of a destination rule.
	// +optional
	Spec v1alpha3.DestinationRule `json:"spec,omitempty"`
}

type ServicePolicyConditionType string

// These are valid conditions of a strategy.
const (
	// StrategyComplete means the strategy has been delivered to istio.
	ServicePolicyComplete ServicePolicyConditionType = "Complete"

	// StrategyFailed means the strategy has failed its delivery to istio.
	ServicePolicyFailed ServicePolicyConditionType = "Failed"
)

// StrategyCondition describes current state of a strategy.
type ServicePolicyCondition struct {
	// Type of strategy condition, Complete or Failed.
	Type ServicePolicyConditionType `json:"type,omitempty"`

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

// ServicePolicyStatus defines the observed state of ServicePolicy
type ServicePolicyStatus struct {
	// The latest available observations of an object's current state.
	// +optional
	Conditions []ServicePolicyCondition `json:"conditions,omitempty"`

	// Represents time when the strategy was acknowledged by the controller.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the strategy was completed.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServicePolicy is the Schema for the servicepolicies API
// +k8s:openapi-gen=true
type ServicePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServicePolicySpec   `json:"spec,omitempty"`
	Status ServicePolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServicePolicyList contains a list of ServicePolicy
type ServicePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServicePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServicePolicy{}, &ServicePolicyList{})
}
