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
	"kubesphere.io/api/notification/v2beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this Config to the Hub version (v2beta1).
func (src *Config) ConvertTo(dstRaw conversion.Hub) error {

	dst := dstRaw.(*v2beta1.Config)
	dst.ObjectMeta = src.ObjectMeta

	if err := src.convertDingTalkTo(dst); err != nil {
		return err
	}

	if err := src.convertEmailTo(dst); err != nil {
		return err
	}

	if err := src.convertSlackTo(dst); err != nil {
		return err
	}

	if err := src.convertWebhookTo(dst); err != nil {
		return err
	}

	if err := src.convertWechatTo(dst); err != nil {
		return err
	}

	return nil
}

// ConvertFrom converts from the Hub version (v2beta1) to this version.
func (dst *Config) ConvertFrom(srcRaw conversion.Hub) error {

	src := srcRaw.(*v2beta1.Config)
	dst.ObjectMeta = src.ObjectMeta

	if err := dst.convertDingTalkFrom(src); err != nil {
		return err
	}

	if err := dst.convertEmailFrom(src); err != nil {
		return err
	}

	if err := dst.convertSlackFrom(src); err != nil {
		return err
	}

	if err := dst.convertWebhookFrom(src); err != nil {
		return err
	}

	if err := dst.convertWechatFrom(src); err != nil {
		return err
	}

	return nil
}

func (src *Config) convertDingTalkTo(dst *v2beta1.Config) error {

	if src.Spec.DingTalk == nil {
		return nil
	}

	dingtalk := src.Spec.DingTalk
	dst.Spec.DingTalk = &v2beta1.DingTalkConfig{
		Labels: dingtalk.Labels,
	}

	if dingtalk.Conversation != nil {
		dst.Spec.DingTalk.Conversation = &v2beta1.DingTalkApplicationConfig{
			AppKey:    credentialToSecretKeySelector(dingtalk.Conversation.AppKey),
			AppSecret: credentialToSecretKeySelector(dingtalk.Conversation.AppSecret),
		}
	}

	return nil
}

func (src *Config) convertEmailTo(dst *v2beta1.Config) error {

	if src.Spec.Email == nil {
		return nil
	}

	email := src.Spec.Email
	dst.Spec.Email = &v2beta1.EmailConfig{
		Labels: email.Labels,
		From:   email.From,
		SmartHost: v2beta1.HostPort{
			Host: email.SmartHost.Host,
			Port: email.SmartHost.Port,
		},
		Hello:        email.Hello,
		AuthUsername: email.AuthUsername,
		AuthPassword: credentialToSecretKeySelector(email.AuthPassword),
		AuthSecret:   credentialToSecretKeySelector(email.AuthSecret),
		AuthIdentify: email.AuthIdentify,
		RequireTLS:   email.RequireTLS,
		TLS:          convertTLSConfigTo(email.TLS),
	}

	return nil
}

func (src *Config) convertSlackTo(dst *v2beta1.Config) error {

	if src.Spec.Slack == nil {
		return nil
	}

	slack := src.Spec.Slack
	dst.Spec.Slack = &v2beta1.SlackConfig{
		Labels:           slack.Labels,
		SlackTokenSecret: credentialToSecretKeySelector(slack.SlackTokenSecret),
	}

	return nil
}

func (src *Config) convertWebhookTo(dst *v2beta1.Config) error {

	if src.Spec.Webhook == nil {
		return nil
	}

	dst.Spec.Webhook = &v2beta1.WebhookConfig{
		Labels: src.Spec.Webhook.Labels,
	}

	return nil
}

func (src *Config) convertWechatTo(dst *v2beta1.Config) error {

	if src.Spec.Wechat == nil {
		return nil
	}

	wechat := src.Spec.Wechat
	dst.Spec.Wechat = &v2beta1.WechatConfig{
		Labels:           wechat.Labels,
		WechatApiUrl:     wechat.WechatApiUrl,
		WechatApiCorpId:  wechat.WechatApiCorpId,
		WechatApiAgentId: wechat.WechatApiAgentId,
		WechatApiSecret:  credentialToSecretKeySelector(wechat.WechatApiSecret),
	}

	return nil
}

