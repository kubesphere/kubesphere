// Copyright 2020 The prometheus-operator Authors
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

package v1alpha1

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	Version = "v1alpha1"

	AlertmanagerConfigKind    = "AlertmanagerConfig"
	AlertmanagerConfigName    = "alertmanagerconfigs"
	AlertmanagerConfigKindKey = "alertmanagerconfig"
)

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:resource:categories="prometheus-operator",shortName="amcfg"
// +kubebuilder:storageversion

// AlertmanagerConfig defines a namespaced AlertmanagerConfig to be aggregated
// across multiple namespaces configuring one Alertmanager cluster.
type AlertmanagerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AlertmanagerConfigSpec `json:"spec"`
}

// AlertmanagerConfigList is a list of AlertmanagerConfig.
// +k8s:openapi-gen=true
type AlertmanagerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of AlertmanagerConfig
	Items []*AlertmanagerConfig `json:"items"`
}

// AlertmanagerConfigSpec is a specification of the desired behavior of the Alertmanager configuration.
// By definition, the Alertmanager configuration only applies to alerts for which
// the `namespace` label is equal to the namespace of the AlertmanagerConfig resource.
type AlertmanagerConfigSpec struct {
	// The Alertmanager route definition for alerts matching the resource's
	// namespace. If present, it will be added to the generated Alertmanager
	// configuration as a first-level route.
	// +optional
	Route *Route `json:"route"`
	// List of receivers.
	// +optional
	Receivers []Receiver `json:"receivers"`
	// List of inhibition rules. The rules will only apply to alerts matching
	// the resource's namespace.
	// +optional
	InhibitRules []InhibitRule `json:"inhibitRules,omitempty"`
	// List of MuteTimeInterval specifying when the routes should be muted.
	// +optional
	MuteTimeIntervals []MuteTimeInterval `json:"muteTimeIntervals,omitempty"`
}

// Route defines a node in the routing tree.
type Route struct {
	// Name of the receiver for this route. If not empty, it should be listed in
	// the `receivers` field.
	// +optional
	Receiver string `json:"receiver"`
	// List of labels to group by.
	// Labels must not be repeated (unique list).
	// Special label "..." (aggregate by all possible labels), if provided, must be the only element in the list.
	// +optional
	GroupBy []string `json:"groupBy,omitempty"`
	// How long to wait before sending the initial notification.
	// Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`
	// Example: "30s"
	// +optional
	GroupWait string `json:"groupWait,omitempty"`
	// How long to wait before sending an updated notification.
	// Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`
	// Example: "5m"
	// +optional
	GroupInterval string `json:"groupInterval,omitempty"`
	// How long to wait before repeating the last notification.
	// Must match the regular expression`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`
	// Example: "4h"
	// +optional
	RepeatInterval string `json:"repeatInterval,omitempty"`
	// List of matchers that the alert's labels should match. For the first
	// level route, the operator removes any existing equality and regexp
	// matcher on the `namespace` label and adds a `namespace: <object
	// namespace>` matcher.
	// +optional
	Matchers []Matcher `json:"matchers,omitempty"`
	// Boolean indicating whether an alert should continue matching subsequent
	// sibling nodes. It will always be overridden to true for the first-level
	// route by the Prometheus operator.
	// +optional
	Continue bool `json:"continue,omitempty"`
	// Child routes.
	Routes []apiextensionsv1.JSON `json:"routes,omitempty"`
	// Note: this comment applies to the field definition above but appears
	// below otherwise it gets included in the generated manifest.
	// CRD schema doesn't support self-referential types for now (see
	// https://github.com/kubernetes/kubernetes/issues/62872). We have to use
	// an alternative type to circumvent the limitation. The downside is that
	// the Kube API can't validate the data beyond the fact that it is a valid
	// JSON representation.
	// MuteTimeIntervals is a list of MuteTimeInterval names that will mute this route when matched,
	// +optional
	MuteTimeIntervals []string `json:"muteTimeIntervals,omitempty"`
	// ActiveTimeIntervals is a list of MuteTimeInterval names when this route should be active.
	// +optional
	ActiveTimeIntervals []string `json:"activeTimeIntervals,omitempty"`
}

