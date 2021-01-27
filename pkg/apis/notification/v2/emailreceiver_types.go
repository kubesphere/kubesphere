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

package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EmailReceiverSpec defines the desired state of EmailReceiver
type EmailReceiverSpec struct {
	// Receivers' email addresses
	To []string `json:"to"`
	// EmailConfig to be selected for this receiver
	EmailConfigSelector *metav1.LabelSelector `json:"emailConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
}

// EmailReceiverStatus defines the observed state of EmailReceiver
type EmailReceiverStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=er
// +genclient
// +genclient:nonNamespaced
// EmailReceiver is the Schema for the emailreceivers API
type EmailReceiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EmailReceiverSpec   `json:"spec,omitempty"`
	Status EmailReceiverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EmailReceiverList contains a list of EmailReceiver
type EmailReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EmailReceiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EmailReceiver{}, &EmailReceiverList{})
}
