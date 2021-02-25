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

// WechatReceiverSpec defines the desired state of WechatReceiver
type WechatReceiverSpec struct {
	// WechatConfig to be selected for this receiver
	WechatConfigSelector *metav1.LabelSelector `json:"wechatConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// +optional
	ToUser  string `json:"toUser,omitempty"`
	ToParty string `json:"toParty,omitempty"`
	ToTag   string `json:"toTag,omitempty"`
}

// WechatReceiverStatus defines the observed state of WechatReceiver
type WechatReceiverStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=wcr
// +kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// WechatReceiver is the Schema for the wechatreceivers API
type WechatReceiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WechatReceiverSpec   `json:"spec,omitempty"`
	Status WechatReceiverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WechatReceiverList contains a list of WechatReceiver
type WechatReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WechatReceiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WechatReceiver{}, &WechatReceiverList{})
}
