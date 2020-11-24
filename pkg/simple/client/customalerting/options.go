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

package customalerting

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	PrometheusEndpoint       string `json:"prometheusEndpoint" yaml:"prometheusEndpoint"`
	ThanosRulerEndpoint      string `json:"thanosRulerEndpoint" yaml:"thanosRulerEndpoint"`
	ThanosRuleResourceLabels string `json:"thanosRuleResourceLabels" yaml:"thanosRuleResourceLabels"`
}

func NewOptions() *Options {
	return &Options{}
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
					errs = append(errs, fmt.Errorf("invalid thanos-rule-resource-labels arg: %s", o.ThanosRuleResourceLabels))
				}
			}
		}
	}

	return errs
}

func (o *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&o.PrometheusEndpoint, "prometheus-endpoint", c.PrometheusEndpoint,
		"Prometheus service endpoint from which built-in alerting rules are gotten.")
	fs.StringVar(&o.ThanosRulerEndpoint, "thanos-ruler-endpoint", c.ThanosRulerEndpoint,
		"Thanos ruler service endpoint from which custom alerting rules are gotten.")
	fs.StringVar(&o.ThanosRuleResourceLabels, "thanos-rule-resource-labels", c.ThanosRuleResourceLabels,
		"The labels will be added to prometheusrule custom resources to be selected by thanos ruler. eg: thanosruler=thanos-ruler,role=custom-alerting-rules")
}
