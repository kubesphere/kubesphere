/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auditing

import (
	"time"

	"k8s.io/apiserver/pkg/apis/audit"

	"github.com/spf13/pflag"
)

type WebhookOptions struct {
	WebhookUrl string `json:"webhookUrl" yaml:"webhookUrl"`
	// The maximum concurrent senders which send auditing events to the auditing webhook.
	EventSendersNum int `json:"eventSendersNum" yaml:"eventSendersNum"`
}

type LogOptions struct {
	Path       string `json:"path" yaml:"path"`
	MaxAge     int    `json:"maxAge" yaml:"maxAge"`
	MaxBackups int    `json:"maxBackups" yaml:"maxBackups"`
	MaxSize    int    `json:"maxSize" yaml:"maxSize"`
}

type Options struct {
	Enable     bool        `json:"enable" yaml:"enable"`
	AuditLevel audit.Level `json:"auditLevel" yaml:"auditLevel"`
	// The batch size of auditing events.
	EventBatchSize int `json:"eventBatchSize" yaml:"eventBatchSize"`
	// The batch interval of auditing events.
	EventBatchInterval time.Duration `json:"eventBatchInterval" yaml:"eventBatchInterval"`

	WebhookOptions WebhookOptions `json:"webhookOptions" yaml:"webhookOptions"`
	LogOptions     LogOptions     `json:"logOptions" yaml:"logOptions"`
}

func NewAuditingOptions() *Options {
	return &Options{}
}

func (s *Options) Validate() []error {
	errs := make([]error, 0)
	return errs
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.BoolVar(&s.Enable, "auditing-enabled", c.Enable, "Enable auditing component or not. ")
	fs.IntVar(&s.EventBatchSize, "auditing-event-batch-size", c.EventBatchSize,
		"The batch size of auditing events.")
	fs.DurationVar(&s.EventBatchInterval, "auditing-event-batch-interval", c.EventBatchInterval,
		"The batch interval of auditing events.")

	fs.StringVar(&s.WebhookOptions.WebhookUrl, "auditing-webhook-url", c.WebhookOptions.WebhookUrl, "Auditing wehook url")
	fs.IntVar(&s.WebhookOptions.EventSendersNum, "auditing-event-senders-num", c.WebhookOptions.EventSendersNum,
		"The maximum concurrent senders which send auditing events to the auditing webhook.")

	fs.StringVar(&s.LogOptions.Path, "audit-log-path", s.LogOptions.Path,
		"If set, all requests coming to the apiserver will be logged to this file.  '-' means standard out.")
	fs.IntVar(&s.LogOptions.MaxAge, "audit-log-maxage", s.LogOptions.MaxAge,
		"The maximum number of days to retain old audit log files based on the timestamp encoded in their filename.")
	fs.IntVar(&s.LogOptions.MaxBackups, "audit-log-maxbackup", s.LogOptions.MaxBackups,
		"The maximum number of old audit log files to retain. Setting a value of 0 will mean there's no restriction on the number of files.")
	fs.IntVar(&s.LogOptions.MaxSize, "audit-log-maxsize", s.LogOptions.MaxSize,
		"The maximum size in megabytes of the audit log file before it gets rotated.")
}
