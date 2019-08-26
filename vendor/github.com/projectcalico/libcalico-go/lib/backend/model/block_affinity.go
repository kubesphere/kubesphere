// Copyright (c) 2016-2019 Tigera, Inc. All rights reserved.

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
	"strings"

	"github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/net"
	log "github.com/sirupsen/logrus"
)

var (
	matchBlockAffinity = regexp.MustCompile("^/?calico/ipam/v2/host/([^/]+)/ipv./block/([^/]+)$")
	typeBlockAff       = reflect.TypeOf(BlockAffinity{})
)

type BlockAffinityState string

const (
	StateConfirmed       BlockAffinityState = "confirmed"
	StatePending         BlockAffinityState = "pending"
	StatePendingDeletion BlockAffinityState = "pendingDeletion"
)

type BlockAffinityKey struct {
	CIDR net.IPNet `json:"-" validate:"required,name"`
	Host string    `json:"-"`
}

type BlockAffinity struct {
	State   BlockAffinityState `json:"state"`
	Deleted bool               `json:"deleted"`
}

func (key BlockAffinityKey) defaultPath() (string, error) {
	if key.CIDR.IP == nil || key.Host == "" {
		return "", errors.ErrorInsufficientIdentifiers{}
	}
	c := strings.Replace(key.CIDR.String(), "/", "-", 1)
	e := fmt.Sprintf("/calico/ipam/v2/host/%s/ipv%d/block/%s", key.Host, key.CIDR.Version(), c)
	return e, nil
}

func (key BlockAffinityKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key BlockAffinityKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key BlockAffinityKey) valueType() (reflect.Type, error) {
	return typeBlockAff, nil
}

func (key BlockAffinityKey) String() string {
	return fmt.Sprintf("BlockAffinityKey(cidr=%s, host=%s)", key.CIDR, key.Host)
}

type BlockAffinityListOptions struct {
	Host      string
	IPVersion int
}

func (options BlockAffinityListOptions) defaultPathRoot() string {
	k := "/calico/ipam/v2/host/"
	if options.Host != "" {
		k = k + options.Host

		if options.IPVersion != 0 {
			k = k + fmt.Sprintf("/ipv%d/block", options.IPVersion)
		}
	}
	return k
}

func (options BlockAffinityListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get Block affinity key from %s", path)
	r := matchBlockAffinity.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("%s didn't match regex", path)
		return nil
	}
	cidrStr := strings.Replace(r[0][2], "-", "/", 1)
	_, cidr, _ := net.ParseCIDR(cidrStr)
	host := r[0][1]

	if options.Host != "" && options.Host != host {
		log.Debugf("Didn't match hostname: %s != %s", options.Host, host)
		return nil
	}
	if options.IPVersion != 0 && options.IPVersion != cidr.Version() {
		log.Debugf("Didn't match IP version. %d != %d", options.IPVersion, cidr.Version())
		return nil
	}
	return BlockAffinityKey{CIDR: *cidr, Host: host}
}
