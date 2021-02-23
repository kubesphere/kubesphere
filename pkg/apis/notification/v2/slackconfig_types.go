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

// SlackConfigSpec defines the desired state of SlackConfig
type SlackConfigSpec struct {
	// The token of user or bot.
	SlackTokenSecret *SecretKeySelector `json:"slackTokenSecret,omitempty"`
}

// SlackConfigStatus defines the observed state of SlackConfig
type SlackConfigStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=sc
// +kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// SlackConfig is the Schema for the slackconfigs API
type SlackConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlackConfigSpec   `json:"spec,omitempty"`
	Status SlackConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SlackConfigList contains a list of SlackConfig
type SlackConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SlackConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SlackConfig{}, &SlackConfigList{})
}
