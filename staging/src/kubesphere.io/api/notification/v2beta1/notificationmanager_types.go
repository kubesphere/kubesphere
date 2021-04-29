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
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretKeySelector selects a key of a Secret.
type SecretKeySelector struct {
	// The namespace of the secret, default to the pod's namespace.
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Name of the secret.
	// +optional
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// The key of the secret to select from.  Must be a valid secret key.
	Key string `json:"key" protobuf:"bytes,2,opt,name=key"`
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
	// Pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run Notification Manager Pods.
	// ServiceAccount 'default' in notification manager's namespace will be used if not specified.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Port name used for the pods and service, defaults to webhook
	PortName string `json:"portName,omitempty"`
	// Default Email/Wechat/Slack/Webhook Config to be selected
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
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will remain unchanged. The $(VAR_NAME) syntax
	// can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded,
	// regardless of whether the variable exists or not.
	// Cannot be updated.
	// +optional
	Args []string `json:"args,omitempty"`
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
	// Template file path, must be a absolute path.
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
	// Type of sending email, bulk or single
	DeliveryType string `json:"deliveryType,omitempty"`
	// The maximum size of receivers in one email.
	MaxEmailReceivers int `json:"maxEmailReceivers,omitempty"`
	// The name of the template to generate email message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
	// The name of the template to generate email subject
	SubjectTemplate string `json:"subjectTemplate,omitempty"`
}

type WechatOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate wechat message.
	Template string `json:"template,omitempty"`
	// The maximum message size that can be sent in a request.
	MessageMaxSize int `json:"messageMaxSize,omitempty"`
	// The time of token expired.
	TokenExpires time.Duration `json:"tokenExpires,omitempty"`
}

type SlackOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate slack message.
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

// The config of flow control.
type Throttle struct {
	// The maximum calls in `Unit`.
	Threshold int           `json:"threshold,omitempty"`
	Unit      time.Duration `json:"unit,omitempty"`
	// The maximum tolerable waiting time when the calls trigger flow control, if the actual waiting time is more than this time, it will
	// return a error, else it will wait for the flow restriction lifted, and send the message.
	// Nil means do not wait, the maximum value is `Unit`.
	MaxWaitTime time.Duration `json:"maxWaitTime,omitempty"`
}

type DingTalkOptions struct {
	// Notification Sending Timeout
	NotificationTimeout *int32 `json:"notificationTimeout,omitempty"`
	// The name of the template to generate DingTalk message.
	// If the global template is not set, it will use default.
	Template string `json:"template,omitempty"`
	// The time of token expired.
	TokenExpires time.Duration `json:"tokenExpires,omitempty"`
	// The maximum message size that can be sent to conversation in a request.
	ConversationMessageMaxSize int `json:"conversationMessageMaxSize,omitempty"`
	// The maximum message size that can be sent to chatbot in a request.
	ChatbotMessageMaxSize int `json:"chatbotMessageMaxSize,omitempty"`
	// The flow control fo chatbot.
	ChatBotThrottle *Throttle `json:"chatBotThrottle,omitempty"`
	// The flow control fo conversation.
	ConversationThrottle *Throttle `json:"conversationThrottle,omitempty"`
}

type Options struct {
	Global   *GlobalOptions   `json:"global,omitempty"`
	Email    *EmailOptions    `json:"email,omitempty"`
	Wechat   *WechatOptions   `json:"wechat,omitempty"`
	Slack    *SlackOptions    `json:"slack,omitempty"`
	Webhook  *WebhookOptions  `json:"webhook,omitempty"`
	DingTalk *DingTalkOptions `json:"dingtalk,omitempty"`
}

// NotificationManagerStatus defines the observed state of NotificationManager
type NotificationManagerStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=nm,categories=notification-manager
// +kubebuilder:subresource:status

// NotificationManager is the Schema for the notificationmanagers API
type NotificationManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotificationManagerSpec   `json:"spec,omitempty"`
	Status NotificationManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NotificationManagerList contains a list of NotificationManager
type NotificationManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NotificationManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NotificationManager{}, &NotificationManagerList{})
}
