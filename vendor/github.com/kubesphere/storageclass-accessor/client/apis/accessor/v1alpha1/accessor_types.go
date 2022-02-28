/*
Copyright 2021 f10atin9.

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

// +kubebuilder:validation:Enum=In;NotIn

type Operator string

// +kubebuilder:validation:Enum=Name;Status

type Field string

const (
	In    Operator = "In"
	NotIn Operator = "NotIn"

	Name   Field = "Name"
	Status Field = "Status"
)

// AccessorSpec defines the desired state of Accessor
type AccessorSpec struct {
	StorageClassName  string        `json:"storageClassName"`
	NameSpaceSelector NameSpaceList `json:"namespaceSelector"`
	WorkSpaceSelector WorkSpaceList `json:"workspaceSelector"`
}

type NameSpaceList struct {
	LabelSelector []MatchExpressions `json:"labelSelector"`
	FieldSelector []FieldExpressions `json:"fieldSelector"`
}

type FieldExpressions struct {
	FieldExpressions []FieldExpression `json:"fieldExpressions"`
}
type MatchExpressions struct {
	MatchExpressions []MatchExpression `json:"matchExpressions"`
}

type MatchExpression struct {
	Key      string   `json:"key"`
	Operator Operator `json:"operator"`
	Values   []string `json:"values"`
}

type FieldExpression struct {
	Field    Field    `json:"field"`
	Operator Operator `json:"operator"`
	Values   []string `json:"values"`
}

type WorkSpaceList struct {
	LabelSelector []MatchExpressions `json:"labelSelector"`
	FieldSelector []FieldExpressions `json:"fieldSelector"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Accessor is the Schema for the accessors API
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="StorageClass",type=string,JSONPath=`.spec.storageClassName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Accessor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AccessorSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// AccessorList contains a list of Accessor
type AccessorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Accessor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Accessor{}, &AccessorList{})
}
