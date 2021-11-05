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

package gateway

import (
	"github.com/spf13/pflag"

	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

// Options contains configuration of the default Gateway
type Options struct {
	WatchesPath string `json:"watchesPath,omitempty" yaml:"watchesPath"`
	Namespace   string `json:"namespace,omitempty" yaml:"namespace"`
	Repository  string `json:"repository,omitempty" yaml:"repository"`
	Tag         string `json:"tag,omitempty" yaml:"tag"`
}

// NewGatewayOptions creates a default Gateway Option
func NewGatewayOptions() *Options {
	return &Options{
		WatchesPath: "",
		Namespace:   "", //constants.KubeSphereControlNamespace
		Repository:  "",
		Tag:         "",
	}
}

func (s *Options) IsEmpty() bool {
	return s.WatchesPath == ""
}

// Validate check options values
func (s *Options) Validate() []error {
	var errors []error

	return errors
}

// ApplyTo overrides options if it's valid, which watchesPath is not empty
func (s *Options) ApplyTo(options *Options) {
	if s.WatchesPath != "" {
		reflectutils.Override(options, s)
	}
}

// AddFlags add options flags to command line flags,
// if watchesPath left empty, following options will be ignored
func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.WatchesPath, "watches-path", c.WatchesPath, "Path to the watches file to use.")
	fs.StringVar(&s.Namespace, "namespace", c.Namespace, "Working Namespace of the Gateway's Ingress Controller.")
	fs.StringVar(&s.Repository, "repository", c.Repository, "The Gateway Controller's image repository")
	fs.StringVar(&s.Tag, "tag", c.Tag, "The Gateway Controller's image tag")
}
