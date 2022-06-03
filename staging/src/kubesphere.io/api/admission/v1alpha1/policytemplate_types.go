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
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindPolicyTemplate      = "PolicyTemplate"
	ResourcesSingularPolicyTemplate = "policytemplate"
	ResourcesPluralPolicyTemplate   = "policytemplates"

	AdmissionPolicyTemplateLabel = "kubesphere.io/admission/policytemplate"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories="admission",scope="Cluster"
// +kubebuilder:printcolumn:name="Name",type="string",JSONPath=".spec.name"

// PolicyTemplate is the Schema for the policy templates API
type PolicyTemplate struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PolicyTemplateSpec `json:"spec"`
	// +optional
	Status PolicyTemplateStatus `json:"status,omitempty"`
}

// PolicyTemplateSpec defines the desired state of PolicyTemplate
type PolicyTemplateSpec struct {
	// Name of policy template
	Name string `json:"name"`
	// +optional
	Description string `json:"description,omitempty"`
	// content of template
	Content PolicyTemplateContent `json:"content"`
}

type PolicyTemplateContent struct {
	// spec of policy content
	Spec PolicyTemplateContentSpec `json:"spec"`
	// target of policy content
	Targets []PolicyTemplateContentSpecTarget `json:"targets"`
}

type PolicyTemplateContentSpec struct {
	// policy rule CRD name spec
	Names Names `json:"names"`
	// policy rule parameters
	Parameters Parameters `json:"parameters"`
}

type Names struct {
	// policy rule CRD name
	Name string `json:"name"`
	// policy rule CRD short names
	ShortNames []string `json:"short_names"`
}

type Parameters struct {
	// validation for policy rule parameters
	Validation *Validation `json:"validation"`
}

type Validation struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	OpenAPIV3Schema *apiextensions.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`
	// +kubebuilder:default=false
	LegacySchema bool `json:"legacySchema,omitempty"`
}

type PolicyTemplateContentSpecTarget struct {
	// target name
	Target string `json:"target,omitempty"`
	// target admission provider
	Provider string `json:"provider,omitempty"`
	// rego etc.
	Expression string `json:"expression,omitempty"`
	// import from other resource
	Import []string `json:"import,omitempty"`
}

// PolicyTemplateStatus defines the observed state of PolicyTemplate
type PolicyTemplateStatus struct{}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyTemplateList contains a list of PolicyTemplate
type PolicyTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyTemplate `json:"items"`
}
