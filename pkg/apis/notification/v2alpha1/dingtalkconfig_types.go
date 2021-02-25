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

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Configuration of conversation
type DingTalkApplicationConfig struct {
	// The key of the application with which to send messages.
	AppKey *SecretKeySelector `json:"appkey,omitempty"`
	// The key in the secret to be used. Must be a valid secret key.
	AppSecret *SecretKeySelector `json:"appsecret,omitempty"`
}

// DingTalkConfigSpec defines the desired state of DingTalkConfig
type DingTalkConfigSpec struct {
	// Only needed when send alerts to the conversation.
	Conversation *DingTalkApplicationConfig `json:"conversation,omitempty"`
}

// DingTalkConfigStatus defines the observed state of DingTalkConfig
type DingTalkConfigStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=dc
// +kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// DingTalkConfig is the Schema for the dingtalkconfigs API
type DingTalkConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DingTalkConfigSpec   `json:"spec,omitempty"`
	Status DingTalkConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DingTalkConfigList contains a list of DingTalkConfig
type DingTalkConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DingTalkConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DingTalkConfig{}, &DingTalkConfigList{})
}
