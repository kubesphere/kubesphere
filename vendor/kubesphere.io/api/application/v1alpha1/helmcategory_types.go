/*
Copyright 2020 The KubeSphere Authors.

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
	ResourceKindHelmCategory     = "HelmCategory"
	ResourceSingularHelmCategory = "helmcategory"
	ResourcePluralHelmCategory   = "helmcategories"
)

// HelmCategorySpec defines the desired state of HelmRepo
type HelmCategorySpec struct {
	// name of the category
	Name string `json:"name"`
	// info from frontend
	Description string `json:"description,omitempty"`
	Locale      string `json:"locale,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=hctg
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="total",type=string,JSONPath=`.status.total`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmCategory is the Schema for the helmcategories API
type HelmCategory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmCategorySpec   `json:"spec,omitempty"`
	Status HelmCategoryStatus `json:"status,omitempty"`
}

type HelmCategoryStatus struct {
	// total helmapplications belong to this category
	Total int `json:"total"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmCategoryList contains a list of HelmCategory
type HelmCategoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmCategory `json:"items"`
}

func (in *HelmCategory) GetTrueName() string {
	if in == nil {
		return ""
	}
	return in.Spec.Name
}

func init() {
	SchemeBuilder.Register(&HelmCategory{}, &HelmCategoryList{})
}
