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
func (src *Receiver) ConvertTo(dstRaw conversion.Hub) error {

	dst := dstRaw.(*v2beta1.Receiver)
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
func (dst *Receiver) ConvertFrom(srcRaw conversion.Hub) error {

	src := srcRaw.(*v2beta1.Receiver)
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

func (src *Receiver) convertDingTalkTo(dst *v2beta1.Receiver) error {

	if src.Spec.DingTalk == nil {
		return nil
	}

	dingtalk := src.Spec.DingTalk
	dst.Spec.DingTalk = &v2beta1.DingTalkReceiver{
		Enabled:                dingtalk.Enabled,
		DingTalkConfigSelector: dingtalk.DingTalkConfigSelector,
		AlertSelector:          dingtalk.AlertSelector,
	}

	if dingtalk.Conversation != nil {
		dst.Spec.DingTalk.Conversation = &v2beta1.DingTalkConversation{
			ChatIDs: dingtalk.Conversation.ChatIDs,
		}
	}

	if dingtalk.ChatBot != nil {
		dst.Spec.DingTalk.ChatBot = &v2beta1.DingTalkChatBot{
			Webhook:  credentialToSecretKeySelector(dingtalk.ChatBot.Webhook),
			Keywords: dingtalk.ChatBot.Keywords,
			Secret:   credentialToSecretKeySelector(dingtalk.ChatBot.Secret),
		}
	}

	return nil
}

func (src *Receiver) convertEmailTo(dst *v2beta1.Receiver) error {

	if src.Spec.Email == nil {
		return nil
	}

	email := src.Spec.Email
	dst.Spec.Email = &v2beta1.EmailReceiver{
		Enabled:             email.Enabled,
		To:                  email.To,
		EmailConfigSelector: email.EmailConfigSelector,
		AlertSelector:       email.AlertSelector,
	}

	return nil
}

func (src *Receiver) convertSlackTo(dst *v2beta1.Receiver) error {

	if src.Spec.Slack == nil {
		return nil
	}

	slack := src.Spec.Slack
	dst.Spec.Slack = &v2beta1.SlackReceiver{
		Enabled:             slack.Enabled,
		SlackConfigSelector: slack.SlackConfigSelector,
		AlertSelector:       slack.AlertSelector,
		Channels:            slack.Channels,
	}

	return nil
}

func (src *Receiver) convertWebhookTo(dst *v2beta1.Receiver) error {

	if src.Spec.Webhook == nil {
		return nil
	}

	webhook := src.Spec.Webhook
	dst.Spec.Webhook = &v2beta1.WebhookReceiver{
		Enabled:               webhook.Enabled,
		WebhookConfigSelector: webhook.WebhookConfigSelector,
		AlertSelector:         webhook.AlertSelector,
		URL:                   webhook.URL,
	}

	if webhook.Service != nil {
		dst.Spec.Webhook.Service = &v2beta1.ServiceReference{
			Namespace: webhook.Service.Namespace,
			Name:      webhook.Service.Name,
			Path:      webhook.Service.Path,
			Port:      webhook.Service.Port,
			Scheme:    webhook.Service.Scheme,
		}
	}

	if webhook.HTTPConfig != nil {
		dst.Spec.Webhook.HTTPConfig = &v2beta1.HTTPClientConfig{
			BearerToken: credentialToSecretKeySelector(webhook.HTTPConfig.BearerToken),
			ProxyURL:    webhook.HTTPConfig.ProxyURL,
			TLSConfig:   convertTLSConfigTo(webhook.HTTPConfig.TLSConfig),
		}

		if webhook.HTTPConfig.BasicAuth != nil {
			dst.Spec.Webhook.HTTPConfig.BasicAuth = &v2beta1.BasicAuth{
				Username: webhook.HTTPConfig.BasicAuth.Username,
				Password: credentialToSecretKeySelector(webhook.HTTPConfig.BasicAuth.Password),
			}
		}
	}

	return nil
}

func (src *Receiver) convertWechatTo(dst *v2beta1.Receiver) error {

	if src.Spec.Wechat == nil {
		return nil
	}

	wechat := src.Spec.Wechat
	dst.Spec.Wechat = &v2beta1.WechatReceiver{
		Enabled:              wechat.Enabled,
		WechatConfigSelector: wechat.WechatConfigSelector,
		AlertSelector:        wechat.AlertSelector,
		ToUser:               wechat.ToUser,
		ToParty:              wechat.ToParty,
		ToTag:                wechat.ToTag,
	}

	return nil
}

func (dst *Receiver) convertDingTalkFrom(src *v2beta1.Receiver) error {

	if src.Spec.DingTalk == nil {
		return nil
	}

	dingtalk := src.Spec.DingTalk
	dst.Spec.DingTalk = &DingTalkReceiver{
		Enabled:                dingtalk.Enabled,
		DingTalkConfigSelector: dingtalk.DingTalkConfigSelector,
		AlertSelector:          dingtalk.AlertSelector,
	}

	if dingtalk.Conversation != nil {
		dst.Spec.DingTalk.Conversation = &DingTalkConversation{
			ChatIDs: dingtalk.Conversation.ChatIDs,
		}
	}

	if dingtalk.ChatBot != nil {
		dst.Spec.DingTalk.ChatBot = &DingTalkChatBot{
			Webhook:  secretKeySelectorToCredential(dingtalk.ChatBot.Webhook),
			Keywords: dingtalk.ChatBot.Keywords,
			Secret:   secretKeySelectorToCredential(dingtalk.ChatBot.Secret),
		}
	}

	return nil
}

func (dst *Receiver) convertEmailFrom(src *v2beta1.Receiver) error {

	if src.Spec.Email == nil {
		return nil
	}

	email := src.Spec.Email
	dst.Spec.Email = &EmailReceiver{
		Enabled:             email.Enabled,
		To:                  email.To,
		EmailConfigSelector: email.EmailConfigSelector,
		AlertSelector:       email.AlertSelector,
	}

	return nil
}

func (dst *Receiver) convertSlackFrom(src *v2beta1.Receiver) error {

	if src.Spec.Slack == nil {
		return nil
	}

	slack := src.Spec.Slack
	dst.Spec.Slack = &SlackReceiver{
		Enabled:             slack.Enabled,
		SlackConfigSelector: slack.SlackConfigSelector,
		AlertSelector:       slack.AlertSelector,
		Channels:            slack.Channels,
	}

	return nil
}

func (dst *Receiver) convertWebhookFrom(src *v2beta1.Receiver) error {

	if src.Spec.Webhook == nil {
		return nil
	}

	webhook := src.Spec.Webhook
	dst.Spec.Webhook = &WebhookReceiver{
		Enabled:               webhook.Enabled,
		WebhookConfigSelector: webhook.WebhookConfigSelector,
		AlertSelector:         webhook.AlertSelector,
		URL:                   webhook.URL,
	}

	if webhook.Service != nil {
		dst.Spec.Webhook.Service = &ServiceReference{
			Namespace: webhook.Service.Namespace,
			Name:      webhook.Service.Name,
			Path:      webhook.Service.Path,
			Port:      webhook.Service.Port,
			Scheme:    webhook.Service.Scheme,
		}
	}

	if webhook.HTTPConfig != nil {
		dst.Spec.Webhook.HTTPConfig = &HTTPClientConfig{
			BearerToken: secretKeySelectorToCredential(webhook.HTTPConfig.BearerToken),
			ProxyURL:    webhook.HTTPConfig.ProxyURL,
			TLSConfig:   convertTLSConfigFrom(webhook.HTTPConfig.TLSConfig),
		}

		if webhook.HTTPConfig.BasicAuth != nil {
			dst.Spec.Webhook.HTTPConfig.BasicAuth = &BasicAuth{
				Username: webhook.HTTPConfig.BasicAuth.Username,
				Password: secretKeySelectorToCredential(webhook.HTTPConfig.BasicAuth.Password),
			}
		}
	}

	return nil
}

func (dst *Receiver) convertWechatFrom(src *v2beta1.Receiver) error {

	if src.Spec.Wechat == nil {
		return nil
	}

	wechat := src.Spec.Wechat
	dst.Spec.Wechat = &WechatReceiver{
		Enabled:              wechat.Enabled,
		WechatConfigSelector: wechat.WechatConfigSelector,
		AlertSelector:        wechat.AlertSelector,
		ToUser:               wechat.ToUser,
		ToParty:              wechat.ToParty,
		ToTag:                wechat.ToTag,
	}

	return nil
}
