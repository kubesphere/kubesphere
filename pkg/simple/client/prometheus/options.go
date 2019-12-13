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
