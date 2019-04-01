/*
Copyright 2019 The Kubesphere Authors.

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

// S2iRunSpec defines the desired state of S2iRun
type S2iRunSpec struct {
	//BuilderName specify the name of s2ibuilder, required
	BuilderName string `json:"builderName"`
	//BackoffLimit limits the restart count of each s2irun. Default is 0
	BackoffLimit int32 `json:"backoffLimit,omitempty"`
	//SecondsAfterFinished if is set and greater than zero, and the job created by s2irun become successful or failed , the job will be auto deleted after SecondsAfterFinished
	SecondsAfterFinished int32 `json:"secondsAfterFinished,omitempty"`
	//NewTag override the default tag in its s2ibuilder, image name cannot be changed.
	NewTag string `json:"newTag,omitempty"`
}

// S2iRunStatus defines the observed state of S2iRun
type S2iRunStatus struct {
	// StartTime represent when this run began
	StartTime *metav1.Time `json:"startTime,omitempty" protobuf:"bytes,2,opt,name=startTime"`

	// Represents time when the job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty" protobuf:"bytes,3,opt,name=completionTime"`
	// RunState  indicates whether this job is done or failed
	RunState RunState `json:"runState,omitempty"`
	//LogURL is uesd for external log handler to let user know where is log located in
	LogURL string `json:"logURL,omitempty"`
	//KubernetesJobName is the job name in k8s
	KubernetesJobName string `json:"kubernetesJobName,omitempty"`
	//ImageName is the name of artifact
	ImageName string `json:"imageName,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S2iRun is the Schema for the s2iruns API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=s2ir
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.runState"
// +kubebuilder:printcolumn:name="K8sJobName",type="string",JSONPath=".status.kubernetesJobName"
// +kubebuilder:printcolumn:name="StartTime",type="date",JSONPath=".status.startTime"
// +kubebuilder:printcolumn:name="CompletionTime",type="date",JSONPath=".status.completionTime"
// +kubebuilder:printcolumn:name="ImageName",type="string",JSONPath=".status.imageName"
type S2iRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S2iRunSpec   `json:"spec,omitempty"`
	Status S2iRunStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S2iRunList contains a list of S2iRun
type S2iRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S2iRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S2iRun{}, &S2iRunList{})
}
