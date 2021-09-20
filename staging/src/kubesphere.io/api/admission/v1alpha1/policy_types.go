/*
Copyright 2019 The KubeSphere Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindPolicy      = "Policy"
	ResourcesSingularPolicy = "policy"
	ResourcesPluralPolicy   = "policies"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// PolicyTemplate is the Schema for the rules API
// +kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.name"
// +kubebuilder:resource:categories="admission",scope="Cluster"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Policy struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PolicySpec `json:"spec"`
	// +optional
	Status PolicyStatus `json:"status,omitempty"`
}

// PolicySpec defines the desired state of PolicyTemplate
type PolicySpec struct {
	// Name of policy
	Name string `json:"name"`
	// +optional
	Description string `json:"description,omitempty"`

	Provider string `json:"provider"`
	// content of policy
	Content PolicyContent `json:"content"`
}

type PolicyContent struct {
	// spec of policy content
	Spec PolicyContentSpec `json:"spec"`
	// target of policy content
	Targets []PolicyContentTarget `json:"targets"`
}

type PolicyContentSpec struct {
	// policy rule CRD name spec
	Names Names `json:"names"`
	// policy rule parameters
	Parameters Parameters `json:"parameters"`
}

type PolicyContentTarget struct {
	// target name
	Target string `json:"target,omitempty"`
	// rego etc.
	Expression string `json:"expression,omitempty"`
	// import from other resource
	Import []string `json:"import,omitempty"`
}

type PolicyState string

// These are the valid phases of a rule.
const (
	// PolicyActive means the policy is active.
	PolicyActive PolicyState = "Active"
	// PolicyInactive means the rule is inactive.
	PolicyInactive PolicyState = "Inactive"
)

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	// The rule status
	// +optional
	State PolicyState `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyTemplate `json:"items"`
}
