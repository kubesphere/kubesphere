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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/constants"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OperatorApplicationVersionSpec defines the desired state of OperatorApplicationVersion
type OperatorApplicationVersionSpec struct {
	// the name of the operator
	AppName         string `json:"name"`
	Screenshots     string `json:"screenshots,omitempty"`
	ScreenshotsEn   string `json:"screenshots_en,omitempty"`
	ChangeLog       string `json:"changeLog"`
	ChangeLogEn     string `json:"changeLog_en,omitempty"`
	OperatorVersion string `json:"operatorVersion"`
	AppVersion      string `json:"appVersion"`
	Owner           string `json:"owner,omitempty"`
}

// OperatorApplicationVersionStatus defines the observed state of OperatorApplicationVersion
type OperatorApplicationVersionStatus struct {
	State string `json:"state,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+genclient:nonNamespaced
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorApplicationVersion is the Schema for the operatorapplicationversions API
type OperatorApplicationVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorApplicationVersionSpec   `json:"spec,omitempty"`
	Status OperatorApplicationVersionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OperatorApplicationVersionList contains a list of OperatorApplicationVersion
type OperatorApplicationVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperatorApplicationVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OperatorApplicationVersion{}, &OperatorApplicationVersionList{})
}

func (in *OperatorApplicationVersion) GetVersionName() string {
	appV := in.Spec.AppVersion
	if appV != "" {
		return fmt.Sprintf("%s [%s]", in.Spec.OperatorVersion, appV)
	} else {
		return in.Spec.AppVersion
	}
}

func (in *OperatorApplicationVersion) GetOperatorVersionType() string {
	return getValue(in.Labels, constants.OperatorAppLabelKey)
}
