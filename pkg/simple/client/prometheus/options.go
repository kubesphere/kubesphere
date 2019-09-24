package prometheus

import (
	"github.com/spf13/pflag"
)

type PrometheusOptions struct {
	Endpoint          string `json:"endpoint,omitempty" yaml:"endpoint"`
	SecondaryEndpoint string `json:"secondaryEndpoint,omitempty" yaml:"secondaryEndpoint"`
}

func NewPrometheusOptions() *PrometheusOptions {
	return &PrometheusOptions{
		Endpoint:          "",
		SecondaryEndpoint: "",
	}
}

func (s *PrometheusOptions) Validate() []error {
	errs := []error{}

	return errs
}

func (s *PrometheusOptions) ApplyTo(options *PrometheusOptions) {
	if s.Endpoint != "" {
		options.Endpoint = s.Endpoint
	}

	if s.SecondaryEndpoint != "" {
		options.SecondaryEndpoint = s.SecondaryEndpoint
	}
}

func (s *PrometheusOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Endpoint, "prometheus-endpoint", s.Endpoint, ""+
		"Prometheus service endpoint which stores KubeSphere monitoring data, if left "+
		"blank, will use builtin metrics-server as data source.")

	fs.StringVar(&s.SecondaryEndpoint, "prometheus-secondary-endpoint", s.SecondaryEndpoint, ""+
		"Prometheus secondary service endpoint, if left empty and endpoint is set, will use endpoint instead.")
}
