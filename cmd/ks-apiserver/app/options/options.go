package options

import (
	"github.com/spf13/pflag"
	genericoptions "kubesphere.io/kubesphere/pkg/options"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions

	// istio pilot discovery service url
	IstioPilotServiceURL string

	// jaeger query service url
	JaegerQueryServiceUrl string

	// prometheus service url for servicemesh metrics
	ServicemeshPrometheusServiceUrl string

	// openpitrix api gateway service url
	OpenPitrixServer string

	// openpitrix service token
	OpenPitrixProxyToken string
}

func NewServerRunOptions() *ServerRunOptions {

	s := ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		IstioPilotServiceURL:    "http://istio-pilot.istio-system.svc:8080/version",
		JaegerQueryServiceUrl:   "http://jaeger-query.istio-system.svc:16686/jaeger",
	}

	return &s
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {

	s.GenericServerRunOptions.AddFlags(fs)

	fs.StringVar(&s.IstioPilotServiceURL, "istio-pilot-service-url", "http://istio-pilot.istio-system.svc:8080/version", "istio pilot discovery service url")
	fs.StringVar(&s.JaegerQueryServiceUrl, "jaeger-query-service-url", "http://jaeger-query.istio-system.svc:16686/jaeger", "jaeger query service url")
	fs.StringVar(&s.ServicemeshPrometheusServiceUrl, "servicemesh-prometheus-service-url", "http://prometheus-k8s-system.kubesphere-monitoring-system.svc:9090", "prometheus service for servicemesh")
}
