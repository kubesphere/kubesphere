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
	ResourceKindS2iBinary     = "S2iBinary"
	ResourceSingularS2iBinary = "s2ibinary"
	ResourcePluralS2iBinary   = "s2ibinaries"
)

const (
	StatusUploading    = "Uploading"
	StatusReady        = "Ready"
	StatusUploadFailed = "UploadFailed"
)

const (
	S2iBinaryFinalizerName = "s2ibinary.finalizers.kubesphere.io"
	S2iBinaryLabelKey      = "s2ibinary-name.kubesphere.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// S2iBinarySpec defines the desired state of S2iBinary
type S2iBinarySpec struct {
	//FileName is filename of binary
	FileName string `json:"fileName,omitempty"`
	//MD5 is Binary's MD5 Hash
	MD5 string `json:"md5,omitempty"`
	//Size is the file size of file
	Size string `json:"size,omitempty"`
	//DownloadURL in KubeSphere
	DownloadURL string `json:"downloadURL,omitempty"`
	// UploadTime is last upload time
	UploadTimeStamp *metav1.Time `json:"uploadTimeStamp,omitempty"`
}

// S2iBinaryStatus defines the observed state of S2iBinary
type S2iBinaryStatus struct {
	//Phase is status of S2iBinary . Possible value is "Ready","UnableToDownload"
	Phase string `json:"phase,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S2iBinary is the Schema for the s2ibinaries API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="FileName",type="string",JSONPath=".spec.fileName"
// +kubebuilder:printcolumn:name="MD5",type="string",JSONPath=".spec.md5"
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=".spec.size"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
type S2iBinary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S2iBinarySpec   `json:"spec,omitempty"`
	Status S2iBinaryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S2iBinaryList contains a list of S2iBinary
type S2iBinaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S2iBinary `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S2iBinary{}, &S2iBinaryList{})
}
