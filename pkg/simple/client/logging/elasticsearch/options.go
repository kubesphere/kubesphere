package elasticsearch

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Host        string `json:"host" yaml:"host"`
	IndexPrefix string `json:"indexPrefix,omitempty" yaml:"indexPrefix"`
	Version     string `json:"version" yaml:"version"`
}

func NewElasticSearchOptions() *Options {
	return &Options{
		Host:        "",
		IndexPrefix: "fluentbit",
		Version:     "",
	}
}

func (s *Options) ApplyTo(options *Options) {
	if s.Host != "" {
		reflectutils.Override(options, s)
	}
}

func (s *Options) Validate() []error {
	errs := []error{}

	return errs
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.Host, "elasticsearch-host", c.Host, ""+
		"Elasticsearch logging service host. KubeSphere is using elastic as log store, "+
		"if this filed left blank, KubeSphere will use kubernetes builtin log API instead, and"+
		" the following elastic search options will be ignored.")

	fs.StringVar(&s.IndexPrefix, "index-prefix", c.IndexPrefix, ""+
		"Index name prefix. KubeSphere will retrieve logs against indices matching the prefix.")

	fs.StringVar(&s.Version, "elasticsearch-version", c.Version, ""+
		"Elasticsearch major version, e.g. 5/6/7, if left blank, will detect automatically."+
		"Currently, minimum supported version is 5.x")
}
