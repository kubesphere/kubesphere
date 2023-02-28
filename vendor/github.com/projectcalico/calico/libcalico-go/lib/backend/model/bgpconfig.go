// Copyright (c) 2016 Tigera, Inc. All rights reserved.
//
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
	"reflect"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
)

var (
	matchGlobalBGPConfig = regexp.MustCompile("^/?calico/bgp/v1/global/(.+)$")
	matchNodeBGPConfig   = regexp.MustCompile("^/?calico/bgp/v1/host/([^/]+)/(.+)$")
	typeGlobalBGPConfig  = rawStringType
	typeNodeBGPConfig    = rawStringType
)

type GlobalBGPConfigKey struct {
	// The name of the global BGP config key.
	Name string `json:"-" validate:"required,name"`
}

func (key GlobalBGPConfigKey) defaultPath() (string, error) {
	return key.defaultDeletePath()
}

func (key GlobalBGPConfigKey) defaultDeletePath() (string, error) {
	if key.Name == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	e := fmt.Sprintf("/calico/bgp/v1/global/%s", key.Name)
	return e, nil
}

func (key GlobalBGPConfigKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key GlobalBGPConfigKey) valueType() (reflect.Type, error) {
	return typeGlobalBGPConfig, nil
}

func (key GlobalBGPConfigKey) String() string {
	return fmt.Sprintf("GlobalBGPConfig(name=%s)", key.Name)
}

type GlobalBGPConfigListOptions struct {
	Name string
}

func (options GlobalBGPConfigListOptions) defaultPathRoot() string {
	k := "/calico/bgp/v1/global"
	if options.Name == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", options.Name)
	return k
}

func (options GlobalBGPConfigListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get GlobalFelixConfig key from %s", path)
	r := matchGlobalBGPConfig.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	name := r[0][1]
	if options.Name != "" && name != options.Name {
		log.Debugf("Didn't match name %s != %s", options.Name, name)
		return nil
	}
	return GlobalBGPConfigKey{Name: name}
}

type NodeBGPConfigKey struct {
	// The hostname for the host specific BGP config
	Nodename string `json:"-" validate:"required,name"`

	// The name of the host specific BGP config key.
	Name string `json:"-" validate:"required,name"`
}

func (key NodeBGPConfigKey) defaultPath() (string, error) {
	return key.defaultDeletePath()
}

func (key NodeBGPConfigKey) defaultDeletePath() (string, error) {
	if key.Nodename == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "node"}
	}
	if key.Name == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	e := fmt.Sprintf("/calico/bgp/v1/host/%s/%s", key.Nodename, key.Name)
	return e, nil
}

func (key NodeBGPConfigKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key NodeBGPConfigKey) valueType() (reflect.Type, error) {
	return typeNodeBGPConfig, nil
}

func (key NodeBGPConfigKey) String() string {
	return fmt.Sprintf("HostBGPConfig(node=%s; name=%s)", key.Nodename, key.Name)
}

type NodeBGPConfigListOptions struct {
	Nodename string
	Name     string
}

func (options NodeBGPConfigListOptions) defaultPathRoot() string {
	k := "/calico/bgp/v1/host/%s"
	if options.Nodename == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", options.Nodename)
	if options.Name == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", options.Name)
	return k
}

func (options NodeBGPConfigListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get HostConfig key from %s", path)
	r := matchNodeBGPConfig.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	nodename := r[0][1]
	name := r[0][2]
	if options.Nodename != "" && nodename != options.Nodename {
		log.Debugf("Didn't match nodename %s != %s", options.Nodename, nodename)
		return nil
	}
	if options.Name != "" && name != options.Name {
		log.Debugf("Didn't match name %s != %s", options.Name, name)
		return nil
	}
	return NodeBGPConfigKey{Nodename: nodename, Name: name}
}
