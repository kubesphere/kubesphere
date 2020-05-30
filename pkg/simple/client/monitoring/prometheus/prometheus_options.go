package prometheus

import (
	"github.com/spf13/pflag"
)

type Options struct {
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint"`
}

func NewPrometheusOptions() *Options {
	return &Options{
		Endpoint: "",
	}
}

func (s *Options) Validate() []error {
	var errs []error
	return errs
}

func (s *Options) ApplyTo(options *Options) {
	if s.Endpoint != "" {
		options.Endpoint = s.Endpoint
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.Endpoint, "prometheus-endpoint", c.Endpoint, ""+
		"Prometheus service endpoint which stores KubeSphere monitoring data, if left "+
		"blank, will use builtin metrics-server as data source.")
}
