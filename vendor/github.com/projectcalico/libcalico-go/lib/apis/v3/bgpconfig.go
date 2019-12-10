// Copyright (c) 2017 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectcalico/libcalico-go/lib/numorstring"
)

const (
	KindBGPConfiguration     = "BGPConfiguration"
	KindBGPConfigurationList = "BGPConfigurationList"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BGPConfiguration contains the configuration for any BGP routing.
type BGPConfiguration struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the BGPConfiguration.
	Spec BGPConfigurationSpec `json:"spec,omitempty"`
}

// BGPConfigurationSpec contains the values of the BGP configuration.
type BGPConfigurationSpec struct {
	// LogSeverityScreen is the log severity above which logs are sent to the stdout. [Default: INFO]
	LogSeverityScreen string `json:"logSeverityScreen,omitempty" validate:"omitempty,logLevel" confignamev1:"loglevel"`

	// NodeToNodeMeshEnabled sets whether full node to node BGP mesh is enabled. [Default: true]
	NodeToNodeMeshEnabled *bool `json:"nodeToNodeMeshEnabled,omitempty" validate:"omitempty" confignamev1:"node_mesh"`

	// ASNumber is the default AS number used by a node. [Default: 64512]
	ASNumber *numorstring.ASNumber `json:"asNumber,omitempty" validate:"omitempty" confignamev1:"as_num"`

	// ServiceExternalIPs are the CIDR blocks for Kubernetes Service External IPs.
	// Kubernetes Service ExternalIPs will only be advertised if they are within one of these blocks.
	ServiceExternalIPs []ServiceExternalIPBlock `json:"serviceExternalIPs,omitempty" validate:"omitempty,dive" confignamev1:"svc_external_ips"`

	// ServiceClusterIPs are the CIDR blocks from which service cluster IPs are allocated.
	// If specified, Calico will advertise these blocks, as well as any cluster IPs within them.
	ServiceClusterIPs []ServiceClusterIPBlock `json:"serviceClusterIPs,omitempty" validate:"omitempty,dive" confignamev1:"svc_cluster_ips"`
}

// ServiceExternalIPBlock represents a single whitelisted CIDR External IP block.
type ServiceExternalIPBlock struct {
	CIDR string `json:"cidr,omitempty" validate:"omitempty,net"`
}

// ServiceClusterIPBlock represents a single whitelisted CIDR block for ClusterIPs.
type ServiceClusterIPBlock struct {
	CIDR string `json:"cidr,omitempty" validate:"omitempty,net"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BGPConfigurationList contains a list of BGPConfiguration resources.
type BGPConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []BGPConfiguration `json:"items"`
}

// New BGPConfiguration creates a new (zeroed) BGPConfiguration struct with the TypeMetadata
// initialized to the current version.
func NewBGPConfiguration() *BGPConfiguration {
	return &BGPConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindBGPConfiguration,
			APIVersion: GroupVersionCurrent,
		},
	}
}

// NewBGPConfigurationList creates a new zeroed) BGPConfigurationList struct with the TypeMetadata
// initialized to the current version.
func NewBGPConfigurationList() *BGPConfigurationList {
	return &BGPConfigurationList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindBGPConfigurationList,
			APIVersion: GroupVersionCurrent,
		},
	}
}
