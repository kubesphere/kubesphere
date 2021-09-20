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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceKindRule      = "PolicyTemplate"
	ResourcesSingularRule = "rule"
	ResourcesPluralRule   = "rules"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// PolicyTemplate is the Schema for the rules API
// +kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:resource:categories="admission",scope="Cluster"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Rule struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RuleSpec `json:"spec"`
	// +optional
	Status RuleStatus `json:"status,omitempty"`
}

// RuleSpec defines the desired state of PolicyTemplate
type RuleSpec struct {
	// Name of rule
	Name string `json:"name"`
	// Name of the policy.
	// +optional
	Policy string `json:"templateName,omitempty"`
	// Name of the admission provider.
	Provider string `json:"provider,omitempty"`
	// Description of the rule.
	// +optional
	Description string `json:"description,omitempty"`
	// Match
	Match Match `json:"match,omitempty"`
	// Parameters
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:EmbeddedResource
	Parameters runtime.RawExtension `json:"parameters"`
}

// Match selects objects to apply mutations to.
// +kubebuilder:object:generate=true
type Match struct {
	Kinds              []Kinds                       `json:"kinds,omitempty"`
	Scope              apiextensionsv1.ResourceScope `json:"scope,omitempty"`
	Namespaces         []string                      `json:"namespaces,omitempty"`
	ExcludedNamespaces []string                      `json:"excludedNamespaces,omitempty"`
	LabelSelector      *metav1.LabelSelector         `json:"labelSelector,omitempty"`
	NamespaceSelector  *metav1.LabelSelector         `json:"namespaceSelector,omitempty"`
}

// Kinds accepts a list of objects with apiGroups and kinds fields
// that list the groups/kinds of objects to which the mutation will apply.
// If multiple groups/kinds objects are specified,
// only one match is needed for the resource to be in scope.
// +kubebuilder:object:generate=true

type Kinds struct {
	// APIGroups is the API groups the resources belong to. '*' is all groups.
	// If '*' is present, the length of the slice must be one.
	// Required.
	APIGroups []string `json:"apiGroups,omitempty" protobuf:"bytes,1,rep,name=apiGroups"`
	Kinds     []string `json:"kinds,omitempty"`
}

type RuleState string

// These are the valid phases of a rule.
const (
	// RuleActive means the rule is available.
	RuleActive RuleState = "Active"
	// RuleInactive means the rule is disabled.
	RuleInactive RuleState = "Inactive"
)

// RuleStatus defines the observed state of Rule
type RuleStatus struct {
	// The rule status
	// +optional
	State RuleState `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleList contains a list of PolicyTemplate
type RuleList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rule `json:"items"`
}
