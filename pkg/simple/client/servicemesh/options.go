/*
Copyright 2019 The KubeSphere Authors.

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
