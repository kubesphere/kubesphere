/*


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

package v2beta1

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
	ChatIDs []string `json:"chatids"`
}

type DingTalkReceiver struct {
	// whether the receiver is enabled
	Enabled *bool `json:"enabled,omitempty"`
	// DingTalkConfig to be selected for this receiver
	DingTalkConfigSelector *metav1.LabelSelector `json:"dingtalkConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// Be careful, a ChatBot only can send 20 message per minute.
	ChatBot *DingTalkChatBot `json:"chatbot,omitempty"`
	// The conversation which message will send to.
	Conversation *DingTalkConversation `json:"conversation,omitempty"`
}

type EmailReceiver struct {
	// whether the receiver is enabled
	Enabled *bool `json:"enabled,omitempty"`
	// Receivers' email addresses
	To []string `json:"to"`
	// EmailConfig to be selected for this receiver
	EmailConfigSelector *metav1.LabelSelector `json:"emailConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
}

type SlackReceiver struct {
	// whether the receiver is enabled
	Enabled *bool `json:"enabled,omitempty"`
	// SlackConfig to be selected for this receiver
	SlackConfigSelector *metav1.LabelSelector `json:"slackConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// The channel or user to send notifications to.
	Channels []string `json:"channels"`
}

// ServiceReference holds a reference to Service.legacy.k8s.io
type ServiceReference struct {
	// `namespace` is the namespace of the service.
	// Required
	Namespace string `json:"namespace"`

	// `name` is the name of the service.
	// Required
	Name string `json:"name"`

	// `path` is an optional URL path which will be sent in any request to
	// this service.
	// +optional
	Path *string `json:"path,omitempty"`

	// If specified, the port on the service that hosting webhook.
	// Default to 443 for backward compatibility.
	// `port` should be a valid port number (1-65535, inclusive).
	// +optional
	Port *int32 `json:"port,omitempty"`

	// Http scheme, default is http.
	// +optional
	Scheme *string `json:"scheme,omitempty"`
}

type WebhookReceiver struct {
	// whether the receiver is enabled
	Enabled *bool `json:"enabled,omitempty"`
	// WebhookConfig to be selected for this receiver
	WebhookConfigSelector *metav1.LabelSelector `json:"webhookConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// `url` gives the location of the webhook, in standard URL form
	// (`scheme://host:port/path`). Exactly one of `url` or `service`
	// must be specified.
	//
	// The `host` should not refer to a service running in the cluster; use
	// the `service` field instead. The host might be resolved via external
	// DNS in some api servers (e.g., `kube-apiserver` cannot resolve
	// in-cluster DNS as that would be a layering violation). `host` may
	// also be an IP address.
	//
	// Please note that using `localhost` or `127.0.0.1` as a `host` is
	// risky unless you take great care to run this webhook on all hosts
	// which run an apiserver which might need to make calls to this
	// webhook. Such installs are likely to be non-portable, i.e., not easy
	// to turn up in a new cluster.
	//
	// A path is optional, and if present may be any string permissible in
	// a URL. You may use the path to pass an arbitrary string to the
	// webhook, for example, a cluster identifier.
	//
	// Attempting to use a user or basic auth e.g. "user:password@" is not
	// allowed. Fragments ("#...") and query parameters ("?...") are not
	// allowed, either.
	//
	// +optional
	URL *string `json:"url,omitempty"`

	// `service` is a reference to the service for this webhook. Either
	// `service` or `url` must be specified.
	//
	// If the webhook is running within the cluster, then you should use `service`.
	//
	// +optional
	Service *ServiceReference `json:"service,omitempty"`

	HTTPConfig *HTTPClientConfig `json:"httpConfig,omitempty"`
}

type WechatReceiver struct {
	// whether the receiver is enabled
	Enabled *bool `json:"enabled,omitempty"`
	// WechatConfig to be selected for this receiver
	WechatConfigSelector *metav1.LabelSelector `json:"wechatConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// +optional
	ToUser  []string `json:"toUser,omitempty"`
	ToParty []string `json:"toParty,omitempty"`
	ToTag   []string `json:"toTag,omitempty"`
}

//ReceiverSpec defines the desired state of Receiver
type ReceiverSpec struct {
	DingTalk *DingTalkReceiver `json:"dingtalk,omitempty"`
	Email    *EmailReceiver    `json:"email,omitempty"`
	Slack    *SlackReceiver    `json:"slack,omitempty"`
	Webhook  *WebhookReceiver  `json:"webhook,omitempty"`
	Wechat   *WechatReceiver   `json:"wechat,omitempty"`
}

// ReceiverStatus defines the observed state of Receiver
type ReceiverStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=nr,categories=notification-manager
// +kubebuilder:subresource:status
// +genclient
// +genclient:nonNamespaced
// Receiver is the Schema for the receivers API
type Receiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReceiverSpec   `json:"spec,omitempty"`
	Status ReceiverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReceiverList contains a list of Receiver
type ReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Receiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Receiver{}, &ReceiverList{})
}
