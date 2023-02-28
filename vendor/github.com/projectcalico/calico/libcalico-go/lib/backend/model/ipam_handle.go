// Copyright (c) 2016,2020 Tigera, Inc. All rights reserved.

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
	matchHandle = regexp.MustCompile("^/?calico/ipam/v2/handle/([^/]+)$")
	typeHandle  = reflect.TypeOf(IPAMHandle{})
)

type IPAMHandleKey struct {
	HandleID string `json:"id"`
}

func (key IPAMHandleKey) defaultPath() (string, error) {
	if key.HandleID == "" {
		return "", errors.ErrorInsufficientIdentifiers{}
	}
	e := fmt.Sprintf("/calico/ipam/v2/handle/%s", key.HandleID)
	return e, nil
}

func (key IPAMHandleKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key IPAMHandleKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key IPAMHandleKey) valueType() (reflect.Type, error) {
	return typeHandle, nil
}

func (key IPAMHandleKey) String() string {
	return fmt.Sprintf("IPAMHandleKey(id=%s)", key.HandleID)
}

type IPAMHandleListOptions struct {
	// TODO: Have some options here?
}

func (options IPAMHandleListOptions) defaultPathRoot() string {
	k := "/calico/ipam/v2/handle/"
	// TODO: Allow filtering on individual host?
	return k
}

func (options IPAMHandleListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get IPAM handle key from %s", path)
	r := matchHandle.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("%s didn't match regex", path)
		return nil
	}
	return IPAMHandleKey{HandleID: r[0][1]}
}

type IPAMHandle struct {
	HandleID string         `json:"-"`
	Block    map[string]int `json:"block"`
	Deleted  bool           `json:"deleted"`
}
