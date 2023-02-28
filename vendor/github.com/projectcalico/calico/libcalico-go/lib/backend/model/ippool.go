// Copyright (c) 2016,2021 Tigera, Inc. All rights reserved.

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

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/calico/libcalico-go/lib/backend/encap"
	"github.com/projectcalico/calico/libcalico-go/lib/errors"
	"github.com/projectcalico/calico/libcalico-go/lib/net"
)

var (
	matchIPPool = regexp.MustCompile("^/?calico/v1/ipam/v./pool/([^/]+)$")
	typeIPPool  = reflect.TypeOf(IPPool{})
)

type IPPoolKey struct {
	CIDR net.IPNet `json:"-" validate:"required,name"`
}

func (key IPPoolKey) defaultPath() (string, error) {
	if key.CIDR.IP == nil {
		return "", errors.ErrorInsufficientIdentifiers{Name: "cidr"}
	}
	c := strings.Replace(key.CIDR.String(), "/", "-", 1)
	e := fmt.Sprintf("/calico/v1/ipam/v%d/pool/%s", key.CIDR.Version(), c)
	return e, nil
}

func (key IPPoolKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key IPPoolKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key IPPoolKey) valueType() (reflect.Type, error) {
	return typeIPPool, nil
}

func (key IPPoolKey) String() string {
	return fmt.Sprintf("IPPool(cidr=%s)", key.CIDR)
}

type IPPoolListOptions struct {
	CIDR net.IPNet
}

func (options IPPoolListOptions) defaultPathRoot() string {
	k := "/calico/v1/ipam/"
	if options.CIDR.IP == nil {
		return k
	}
	c := strings.Replace(options.CIDR.String(), "/", "-", 1)
	k = k + fmt.Sprintf("v%d/pool/", options.CIDR.Version()) + fmt.Sprintf("%s", c)
	return k
}

func (options IPPoolListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get Pool key from %s", path)
	r := matchIPPool.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("%s didn't match regex", path)
		return nil
	}
	cidrStr := strings.Replace(r[0][1], "-", "/", 1)
	_, cidr, err := net.ParseCIDR(cidrStr)
	if err != nil {
		log.WithError(err).Warningf("Failed to parse CIDR %s", cidrStr)
		return nil
	}
	if options.CIDR.IP != nil && !reflect.DeepEqual(*cidr, options.CIDR) {
		log.Debugf("Didn't match cidr %s != %s", options.CIDR.String(), cidr.String())
		return nil
	}
	return IPPoolKey{CIDR: *cidr}
}

type IPPool struct {
	CIDR             net.IPNet  `json:"cidr"`
	IPIPInterface    string     `json:"ipip"`
	IPIPMode         encap.Mode `json:"ipip_mode"`
	VXLANMode        encap.Mode `json:"vxlan_mode"`
	Masquerade       bool       `json:"masquerade"`
	IPAM             bool       `json:"ipam"`
	Disabled         bool       `json:"disabled"`
	DisableBGPExport bool       `json:"disableBGPExport"`
}
