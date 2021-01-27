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

// SlackReceiverSpec defines the desired state of SlackReceiver
type SlackReceiverSpec struct {
	// SlackConfig to be selected for this receiver
	SlackConfigSelector *metav1.LabelSelector `json:"slackConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// The channel or user to send notifications to.
	Channel string `json:"channel"`
}

// SlackReceiverStatus defines the observed state of SlackReceiver
type SlackReceiverStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=sr
// +genclient
// +genclient:nonNamespaced
// SlackReceiver is the Schema for the slackreceivers API
type SlackReceiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlackReceiverSpec   `json:"spec,omitempty"`
	Status SlackReceiverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SlackReceiverList contains a list of SlackReceiver
type SlackReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SlackReceiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SlackReceiver{}, &SlackReceiverList{})
}
