// Copyright (c) 2020 Tigera, Inc. All rights reserved.

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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/api/pkg/lib/numorstring"

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
	"github.com/projectcalico/calico/libcalico-go/lib/net"
)

var (
	matchGlobalBGPPeer        = regexp.MustCompile("^/?calico/bgp/v1/global/peer_v./([^/]+)$")
	matchHostBGPPeer          = regexp.MustCompile("^/?calico/bgp/v1/host/([^/]+)/peer_v./([^/]+)$")
	typeBGPPeer               = reflect.TypeOf(BGPPeer{})
	ipPortSeparator           = "-"
	defaultPort        uint16 = 179
)

type NodeBGPPeerKey struct {
	Nodename string `json:"-" validate:"omitempty"`
	PeerIP   net.IP `json:"-" validate:"required"`
	Port     uint16 `json:"-" validate:"omitempty"`
}

func (key NodeBGPPeerKey) defaultPath() (string, error) {
	if key.PeerIP.IP == nil {
		return "", errors.ErrorInsufficientIdentifiers{Name: "peerIP"}
	}
	if key.Nodename == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "node"}
	}
	e := fmt.Sprintf("/calico/bgp/v1/host/%s/peer_v%d/%s",
		key.Nodename, key.PeerIP.Version(), combineIPAndPort(key.PeerIP, key.Port))
	return e, nil
}

func (key NodeBGPPeerKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key NodeBGPPeerKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key NodeBGPPeerKey) valueType() (reflect.Type, error) {
	return typeBGPPeer, nil
}

func (key NodeBGPPeerKey) String() string {
	return fmt.Sprintf("BGPPeer(node=%s, ip=%s, port=%d)", key.Nodename, key.PeerIP, key.Port)
}

type NodeBGPPeerListOptions struct {
	Nodename string
	PeerIP   net.IP
	Port     uint16
}

func (options NodeBGPPeerListOptions) defaultPathRoot() string {
	if options.Nodename == "" {
		return "/calico/bgp/v1/host"
	} else if options.PeerIP.IP == nil {
		return fmt.Sprintf("/calico/bgp/v1/host/%s",
			options.Nodename)
	} else {
		return fmt.Sprintf("/calico/bgp/v1/host/%s/peer_v%d/%s",
			options.Nodename, options.PeerIP.Version(), combineIPAndPort(options.PeerIP, options.Port))
	}
}

func (options NodeBGPPeerListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get BGPPeer key from %s", path)
	nodename := ""
	var port uint16
	peerIP := net.IP{}
	ekeyb := []byte(path)
	if r := matchHostBGPPeer.FindAllSubmatch(ekeyb, -1); len(r) == 1 {
		var ipBytes []byte
		ipBytes, port = extractIPAndPort(string(r[0][2]))
		nodename = string(r[0][1])
		if err := peerIP.UnmarshalText(ipBytes); err != nil {
			log.WithError(err).WithField("PeerIP", r[0][2]).Error("Error unmarshalling GlobalBGPPeer IP address")
			return nil
		}
	} else {
		log.Debugf("%s didn't match regex", path)
		return nil
	}

	if options.PeerIP.IP != nil && !options.PeerIP.Equal(peerIP.IP) {
		log.Debugf("Didn't match peerIP %s != %s", options.PeerIP.String(), peerIP.String())
		return nil
	}
	if options.Nodename != "" && nodename != options.Nodename {
		log.Debugf("Didn't match hostname %s != %s", options.Nodename, nodename)
		return nil
	}

	if port == 0 {
		return NodeBGPPeerKey{PeerIP: peerIP, Nodename: nodename}
	}
	return NodeBGPPeerKey{PeerIP: peerIP, Nodename: nodename, Port: port}
}

type GlobalBGPPeerKey struct {
	PeerIP net.IP `json:"-" validate:"required"`
	Port   uint16 `json:"-" validate:"omitempty"`
}

func (key GlobalBGPPeerKey) defaultPath() (string, error) {
	if key.PeerIP.IP == nil {
		return "", errors.ErrorInsufficientIdentifiers{Name: "peerIP"}
	}
	e := fmt.Sprintf("/calico/bgp/v1/global/peer_v%d/%s",
		key.PeerIP.Version(), combineIPAndPort(key.PeerIP, key.Port))
	return e, nil
}

func (key GlobalBGPPeerKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key GlobalBGPPeerKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key GlobalBGPPeerKey) valueType() (reflect.Type, error) {
	return typeBGPPeer, nil
}

func (key GlobalBGPPeerKey) String() string {
	return fmt.Sprintf("BGPPeer(global, ip=%s, port=%d)", key.PeerIP, key.Port)
}

type GlobalBGPPeerListOptions struct {
	PeerIP net.IP
	Port   uint16
}

func (options GlobalBGPPeerListOptions) defaultPathRoot() string {
	if options.PeerIP.IP == nil {
		return "/calico/bgp/v1/global"
	} else {
		return fmt.Sprintf("/calico/bgp/v1/global/peer_v%d/%s",
			options.PeerIP.Version(), combineIPAndPort(options.PeerIP, options.Port))
	}
}

func (options GlobalBGPPeerListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get BGPPeer key from %s", path)
	peerIP := net.IP{}
	ekeyb := []byte(path)
	var port uint16

	if r := matchGlobalBGPPeer.FindAllSubmatch(ekeyb, -1); len(r) == 1 {
		var ipBytes []byte
		ipBytes, port = extractIPAndPort(string(r[0][1]))
		if err := peerIP.UnmarshalText(ipBytes); err != nil {
			log.WithError(err).WithField("PeerIP", r[0][1]).Error("Error unmarshalling GlobalBGPPeer IP address")
			return nil
		}
	} else {
		log.Debugf("%s didn't match regex", path)
		return nil
	}

	if options.PeerIP.IP != nil && !options.PeerIP.Equal(peerIP.IP) {
		log.Debugf("Didn't match peerIP %s != %s", options.PeerIP.String(), peerIP.String())
		return nil
	}

	if port == 0 {
		return GlobalBGPPeerKey{PeerIP: peerIP, Port: port}
	}
	return GlobalBGPPeerKey{PeerIP: peerIP, Port: port}
}

type BGPPeer struct {
	// PeerIP is the IP address of the BGP peer.
	PeerIP net.IP `json:"ip"`

	// ASNum is the AS number of the peer.  Note that we write out the
	// value as a string in to the backend, because confd templating
	// converts large uints to float e notation which breaks the BIRD
	// configuration.
	ASNum numorstring.ASNumber `json:"as_num,string"`
}

func extractIPAndPort(ipPort string) ([]byte, uint16) {
	arr := strings.Split(ipPort, ipPortSeparator)
	if len(arr) == 2 {
		port, err := strconv.ParseUint(arr[1], 0, 16)
		if err != nil {
			log.Warningf("Error extracting port. %#v", err)
			return []byte(ipPort), defaultPort
		}
		return []byte(arr[0]), uint16(port)
	}
	return []byte(ipPort), defaultPort
}

func combineIPAndPort(ip net.IP, port uint16) string {
	if port == 0 || port == defaultPort {
		return ip.String()
	} else {
		strPort := strconv.Itoa(int(port))
		return ip.String() + ipPortSeparator + strPort
	}
}
