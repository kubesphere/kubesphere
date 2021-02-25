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

// WechatConfigSpec defines the desired state of WechatConfig
type WechatConfigSpec struct {
	// The WeChat API URL.
	WechatApiUrl string `json:"wechatApiUrl,omitempty"`
	// The corp id for authentication.
	WechatApiCorpId string `json:"wechatApiCorpId"`
	// The id of the application which sending message.
	WechatApiAgentId string `json:"wechatApiAgentId"`
	// The API key to use when talking to the WeChat API.
	WechatApiSecret *SecretKeySelector `json:"wechatApiSecret"`
}

// WechatConfigStatus defines the observed state of WechatConfig
type WechatConfigStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=wcc
// +kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// WechatConfig is the Schema for the wechatconfigs API
type WechatConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WechatConfigSpec   `json:"spec,omitempty"`
	Status WechatConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WechatConfigList contains a list of WechatConfig
type WechatConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WechatConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WechatConfig{}, &WechatConfigList{})
}
