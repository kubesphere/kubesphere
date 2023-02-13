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

func convertRouteTo(in *Route) (*v1alpha1.Route, error) {
	if in == nil {
		return nil, nil
	}

	out := &v1alpha1.Route{
		Receiver:            in.Receiver,
		GroupBy:             in.GroupBy,
		GroupWait:           in.GroupWait,
		GroupInterval:       in.GroupInterval,
		RepeatInterval:      in.RepeatInterval,
		Matchers:            convertMatchersTo(in.Matchers),
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
		cr, err := convertRouteTo(&crs[i])
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

func convertMatchersTo(in []Matcher) []v1alpha1.Matcher {
	out := make([]v1alpha1.Matcher, 0, len(in))

	for _, m := range in {
		out = append(
			out,
			v1alpha1.Matcher{
				Name:      m.Name,
				Value:     m.Value,
				MatchType: v1alpha1.MatchType(m.MatchType),
			},
		)
	}

	return out
}

func convertTimeIntervalsTo(in []TimePeriod) []v1alpha1.TimeInterval {
	out := make([]v1alpha1.TimeInterval, 0, len(in))

	for _, ti := range in {
		var (
			trs  = make([]v1alpha1.TimeRange, 0, len(ti.Times))
			wds  = make([]v1alpha1.WeekdayRange, 0, len(ti.Weekdays))
			doms = make([]v1alpha1.DayOfMonthRange, 0, len(ti.DaysOfMonth))
			mrs  = make([]v1alpha1.MonthRange, 0, len(ti.Months))
			yrs  = make([]v1alpha1.YearRange, 0, len(ti.Years))
		)

		for _, tr := range ti.Times {
			trs = append(trs, v1alpha1.TimeRange{StartTime: v1alpha1.Time(tr.StartTime), EndTime: v1alpha1.Time(tr.EndTime)})
		}

		for _, wd := range ti.Weekdays {
			wds = append(wds, v1alpha1.WeekdayRange(wd))
		}

		for _, dm := range ti.DaysOfMonth {
			doms = append(doms, v1alpha1.DayOfMonthRange{Start: dm.Start, End: dm.End})
		}

		for _, mr := range ti.Months {
			mrs = append(mrs, v1alpha1.MonthRange(mr))
		}

		for _, yr := range ti.Years {
			yrs = append(yrs, v1alpha1.YearRange(yr))
		}

		out = append(
			out,
			v1alpha1.TimeInterval{
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

func convertHTTPConfigTo(in *HTTPConfig) *v1alpha1.HTTPConfig {
	if in == nil {
		return nil
	}

	return &v1alpha1.HTTPConfig{
		Authorization:     in.Authorization,
		BasicAuth:         in.BasicAuth,
		OAuth2:            in.OAuth2,
		BearerTokenSecret: convertSecretKeySelectorTo(in.BearerTokenSecret),
		TLSConfig:         in.TLSConfig,
		ProxyURL:          in.ProxyURL,
		FollowRedirects:   in.FollowRedirects,
	}
}

func convertKeyValuesTo(in []KeyValue) []v1alpha1.KeyValue {
	out := make([]v1alpha1.KeyValue, len(in))

	for i := range in {
		out[i] = v1alpha1.KeyValue{
			Key:   in[i].Key,
			Value: in[i].Value,
		}
	}

	return out

}

func convertSecretKeySelectorTo(in *SecretKeySelector) *v1.SecretKeySelector {
	if in == nil {
		return nil
	}

	return &v1.SecretKeySelector{
		LocalObjectReference: v1.LocalObjectReference{
			Name: in.Name,
		},
		Key: in.Key,
	}
}

func convertOpsGenieConfigRespondersTo(in []OpsGenieConfigResponder) []v1alpha1.OpsGenieConfigResponder {
	out := make([]v1alpha1.OpsGenieConfigResponder, len(in))

	for i := range in {
		out[i] = v1alpha1.OpsGenieConfigResponder{
			ID:       in[i].ID,
			Name:     in[i].Name,
			Username: in[i].Username,
			Type:     in[i].Type,
		}
	}

	return out
}

func convertOpsGenieConfigTo(in OpsGenieConfig) v1alpha1.OpsGenieConfig {
	return v1alpha1.OpsGenieConfig{
		SendResolved: in.SendResolved,
		APIKey:       convertSecretKeySelectorTo(in.APIKey),
		APIURL:       in.APIURL,
		Message:      in.Message,
		Description:  in.Description,
		Source:       in.Source,
		Tags:         in.Tags,
		Note:         in.Note,
		Priority:     in.Priority,
		Details:      convertKeyValuesTo(in.Details),
		Responders:   convertOpsGenieConfigRespondersTo(in.Responders),
		HTTPConfig:   convertHTTPConfigTo(in.HTTPConfig),
		Entity:       in.Entity,
		Actions:      in.Actions,
	}
}

func convertPagerDutyImageConfigsTo(in []PagerDutyImageConfig) []v1alpha1.PagerDutyImageConfig {
	out := make([]v1alpha1.PagerDutyImageConfig, len(in))

	for i := range in {
		out[i] = v1alpha1.PagerDutyImageConfig{
			Src:  in[i].Src,
			Href: in[i].Href,
			Alt:  in[i].Alt,
		}
	}

	return out
}

func convertPagerDutyLinkConfigsTo(in []PagerDutyLinkConfig) []v1alpha1.PagerDutyLinkConfig {
	out := make([]v1alpha1.PagerDutyLinkConfig, len(in))

	for i := range in {
		out[i] = v1alpha1.PagerDutyLinkConfig{
			Href: in[i].Href,
			Text: in[i].Text,
		}
	}

	return out
}

func convertPagerDutyConfigTo(in PagerDutyConfig) v1alpha1.PagerDutyConfig {
	return v1alpha1.PagerDutyConfig{
		SendResolved:          in.SendResolved,
		RoutingKey:            convertSecretKeySelectorTo(in.RoutingKey),
		ServiceKey:            convertSecretKeySelectorTo(in.ServiceKey),
		URL:                   in.URL,
		Client:                in.Client,
		ClientURL:             in.ClientURL,
		Description:           in.Description,
		Severity:              in.Severity,
		Class:                 in.Class,
		Group:                 in.Group,
		Component:             in.Component,
		Details:               convertKeyValuesTo(in.Details),
		PagerDutyImageConfigs: convertPagerDutyImageConfigsTo(in.PagerDutyImageConfigs),
		PagerDutyLinkConfigs:  convertPagerDutyLinkConfigsTo(in.PagerDutyLinkConfigs),
		HTTPConfig:            convertHTTPConfigTo(in.HTTPConfig),
	}
}

func convertSlackFieldsTo(in []SlackField) []v1alpha1.SlackField {
	out := make([]v1alpha1.SlackField, len(in))

	for i := range in {
		out[i] = v1alpha1.SlackField{
			Title: in[i].Title,
			Value: in[i].Value,
			Short: in[i].Short,
		}
	}

	return out
}

func convertSlackActionsTo(in []SlackAction) []v1alpha1.SlackAction {
	out := make([]v1alpha1.SlackAction, len(in))

	for i := range in {
		out[i] = v1alpha1.SlackAction{
			Type:  in[i].Type,
			Text:  in[i].Text,
			URL:   in[i].URL,
			Style: in[i].Style,
			Name:  in[i].Name,
			Value: in[i].Value,
		}
		if in[i].ConfirmField != nil {
			out[i].ConfirmField = &v1alpha1.SlackConfirmationField{
				Text:        in[i].ConfirmField.Text,
				Title:       in[i].ConfirmField.Title,
				OkText:      in[i].ConfirmField.OkText,
				DismissText: in[i].ConfirmField.DismissText,
			}
		}
	}

	return out
}

func convertSlackConfigTo(in SlackConfig) v1alpha1.SlackConfig {
	return v1alpha1.SlackConfig{
		SendResolved: in.SendResolved,
		APIURL:       convertSecretKeySelectorTo(in.APIURL),
		Channel:      in.Channel,
		Username:     in.Username,
		Color:        in.Color,
		Title:        in.Title,
		TitleLink:    in.TitleLink,
		Pretext:      in.Pretext,
		Text:         in.Text,
		Fields:       convertSlackFieldsTo(in.Fields),
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
		Actions:      convertSlackActionsTo(in.Actions),
		HTTPConfig:   convertHTTPConfigTo(in.HTTPConfig),
	}
}

func convertWebhookConfigTo(in WebhookConfig) v1alpha1.WebhookConfig {
	return v1alpha1.WebhookConfig{
		SendResolved: in.SendResolved,
		URL:          in.URL,
		URLSecret:    convertSecretKeySelectorTo(in.URLSecret),
		HTTPConfig:   convertHTTPConfigTo(in.HTTPConfig),
		MaxAlerts:    in.MaxAlerts,
	}
}

func convertWeChatConfigTo(in WeChatConfig) v1alpha1.WeChatConfig {
	return v1alpha1.WeChatConfig{
		SendResolved: in.SendResolved,
		APISecret:    convertSecretKeySelectorTo(in.APISecret),
		APIURL:       in.APIURL,
		CorpID:       in.CorpID,
		AgentID:      in.AgentID,
		ToUser:       in.ToUser,
		ToParty:      in.ToParty,
		ToTag:        in.ToTag,
		Message:      in.Message,
		MessageType:  in.MessageType,
		HTTPConfig:   convertHTTPConfigTo(in.HTTPConfig),
	}
}

func convertEmailConfigTo(in EmailConfig) v1alpha1.EmailConfig {
	return v1alpha1.EmailConfig{
		SendResolved: in.SendResolved,
		To:           in.To,
		From:         in.From,
		Hello:        in.Hello,
		Smarthost:    in.Smarthost,
		AuthUsername: in.AuthUsername,
		AuthPassword: convertSecretKeySelectorTo(in.AuthPassword),
		AuthSecret:   convertSecretKeySelectorTo(in.AuthSecret),
		AuthIdentity: in.AuthIdentity,
		Headers:      convertKeyValuesTo(in.Headers),
		HTML:         in.HTML,
		Text:         in.Text,
		RequireTLS:   in.RequireTLS,
		TLSConfig:    in.TLSConfig,
	}
}

func convertVictorOpsConfigTo(in VictorOpsConfig) v1alpha1.VictorOpsConfig {
	return v1alpha1.VictorOpsConfig{
		SendResolved:      in.SendResolved,
		APIKey:            convertSecretKeySelectorTo(in.APIKey),
		APIURL:            in.APIURL,
		RoutingKey:        in.RoutingKey,
		MessageType:       in.MessageType,
		EntityDisplayName: in.EntityDisplayName,
		StateMessage:      in.StateMessage,
		MonitoringTool:    in.MonitoringTool,
		CustomFields:      convertKeyValuesTo(in.CustomFields),
		HTTPConfig:        convertHTTPConfigTo(in.HTTPConfig),
	}
}

func convertPushoverConfigTo(in PushoverConfig) v1alpha1.PushoverConfig {
	return v1alpha1.PushoverConfig{
		SendResolved: in.SendResolved,
		UserKey:      convertSecretKeySelectorTo(in.UserKey),
		Token:        convertSecretKeySelectorTo(in.Token),
		Title:        in.Title,
		Message:      in.Message,
		URL:          in.URL,
		URLTitle:     in.URLTitle,
		Sound:        in.Sound,
		Priority:     in.Priority,
		Retry:        in.Retry,
		Expire:       in.Expire,
		HTML:         in.HTML,
		HTTPConfig:   convertHTTPConfigTo(in.HTTPConfig),
	}
}

func convertSNSConfigTo(in SNSConfig) v1alpha1.SNSConfig {
	return v1alpha1.SNSConfig{
		SendResolved: in.SendResolved,
		ApiURL:       in.ApiURL,
		Sigv4:        in.Sigv4,
		TopicARN:     in.TopicARN,
		Subject:      in.Subject,
		PhoneNumber:  in.PhoneNumber,
		TargetARN:    in.TargetARN,
		Message:      in.Message,
		Attributes:   in.Attributes,
		HTTPConfig:   convertHTTPConfigTo(in.HTTPConfig),
	}
}

func convertTelegramConfigTo(in TelegramConfig) v1alpha1.TelegramConfig {
	return v1alpha1.TelegramConfig{
		SendResolved:         in.SendResolved,
		APIURL:               in.APIURL,
		BotToken:             convertSecretKeySelectorTo(in.BotToken),
		ChatID:               in.ChatID,
		Message:              in.Message,
		DisableNotifications: in.DisableNotifications,
		ParseMode:            in.ParseMode,
		HTTPConfig:           convertHTTPConfigTo(in.HTTPConfig),
	}
}

// ConvertTo converts from this version (v1beta1) to the Hub version (v1alpha1).
func (src *AlertmanagerConfig) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.AlertmanagerConfig)

	dst.ObjectMeta = src.ObjectMeta

	for _, in := range src.Spec.Receivers {
		out := v1alpha1.Receiver{
			Name: in.Name,
		}

		for _, in := range in.OpsGenieConfigs {
			out.OpsGenieConfigs = append(
				out.OpsGenieConfigs,
				convertOpsGenieConfigTo(in),
			)
		}

		for _, in := range in.PagerDutyConfigs {
			out.PagerDutyConfigs = append(
				out.PagerDutyConfigs,
				convertPagerDutyConfigTo(in),
			)
		}

		for _, in := range in.SlackConfigs {
			out.SlackConfigs = append(
				out.SlackConfigs,
				convertSlackConfigTo(in),
			)
		}

		for _, in := range in.WebhookConfigs {
			out.WebhookConfigs = append(
				out.WebhookConfigs,
				convertWebhookConfigTo(in),
			)
		}

		for _, in := range in.WeChatConfigs {
			out.WeChatConfigs = append(
				out.WeChatConfigs,
				convertWeChatConfigTo(in),
			)
		}

		for _, in := range in.EmailConfigs {
			out.EmailConfigs = append(
				out.EmailConfigs,
				convertEmailConfigTo(in),
			)
		}

		for _, in := range in.VictorOpsConfigs {
			out.VictorOpsConfigs = append(
				out.VictorOpsConfigs,
				convertVictorOpsConfigTo(in),
			)
		}

		for _, in := range in.PushoverConfigs {
			out.PushoverConfigs = append(
				out.PushoverConfigs,
				convertPushoverConfigTo(in),
			)
		}

		for _, in := range in.SNSConfigs {
			out.SNSConfigs = append(
				out.SNSConfigs,
				convertSNSConfigTo(in),
			)
		}

		for _, in := range in.TelegramConfigs {
			out.TelegramConfigs = append(
				out.TelegramConfigs,
				convertTelegramConfigTo(in),
			)
		}

		dst.Spec.Receivers = append(dst.Spec.Receivers, out)
	}

	for _, in := range src.Spec.InhibitRules {
		dst.Spec.InhibitRules = append(
			dst.Spec.InhibitRules,
			v1alpha1.InhibitRule{
				TargetMatch: convertMatchersTo(in.TargetMatch),
				SourceMatch: convertMatchersTo(in.SourceMatch),
				Equal:       in.Equal,
			},
		)

	}

	for _, in := range src.Spec.TimeIntervals {
		dst.Spec.MuteTimeIntervals = append(
			dst.Spec.MuteTimeIntervals,
			v1alpha1.MuteTimeInterval{
				Name:          in.Name,
				TimeIntervals: convertTimeIntervalsTo(in.TimeIntervals),
			},
		)
	}

	r, err := convertRouteTo(src.Spec.Route)
	if err != nil {
		return err
	}
	dst.Spec.Route = r

	return nil
}
