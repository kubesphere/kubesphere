package servicemesh

import "github.com/spf13/pflag"

type Options struct {

	// istio pilot discovery service url
	IstioPilotHost string `json:"istioPilotHost,omitempty" yaml:"istioPilotHost"`

	// jaeger query service url
	JaegerQueryHost string `json:"jaegerQueryHost,omitempty" yaml:"jaegerQueryHost"`

	// prometheus service url for servicemesh metrics
	ServicemeshPrometheusHost string `json:"servicemeshPrometheusHost,omitempty" yaml:"servicemeshPrometheusHost"`
}

// NewServiceMeshOptions returns a `zero` instance
func NewServiceMeshOptions() *Options {
	return &Options{
		IstioPilotHost:            "",
		JaegerQueryHost:           "",
		ServicemeshPrometheusHost: "",
	}
}

func (s *Options) Validate() []error {
	errors := []error{}

	return errors
}

func (s *Options) ApplyTo(options *Options) {
	if s.ServicemeshPrometheusHost != "" {
		options.ServicemeshPrometheusHost = s.ServicemeshPrometheusHost
	}

	if s.JaegerQueryHost != "" {
		options.JaegerQueryHost = s.JaegerQueryHost
	}

	if s.IstioPilotHost != "" {
		options.IstioPilotHost = s.IstioPilotHost
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.IstioPilotHost, "istio-pilot-host", c.IstioPilotHost, ""+
		"istio pilot discovery service url")

	fs.StringVar(&s.JaegerQueryHost, "jaeger-query-host", c.JaegerQueryHost, ""+
		"jaeger query service url")

	fs.StringVar(&s.ServicemeshPrometheusHost, "servicemesh-prometheus-host", c.ServicemeshPrometheusHost, ""+
		"prometheus service for servicemesh")
}