func (dst *Config) convertDingTalkFrom(src *v2beta1.Config) error {

	if src.Spec.DingTalk == nil {
		return nil
	}

	dingtalk := src.Spec.DingTalk
	dst.Spec.DingTalk = &DingTalkConfig{
		Labels: dingtalk.Labels,
	}

	if dingtalk.Conversation != nil {
		dst.Spec.DingTalk.Conversation = &DingTalkApplicationConfig{
			AppKey:    secretKeySelectorToCredential(dingtalk.Conversation.AppKey),
			AppSecret: secretKeySelectorToCredential(dingtalk.Conversation.AppSecret),
		}
	}

	return nil
}

func (dst *Config) convertEmailFrom(src *v2beta1.Config) error {

	if src.Spec.Email == nil {
		return nil
	}

	email := src.Spec.Email
	dst.Spec.Email = &EmailConfig{
		Labels: email.Labels,
		From:   email.From,
		SmartHost: HostPort{
			Host: email.SmartHost.Host,
			Port: email.SmartHost.Port,
		},
		Hello:        email.Hello,
		AuthUsername: email.AuthUsername,
		AuthPassword: secretKeySelectorToCredential(email.AuthPassword),
		AuthSecret:   secretKeySelectorToCredential(email.AuthSecret),
		AuthIdentify: email.AuthIdentify,
		RequireTLS:   email.RequireTLS,
		TLS:          convertTLSConfigFrom(email.TLS),
	}

	return nil
}

func (dst *Config) convertSlackFrom(src *v2beta1.Config) error {

	if src.Spec.Slack == nil {
		return nil
	}

	slack := src.Spec.Slack
	dst.Spec.Slack = &SlackConfig{
		Labels:           slack.Labels,
		SlackTokenSecret: secretKeySelectorToCredential(slack.SlackTokenSecret),
	}

	return nil
}

func (dst *Config) convertWebhookFrom(src *v2beta1.Config) error {

	if src.Spec.Webhook == nil {
		return nil
	}

	dst.Spec.Webhook = &WebhookConfig{
		Labels: src.Spec.Webhook.Labels,
	}

	return nil
}

func (dst *Config) convertWechatFrom(src *v2beta1.Config) error {

	if src.Spec.Wechat == nil {
		return nil
	}

	wechat := src.Spec.Wechat
	dst.Spec.Wechat = &WechatConfig{
		Labels:           wechat.Labels,
		WechatApiUrl:     wechat.WechatApiUrl,
		WechatApiCorpId:  wechat.WechatApiCorpId,
		WechatApiAgentId: wechat.WechatApiAgentId,
		WechatApiSecret:  secretKeySelectorToCredential(wechat.WechatApiSecret),
	}

	return nil
}

func convertTLSConfigTo(src *TLSConfig) *v2beta1.TLSConfig {

	if src == nil {
		return nil
	}

	dst := &v2beta1.TLSConfig{
		RootCA:             credentialToSecretKeySelector(src.RootCA),
		ServerName:         src.ServerName,
		InsecureSkipVerify: src.InsecureSkipVerify,
	}

	if src.ClientCertificate != nil {
		dst.ClientCertificate = &v2beta1.ClientCertificate{
			Cert: credentialToSecretKeySelector(src.Cert),
			Key:  credentialToSecretKeySelector(src.Key),
		}
	}

	return dst
}

func convertTLSConfigFrom(src *v2beta1.TLSConfig) *TLSConfig {

	if src == nil {
		return nil
	}

	dst := &TLSConfig{
		RootCA:             secretKeySelectorToCredential(src.RootCA),
		ServerName:         src.ServerName,
		InsecureSkipVerify: src.InsecureSkipVerify,
	}

	if src.ClientCertificate != nil {
		dst.ClientCertificate = &ClientCertificate{
			Cert: secretKeySelectorToCredential(src.Cert),
			Key:  secretKeySelectorToCredential(src.Key),
		}
	}

	return dst
}

func credentialToSecretKeySelector(src *Credential) *v2beta1.SecretKeySelector {

	if src == nil || src.ValueFrom == nil || src.ValueFrom.SecretKeyRef == nil {
		return nil
	}

	return &v2beta1.SecretKeySelector{
		Key:       src.ValueFrom.SecretKeyRef.Key,
		Name:      src.ValueFrom.SecretKeyRef.Name,
		Namespace: src.ValueFrom.SecretKeyRef.Namespace,
	}
}

func secretKeySelectorToCredential(selector *v2beta1.SecretKeySelector) *Credential {

	if selector == nil {
		return nil
	}

	return &Credential{
		ValueFrom: &ValueSource{
			SecretKeyRef: &SecretKeySelector{
				Key:       selector.Key,
				Name:      selector.Name,
				Namespace: selector.Namespace,
			},
		},
	}
}
