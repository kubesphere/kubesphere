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

package v2beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DingTalkChatBot is the configuration of ChatBot
type DingTalkChatBot struct {
	// The webhook of ChatBot which the message will send to.
	Webhook *Credential `json:"webhook"`

	// Custom keywords of ChatBot
	Keywords []string `json:"keywords,omitempty"`

	// Secret of ChatBot, you can get it after enabled Additional Signature of ChatBot.
	Secret *Credential `json:"secret,omitempty"`
	// The phone numbers of the users which will be @.
	AtMobiles []string `json:"atMobiles,omitempty"`
	// The users who will be @.
	AtUsers []string `json:"atUsers,omitempty"`
	// Whether @everyone.
	AtAll bool `json:"atAll,omitempty"`
}

// DingTalkConversation of conversation
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
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template *string `json:"template,omitempty"`
	// The name of the template to generate markdown title
	TitleTemplate *string `json:"titleTemplate,omitempty"`
	// template type: text or markdown
	TmplType *string `json:"tmplType,omitempty"`
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
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template *string `json:"template,omitempty"`
	// The name of the template to generate email subject
	SubjectTemplate *string `json:"subjectTemplate,omitempty"`
	// template type: text or html, default type is html
	TmplType *string `json:"tmplType,omitempty"`
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
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template *string `json:"template,omitempty"`
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
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template *string `json:"template,omitempty"`
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
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template *string `json:"template,omitempty"`
	// template type: text or markdown, default type is text
	TmplType *string `json:"tmplType,omitempty"`
}

type SmsReceiver struct {
	// whether the receiver is enabled
	Enabled *bool `json:"enabled,omitempty"`
	// SmsConfig to be selected for this receiver
	SmsConfigSelector *metav1.LabelSelector `json:"smsConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// Receivers' phone numbers
	PhoneNumbers []string `json:"phoneNumbers"`
	// The name of the template to generate Sms message.
	// If the global template is not set, it will use default.
	Template *string `json:"template,omitempty"`
}

// PushoverUserProfile includes userKey and other preferences
type PushoverUserProfile struct {
	// UserKey is the user (Pushover User Key) to send notifications to.
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9]{30}$`
	UserKey *string `json:"userKey"`
	// Devices refers to device name to send the message directly to that device, rather than all of the user's devices
	Devices []string `json:"devices,omitempty"`
	// Title refers to message's title, otherwise your app's name is used.
	Title *string `json:"title,omitempty"`
	// Sound refers to the name of one of the sounds (https://pushover.net/api#sounds) supported by device clients
	Sound *string `json:"sound,omitempty"`
}

type PushoverReceiver struct {
	// whether the receiver is enabled
	Enabled *bool `json:"enabled,omitempty"`
	// PushoverConfig to be selected for this receiver
	PushoverConfigSelector *metav1.LabelSelector `json:"pushoverConfigSelector,omitempty"`
	// Selector to filter alerts.
	AlertSelector *metav1.LabelSelector `json:"alertSelector,omitempty"`
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template *string `json:"template,omitempty"`
	// The users profile.
	Profiles []*PushoverUserProfile `json:"profiles"`
}

//ReceiverSpec defines the desired state of Receiver
type ReceiverSpec struct {
	DingTalk *DingTalkReceiver `json:"dingtalk,omitempty"`
	Email    *EmailReceiver    `json:"email,omitempty"`
	Slack    *SlackReceiver    `json:"slack,omitempty"`
	Webhook  *WebhookReceiver  `json:"webhook,omitempty"`
	Wechat   *WechatReceiver   `json:"wechat,omitempty"`
	Sms      *SmsReceiver      `json:"sms,omitempty"`
	Pushover *PushoverReceiver `json:"pushover,omitempty"`
}

// ReceiverStatus defines the observed state of Receiver
type ReceiverStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=nr,categories=notification-manager
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

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
