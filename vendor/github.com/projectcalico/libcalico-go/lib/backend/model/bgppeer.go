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

	"github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/net"
	"github.com/projectcalico/libcalico-go/lib/numorstring"
	log "github.com/sirupsen/logrus"
)

var (
	matchGlobalBGPPeer = regexp.MustCompile("^/?calico/bgp/v1/global/peer_v./([^/]+)$")
	matchHostBGPPeer   = regexp.MustCompile("^/?calico/bgp/v1/host/([^/]+)/peer_v./([^/]+)$")
	typeBGPPeer        = reflect.TypeOf(BGPPeer{})
)

type NodeBGPPeerKey struct {
	Nodename string `json:"-" validate:"omitempty"`
	PeerIP   net.IP `json:"-" validate:"required"`
}

func (key NodeBGPPeerKey) defaultPath() (string, error) {
	if key.PeerIP.IP == nil {
		return "", errors.ErrorInsufficientIdentifiers{Name: "peerIP"}
	}
	if key.Nodename == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "node"}
	}
	e := fmt.Sprintf("/calico/bgp/v1/host/%s/peer_v%d/%s",
		key.Nodename, key.PeerIP.Version(), key.PeerIP)
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
	return fmt.Sprintf("BGPPeer(node=%s, ip=%s)", key.Nodename, key.PeerIP)
}

type NodeBGPPeerListOptions struct {
	Nodename string
	PeerIP   net.IP
}

func (options NodeBGPPeerListOptions) defaultPathRoot() string {
	if options.Nodename == "" {
		return "/calico/bgp/v1/host"
	} else if options.PeerIP.IP == nil {
		return fmt.Sprintf("/calico/bgp/v1/host/%s",
			options.Nodename)
	} else {
		return fmt.Sprintf("/calico/bgp/v1/host/%s/peer_v%d/%s",
			options.Nodename, options.PeerIP.Version(), options.PeerIP)
	}
}

func (options NodeBGPPeerListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get BGPPeer key from %s", path)
	nodename := ""
	peerIP := net.IP{}
	ekeyb := []byte(path)

	if r := matchHostBGPPeer.FindAllSubmatch(ekeyb, -1); len(r) == 1 {
		nodename = string(r[0][1])
		if err := peerIP.UnmarshalText(r[0][2]); err != nil {
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
	return NodeBGPPeerKey{PeerIP: peerIP, Nodename: nodename}
}

type GlobalBGPPeerKey struct {
	PeerIP net.IP `json:"-" validate:"required"`
}

func (key GlobalBGPPeerKey) defaultPath() (string, error) {
	if key.PeerIP.IP == nil {
		return "", errors.ErrorInsufficientIdentifiers{Name: "peerIP"}
	}
	e := fmt.Sprintf("/calico/bgp/v1/global/peer_v%d/%s",
		key.PeerIP.Version(), key.PeerIP)
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
	return fmt.Sprintf("BGPPeer(global, ip=%s)", key.PeerIP)
}

type GlobalBGPPeerListOptions struct {
	PeerIP net.IP
}

func (options GlobalBGPPeerListOptions) defaultPathRoot() string {
	if options.PeerIP.IP == nil {
		return "/calico/bgp/v1/global"
	} else {
		return fmt.Sprintf("/calico/bgp/v1/global/peer_v%d/%s",
			options.PeerIP.Version(), options.PeerIP)
	}
}

func (options GlobalBGPPeerListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get BGPPeer key from %s", path)
	peerIP := net.IP{}
	ekeyb := []byte(path)

	if r := matchGlobalBGPPeer.FindAllSubmatch(ekeyb, -1); len(r) == 1 {
		if err := peerIP.UnmarshalText(r[0][1]); err != nil {
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
	return GlobalBGPPeerKey{PeerIP: peerIP}
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
