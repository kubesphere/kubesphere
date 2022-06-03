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

package templates

import (
	"reflect"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConstraintTemplateSpec defines the desired state of ConstraintTemplate
type ConstraintTemplateSpec struct {
	CRD     CRD      `json:"crd,omitempty"`
	Targets []Target `json:"targets,omitempty"`
}

type CRD struct {
	Spec CRDSpec `json:"spec,omitempty"`
}

type CRDSpec struct {
	Names      Names       `json:"names,omitempty"`
	Validation *Validation `json:"validation,omitempty"`
}

type Names struct {
	Kind       string   `json:"kind,omitempty"`
	ShortNames []string `json:"shortNames,omitempty"`
}

type Validation struct {
	// +kubebuilder:validation:Schemaless
	OpenAPIV3Schema *apiextensions.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`
}

type Target struct {
	Target string   `json:"target,omitempty"`
	Rego   string   `json:"rego,omitempty"`
	Libs   []string `json:"libs,omitempty"`
}

// CreateCRDError represents a single error caught during parsing, compiling, etc.
type CreateCRDError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
}

// ByPodStatus defines the observed state of ConstraintTemplate as seen by
// an individual controller
type ByPodStatus struct {
	// a unique identifier for the pod that wrote the status
	ID                 string           `json:"id,omitempty"`
	ObservedGeneration int64            `json:"observedGeneration,omitempty"`
	Errors             []CreateCRDError `json:"errors,omitempty"`
}

// ConstraintTemplateStatus defines the observed state of ConstraintTemplate
type ConstraintTemplateStatus struct {
	Created bool          `json:"created,omitempty"`
	ByPod   []ByPodStatus `json:"byPod,omitempty"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:skip

// ConstraintTemplate is the Schema for the constrainttemplates API
// +k8s:openapi-gen=true
type ConstraintTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConstraintTemplateSpec   `json:"spec,omitempty"`
	Status ConstraintTemplateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConstraintTemplateList contains a list of ConstraintTemplate
type ConstraintTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConstraintTemplate `json:"items"`
}

// SemanticEqual returns whether there have been changes to a constraint that
// the framework should know about. It can ignore metadata as it assumes the
// two comparables share the same identity
func (ct *ConstraintTemplate) SemanticEqual(other *ConstraintTemplate) bool {
	return reflect.DeepEqual(ct.Spec, other.Spec)
}
