/*
Copyright 2018 The Kubernetes Authors.

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

package options

import (
	"github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/apis/componentconfig"
)

// NodeIpamControllerOptions holds the NodeIpamController options.
type NodeIpamControllerOptions struct {
	ServiceCIDR      string
	NodeCIDRMaskSize int32
}

// AddFlags adds flags related to NodeIpamController for controller manager to the specified FlagSet.
func (o *NodeIpamControllerOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.ServiceCIDR, "service-cluster-ip-range", o.ServiceCIDR, "CIDR Range for Services in cluster. Requires --allocate-node-cidrs to be true")
	fs.Int32Var(&o.NodeCIDRMaskSize, "node-cidr-mask-size", o.NodeCIDRMaskSize, "Mask size for node cidr in cluster.")
}

// ApplyTo fills up NodeIpamController config with options.
func (o *NodeIpamControllerOptions) ApplyTo(cfg *componentconfig.NodeIpamControllerConfiguration) error {
	if o == nil {
		return nil
	}

	cfg.ServiceCIDR = o.ServiceCIDR
	cfg.NodeCIDRMaskSize = o.NodeCIDRMaskSize

	return nil
}

// Validate checks validation of NodeIpamControllerOptions.
func (o *NodeIpamControllerOptions) Validate() []error {
	if o == nil {
		return nil
	}

	errs := []error{}
	return errs
}