// ChildRoutes extracts the child routes.
func (r *Route) ChildRoutes() ([]Route, error) {
	out := make([]Route, len(r.Routes))

	for i, v := range r.Routes {
		if err := json.Unmarshal(v.Raw, &out[i]); err != nil {
			return nil, fmt.Errorf("route[%d]: %w", i, err)
		}
	}

	return out, nil
}

// Receiver defines one or more notification integrations.
type Receiver struct {
	// Name of the receiver. Must be unique across all items from the list.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// List of OpsGenie configurations.
	OpsGenieConfigs []OpsGenieConfig `json:"opsgenieConfigs,omitempty"`
	// List of PagerDuty configurations.
	PagerDutyConfigs []PagerDutyConfig `json:"pagerdutyConfigs,omitempty"`
	// List of Slack configurations.
	SlackConfigs []SlackConfig `json:"slackConfigs,omitempty"`
	// List of webhook configurations.
	WebhookConfigs []WebhookConfig `json:"webhookConfigs,omitempty"`
	// List of WeChat configurations.
	WeChatConfigs []WeChatConfig `json:"wechatConfigs,omitempty"`
	// List of Email configurations.
	EmailConfigs []EmailConfig `json:"emailConfigs,omitempty"`
	// List of VictorOps configurations.
	VictorOpsConfigs []VictorOpsConfig `json:"victoropsConfigs,omitempty"`
	// List of Pushover configurations.
	PushoverConfigs []PushoverConfig `json:"pushoverConfigs,omitempty"`
	// List of SNS configurations
	SNSConfigs []SNSConfig `json:"snsConfigs,omitempty"`
	// List of Telegram configurations.
	TelegramConfigs []TelegramConfig `json:"telegramConfigs,omitempty"`
}

