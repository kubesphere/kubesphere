/*
Copyright 2022 The KubeSphere Authors.

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

type APIServiceSpec struct {
	Group    string `json:"group,omitempty"`
	Version  string `json:"version,omitempty"`
	Endpoint `json:",inline"`
}

type APIServiceStatus struct {
	State      string             `json:"state,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"

// APIService is a special resource used in Ks-apiserver
// declares a directional proxy path for a resource type API，
// it's similar to Kubernetes API Aggregation Layer.
type APIService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIServiceSpec   `json:"spec,omitempty"`
	Status APIServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type APIServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIService `json:"items"`
}
