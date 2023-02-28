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

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
	"github.com/projectcalico/calico/libcalico-go/lib/net"
)

var (
	matchHostEndpoint = regexp.MustCompile("^/?calico/v1/host/([^/]+)/endpoint/([^/]+)$")
	typeHostEndpoint  = reflect.TypeOf(HostEndpoint{})
)

type HostEndpointKey struct {
	Hostname   string `json:"-" validate:"required,hostname"`
	EndpointID string `json:"-" validate:"required,namespacedName"`
}

func (key HostEndpointKey) defaultPath() (string, error) {
	if key.Hostname == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "node"}
	}
	if key.EndpointID == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	e := fmt.Sprintf("/calico/v1/host/%s/endpoint/%s",
		key.Hostname, escapeName(key.EndpointID))
	return e, nil
}

func (key HostEndpointKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key HostEndpointKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key HostEndpointKey) valueType() (reflect.Type, error) {
	return typeHostEndpoint, nil
}

func (key HostEndpointKey) String() string {
	return fmt.Sprintf("HostEndpoint(node=%s, name=%s)", key.Hostname, key.EndpointID)
}

type HostEndpointListOptions struct {
	Hostname   string
	EndpointID string
}

func (options HostEndpointListOptions) defaultPathRoot() string {
	k := "/calico/v1/host"
	if options.Hostname == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s/endpoint", options.Hostname)
	if options.EndpointID == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", escapeName(options.EndpointID))
	return k
}

func (options HostEndpointListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get HostEndpoint key from %s", path)
	r := matchHostEndpoint.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	hostname := r[0][1]
	endpointID := unescapeName(r[0][2])
	if options.Hostname != "" && hostname != options.Hostname {
		log.Debugf("Didn't match hostname %s != %s", options.Hostname, hostname)
		return nil
	}
	if options.EndpointID != "" && endpointID != options.EndpointID {
		log.Debugf("Didn't match endpointID %s != %s", options.EndpointID, endpointID)
		return nil
	}
	return HostEndpointKey{Hostname: hostname, EndpointID: endpointID}
}

type HostEndpoint struct {
	Name              string            `json:"name,omitempty" validate:"omitempty,interface"`
	ExpectedIPv4Addrs []net.IP          `json:"expected_ipv4_addrs,omitempty" validate:"omitempty,dive,ipv4"`
	ExpectedIPv6Addrs []net.IP          `json:"expected_ipv6_addrs,omitempty" validate:"omitempty,dive,ipv6"`
	Labels            map[string]string `json:"labels,omitempty" validate:"omitempty,labels"`
	ProfileIDs        []string          `json:"profile_ids,omitempty" validate:"omitempty,dive,name"`
	Ports             []EndpointPort    `json:"ports,omitempty" validate:"dive"`
}
