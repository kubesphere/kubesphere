/*
Copyright 2021 KubeSphere Authors

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

package kubeedge

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
}

func NewKubeEdgeOptions() *Options {
	return &Options{
		Endpoint: "",
	}
}

func (o *Options) ApplyTo(options *Options) {
	reflectutils.Override(options, o)
}

func (o *Options) Validate() []error {
	errs := []error{}

	return errs
}

func (o *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&o.Endpoint, "edge-watcher-endpoint", c.Endpoint,
		"edge watcher endpoint for kubeedge v1alpha1.")
}
