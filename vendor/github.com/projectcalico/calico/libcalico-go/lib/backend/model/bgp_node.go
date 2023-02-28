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

package model

import (
	"fmt"
	"reflect"

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
)

var (
	typeBGPNode = reflect.TypeOf(BGPNode{})
)

type BGPNodeKey struct {
	Host string
}

func (key BGPNodeKey) defaultPath() (string, error) {
	if key.Host == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "host"}
	}

	k := "/calico/bgp/v1/host/" + key.Host
	return k, nil
}

func (key BGPNodeKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key BGPNodeKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key BGPNodeKey) valueType() (reflect.Type, error) {
	return typeBGPNode, nil
}

func (key BGPNodeKey) String() string {
	return fmt.Sprintf("BGPNodeKey(host=%s)", key.Host)
}

type BGPNode struct {
}
