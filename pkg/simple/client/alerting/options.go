/*
Copyright 2020 KubeSphere Authors

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

package alerting

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// The following options are for the alerting with v2alpha1 version or higher versions
	PrometheusEndpoint       string `json:"prometheusEndpoint" yaml:"prometheusEndpoint"`
	ThanosRulerEndpoint      string `json:"thanosRulerEndpoint" yaml:"thanosRulerEndpoint"`
	ThanosRuleResourceLabels string `json:"thanosRuleResourceLabels" yaml:"thanosRuleResourceLabels"`
}

func NewAlertingOptions() *Options {
	return &Options{
		Endpoint: "",
	}
}

func (o *Options) ApplyTo(options *Options) {
	reflectutils.Override(options, o)
}

func (o *Options) Validate() []error {
	errs := []error{}

	if len(o.ThanosRuleResourceLabels) > 0 {
		lblStrings := strings.Split(o.ThanosRuleResourceLabels, ",")
		for _, lblString := range lblStrings {
			if len(lblString) > 0 {
				lbl := strings.Split(lblString, "=")
				if len(lbl) != 2 {
					errs = append(errs, fmt.Errorf("invalid alerting-thanos-rule-resource-labels arg: %s", o.ThanosRuleResourceLabels))
					break
				}
			}
		}
	}

	return errs
}

func (o *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&o.Endpoint, "alerting-server-endpoint", c.Endpoint,
		"alerting server endpoint for alerting v1.")

	fs.StringVar(&o.PrometheusEndpoint, "alerting-prometheus-endpoint", c.PrometheusEndpoint,
		"Prometheus service endpoint from which built-in alerting rules are fetched(alerting v2alpha1 or higher required)")
	fs.StringVar(&o.ThanosRulerEndpoint, "alerting-thanos-ruler-endpoint", c.ThanosRulerEndpoint,
		"Thanos ruler service endpoint from which custom alerting rules are fetched(alerting v2alpha1 or higher required)")
	fs.StringVar(&o.ThanosRuleResourceLabels, "alerting-thanos-rule-resource-labels", c.ThanosRuleResourceLabels,
		"Labels used by Thanos Ruler to select PrometheusRule custom resources. eg: thanosruler=thanos-ruler,role=custom-alerting-rules (alerting v2alpha1 or higher required)")
}
