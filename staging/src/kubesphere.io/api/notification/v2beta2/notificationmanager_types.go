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
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Tenant = "tenant"
)

// SecretKeySelector selects a key of a Secret.
type SecretKeySelector struct {
	// The namespace of the secret, default to the `defaultSecretNamespace` of `NotificationManager` crd.
	// If the `defaultSecretNamespace` does not set, default to the pod's namespace.
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Name of the secret.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// The key of the secret to select from.  Must be a valid secret key.
	Key string `json:"key" protobuf:"bytes,2,opt,name=key"`
}

// ConfigmapKeySelector selects a key of a Configmap.
type ConfigmapKeySelector struct {
	// The namespace of the configmap, default to the `defaultSecretNamespace` of `NotificationManager` crd.
	// If the `defaultSecretNamespace` does not set, default to the pod's namespace.
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Name of the configmap.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// The key of the configmap to select from.  Must be a valid configmap key.
	Key string `json:"key,omitempty" protobuf:"bytes,2,opt,name=key"`
}

type ValueSource struct {
	// Selects a key of a secret in the pod's namespace
	// +optional
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty" protobuf:"bytes,4,opt,name=secretKeyRef"`
}

type Credential struct {
	// +optional
	Value     string       `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	ValueFrom *ValueSource `json:"valueFrom,omitempty" protobuf:"bytes,3,opt,name=valueFrom"`
}

// Sidecar defines a sidecar container which will be added to the notification manager deployment pod.
type Sidecar struct {
	// The type of sidecar, it can be specified to any value.
	// Notification manager built-in sidecar for KubeSphere,
	// It can be used with set `type` to `kubesphere`.
	Type string `json:"type" protobuf:"bytes,2,opt,name=type"`
	// Container of sidecar.
	*v1.Container `json:",inline"`
}

// HistoryReceiver used to collect notification history.
type HistoryReceiver struct {
	// Use a webhook to collect notification history, it will create a virtual receiver.
	Webhook *WebhookReceiver `json:"webhook"`
}

type Template struct {
	// Template file.
	Text *ConfigmapKeySelector `json:"text,omitempty"`
	// Time to reload template file.
	//
	// +kubebuilder:default="1m"
	ReloadCycle metav1.Duration `json:"reloadCycle,omitempty"`
	// Configmap which the i18n file be in.
	LanguagePack []*ConfigmapKeySelector `json:"languagePack,omitempty"`
	// The language used to send notification.
	//
	// +kubebuilder:default="English"
	Language string `json:"language,omitempty"`
}

// NotificationManagerSpec defines the desired state of NotificationManager
type NotificationManagerSpec struct {
	// Compute Resources required by container.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Docker Image used to start Notification Manager container,
	// for example kubesphere/notification-manager:v0.1.0
	Image *string `json:"image,omitempty"`
	// Image pull policy. One of Always, Never, IfNotPresent.
	// Defaults to IfNotPresent if not specified
	ImagePullPolicy *v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Number of instances to deploy for Notification Manager deployment.
	Replicas *int32 `json:"replicas,omitempty"`
	// Define which Nodes the Pods will be scheduled to.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Pod's scheduling constraints.
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Pod's toleration.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run Notification Manager Pods.
	// ServiceAccount 'default' in notification manager's namespace will be used if not specified.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Port name used for the pods and service, defaults to webhook
	PortName string `json:"portName,omitempty"`
	// Default Email/WeChat/Slack/Webhook Config to be selected
	DefaultConfigSelector *metav1.LabelSelector `json:"defaultConfigSelector,omitempty"`
	// Receivers to send notifications to
	Receivers *ReceiversSpec `json:"receivers"`
	// The default namespace to which notification manager secrets belong.
	DefaultSecretNamespace string `json:"defaultSecretNamespace,omitempty"`
	// List of volumes that can be mounted by containers belonging to the pod.
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// Pod volumes to mount into the container's filesystem.
	// Cannot be updated.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Arguments to the entrypoint.
	// The docker image's CMD is used if this is not provided.
	// +optional
	Args []string `json:"args,omitempty"`
	// Sidecar containers. The key is the type of sidecar, known value include: tenant.
	// Tenant sidecar used to manage the tenants which will receive notifications.
	// It needs to provide the API `/api/v2/tenant` at port `19094`, this api receives
	// a parameter `namespace` and return all tenants which need to receive notifications in this namespace.
	Sidecars map[string]*Sidecar `json:"sidecars,omitempty"`
	// History used to collect notification history.
	History *HistoryReceiver `json:"history,omitempty"`
	// Labels for grouping notifiations.
	GroupLabels []string `json:"groupLabels,omitempty"`
	// The maximum size of a batch. A batch used to buffer alerts and asynchronously process them.
	//
	// +kubebuilder:default=100
	BatchMaxSize int `json:"batchMaxSize,omitempty"`
	// The amount of time to wait before force processing the batch that hadn't reached the max size.
	//
	// +kubebuilder:default="1m"
	BatchMaxWait metav1.Duration `json:"batchMaxWait,omitempty"`
	// The RoutePolicy determines how to find receivers to which notifications will be sent.
	// Valid RoutePolicy include All, RouterFirst, and RouterOnly.
	// All: The alerts will be sent to the receivers that match any router,
	// and also will be sent to the receivers of those tenants with the right to access the namespace to which the alert belongs.
	// RouterFirst: The alerts will be sent to the receivers that match any router first.
	// If no receivers match any router, alerts will send to the receivers of those tenants with the right to access the namespace to which the alert belongs.
	// RouterOnly: The alerts will only be sent to the receivers that match any router.
	//
	// +kubebuilder:default=All
	RoutePolicy string `json:"routePolicy,omitempty"`
	// Template used to define information about templates
	Template *Template `json:"template,omitempty"`
}

type ReceiversSpec struct {
	// Key used to identify tenant, default to be "namespace" if not specified
	TenantKey string `json:"tenantKey"`
	// Selector to find global notification receivers
	// which will be used when tenant receivers cannot be found.
	// Only matchLabels expression is allowed.
	GlobalReceiverSelector *metav1.LabelSelector `json:"globalReceiverSelector"`
	// Selector to find tenant notification receivers.
	// Only matchLabels expression is allowed.
	TenantReceiverSelector *metav1.LabelSelector `json:"tenantReceiverSelector"`
	// Various receiver options
	Options *Options `json:"options,omitempty"`
}

type GlobalOptions struct {
	// Template file path, must be an absolute path.
	//
	// Deprecated
	TemplateFiles []string `json:"templateFile,omitempty"`
	// The name of the template to generate message.
	// If the receiver dose not setup template, it will use this.
	Template string `json:"template,omitempty"`
	// The name of the cluster in which the notification manager is deployed.
	Cluster string `json:"cluster,omitempty"`
}

type EmailOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// Deprecated
	DeliveryType string `json:"deliveryType,omitempty"`
	// The maximum size of receivers in one email.
	MaxEmailReceivers int `json:"maxEmailReceivers,omitempty"`
	// The name of the template to generate email message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
	// The name of the template to generate email subject
	SubjectTemplate string `json:"subjectTemplate,omitempty"`
	// template type: text or html, default type is html
	TmplType string `json:"tmplType,omitempty"`
}

type WechatOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate WeChat message.
	Template string `json:"template,omitempty"`
	// template type: text or markdown, default type is text
	TmplType string `json:"tmplType,omitempty"`
	// The maximum message size that can be sent in a request.
	MessageMaxSize int `json:"messageMaxSize,omitempty"`
	// The time of token expired.
	TokenExpires time.Duration `json:"tokenExpires,omitempty"`
}

type SlackOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate Slack message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
}

type WebhookOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate webhook message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
}

// Throttle is the config of flow control.
type Throttle struct {
	// The maximum calls in `Unit`.
	Threshold int           `json:"threshold,omitempty"`
	Unit      time.Duration `json:"unit,omitempty"`
	// The maximum tolerable waiting time when the calls trigger flow control, if the actual waiting time is more than this time, it will
	// return an error, else it will wait for the flow restriction lifted, and send the message.
	// Nil means do not wait, the maximum value is `Unit`.
	MaxWaitTime time.Duration `json:"maxWaitTime,omitempty"`
}

type DingTalkOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
	// The name of the template to generate markdown title
	TitleTemplate string `json:"titleTemplate,omitempty"`
	// template type: text or markdown, default type is text
	TmplType string `json:"tmplType,omitempty"`
	// The time of token expired.
	TokenExpires time.Duration `json:"tokenExpires,omitempty"`
	// The maximum message size that can be sent to conversation in a request.
	ConversationMessageMaxSize int `json:"conversationMessageMaxSize,omitempty"`
	// The maximum message size that can be sent to chatbot in a request.
	ChatbotMessageMaxSize int `json:"chatbotMessageMaxSize,omitempty"`
	// The flow control for chatbot.
	ChatBotThrottle *Throttle `json:"chatBotThrottle,omitempty"`
	// The flow control for conversation.
	ConversationThrottle *Throttle `json:"conversationThrottle,omitempty"`
}

type SmsOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate sms message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
}

type PushoverOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate pushover message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
	// The name of the template to generate message title
	TitleTemplate string `json:"titleTemplate,omitempty"`
}

type FeishuOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
	// template type: text or post, default type is post
	TmplType string `json:"tmplType,omitempty"`
	// The time of token expired.
	TokenExpires time.Duration `json:"tokenExpires,omitempty"`
}

type Options struct {
	Global   *GlobalOptions   `json:"global,omitempty"`
	Email    *EmailOptions    `json:"email,omitempty"`
	Wechat   *WechatOptions   `json:"wechat,omitempty"`
	Slack    *SlackOptions    `json:"slack,omitempty"`
	Webhook  *WebhookOptions  `json:"webhook,omitempty"`
	DingTalk *DingTalkOptions `json:"dingtalk,omitempty"`
	Sms      *SmsOptions      `json:"sms,omitempty"`
	Pushover *PushoverOptions `json:"pushover,omitempty"`
	Feishu   *FeishuOptions   `json:"feishu,omitempty"`
}

// NotificationManagerStatus defines the observed state of NotificationManager
type NotificationManagerStatus struct {
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// NotificationManager is the Schema for the notificationmanagers API
type NotificationManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotificationManagerSpec   `json:"spec,omitempty"`
	Status NotificationManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true

// NotificationManagerList contains a list of NotificationManager
type NotificationManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotificationManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NotificationManager{}, &NotificationManagerList{})
}
