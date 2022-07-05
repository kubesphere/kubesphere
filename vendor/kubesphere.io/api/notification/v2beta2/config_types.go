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

// DingTalkApplicationConfig it th configuration of conversation
type DingTalkApplicationConfig struct {
	// The key of the application with which to send messages.
	AppKey *Credential `json:"appkey"`
	// The key in the secret to be used. Must be a valid secret key.
	AppSecret *Credential `json:"appsecret"`
}

type DingTalkConfig struct {
	Labels map[string]string `json:"labels,omitempty"`
	// Only needed when send alerts to the conversation.
	Conversation *DingTalkApplicationConfig `json:"conversation,omitempty"`
}

type ClientCertificate struct {
	// The client cert file for the targets.
	Cert *Credential `json:"cert"`
	// The client key file for the targets.
	Key *Credential `json:"key"`
}

// TLSConfig configures the options for TLS connections.
type TLSConfig struct {
	// RootCA defines the root certificate authorities
	// that clients use when verifying server certificates.
	RootCA *Credential `json:"rootCA,omitempty"`
	// The certificate of the client.
	*ClientCertificate `json:"clientCertificate,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `json:"serverName,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// BasicAuth contains basic HTTP authentication credentials.
type BasicAuth struct {
	Username string      `json:"username"`
	Password *Credential `json:"password,omitempty"`
}

// HTTPClientConfig configures an HTTP client.
type HTTPClientConfig struct {
	// The HTTP basic authentication credentials for the targets.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// The bearer token for the targets.
	BearerToken *Credential `json:"bearerToken,omitempty"`
	// HTTP proxy server to use to connect to the targets.
	ProxyURL string `json:"proxyUrl,omitempty"`
	// TLSConfig to use to connect to the targets.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
}

type HostPort struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type EmailConfig struct {
	Labels map[string]string `json:"labels,omitempty"`
	// The sender address.
	From string `json:"from"`
	// The address of the SMTP server.
	SmartHost HostPort `json:"smartHost"`
	// The hostname to use when identifying to the SMTP server.
	Hello *string `json:"hello,omitempty"`
	// The username for CRAM-MD5, LOGIN and PLAIN authentications.
	AuthUsername *string `json:"authUsername,omitempty"`
	// The identity for PLAIN authentication.
	AuthIdentify *string `json:"authIdentify,omitempty"`
	// The secret contains the SMTP password for LOGIN and PLAIN authentications.
	AuthPassword *Credential `json:"authPassword,omitempty"`
	// The secret contains the SMTP secret for CRAM-MD5 authentication.
	AuthSecret *Credential `json:"authSecret,omitempty"`
	// The default SMTP TLS requirement.
	RequireTLS *bool      `json:"requireTLS,omitempty"`
	TLS        *TLSConfig `json:"tls,omitempty"`
}

type SlackConfig struct {
	Labels map[string]string `json:"labels,omitempty"`
	// The token of user or bot.
	SlackTokenSecret *Credential `json:"slackTokenSecret"`
}

type WebhookConfig struct {
	Labels map[string]string `json:"labels,omitempty"`
}

type WechatConfig struct {
	Labels map[string]string `json:"labels,omitempty"`
	// The WeChat API URL.
	WechatApiUrl string `json:"wechatApiUrl,omitempty"`
	// The corp id for authentication.
	WechatApiCorpId string `json:"wechatApiCorpId"`
	// The id of the application which sending message.
	WechatApiAgentId string `json:"wechatApiAgentId"`
	// The API key to use when talking to the WeChat API.
	WechatApiSecret *Credential `json:"wechatApiSecret"`
}

// Sms Aliyun provider parameters
type AliyunSMS struct {
	SignName        string      `json:"signName"`
	TemplateCode    string      `json:"templateCode,omitempty"`
	AccessKeyId     *Credential `json:"accessKeyId"`
	AccessKeySecret *Credential `json:"accessKeySecret"`
}

// Sms tencent provider parameters
type TencentSMS struct {
	Sign        string      `json:"sign"`
	TemplateID  string      `json:"templateID"`
	SmsSdkAppid string      `json:"smsSdkAppid"`
	SecretId    *Credential `json:"secretId"`
	SecretKey   *Credential `json:"secretKey"`
}

// Sms huawei provider parameters
type HuaweiSMS struct {
	Url        string      `json:"url,omitempty"`
	Signature  string      `json:"signature"`
	TemplateId string      `json:"templateId"`
	Sender     string      `json:"sender"`
	AppSecret  *Credential `json:"appSecret"`
	AppKey     *Credential `json:"appKey"`
}

type Providers struct {
	Aliyun  *AliyunSMS  `json:"aliyun,omitempty"`
	Tencent *TencentSMS `json:"tencent,omitempty"`
	Huawei  *HuaweiSMS  `json:"huawei,omitempty"`
}

type SmsConfig struct {
	// The default sms provider, optional, use the first provider if not set
	DefaultProvider string `json:"defaultProvider,omitempty"`
	// All sms providers
	Providers *Providers `json:"providers"`
}

type PushoverConfig struct {
	Labels map[string]string `json:"labels,omitempty"`
	// The token of a pushover application.
	PushoverTokenSecret *Credential `json:"pushoverTokenSecret"`
}

//ConfigSpec defines the desired state of Config
type ConfigSpec struct {
	DingTalk *DingTalkConfig `json:"dingtalk,omitempty"`
	Email    *EmailConfig    `json:"email,omitempty"`
	Slack    *SlackConfig    `json:"slack,omitempty"`
	Webhook  *WebhookConfig  `json:"webhook,omitempty"`
	Wechat   *WechatConfig   `json:"wechat,omitempty"`
	Sms      *SmsConfig      `json:"sms,omitempty"`
	Pushover *PushoverConfig `json:"pushover,omitempty"`
}

// ConfigStatus defines the observed state of Config
type ConfigStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=nc,categories=notification-manager
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Config is the Schema for the dingtalkconfigs API
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}
