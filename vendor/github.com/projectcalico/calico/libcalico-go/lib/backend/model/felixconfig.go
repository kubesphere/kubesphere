// Copyright (c) 2016-2018 Tigera, Inc. All rights reserved.
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
	matchGlobalConfig = regexp.MustCompile("^/?calico/v1/config/(.+)$")
	matchHostConfig   = regexp.MustCompile("^/?calico/v1/host/([^/]+)/config/(.+)$")
	matchReadyFlag    = regexp.MustCompile("^/calico/v1/Ready$")
	typeGlobalConfig  = rawStringType
	typeHostConfig    = rawStringType
	typeReadyFlag     = rawBoolType
)

type ReadyFlagKey struct {
}

func (key ReadyFlagKey) defaultPath() (string, error) {
	return "/calico/v1/Ready", nil
}

func (key ReadyFlagKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key ReadyFlagKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key ReadyFlagKey) valueType() (reflect.Type, error) {
	return typeReadyFlag, nil
}

func (key ReadyFlagKey) String() string {
	return "ReadyFlagKey()"
}

type GlobalConfigKey struct {
	Name string `json:"-" validate:"required,name"`
}

func (key GlobalConfigKey) defaultPath() (string, error) {
	if key.Name == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	e := fmt.Sprintf("/calico/v1/config/%s", key.Name)
	return e, nil
}

func (key GlobalConfigKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key GlobalConfigKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key GlobalConfigKey) valueType() (reflect.Type, error) {
	return typeGlobalConfig, nil
}

func (key GlobalConfigKey) String() string {
	return fmt.Sprintf("GlobalFelixConfig(name=%s)", key.Name)
}

type GlobalConfigListOptions struct {
	Name string
}

func (options GlobalConfigListOptions) defaultPathRoot() string {
	k := "/calico/v1/config"
	if options.Name == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", options.Name)
	return k
}

func (options GlobalConfigListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get GlobalConfig key from %s", path)
	r := matchGlobalConfig.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	name := r[0][1]
	if options.Name != "" && name != options.Name {
		log.Debugf("Didn't match name %s != %s", options.Name, name)
		return nil
	}
	return GlobalConfigKey{Name: name}
}

type HostConfigKey struct {
	Hostname string `json:"-" validate:"required,name"`
	Name     string `json:"-" validate:"required,name"`
}

func (key HostConfigKey) defaultPath() (string, error) {
	if key.Name == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	if key.Hostname == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "node"}
	}
	e := fmt.Sprintf("/calico/v1/host/%s/config/%s", key.Hostname, key.Name)
	return e, nil
}

func (key HostConfigKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key HostConfigKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key HostConfigKey) valueType() (reflect.Type, error) {
	return typeHostConfig, nil
}

func (key HostConfigKey) String() string {
	return fmt.Sprintf("HostConfig(node=%s,name=%s)", key.Hostname, key.Name)
}

type HostConfigListOptions struct {
	Hostname string
	Name     string
}

func (options HostConfigListOptions) defaultPathRoot() string {
	k := "/calico/v1/host"
	if options.Hostname == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s/config", options.Hostname)
	if options.Name == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", options.Name)
	return k
}

func (options HostConfigListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get HostConfig key from %s", path)
	r := matchHostConfig.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	hostname := r[0][1]
	name := r[0][2]
	if options.Hostname != "" && hostname != options.Hostname {
		log.Debugf("Didn't match hostname %s != %s", options.Hostname, hostname)
		return nil
	}
	if options.Name != "" && name != options.Name {
		log.Debugf("Didn't match name %s != %s", options.Name, name)
		return nil
	}
	return HostConfigKey{Hostname: hostname, Name: name}
}
