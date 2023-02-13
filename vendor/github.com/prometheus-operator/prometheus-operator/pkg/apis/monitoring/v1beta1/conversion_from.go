// Copyright 2022 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta1

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
)

func convertRouteFrom(in *v1alpha1.Route) (*Route, error) {
	if in == nil {
		return nil, nil
	}

	out := &Route{
		Receiver:            in.Receiver,
		GroupBy:             in.GroupBy,
		GroupWait:           in.GroupWait,
		GroupInterval:       in.GroupInterval,
		RepeatInterval:      in.RepeatInterval,
		Matchers:            convertMatchersFrom(in.Matchers),
		MuteTimeIntervals:   in.MuteTimeIntervals,
		ActiveTimeIntervals: in.ActiveTimeIntervals,
	}

	// Deserialize child routes to convert them to v1alpha1 and serialize back.
	crs, err := in.ChildRoutes()
	if err != nil {
		return nil, err
	}

	out.Routes = make([]apiextensionsv1.JSON, 0, len(in.Routes))
	for i := range crs {
		cr, err := convertRouteFrom(&crs[i])
		if err != nil {
			return nil, fmt.Errorf("route[%d]: %w", i, err)
		}

		b, err := json.Marshal(cr)
		if err != nil {
			return nil, fmt.Errorf("route[%d]: %w", i, err)
		}

		out.Routes = append(out.Routes, apiextensionsv1.JSON{Raw: b})
	}

	return out, nil
}

func convertMatchersFrom(in []v1alpha1.Matcher) []Matcher {
	out := make([]Matcher, 0, len(in))

	for _, m := range in {
		mt := m.MatchType
		if mt == "" {
			mt = "="
			if m.Regex {
				mt = "=~"
			}
		}
		out = append(
			out,
			Matcher{
				Name:      m.Name,
				Value:     m.Value,
				MatchType: MatchType(mt),
			},
		)
	}

	return out
}

func convertTimeIntervalsFrom(in []v1alpha1.TimeInterval) []TimePeriod {
	out := make([]TimePeriod, 0, len(in))

	for _, ti := range in {
		var (
			trs  = make([]TimeRange, 0, len(ti.Times))
			wds  = make([]WeekdayRange, 0, len(ti.Weekdays))
			doms = make([]DayOfMonthRange, 0, len(ti.DaysOfMonth))
			mrs  = make([]MonthRange, 0, len(ti.Months))
			yrs  = make([]YearRange, 0, len(ti.Years))
		)

		for _, tr := range ti.Times {
			trs = append(trs, TimeRange{StartTime: Time(tr.StartTime), EndTime: Time(tr.EndTime)})
		}

		for _, wd := range ti.Weekdays {
			wds = append(wds, WeekdayRange(wd))
		}

		for _, dm := range ti.DaysOfMonth {
			doms = append(doms, DayOfMonthRange{Start: dm.Start, End: dm.End})
		}

		for _, mr := range ti.Months {
			mrs = append(mrs, MonthRange(mr))
		}

		for _, yr := range ti.Years {
			yrs = append(yrs, YearRange(yr))
		}

		out = append(
			out,
			TimePeriod{
				Times:       trs,
				Weekdays:    wds,
				DaysOfMonth: doms,
				Months:      mrs,
				Years:       yrs,
			},
		)
	}

	return out
}

func convertHTTPConfigFrom(in *v1alpha1.HTTPConfig) *HTTPConfig {
	if in == nil {
		return nil
	}

	return &HTTPConfig{
		Authorization:     in.Authorization,
		BasicAuth:         in.BasicAuth,
		OAuth2:            in.OAuth2,
		BearerTokenSecret: convertSecretKeySelectorFrom(in.BearerTokenSecret),
		TLSConfig:         in.TLSConfig,
		ProxyURL:          in.ProxyURL,
		FollowRedirects:   in.FollowRedirects,
	}
}

func convertKeyValuesFrom(in []v1alpha1.KeyValue) []KeyValue {
	out := make([]KeyValue, len(in))

	for i := range in {
		out[i] = KeyValue{
			Key:   in[i].Key,
			Value: in[i].Value,
		}
	}

	return out
}

