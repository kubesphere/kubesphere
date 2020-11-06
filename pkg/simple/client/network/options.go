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

package network

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/simple/client/network/ippool"
)

type NSNPOptions struct {
	AllowedIngressNamespaces []string `json:"allowedIngressNamespaces,omitempty" yaml:"allowedIngressNamespaces,omitempty"`
}

type Options struct {
	EnableNetworkPolicy bool           `json:"enableNetworkPolicy,omitempty" yaml:"enableNetworkPolicy"`
	NSNPOptions         NSNPOptions    `json:"nsnpOptions,omitempty" yaml:"nsnpOptions,omitempty"`
	EnableIPPool        bool           `json:"enableIPPool,omitempty" yaml:"enableIPPool"`
	IPPoolOptions       ippool.Options `json:"ippoolOptions,omitempty" yaml:"ippoolOptions,omitempty"`
}

// NewNetworkOptions returns a `zero` instance
func NewNetworkOptions() *Options {
	return &Options{
		EnableNetworkPolicy: false,
		EnableIPPool:        false,
		NSNPOptions: NSNPOptions{
			AllowedIngressNamespaces: []string{},
		},
		IPPoolOptions: ippool.Options{
			Calico: nil,
		},
	}
}

func (s *Options) Validate() []error {
	var errors []error
	return errors
}

func (s *Options) ApplyTo(options *Options) {
	options.EnableNetworkPolicy = s.EnableNetworkPolicy
	options.EnableIPPool = s.EnableIPPool
	options.NSNPOptions = s.NSNPOptions
	options.IPPoolOptions = s.IPPoolOptions
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.BoolVar(&s.EnableNetworkPolicy, "enable-network-policy", c.EnableNetworkPolicy,
		"This field instructs KubeSphere to enable network policy or not.")
	fs.BoolVar(&s.EnableIPPool, "enable-ippool", c.EnableIPPool,
		"This field instructs KubeSphere to enable ippool or not.")
}
