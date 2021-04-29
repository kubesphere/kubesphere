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

	networkv1alpha1 "kubesphere.io/api/network/v1alpha1"
)

type NSNPOptions struct {
	AllowedIngressNamespaces []string `json:"allowedIngressNamespaces,omitempty" yaml:"allowedIngressNamespaces,omitempty"`
}

type Options struct {
	EnableNetworkPolicy bool        `json:"enableNetworkPolicy,omitempty" yaml:"enableNetworkPolicy"`
	NSNPOptions         NSNPOptions `json:"nsnpOptions,omitempty" yaml:"nsnpOptions,omitempty"`
	WeaveScopeHost      string      `json:"weaveScopeHost,omitempty" yaml:"weaveScopeHost,omitempty"`
	IPPoolType          string      `json:"ippoolType,omitempty" yaml:"ippoolType,omitempty"`
}

// NewNetworkOptions returns a `zero` instance
func NewNetworkOptions() *Options {
	return &Options{
		EnableNetworkPolicy: false,
		IPPoolType:          networkv1alpha1.IPPoolTypeNone,
		NSNPOptions: NSNPOptions{
			AllowedIngressNamespaces: []string{},
		},
		WeaveScopeHost: "",
	}
}

func (s *Options) IsEmpty() bool {
	return s.EnableNetworkPolicy == false &&
		s.WeaveScopeHost == "" &&
		s.IPPoolType == networkv1alpha1.IPPoolTypeNone
}

func (s *Options) Validate() []error {
	var errors []error
	return errors
}

func (s *Options) ApplyTo(options *Options) {
	options.EnableNetworkPolicy = s.EnableNetworkPolicy
	options.IPPoolType = s.IPPoolType
	options.NSNPOptions = s.NSNPOptions
	options.WeaveScopeHost = s.WeaveScopeHost
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.BoolVar(&s.EnableNetworkPolicy, "enable-network-policy", c.EnableNetworkPolicy,
		"This field instructs KubeSphere to enable network policy or not.")
	fs.StringVar(&s.IPPoolType, "ippool-type", c.IPPoolType,
		"This field instructs KubeSphere to enable ippool or not.")
	fs.StringVar(&s.WeaveScopeHost, "weave-scope-host", c.WeaveScopeHost,
		"Weave Scope service endpoint which build a topology API of the applications and the containers running on the hosts")
}