func convertSecretKeySelectorFrom(in *v1.SecretKeySelector) *SecretKeySelector {
	if in == nil {
		return nil
	}

	return &SecretKeySelector{
		Name: in.Name,
		Key:  in.Key,
	}
}

func convertOpsGenieConfigRespondersFrom(in []v1alpha1.OpsGenieConfigResponder) []OpsGenieConfigResponder {
	out := make([]OpsGenieConfigResponder, len(in))

	for i := range in {
		out[i] = OpsGenieConfigResponder{
			ID:       in[i].ID,
			Name:     in[i].Name,
			Username: in[i].Username,
			Type:     in[i].Type,
		}
	}

	return out
}

func convertOpsGenieConfigFrom(in v1alpha1.OpsGenieConfig) OpsGenieConfig {
	return OpsGenieConfig{
		SendResolved: in.SendResolved,
		APIKey:       convertSecretKeySelectorFrom(in.APIKey),
		APIURL:       in.APIURL,
		Message:      in.Message,
		Description:  in.Description,
		Source:       in.Source,
		Tags:         in.Tags,
		Note:         in.Note,
		Priority:     in.Priority,
		Details:      convertKeyValuesFrom(in.Details),
		Responders:   convertOpsGenieConfigRespondersFrom(in.Responders),
		HTTPConfig:   convertHTTPConfigFrom(in.HTTPConfig),
		Entity:       in.Entity,
		Actions:      in.Actions,
	}
}

func convertPagerDutyImageConfigsFrom(in []v1alpha1.PagerDutyImageConfig) []PagerDutyImageConfig {
	out := make([]PagerDutyImageConfig, len(in))

	for i := range in {
		out[i] = PagerDutyImageConfig{
			Src:  in[i].Src,
			Href: in[i].Href,
			Alt:  in[i].Alt,
		}
	}

	return out
}

func convertPagerDutyLinkConfigsFrom(in []v1alpha1.PagerDutyLinkConfig) []PagerDutyLinkConfig {
	out := make([]PagerDutyLinkConfig, len(in))

	for i := range in {
		out[i] = PagerDutyLinkConfig{
			Href: in[i].Href,
			Text: in[i].Text,
		}
	}

	return out
}

func convertPagerDutyConfigFrom(in v1alpha1.PagerDutyConfig) PagerDutyConfig {
	return PagerDutyConfig{
		SendResolved:          in.SendResolved,
		RoutingKey:            convertSecretKeySelectorFrom(in.RoutingKey),
		ServiceKey:            convertSecretKeySelectorFrom(in.ServiceKey),
		URL:                   in.URL,
		Client:                in.Client,
		ClientURL:             in.ClientURL,
		Description:           in.Description,
		Severity:              in.Severity,
		Class:                 in.Class,
		Group:                 in.Group,
		Component:             in.Component,
		Details:               convertKeyValuesFrom(in.Details),
		PagerDutyImageConfigs: convertPagerDutyImageConfigsFrom(in.PagerDutyImageConfigs),
		PagerDutyLinkConfigs:  convertPagerDutyLinkConfigsFrom(in.PagerDutyLinkConfigs),
		HTTPConfig:            convertHTTPConfigFrom(in.HTTPConfig),
	}
}

func convertSlackFieldsFrom(in []v1alpha1.SlackField) []SlackField {
	out := make([]SlackField, len(in))

	for i := range in {
		out[i] = SlackField{
			Title: in[i].Title,
			Value: in[i].Value,
			Short: in[i].Short,
		}
	}

	return out
}

func convertSlackActionsFrom(in []v1alpha1.SlackAction) []SlackAction {
	out := make([]SlackAction, len(in))

	for i := range in {
		out[i] = SlackAction{
			Type:  in[i].Type,
			Text:  in[i].Text,
			URL:   in[i].URL,
			Style: in[i].Style,
			Name:  in[i].Name,
			Value: in[i].Value,
		}
		if in[i].ConfirmField != nil {
			out[i].ConfirmField = &SlackConfirmationField{
				Text:        in[i].ConfirmField.Text,
				Title:       in[i].ConfirmField.Title,
				OkText:      in[i].ConfirmField.OkText,
				DismissText: in[i].ConfirmField.DismissText,
			}
		}
	}

	return out
}

