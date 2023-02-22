// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.

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

package model

import (
	"fmt"

	"regexp"

	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/api/pkg/lib/numorstring"

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
	"github.com/projectcalico/calico/libcalico-go/lib/net"
)

var (
	matchWorkloadEndpoint = regexp.MustCompile("^/?calico/v1/host/([^/]+)/workload/([^/]+)/([^/]+)/endpoint/([^/]+)$")
)

type WorkloadEndpointKey struct {
	Hostname       string `json:"-"`
	OrchestratorID string `json:"-"`
	WorkloadID     string `json:"-"`
	EndpointID     string `json:"-"`
}

func (key WorkloadEndpointKey) defaultPath() (string, error) {
	if key.Hostname == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "node"}
	}
	if key.OrchestratorID == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "orchestrator"}
	}
	if key.WorkloadID == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "workload"}
	}
	if key.EndpointID == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	return fmt.Sprintf("/calico/v1/host/%s/workload/%s/%s/endpoint/%s",
		key.Hostname, escapeName(key.OrchestratorID), escapeName(key.WorkloadID), escapeName(key.EndpointID)), nil
}

func (key WorkloadEndpointKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key WorkloadEndpointKey) defaultDeleteParentPaths() ([]string, error) {
	if key.Hostname == "" {
		return nil, errors.ErrorInsufficientIdentifiers{Name: "node"}
	}
	if key.OrchestratorID == "" {
		return nil, errors.ErrorInsufficientIdentifiers{Name: "orchestrator"}
	}
	if key.WorkloadID == "" {
		return nil, errors.ErrorInsufficientIdentifiers{Name: "workload"}
	}
	workload := fmt.Sprintf("/calico/v1/host/%s/workload/%s/%s",
		key.Hostname, escapeName(key.OrchestratorID), escapeName(key.WorkloadID))
	endpoints := workload + "/endpoint"
	return []string{endpoints, workload}, nil
}

func (key WorkloadEndpointKey) valueType() (reflect.Type, error) {
	return reflect.TypeOf(WorkloadEndpoint{}), nil
}

func (key WorkloadEndpointKey) String() string {
	return fmt.Sprintf("WorkloadEndpoint(node=%s, orchestrator=%s, workload=%s, name=%s)",
		key.Hostname, key.OrchestratorID, key.WorkloadID, key.EndpointID)
}

type WorkloadEndpointListOptions struct {
	Hostname       string
	OrchestratorID string
	WorkloadID     string
	EndpointID     string
}

func (options WorkloadEndpointListOptions) defaultPathRoot() string {
	k := "/calico/v1/host"
	if options.Hostname == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s/workload", options.Hostname)
	if options.OrchestratorID == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", escapeName(options.OrchestratorID))
	if options.WorkloadID == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s/endpoint", escapeName(options.WorkloadID))
	if options.EndpointID == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", escapeName(options.EndpointID))
	return k
}

func (options WorkloadEndpointListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get WorkloadEndpoint key from %s", path)
	r := matchWorkloadEndpoint.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	hostname := r[0][1]
	orch := unescapeName(r[0][2])
	workload := unescapeName(r[0][3])
	endpointID := unescapeName(r[0][4])
	if options.Hostname != "" && hostname != options.Hostname {
		log.Debugf("Didn't match hostname %s != %s", options.Hostname, hostname)
		return nil
	}
	if options.OrchestratorID != "" && orch != options.OrchestratorID {
		log.Debugf("Didn't match orchestrator %s != %s", options.OrchestratorID, orch)
		return nil
	}
	if options.WorkloadID != "" && workload != options.WorkloadID {
		log.Debugf("Didn't match workload %s != %s", options.WorkloadID, workload)
		return nil
	}
	if options.EndpointID != "" && endpointID != options.EndpointID {
		log.Debugf("Didn't match endpoint ID %s != %s", options.EndpointID, endpointID)
		return nil
	}
	return WorkloadEndpointKey{
		Hostname:       hostname,
		OrchestratorID: orch,
		WorkloadID:     workload,
		EndpointID:     endpointID,
	}
}

type WorkloadEndpoint struct {
	State                      string            `json:"state"`
	Name                       string            `json:"name"`
	ActiveInstanceID           string            `json:"active_instance_id"`
	Mac                        *net.MAC          `json:"mac"`
	ProfileIDs                 []string          `json:"profile_ids"`
	IPv4Nets                   []net.IPNet       `json:"ipv4_nets"`
	IPv6Nets                   []net.IPNet       `json:"ipv6_nets"`
	IPv4NAT                    []IPNAT           `json:"ipv4_nat,omitempty"`
	IPv6NAT                    []IPNAT           `json:"ipv6_nat,omitempty"`
	Labels                     map[string]string `json:"labels,omitempty"`
	IPv4Gateway                *net.IP           `json:"ipv4_gateway,omitempty" validate:"omitempty,ipv4"`
	IPv6Gateway                *net.IP           `json:"ipv6_gateway,omitempty" validate:"omitempty,ipv6"`
	Ports                      []EndpointPort    `json:"ports,omitempty" validate:"dive"`
	GenerateName               string            `json:"generate_name,omitempty"`
	AllowSpoofedSourcePrefixes []net.IPNet       `json:"allow_spoofed_source_ips,omitempty"`
	Annotations                map[string]string `json:"annotations,omitempty"`
}

type EndpointPort struct {
	Name     string               `json:"name" validate:"name"`
	Protocol numorstring.Protocol `json:"protocol"`
	Port     uint16               `json:"port" validate:"gt=0"`
}

// IPNat contains a single NAT mapping for a WorkloadEndpoint resource.
type IPNAT struct {
	// The internal IP address which must be associated with the owning endpoint via the
	// configured IPNetworks for the endpoint.
	IntIP net.IP `json:"int_ip" validate:"ip"`

	// The external IP address.
	ExtIP net.IP `json:"ext_ip" validate:"ip"`
}
