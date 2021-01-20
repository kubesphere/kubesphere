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

package kubeovn

import "github.com/spf13/pflag"

type Options struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

func NewOptions() *Options {
	return &Options{
		Enabled: false,
	}
}

func (o *Options) Validate() []error {
	return nil
}

func (o *Options) ApplyTo(options *Options) {
	options.Enabled = o.Enabled
}

func (o *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.BoolVar(&o.Enabled, "kubeovn", s.Enabled, ""+
		"This field enable the kubesphere support the kubeovn")
}