func convertSlackConfigFrom(in v1alpha1.SlackConfig) SlackConfig {
	return SlackConfig{
		SendResolved: in.SendResolved,
		APIURL:       convertSecretKeySelectorFrom(in.APIURL),
		Channel:      in.Channel,
		Username:     in.Username,
		Color:        in.Color,
		Title:        in.Title,
		TitleLink:    in.TitleLink,
		Pretext:      in.Pretext,
		Text:         in.Text,
		Fields:       convertSlackFieldsFrom(in.Fields),
		ShortFields:  in.ShortFields,
		Footer:       in.Footer,
		Fallback:     in.Fallback,
		CallbackID:   in.CallbackID,
		IconEmoji:    in.IconEmoji,
		IconURL:      in.IconURL,
		ImageURL:     in.ImageURL,
		ThumbURL:     in.ThumbURL,
		LinkNames:    in.LinkNames,
		MrkdwnIn:     in.MrkdwnIn,
		Actions:      convertSlackActionsFrom(in.Actions),
		HTTPConfig:   convertHTTPConfigFrom(in.HTTPConfig),
	}
}

func convertWebhookConfigFrom(in v1alpha1.WebhookConfig) WebhookConfig {
	return WebhookConfig{
		SendResolved: in.SendResolved,
		URL:          in.URL,
		URLSecret:    convertSecretKeySelectorFrom(in.URLSecret),
		HTTPConfig:   convertHTTPConfigFrom(in.HTTPConfig),
		MaxAlerts:    in.MaxAlerts,
	}
}

func convertWeChatConfigFrom(in v1alpha1.WeChatConfig) WeChatConfig {
	return WeChatConfig{
		SendResolved: in.SendResolved,
		APISecret:    convertSecretKeySelectorFrom(in.APISecret),
		APIURL:       in.APIURL,
		CorpID:       in.CorpID,
		AgentID:      in.AgentID,
		ToUser:       in.ToUser,
		ToParty:      in.ToParty,
		ToTag:        in.ToTag,
		Message:      in.Message,
		MessageType:  in.MessageType,
		HTTPConfig:   convertHTTPConfigFrom(in.HTTPConfig),
	}
}

func convertEmailConfigFrom(in v1alpha1.EmailConfig) EmailConfig {
	return EmailConfig{
		SendResolved: in.SendResolved,
		To:           in.To,
		From:         in.From,
		Hello:        in.Hello,
		Smarthost:    in.Smarthost,
		AuthUsername: in.AuthUsername,
		AuthPassword: convertSecretKeySelectorFrom(in.AuthPassword),
		AuthSecret:   convertSecretKeySelectorFrom(in.AuthSecret),
		AuthIdentity: in.AuthIdentity,
		Headers:      convertKeyValuesFrom(in.Headers),
		HTML:         in.HTML,
		Text:         in.Text,
		RequireTLS:   in.RequireTLS,
		TLSConfig:    in.TLSConfig,
	}
}

func convertVictorOpsConfigFrom(in v1alpha1.VictorOpsConfig) VictorOpsConfig {
	return VictorOpsConfig{
		SendResolved:      in.SendResolved,
		APIKey:            convertSecretKeySelectorFrom(in.APIKey),
		APIURL:            in.APIURL,
		RoutingKey:        in.RoutingKey,
		MessageType:       in.MessageType,
		EntityDisplayName: in.EntityDisplayName,
		StateMessage:      in.StateMessage,
		MonitoringTool:    in.MonitoringTool,
		CustomFields:      convertKeyValuesFrom(in.CustomFields),
		HTTPConfig:        convertHTTPConfigFrom(in.HTTPConfig),
	}
}

func convertPushoverConfigFrom(in v1alpha1.PushoverConfig) PushoverConfig {
	return PushoverConfig{
		SendResolved: in.SendResolved,
		UserKey:      convertSecretKeySelectorFrom(in.UserKey),
		Token:        convertSecretKeySelectorFrom(in.Token),
		Title:        in.Title,
		Message:      in.Message,
		URL:          in.URL,
		URLTitle:     in.URLTitle,
		Sound:        in.Sound,
		Priority:     in.Priority,
		Retry:        in.Retry,
		Expire:       in.Expire,
		HTML:         in.HTML,
		HTTPConfig:   convertHTTPConfigFrom(in.HTTPConfig),
	}
}

