package servicemesh

import "github.com/spf13/pflag"

type ServiceMeshOptions struct {

	// istio pilot discovery service url
	IstioPilotHost string `json:"istioPilotHost,omitempty" yaml:"istioPilotHost"`

	// jaeger query service url
	JaegerQueryHost string `json:"jaegerQueryHost,omitempty" yaml:"jaegerQueryHost"`

	// prometheus service url for servicemesh metrics
	ServicemeshPrometheusHost string `json:"servicemeshPrometheusHost,omitempty" yaml:"servicemeshPrometheusHost"`
}

// NewServiceMeshOptions returns a `zero` instance
func NewServiceMeshOptions() *ServiceMeshOptions {
	return &ServiceMeshOptions{
		IstioPilotHost:            "",
		JaegerQueryHost:           "",
		ServicemeshPrometheusHost: "",
	}
}

func (s *ServiceMeshOptions) Validate() []error {
	errors := []error{}

	return errors
}

func (s *ServiceMeshOptions) ApplyTo(options *ServiceMeshOptions) {
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

func (s *ServiceMeshOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.IstioPilotHost, "istio-pilot-host", s.IstioPilotHost, ""+
		"istio pilot discovery service url")

	fs.StringVar(&s.JaegerQueryHost, "jaeger-query-host", s.JaegerQueryHost, ""+
		"jaeger query service url")

	fs.StringVar(&s.ServicemeshPrometheusHost, "servicemesh-prometheus-host", s.ServicemeshPrometheusHost, ""+
		"prometheus service for servicemesh")
}
