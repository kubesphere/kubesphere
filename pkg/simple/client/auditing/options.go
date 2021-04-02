/*
Copyright 2020 The KubeSphere Authors.

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

package auditing

import (
	"time"

	"github.com/spf13/pflag"

	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Enable     bool   `json:"enable" yaml:"enable"`
	WebhookUrl string `json:"webhookUrl" yaml:"webhookUrl"`
	// The maximum concurrent senders which send auditing events to the auditing webhook.
	EventSendersNum int `json:"eventSendersNum" yaml:"eventSendersNum"`
	// The batch size of auditing events.
	EventBatchSize int `json:"eventBatchSize" yaml:"eventBatchSize"`
	// The batch interval of auditing events.
	EventBatchInterval time.Duration `json:"eventBatchInterval" yaml:"eventBatchInterval"`
	Host               string        `json:"host" yaml:"host"`
	BasicAuth          bool          `json:"basicAuth" yaml:"basicAuth"`
	Username           string        `json:"username" yaml:"username"`
	Password           string        `json:"password" yaml:"password"`
	IndexPrefix        string        `json:"indexPrefix,omitempty" yaml:"indexPrefix"`
	Version            string        `json:"version" yaml:"version"`
}

func NewAuditingOptions() *Options {
	return &Options{
		Host:        "",
		IndexPrefix: "ks-logstash-auditing",
		Version:     "",
	}
}

func (s *Options) ApplyTo(options *Options) {
	if s.Host != "" {
		reflectutils.Override(options, s)
	}
}

func (s *Options) Validate() []error {
	errs := make([]error, 0)
	return errs
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.BoolVar(&s.Enable, "auditing-enabled", c.Enable, "Enable auditing component or not. ")

	fs.StringVar(&s.WebhookUrl, "auditing-webhook-url", c.WebhookUrl, "Auditing wehook url")

	fs.BoolVar(&s.BasicAuth, "auditing-elasticsearch-basicAuth", c.BasicAuth, ""+
		"Elasticsearch auditing service basic auth enabled. KubeSphere is using elastic as auditing store, "+
		"if it is set to true, KubeSphere will connect to ElasticSearch using provided username and password by "+
		"auditing-elasticsearch-username and auditing-elasticsearch-username. Otherwise, KubeSphere will "+
		"anonymously access the Elasticsearch.")

	fs.StringVar(&s.Username, "auditing-elasticsearch-username", c.Username, ""+
		"ElasticSearch authentication username, only needed when auditing-elasticsearch-basicAuth is"+
		"set to true. ")

	fs.StringVar(&s.Password, "auditing-elasticsearch-password", c.Password, ""+
		"ElasticSearch authentication password, only needed when auditing-elasticsearch-basicAuth is"+
		"set to true. ")

	fs.IntVar(&s.EventSendersNum, "auditing-event-senders-num", c.EventSendersNum,
		"The maximum concurrent senders which send auditing events to the auditing webhook.")
	fs.IntVar(&s.EventBatchSize, "auditing-event-batch-size", c.EventBatchSize,
		"The batch size of auditing events.")
	fs.DurationVar(&s.EventBatchInterval, "auditing-event-batch-interval", c.EventBatchInterval,
		"The batch interval of auditing events.")

	fs.StringVar(&s.Host, "auditing-elasticsearch-host", c.Host, ""+
		"Elasticsearch service host. KubeSphere is using elastic as auditing store, "+
		"if this filed left blank, KubeSphere will use kubernetes builtin event API instead, and"+
		" the following elastic search options will be ignored.")

	fs.StringVar(&s.IndexPrefix, "auditing-index-prefix", c.IndexPrefix, ""+
		"Index name prefix. KubeSphere will retrieve auditing against indices matching the prefix.")

	fs.StringVar(&s.Version, "auditing-elasticsearch-version", c.Version, ""+
		"Elasticsearch major version, e.g. 5/6/7, if left blank, will detect automatically."+
		"Currently, minimum supported version is 5.x")
}