// PagerDutyConfig configures notifications via PagerDuty.
// See https://prometheus.io/docs/alerting/latest/configuration/#pagerduty_config
type PagerDutyConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the PagerDuty integration key (when using
	// Events API v2). Either this field or `serviceKey` needs to be defined.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	RoutingKey *v1.SecretKeySelector `json:"routingKey,omitempty"`
	// The secret's key that contains the PagerDuty service key (when using
	// integration type "Prometheus"). Either this field or `routingKey` needs to
	// be defined.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	ServiceKey *v1.SecretKeySelector `json:"serviceKey,omitempty"`
	// The URL to send requests to.
	// +optional
	URL string `json:"url,omitempty"`
	// Client identification.
	// +optional
	Client string `json:"client,omitempty"`
	// Backlink to the sender of notification.
	// +optional
	ClientURL string `json:"clientURL,omitempty"`
	// Description of the incident.
	// +optional
	Description string `json:"description,omitempty"`
	// Severity of the incident.
	// +optional
	Severity string `json:"severity,omitempty"`
	// The class/type of the event.
	// +optional
	Class string `json:"class,omitempty"`
	// A cluster or grouping of sources.
	// +optional
	Group string `json:"group,omitempty"`
	// The part or component of the affected system that is broken.
	// +optional
	Component string `json:"component,omitempty"`
	// Arbitrary key/value pairs that provide further detail about the incident.
	// +optional
	Details []KeyValue `json:"details,omitempty"`
	// A list of image details to attach that provide further detail about an incident.
	// +optional
	PagerDutyImageConfigs []PagerDutyImageConfig `json:"pagerDutyImageConfigs,omitempty"`
	// A list of link details to attach that provide further detail about an incident.
	// +optional
	PagerDutyLinkConfigs []PagerDutyLinkConfig `json:"pagerDutyLinkConfigs,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// PagerDutyImageConfig attaches images to an incident
type PagerDutyImageConfig struct {
	// Src of the image being attached to the incident
	// +optional
	Src string `json:"src,omitempty"`
	// Optional URL; makes the image a clickable link.
	// +optional
	Href string `json:"href,omitempty"`
	// Alt is the optional alternative text for the image.
	// +optional
	Alt string `json:"alt,omitempty"`
}

// PagerDutyLinkConfig attaches text links to an incident
type PagerDutyLinkConfig struct {
	// Href is the URL of the link to be attached
	// +optional
	Href string `json:"href,omitempty"`
	// Text that describes the purpose of the link, and can be used as the link's text.
	// +optional
	Text string `json:"alt,omitempty"`
}

// SlackConfig configures notifications via Slack.
// See https://prometheus.io/docs/alerting/latest/configuration/#slack_config
type SlackConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the Slack webhook URL.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	APIURL *v1.SecretKeySelector `json:"apiURL,omitempty"`
	// The channel or user to send notifications to.
	// +optional
	Channel string `json:"channel,omitempty"`
	// +optional
	Username string `json:"username,omitempty"`
	// +optional
	Color string `json:"color,omitempty"`
	// +optional
	Title string `json:"title,omitempty"`
	// +optional
	TitleLink string `json:"titleLink,omitempty"`
	// +optional
	Pretext string `json:"pretext,omitempty"`
	// +optional
	Text string `json:"text,omitempty"`
	// A list of Slack fields that are sent with each notification.
	// +optional
	Fields []SlackField `json:"fields,omitempty"`
	// +optional
	ShortFields bool `json:"shortFields,omitempty"`
	// +optional
	Footer string `json:"footer,omitempty"`
	// +optional
	Fallback string `json:"fallback,omitempty"`
	// +optional
	CallbackID string `json:"callbackId,omitempty"`
	// +optional
	IconEmoji string `json:"iconEmoji,omitempty"`
	// +optional
	IconURL string `json:"iconURL,omitempty"`
	// +optional
	ImageURL string `json:"imageURL,omitempty"`
	// +optional
	ThumbURL string `json:"thumbURL,omitempty"`
	// +optional
	LinkNames bool `json:"linkNames,omitempty"`
	// +optional
	MrkdwnIn []string `json:"mrkdwnIn,omitempty"`
	// A list of Slack actions that are sent with each notification.
	// +optional
	Actions []SlackAction `json:"actions,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// Validate ensures SlackConfig is valid.
func (sc *SlackConfig) Validate() error {
	for _, action := range sc.Actions {
		if err := action.Validate(); err != nil {
			return err
		}
	}

	for _, field := range sc.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// SlackAction configures a single Slack action that is sent with each
// notification.
// See https://api.slack.com/docs/message-attachments#action_fields and
// https://api.slack.com/docs/message-buttons for more information.
type SlackAction struct {
	// +kubebuilder:validation:MinLength=1
	Type string `json:"type"`
	// +kubebuilder:validation:MinLength=1
	Text string `json:"text"`
	// +optional
	URL string `json:"url,omitempty"`
	// +optional
	Style string `json:"style,omitempty"`
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Value string `json:"value,omitempty"`
	// +optional
	ConfirmField *SlackConfirmationField `json:"confirm,omitempty"`
}

// Validate ensures SlackAction is valid.
func (sa *SlackAction) Validate() error {
	if sa.Type == "" {
		return errors.New("missing type in Slack action configuration")
	}

	if sa.Text == "" {
		return errors.New("missing text in Slack action configuration")
	}

	if sa.URL == "" && sa.Name == "" {
		return errors.New("missing name or url in Slack action configuration")
	}

	if sa.ConfirmField != nil {
		if err := sa.ConfirmField.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// SlackConfirmationField protect users from destructive actions or
// particularly distinguished decisions by asking them to confirm their button
// click one more time.
// See https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields
// for more information.
type SlackConfirmationField struct {
	// +kubebuilder:validation:MinLength=1
	Text string `json:"text"`
	// +optional
	Title string `json:"title,omitempty"`
	// +optional
	OkText string `json:"okText,omitempty"`
	// +optional
	DismissText string `json:"dismissText,omitempty"`
}

// Validate ensures SlackConfirmationField is valid.
func (scf *SlackConfirmationField) Validate() error {
	if scf.Text == "" {
		return errors.New("missing text in Slack confirmation configuration")
	}
	return nil
}

// SlackField configures a single Slack field that is sent with each notification.
// Each field must contain a title, value, and optionally, a boolean value to indicate if the field
// is short enough to be displayed next to other fields designated as short.
// See https://api.slack.com/docs/message-attachments#fields for more information.
type SlackField struct {
	// +kubebuilder:validation:MinLength=1
	Title string `json:"title"`
	// +kubebuilder:validation:MinLength=1
	Value string `json:"value"`
	// +optional
	Short *bool `json:"short,omitempty"`
}

// Validate ensures SlackField is valid
func (sf *SlackField) Validate() error {
	if sf.Title == "" {
		return errors.New("missing title in Slack field configuration")
	}

	if sf.Value == "" {
		return errors.New("missing value in Slack field configuration")
	}

	return nil
}

// WebhookConfig configures notifications via a generic receiver supporting the webhook payload.
// See https://prometheus.io/docs/alerting/latest/configuration/#webhook_config
type WebhookConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The URL to send HTTP POST requests to. `urlSecret` takes precedence over
	// `url`. One of `urlSecret` and `url` should be defined.
	// +optional
	URL *string `json:"url,omitempty"`
	// The secret's key that contains the webhook URL to send HTTP requests to.
	// `urlSecret` takes precedence over `url`. One of `urlSecret` and `url`
	// should be defined.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	URLSecret *v1.SecretKeySelector `json:"urlSecret,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
	// Maximum number of alerts to be sent per webhook message. When 0, all alerts are included.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxAlerts int32 `json:"maxAlerts,omitempty"`
}

// OpsGenieConfig configures notifications via OpsGenie.
// See https://prometheus.io/docs/alerting/latest/configuration/#opsgenie_config
type OpsGenieConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the OpsGenie API key.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	APIKey *v1.SecretKeySelector `json:"apiKey,omitempty"`
	// The URL to send OpsGenie API requests to.
	// +optional
	APIURL string `json:"apiURL,omitempty"`
	// Alert text limited to 130 characters.
	// +optional
	Message string `json:"message,omitempty"`
	// Description of the incident.
	// +optional
	Description string `json:"description,omitempty"`
	// Backlink to the sender of the notification.
	// +optional
	Source string `json:"source,omitempty"`
	// Comma separated list of tags attached to the notifications.
	// +optional
	Tags string `json:"tags,omitempty"`
	// Additional alert note.
	// +optional
	Note string `json:"note,omitempty"`
	// Priority level of alert. Possible values are P1, P2, P3, P4, and P5.
	// +optional
	Priority string `json:"priority,omitempty"`
	// Whether to update message and description of the alert in OpsGenie if it already exists
	// By default, the alert is never updated in OpsGenie, the new message only appears in activity log.
	// +optional
	UpdateAlerts *bool `json:"updateAlerts,omitempty"`
	// A set of arbitrary key/value pairs that provide further detail about the incident.
	// +optional
	Details []KeyValue `json:"details,omitempty"`
	// List of responders responsible for notifications.
	// +optional
	Responders []OpsGenieConfigResponder `json:"responders,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
	// Optional field that can be used to specify which domain alert is related to.
	// +optional
	Entity string `json:"entity,omitempty"`
	// Comma separated list of actions that will be available for the alert.
	// +optional
	Actions string `json:"actions,omitempty"`
}

// Validate ensures OpsGenieConfig is valid
func (o *OpsGenieConfig) Validate() error {
	for _, responder := range o.Responders {
		if err := responder.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// OpsGenieConfigResponder defines a responder to an incident.
// One of `id`, `name` or `username` has to be defined.
type OpsGenieConfigResponder struct {
	// ID of the responder.
	// +optional
	ID string `json:"id,omitempty"`
	// Name of the responder.
	// +optional
	Name string `json:"name,omitempty"`
	// Username of the responder.
	// +optional
	Username string `json:"username,omitempty"`
	// Type of responder.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Enum=team;teams;user;escalation;schedule
	Type string `json:"type"`
}

// Validate ensures OpsGenieConfigResponder is valid.
func (r *OpsGenieConfigResponder) Validate() error {
	if r.ID == "" && r.Name == "" && r.Username == "" {
		return errors.New("responder must have at least an ID, a Name or an Username defined")
	}

	return nil
}

// HTTPConfig defines a client HTTP configuration.
// See https://prometheus.io/docs/alerting/latest/configuration/#http_config
type HTTPConfig struct {
	// Authorization header configuration for the client.
	// This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.
	// +optional
	Authorization *monitoringv1.SafeAuthorization `json:"authorization,omitempty"`
	// BasicAuth for the client.
	// This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.
	// +optional
	BasicAuth *monitoringv1.BasicAuth `json:"basicAuth,omitempty"`
	// OAuth2 client credentials used to fetch a token for the targets.
	// +optional
	OAuth2 *monitoringv1.OAuth2 `json:"oauth2,omitempty"`
	// The secret's key that contains the bearer token to be used by the client
	// for authentication.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	BearerTokenSecret *v1.SecretKeySelector `json:"bearerTokenSecret,omitempty"`
	// TLS configuration for the client.
	// +optional
	TLSConfig *monitoringv1.SafeTLSConfig `json:"tlsConfig,omitempty"`
	// Optional proxy URL.
	// +optional
	ProxyURL string `json:"proxyURL,omitempty"`
	// FollowRedirects specifies whether the client should follow HTTP 3xx redirects.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`
}

// WeChatConfig configures notifications via WeChat.
// See https://prometheus.io/docs/alerting/latest/configuration/#wechat_config
type WeChatConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the WeChat API key.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	APISecret *v1.SecretKeySelector `json:"apiSecret,omitempty"`
	// The WeChat API URL.
	// +optional
	APIURL string `json:"apiURL,omitempty"`
	// The corp id for authentication.
	// +optional
	CorpID string `json:"corpID,omitempty"`
	// +optional
	AgentID string `json:"agentID,omitempty"`
	// +optional
	ToUser string `json:"toUser,omitempty"`
	// +optional
	ToParty string `json:"toParty,omitempty"`
	// +optional
	ToTag string `json:"toTag,omitempty"`
	// API request data as defined by the WeChat API.
	Message string `json:"message,omitempty"`
	// +optional
	MessageType string `json:"messageType,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// EmailConfig configures notifications via Email.
type EmailConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The email address to send notifications to.
	// +optional
	To string `json:"to,omitempty"`
	// The sender address.
	// +optional
	From string `json:"from,omitempty"`
	// The hostname to identify to the SMTP server.
	// +optional
	Hello string `json:"hello,omitempty"`
	// The SMTP host and port through which emails are sent. E.g. example.com:25
	// +optional
	Smarthost string `json:"smarthost,omitempty"`
	// The username to use for authentication.
	// +optional
	AuthUsername string `json:"authUsername,omitempty"`
	// The secret's key that contains the password to use for authentication.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	AuthPassword *v1.SecretKeySelector `json:"authPassword,omitempty"`
	// The secret's key that contains the CRAM-MD5 secret.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	AuthSecret *v1.SecretKeySelector `json:"authSecret,omitempty"`
	// The identity to use for authentication.
	// +optional
	AuthIdentity string `json:"authIdentity,omitempty"`
	// Further headers email header key/value pairs. Overrides any headers
	// previously set by the notification implementation.
	Headers []KeyValue `json:"headers,omitempty"`
	// The HTML body of the email notification.
	// +optional
	HTML string `json:"html,omitempty"`
	// The text body of the email notification.
	// +optional
	Text string `json:"text,omitempty"`
	// The SMTP TLS requirement.
	// Note that Go does not support unencrypted connections to remote SMTP endpoints.
	// +optional
	RequireTLS *bool `json:"requireTLS,omitempty"`
	// TLS configuration
	// +optional
	TLSConfig *monitoringv1.SafeTLSConfig `json:"tlsConfig,omitempty"`
}

// VictorOpsConfig configures notifications via VictorOps.
// See https://prometheus.io/docs/alerting/latest/configuration/#victorops_config
type VictorOpsConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the API key to use when talking to the VictorOps API.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +optional
	APIKey *v1.SecretKeySelector `json:"apiKey,omitempty"`
	// The VictorOps API URL.
	// +optional
	APIURL string `json:"apiUrl,omitempty"`
	// A key used to map the alert to a team.
	// +optional
	RoutingKey string `json:"routingKey"`
	// Describes the behavior of the alert (CRITICAL, WARNING, INFO).
	// +optional
	MessageType string `json:"messageType,omitempty"`
	// Contains summary of the alerted problem.
	// +optional
	EntityDisplayName string `json:"entityDisplayName,omitempty"`
	// Contains long explanation of the alerted problem.
	// +optional
	StateMessage string `json:"stateMessage,omitempty"`
	// The monitoring tool the state message is from.
	// +optional
	MonitoringTool string `json:"monitoringTool,omitempty"`
	// Additional custom fields for notification.
	// +optional
	CustomFields []KeyValue `json:"customFields,omitempty"`
	// The HTTP client's configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// PushoverConfig configures notifications via Pushover.
// See https://prometheus.io/docs/alerting/latest/configuration/#pushover_config
type PushoverConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The secret's key that contains the recipient user's user key.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +kubebuilder:validation:Required
	UserKey *v1.SecretKeySelector `json:"userKey,omitempty"`
	// The secret's key that contains the registered application's API token, see https://pushover.net/apps.
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	// +kubebuilder:validation:Required
	Token *v1.SecretKeySelector `json:"token,omitempty"`
	// Notification title.
	// +optional
	Title string `json:"title,omitempty"`
	// Notification message.
	// +optional
	Message string `json:"message,omitempty"`
	// A supplementary URL shown alongside the message.
	// +optional
	URL string `json:"url,omitempty"`
	// A title for supplementary URL, otherwise just the URL is shown
	// +optional
	URLTitle string `json:"urlTitle,omitempty"`
	// The name of one of the sounds supported by device clients to override the user's default sound choice
	// +optional
	Sound string `json:"sound,omitempty"`
	// Priority, see https://pushover.net/api#priority
	// +optional
	Priority string `json:"priority,omitempty"`
	// How often the Pushover servers will send the same notification to the user.
	// Must be at least 30 seconds.
	// +kubebuilder:validation:Pattern=`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`
	// +optional
	Retry string `json:"retry,omitempty"`
	// How long your notification will continue to be retried for, unless the user
	// acknowledges the notification.
	// +kubebuilder:validation:Pattern=`^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$`
	// +optional
	Expire string `json:"expire,omitempty"`
	// Whether notification message is HTML or plain text.
	// +optional
	HTML bool `json:"html,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// SNSConfig configures notifications via AWS SNS.
// See https://prometheus.io/docs/alerting/latest/configuration/#sns_configs
type SNSConfig struct {
	// Whether or not to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The SNS API URL i.e. https://sns.us-east-2.amazonaws.com.
	// If not specified, the SNS API URL from the SNS SDK will be used.
	// +optional
	ApiURL string `json:"apiURL,omitempty"`
	// Configures AWS's Signature Verification 4 signing process to sign requests.
	// +optional
	Sigv4 *monitoringv1.Sigv4 `json:"sigv4,omitempty"`
	// SNS topic ARN, i.e. arn:aws:sns:us-east-2:698519295917:My-Topic
	// If you don't specify this value, you must specify a value for the PhoneNumber or TargetARN.
	// +optional
	TopicARN string `json:"topicARN,omitempty"`
	// Subject line when the message is delivered to email endpoints.
	// +optional
	Subject string `json:"subject,omitempty"`
	// Phone number if message is delivered via SMS in E.164 format.
	// If you don't specify this value, you must specify a value for the TopicARN or TargetARN.
	// +optional
	PhoneNumber string `json:"phoneNumber,omitempty"`
	// The  mobile platform endpoint ARN if message is delivered via mobile notifications.
	// If you don't specify this value, you must specify a value for the topic_arn or PhoneNumber.
	// +optional
	TargetARN string `json:"targetARN,omitempty"`
	// The message content of the SNS notification.
	// +optional
	Message string `json:"message,omitempty"`
	// SNS message attributes.
	// +optional
	Attributes map[string]string `json:"attributes,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// TelegramConfig configures notifications via Telegram.
// See https://prometheus.io/docs/alerting/latest/configuration/#telegram_config
type TelegramConfig struct {
	// Whether to notify about resolved alerts.
	// +optional
	SendResolved *bool `json:"sendResolved,omitempty"`
	// The Telegram API URL i.e. https://api.telegram.org.
	// If not specified, default API URL will be used.
	// +optional
	APIURL string `json:"apiURL,omitempty"`
	// Telegram bot token
	// The secret needs to be in the same namespace as the AlertmanagerConfig
	// object and accessible by the Prometheus Operator.
	BotToken *v1.SecretKeySelector `json:"botToken,omitempty"`
	// The Telegram chat ID.
	ChatID int64 `json:"chatID,omitempty"`
	// Message template
	// +optional
	Message string `json:"message,omitempty"`
	// Disable telegram notifications
	// +optional
	DisableNotifications *bool `json:"disableNotifications,omitempty"`
	// Parse mode for telegram message
	//+kubebuilder:validation:Enum=MarkdownV2;Markdown;HTML
	// +optional
	ParseMode string `json:"parseMode,omitempty"`
	// HTTP client configuration.
	// +optional
	HTTPConfig *HTTPConfig `json:"httpConfig,omitempty"`
}

// InhibitRule defines an inhibition rule that allows to mute alerts when other
// alerts are already firing.
// See https://prometheus.io/docs/alerting/latest/configuration/#inhibit_rule
type InhibitRule struct {
	// Matchers that have to be fulfilled in the alerts to be muted. The
	// operator enforces that the alert matches the resource's namespace.
	TargetMatch []Matcher `json:"targetMatch,omitempty"`
	// Matchers for which one or more alerts have to exist for the inhibition
	// to take effect. The operator enforces that the alert matches the
	// resource's namespace.
	SourceMatch []Matcher `json:"sourceMatch,omitempty"`
	// Labels that must have an equal value in the source and target alert for
	// the inhibition to take effect.
	Equal []string `json:"equal,omitempty"`
}

// KeyValue defines a (key, value) tuple.
type KeyValue struct {
	// Key of the tuple.
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
	// Value of the tuple.
	Value string `json:"value"`
}

// Matcher defines how to match on alert's labels.
type Matcher struct {
	// Label to match.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Label value to match.
	// +optional
	Value string `json:"value"`
	// Match operation available with AlertManager >= v0.22.0 and
	// takes precedence over Regex (deprecated) if non-empty.
	// +kubebuilder:validation:Enum=!=;=;=~;!~
	// +optional
	MatchType MatchType `json:"matchType,omitempty"`
	// Whether to match on equality (false) or regular-expression (true).
	// Deprecated as of AlertManager >= v0.22.0 where a user should use MatchType instead.
	// +optional
	Regex bool `json:"regex,omitempty"`
}

// String returns Matcher as a string
// Use only for MatchType Matcher
func (in Matcher) String() string {
	return fmt.Sprintf(`%s%s"%s"`, in.Name, in.MatchType, openMetricsEscape(in.Value))
}

// Validate the Matcher returns an error if the matcher is invalid
// Validates only non-deprecated matching fields
func (in Matcher) Validate() error {
	// nothing to do
	if in.MatchType == "" {
		return nil
	}

	if !in.MatchType.Valid() {
		return fmt.Errorf("invalid 'matchType' '%s' provided'", in.MatchType)
	}

	if strings.TrimSpace(in.Name) == "" {
		return errors.New("matcher 'name' is required")
	}

	return nil
}

// MatchType is a comparison operator on a Matcher
type MatchType string

// Valid MatchType returns true if the operator is acceptable
func (mt MatchType) Valid() bool {
	_, ok := validMatchTypes[mt]
	return ok
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerConfig) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

// DeepCopyObject implements the runtime.Object interface.
func (l *AlertmanagerConfigList) DeepCopyObject() runtime.Object {
	return l.DeepCopy()
}

const (
	MatchEqual     MatchType = "="
	MatchNotEqual  MatchType = "!="
	MatchRegexp    MatchType = "=~"
	MatchNotRegexp MatchType = "!~"
)

var validMatchTypes = map[MatchType]bool{
	MatchEqual:     true,
	MatchNotEqual:  true,
	MatchRegexp:    true,
	MatchNotRegexp: true,
}

// openMetricsEscape is similar to the usual string escaping, but more
// restricted. It merely replaces a new-line character with '\n', a double-quote
// character with '\"', and a backslash with '\\', which is the escaping used by
// OpenMetrics.
// * Copied from alertmanager codebase pkg/labels *
func openMetricsEscape(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		"\n", `\n`,
		`"`, `\"`,
	)
	return r.Replace(s)
}

// MuteTimeInterval specifies the periods in time when notifications will be muted
type MuteTimeInterval struct {
	// Name of the time interval
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// TimeIntervals is a list of TimeInterval
	TimeIntervals []TimeInterval `json:"timeIntervals,omitempty"`
}

// TimeInterval describes intervals of time
type TimeInterval struct {
	// Times is a list of TimeRange
	// +optional
	Times []TimeRange `json:"times,omitempty"`
	// Weekdays is a list of WeekdayRange
	// +optional
	Weekdays []WeekdayRange `json:"weekdays,omitempty"`
	// DaysOfMonth is a list of DayOfMonthRange
	// +optional
	DaysOfMonth []DayOfMonthRange `json:"daysOfMonth,omitempty"`
	// Months is a list of MonthRange
	// +optional
	Months []MonthRange `json:"months,omitempty"`
	// Years is a list of YearRange
	// +optional
	Years []YearRange `json:"years,omitempty"`
}

// Time defines a time in 24hr format
// +kubebuilder:validation:Pattern=`^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)`
type Time string

// TimeRange defines a start and end time in 24hr format
type TimeRange struct {
	// StartTime is the start time in 24hr format.
	StartTime Time `json:"startTime,omitempty"`
	// EndTime is the end time in 24hr format.
	EndTime Time `json:"endTime,omitempty"`
}

// WeekdayRange is an inclusive range of days of the week beginning on Sunday
// Days can be specified by name (e.g 'Sunday') or as an inclusive range (e.g 'Monday:Friday')
// +kubebuilder:validation:Pattern=`^((?i)sun|mon|tues|wednes|thurs|fri|satur)day(?:((:(sun|mon|tues|wednes|thurs|fri|satur)day)$)|$)`
type WeekdayRange string

// DayOfMonthRange is an inclusive range of days of the month beginning at 1
type DayOfMonthRange struct {
	// Start of the inclusive range
	// +kubebuilder:validation:Minimum=-31
	// +kubebuilder:validation:Maximum=31
	Start int `json:"start,omitempty"`
	// End of the inclusive range
	// +kubebuilder:validation:Minimum=-31
	// +kubebuilder:validation:Maximum=31
	End int `json:"end,omitempty"`
}

// MonthRange is an inclusive range of months of the year beginning in January
// Months can be specified by name (e.g 'January') by numerical month (e.g '1') or as an inclusive range (e.g 'January:March', '1:3', '1:March')
// +kubebuilder:validation:Pattern=`^((?i)january|february|march|april|may|june|july|august|september|october|november|december|[1-12])(?:((:((?i)january|february|march|april|may|june|july|august|september|october|november|december|[1-12]))$)|$)`
type MonthRange string

// YearRange is an inclusive range of years
// +kubebuilder:validation:Pattern=`^2\d{3}(?::2\d{3}|$)`
type YearRange string

// Weekday is day of the week
type Weekday string

const (
	Sunday    Weekday = "sunday"
	Monday    Weekday = "monday"
	Tuesday   Weekday = "tuesday"
	Wednesday Weekday = "wednesday"
	Thursday  Weekday = "thursday"
	Friday    Weekday = "friday"
	Saturday  Weekday = "saturday"
)

var daysOfWeek = map[Weekday]int{
	Sunday:    0,
	Monday:    1,
	Tuesday:   2,
	Wednesday: 3,
	Thursday:  4,
	Friday:    5,
	Saturday:  6,
}

var daysOfWeekInv = map[int]Weekday{
	0: Sunday,
	1: Monday,
	2: Tuesday,
	3: Wednesday,
	4: Thursday,
	5: Friday,
	6: Saturday,
}

// Month of the year
type Month string

const (
	January   Month = "january"
	February  Month = "february"
	March     Month = "march"
	April     Month = "april"
	May       Month = "may"
	June      Month = "june"
	July      Month = "july"
	August    Month = "august"
	September Month = "september"
	October   Month = "october"
	November  Month = "november"
	December  Month = "december"
)

var months = map[Month]int{
	January:   1,
	February:  2,
	March:     3,
	April:     4,
	May:       5,
	June:      6,
	July:      7,
	August:    8,
	September: 9,
	October:   10,
	November:  11,
	December:  12,
}

var monthsInv = map[int]Month{
	1:  January,
	2:  February,
	3:  March,
	4:  April,
	5:  May,
	6:  June,
	7:  July,
	8:  August,
	9:  September,
	10: October,
	11: November,
	12: December,
}
