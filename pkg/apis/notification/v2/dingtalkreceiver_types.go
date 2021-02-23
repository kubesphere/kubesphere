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

// Configuration of ChatBot
type DingTalkChatBot struct {
	// The webhook of ChatBot which the message will send to.
	Webhook *SecretKeySelector `json:"webhook"`

	// Custom keywords of ChatBot
	Keywords []string `json:"keywords,omitempty"`

	// Secret of ChatBot, you can get it after enabled Additional Signature of ChatBot.
	Secret *SecretKeySelector `json:"secret,omitempty"`
}

// Configuration of conversation
type DingTalkConversation struct {
	ChatID string `json:"chatid"`
}

// DingTalkReceiverSpec defines the desired state of DingTalkReceiver
type DingTalkReceiverSpec struct {
	// DingTalkConfig to be selected for this receiver
	DingTalkConfigSelector *metav1.LabelSelector `json:"dingtalkConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// Be careful, a ChatBot only can send 20 message per minute.
	ChatBot *DingTalkChatBot `json:"chatbot,omitempty"`
	// The conversation which message will send to.
	Conversation *DingTalkConversation `json:"conversation,omitempty"`
}

// DingTalkReceiverStatus defines the observed state of DingTalkReceiver
type DingTalkReceiverStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=dr
// +kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// DingTalkReceiver is the Schema for the dingtalkreceivers API
type DingTalkReceiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DingTalkReceiverSpec   `json:"spec,omitempty"`
	Status DingTalkReceiverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DingTalkReceiverList contains a list of DingTalkReceiver
type DingTalkReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DingTalkReceiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DingTalkReceiver{}, &DingTalkReceiverList{})
}
