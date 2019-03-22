package options

import (
	"github.com/spf13/pflag"
	genericoptions "kubesphere.io/kubesphere/pkg/options"
)

type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions

	// istio pilot discovery service url
	IstioPilotServiceURL string
}

func NewServerRunOptions() *ServerRunOptions {

	s := ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		IstioPilotServiceURL:    "http://istio-pilot.istio-system.svc:8080/version",
	}

	return &s
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {

	s.GenericServerRunOptions.AddFlags(fs)

	fs.StringVar(&s.IstioPilotServiceURL, "istio-pilot-service-url", "http://istio-pilot.istio-system.svc:8080/version", "istio pilot discovery service url")
}
