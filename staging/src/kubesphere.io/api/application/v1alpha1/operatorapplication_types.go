/*
Copyright 2021.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OperatorApplicationSpec defines the desired state of OperatorApplication
type OperatorApplicationSpec struct {
	// the name of the operator application
	AppName string `json:"name"`
	// description from operator's description or frontend
	Description   string `json:"description,omitempty"`
	DescriptionEn string `json:"description_en,omitempty"`
	Abstraction   string `json:"abstraction,omitempty"`
	AbstractionEn string `json:"abstraction_en,omitempty"`
	Operator      string `json:"operator"`
	Icon          string `json:"icon,omitempty"`
	Owner         string `json:"owner,omitempty"`
}

// OperatorApplicationStatus defines the observed state of OperatorApplication
type OperatorApplicationStatus struct {
	State string `json:"state,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+genclient:nonNamespaced
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorApplication is the Schema for the operatorapplications API
type OperatorApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorApplicationSpec   `json:"spec,omitempty"`
	Status OperatorApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorApplicationList contains a list of OperatorApplication
type OperatorApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperatorApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OperatorApplication{}, &OperatorApplicationList{})
}
