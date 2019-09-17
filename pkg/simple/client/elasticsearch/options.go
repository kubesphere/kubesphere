package esclient

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type ElasticSearchOptions struct {
	Host           string `json:"host,omitempty" yaml:"host,omitempty"`
	LogstashFormat bool   `json:"logstashFormat,omitempty" yaml:"logstashFormat,omitempty"`
	Index          string `json:",omitempty" yaml:",omitempty"`
	LogstashPrefix string `json:"logstashPrefix,omitempty" yaml:"logstashPrefix,omitempty"`
	Match          string `json:",omitempty" yaml:",omitempty"`
	Version        string `json:",omitempty" yaml:",omitempty"`
}

func NewElasticSearchOptions() *ElasticSearchOptions {
	return &ElasticSearchOptions{
		Host:           "",
		LogstashFormat: false,
		Index:          "fluentbit",
		LogstashPrefix: "",
		Match:          "kube.*",
		Version:        "6",
	}
}

func (s *ElasticSearchOptions) ApplyTo(options *ElasticSearchOptions) {
	if s.Host != "" {
		reflectutils.Override(options, s)
	}
}

func (s *ElasticSearchOptions) Validate() []error {
	errs := []error{}

	return errs
}

func (s *ElasticSearchOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Host, "elasticsearch-host", s.Host, ""+
		"ElasticSearch logging service host. KubeSphere is using elastic as log store, "+
		"if this filed left blank, KubeSphere will use kubernetes builtin log API instead, and"+
		" the following elastic search options will be ignored.")

	fs.BoolVar(&s.LogstashFormat, "logstash-format", s.LogstashFormat, ""+
		"Whether to toggle logstash format compatibility.")

	fs.StringVar(&s.LogstashPrefix, "logstash-prefix", s.LogstashPrefix, ""+
		"If logstash-format is enabled, the Index name is composed using a prefix and the date,"+
		"e.g: If logstash-prefix is equals to 'mydata' your index will become 'mydata-YYYY.MM.DD'."+
		"The last string appended belongs to the date when the data is being generated.")

	fs.StringVar(&s.Match, "elasticsearch-match", s.Match, ""+
		"The regex match for index, eg. kube.*")

	fs.StringVar(&s.Index, "elasticsearch-index", s.Index, ""+
		"Index name.")

	fs.StringVar(&s.Version, "elasticsearch-version", s.Version, ""+
		"ElasticSearch major version, e.g. 5/6/7, if left blank, will detect automatically."+
		"Currently, minimum supported version is 5.x")
}
