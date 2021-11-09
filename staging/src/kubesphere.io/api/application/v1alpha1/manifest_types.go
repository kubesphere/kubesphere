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
	"kubesphere.io/api/constants"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ManifestSpec defines the desired state of Manifest
type ManifestSpec struct {
	// cluster name
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	// kind of the database cluster
	Kind string `json:"kind"`
	// info from frontend
	Description    string `json:"description,omitempty"`
	AppName        string `json:"app,omitempty"`
	AppVersion     string `json:"appVersion"`
	CustomResource string `json:"customResource" yaml:"customResource"`
	// expected release version, when this version is not equal status.version, the release need upgrade
	// this filed should be modified when any filed of the spec modified.
	Version int `json:"version"`
}

// ManifestStatus defines the observed state of Manifest
type ManifestStatus struct {
	State         string      `json:"state,omitempty"`
	ResourceState string      `json:"resourceState,omitempty"`
	Condition     []ApiResult `json:"condition,omitempty"`
	// current manifest version
	Version    int          `json:"version,omitempty"`
	LastUpdate *metav1.Time `json:"lastUpdate,omitempty"`
}

// ApiResult defines the result of pg operator ApiServer
type ApiResult struct {
	Api  string `json:"api,omitempty"`
	Code string `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
	Data string `json:"data,omitempty"`
}

// +kubebuilder:printcolumn:name="Kind",type="string",JSONPath=".spec.kind"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="Application",type="string",JSONPath=".spec.application"
// +kubebuilder:printcolumn:name="AppVersion",type="string",JSONPath=".spec.appVersion"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Manifest is the Schema for the manifests API
type Manifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManifestSpec   `json:"spec,omitempty"`
	Status ManifestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManifestList contains a list of Manifest
type ManifestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Manifest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Manifest{}, &ManifestList{})
}

func (in *Manifest) GetManifestCluster() string {
	return getValue(in.Labels, constants.ClusterNameLabelKey)
}

func (in *Manifest) GetManifestWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

func (in *Manifest) GetManifestNamespace() string {
	return getValue(in.Labels, constants.NamespaceLabelKey)
}

func (in *Manifest) GetCreator() string {
	return getValue(in.Annotations, constants.CreatorAnnotationKey)
}