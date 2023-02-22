// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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
)

var (
	matchHostEndpointStatus = regexp.MustCompile("^/?calico/felix/v1/host/([^/]+)/endpoint/([^/]+)$")
	typeHostEndpointStatus  = reflect.TypeOf(HostEndpointStatus{})
)

type HostEndpointStatusKey struct {
	Hostname   string `json:"-" validate:"required,hostname"`
	EndpointID string `json:"-" validate:"required,namespacedName"`
}

func (key HostEndpointStatusKey) defaultPath() (string, error) {
	if key.Hostname == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "node"}
	}
	if key.EndpointID == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	e := fmt.Sprintf("/calico/felix/v1/host/%s/endpoint/%s",
		key.Hostname, escapeName(key.EndpointID))
	return e, nil
}

func (key HostEndpointStatusKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key HostEndpointStatusKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key HostEndpointStatusKey) valueType() (reflect.Type, error) {
	return typeHostEndpointStatus, nil
}

func (key HostEndpointStatusKey) String() string {
	return fmt.Sprintf("HostEndpointStatus(hostname=%s, name=%s)", key.Hostname, key.EndpointID)
}

type HostEndpointStatusListOptions struct {
	Hostname   string
	EndpointID string
}

func (options HostEndpointStatusListOptions) defaultPathRoot() string {
	k := "/calico/felix/v1/host"
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

func (options HostEndpointStatusListOptions) KeyFromDefaultPath(ekey string) Key {
	log.Debugf("Get HostEndpointStatus key from %s", ekey)
	r := matchHostEndpointStatus.FindAllStringSubmatch(ekey, -1)
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
	return HostEndpointStatusKey{Hostname: hostname, EndpointID: endpointID}
}

type HostEndpointStatus struct {
	Status string `json:"status"`
}