func convertSNSConfigFrom(in v1alpha1.SNSConfig) SNSConfig {
	return SNSConfig{
		SendResolved: in.SendResolved,
		ApiURL:       in.ApiURL,
		Sigv4:        in.Sigv4,
		TopicARN:     in.TopicARN,
		Subject:      in.Subject,
		PhoneNumber:  in.PhoneNumber,
		TargetARN:    in.TargetARN,
		Message:      in.Message,
		Attributes:   in.Attributes,
		HTTPConfig:   convertHTTPConfigFrom(in.HTTPConfig),
	}
}

func convertTelegramConfigFrom(in v1alpha1.TelegramConfig) TelegramConfig {
	return TelegramConfig{
		SendResolved:         in.SendResolved,
		APIURL:               in.APIURL,
		BotToken:             convertSecretKeySelectorFrom(in.BotToken),
		ChatID:               in.ChatID,
		Message:              in.Message,
		DisableNotifications: in.DisableNotifications,
		ParseMode:            in.ParseMode,
		HTTPConfig:           convertHTTPConfigFrom(in.HTTPConfig),
	}
}

// ConvertFrom converts from the Hub version (v1alpha1) to this version (v1beta1).
func (dst *AlertmanagerConfig) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.AlertmanagerConfig)

	dst.ObjectMeta = src.ObjectMeta

	for _, in := range src.Spec.Receivers {
		out := Receiver{
			Name: in.Name,
		}

		for _, in := range in.OpsGenieConfigs {
			out.OpsGenieConfigs = append(
				out.OpsGenieConfigs,
				convertOpsGenieConfigFrom(in),
			)
		}

		for _, in := range in.PagerDutyConfigs {
			out.PagerDutyConfigs = append(
				out.PagerDutyConfigs,
				convertPagerDutyConfigFrom(in),
			)
		}

		for _, in := range in.SlackConfigs {
			out.SlackConfigs = append(
				out.SlackConfigs,
				convertSlackConfigFrom(in),
			)
		}

		for _, in := range in.WebhookConfigs {
			out.WebhookConfigs = append(
				out.WebhookConfigs,
				convertWebhookConfigFrom(in),
			)
		}

		for _, in := range in.WeChatConfigs {
			out.WeChatConfigs = append(
				out.WeChatConfigs,
				convertWeChatConfigFrom(in),
			)
		}

		for _, in := range in.EmailConfigs {
			out.EmailConfigs = append(
				out.EmailConfigs,
				convertEmailConfigFrom(in),
			)
		}

		for _, in := range in.VictorOpsConfigs {
			out.VictorOpsConfigs = append(
				out.VictorOpsConfigs,
				convertVictorOpsConfigFrom(in),
			)
		}

		for _, in := range in.PushoverConfigs {
			out.PushoverConfigs = append(
				out.PushoverConfigs,
				convertPushoverConfigFrom(in),
			)
		}

		for _, in := range in.SNSConfigs {
			out.SNSConfigs = append(
				out.SNSConfigs,
				convertSNSConfigFrom(in),
			)
		}

		for _, in := range in.TelegramConfigs {
			out.TelegramConfigs = append(
				out.TelegramConfigs,
				convertTelegramConfigFrom(in),
			)
		}

		dst.Spec.Receivers = append(dst.Spec.Receivers, out)
	}

	for _, in := range src.Spec.InhibitRules {
		dst.Spec.InhibitRules = append(
			dst.Spec.InhibitRules,
			InhibitRule{
				TargetMatch: convertMatchersFrom(in.TargetMatch),
				SourceMatch: convertMatchersFrom(in.SourceMatch),
				Equal:       in.Equal,
			},
		)
	}

	for _, in := range src.Spec.MuteTimeIntervals {
		dst.Spec.TimeIntervals = append(
			dst.Spec.TimeIntervals,
			TimeInterval{
				Name:          in.Name,
				TimeIntervals: convertTimeIntervalsFrom(in.TimeIntervals),
			},
		)
	}

	r, err := convertRouteFrom(src.Spec.Route)
	if err != nil {
		return err
	}
	dst.Spec.Route = r

	return nil
}
