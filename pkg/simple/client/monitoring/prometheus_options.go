package monitoring

import (
    "github.com/spf13/pflag"
)

type Options struct {
    Endpoint          string `json:"endpoint,omitempty" yaml:"endpoint"`
    SecondaryEndpoint string `json:"secondaryEndpoint,omitempty" yaml:"secondaryEndpoint"`
}

func NewPrometheusOptions() *Options {
    return &Options{
        Endpoint:          "",
        SecondaryEndpoint: "",
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

    if s.SecondaryEndpoint != "" {
        options.SecondaryEndpoint = s.SecondaryEndpoint
    }
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
    fs.StringVar(&s.Endpoint, "prometheus-endpoint", c.Endpoint, ""+
        "Prometheus service endpoint which stores KubeSphere monitoring data, if left "+
        "blank, will use builtin metrics-server as data source.")

    fs.StringVar(&s.SecondaryEndpoint, "prometheus-secondary-endpoint", c.SecondaryEndpoint, ""+
        "Prometheus secondary service endpoint, if left empty and endpoint is set, will use endpoint instead.")
}
