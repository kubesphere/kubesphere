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

package multicluster

import "github.com/spf13/pflag"

type Options struct {
	// Enable
	Enable           bool `json:"enable"`
	EnableFederation bool `json:"enableFederation,omitempty"`

	// ProxyPublishService is the service name of multicluster component tower.
	//   If this field provided, apiserver going to use the ingress.ip of this service.
	// This field will be used when generating agent deployment yaml for joining clusters.
	ProxyPublishService string `json:"proxyPublishService,omitempty"`

	// ProxyPublishAddress is the public address of tower for all cluster agents.
	//   This field takes precedence over field ProxyPublishService.
	// If both field ProxyPublishService and ProxyPublishAddress are empty, apiserver will
	// return 404 Not Found for all cluster agent yaml requests.
	ProxyPublishAddress string `json:"proxyPublishAddress,omitempty"`

	// AgentImage is the image used when generating deployment for all cluster agents.
	AgentImage string `json:"agentImage,omitempty"`
}

// NewOptions() returns a default nil options
func NewOptions() *Options {
	return &Options{
		Enable:              false,
		EnableFederation:    false,
		ProxyPublishAddress: "",
		ProxyPublishService: "",
		AgentImage:          "kubesphere/tower:v1.0",
	}
}

func (o *Options) Validate() []error {
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.BoolVar(&o.Enable, "multiple-clusters", s.Enable, ""+
		"This field instructs KubeSphere to enter multiple-cluster mode or not.")

	fs.StringVar(&o.ProxyPublishService, "proxy-publish-service", s.ProxyPublishService, ""+
		"Service name of tower. APIServer will use its ingress address as proxy publish address."+
		"For example, tower.kubesphere-system.svc.")

	fs.StringVar(&o.ProxyPublishAddress, "proxy-publish-address", s.ProxyPublishAddress, ""+
		"Public address of tower, APIServer will use this field as proxy publish address. This field "+
		"takes precedence over field proxy-publish-service. For example, http://139.198.121.121:8080.")

	fs.StringVar(&o.AgentImage, "agent-image", s.AgentImage, ""+
		"This field is used when generating deployment yaml for agent.")
}
