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
	goerrors "errors"
	"fmt"
	"regexp"

	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/api/pkg/lib/numorstring"

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
	"github.com/projectcalico/calico/libcalico-go/lib/net"
)

var (
	typeNode          = reflect.TypeOf(Node{})
	typeHostMetadata  = reflect.TypeOf(HostMetadata{})
	typeOrchRefs      = reflect.TypeOf([]OrchRef{})
	typeHostIp        = rawIPType
	typeWireguard     = reflect.TypeOf(Wireguard{})
	matchHostMetadata = regexp.MustCompile(`^/?calico/v1/host/([^/]+)/metadata$`)
	matchHostIp       = regexp.MustCompile(`^/?calico/v1/host/([^/]+)/bird_ip$`)
	matchWireguard    = regexp.MustCompile(`^/?calico/v1/host/([^/]+)/wireguard$`)
)

type Node struct {
	// Felix specific configuration
	FelixIPv4 *net.IP

	// Node specific labels
	Labels map[string]string `json:"labels,omitempty"`

	// BGP specific configuration
	BGPIPv4Addr *net.IP
	BGPIPv6Addr *net.IP
	BGPIPv4Net  *net.IPNet
	BGPIPv6Net  *net.IPNet
	BGPASNumber *numorstring.ASNumber
	OrchRefs    []OrchRef `json:"orchRefs,omitempty"`
}

type OrchRef struct {
	Orchestrator string `json:"orchestrator,omitempty"`
	NodeName     string `json:"nodeName,omitempty"`
}

type Wireguard struct {
	InterfaceIPv4Addr *net.IP `json:"interfaceIPv4Addr,omitempty"`
	PublicKey         string  `json:"publicKey,omitempty"`
	InterfaceIPv6Addr *net.IP `json:"interfaceIPv6Addr,omitempty"`
	PublicKeyV6       string  `json:"publicKeyV6,omitempty"`
}

type NodeKey struct {
	Hostname string
}

func (key NodeKey) defaultPath() (string, error) {
	return "", goerrors.New("Node is a composite type, so not handled with a single path")
}

func (key NodeKey) defaultDeletePath() (string, error) {
	return "", goerrors.New("Node is a composite type, so not handled with a single path")
}

func (key NodeKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, goerrors.New("Node is composite type, so not handled with a single path")
}

func (key NodeKey) valueType() (reflect.Type, error) {
	return typeNode, nil
}

func (key NodeKey) String() string {
	return fmt.Sprintf("Node(name=%s)", key.Hostname)
}

type NodeListOptions struct {
	Hostname string
}

func (options NodeListOptions) defaultPathRoot() string {
	return ""
}

func (options NodeListOptions) KeyFromDefaultPath(path string) Key {
	return nil
}

// The node is a composite of the following subcomponents:
// -  The host metadata.  This is the primary subcomponent and is used to enumerate
//    hosts.  However, for backwards compatibility, the etcd driver needs to handle
//    that this may not exist, and instead need to enumerate based on directory.
// -  The host IPv4 address used by Calico to lock down IPIP traffic.
// -  The BGP IPv4 and IPv6 addresses
// -  The BGP ASN.

type HostMetadata struct {
}

type HostMetadataKey struct {
	Hostname string
}

func (key HostMetadataKey) defaultPath() (string, error) {
	if key.Hostname == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	return fmt.Sprintf("/calico/v1/host/%s/metadata", key.Hostname), nil
}

func (key HostMetadataKey) defaultDeletePath() (string, error) {
	if key.Hostname == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	return fmt.Sprintf("/calico/v1/host/%s", key.Hostname), nil
}

func (key HostMetadataKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key HostMetadataKey) valueType() (reflect.Type, error) {
	return typeHostMetadata, nil
}

func (key HostMetadataKey) String() string {
	return fmt.Sprintf("Node(name=%s)", key.Hostname)
}

type HostMetadataListOptions struct {
	Hostname string
}

func (options HostMetadataListOptions) defaultPathRoot() string {
	if options.Hostname == "" {
		return "/calico/v1/host"
	} else {
		return fmt.Sprintf("/calico/v1/host/%s/metadata", options.Hostname)
	}
}

func (options HostMetadataListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get Node key from %s", path)
	if r := matchHostMetadata.FindAllStringSubmatch(path, -1); len(r) == 1 {
		return HostMetadataKey{Hostname: r[0][1]}
	} else {
		log.Debugf("%s didn't match regex", path)
		return nil
	}
}

// The Felix Host IP Key.
type HostIPKey struct {
	Hostname string
}

func (key HostIPKey) defaultPath() (string, error) {
	return fmt.Sprintf("/calico/v1/host/%s/bird_ip",
		key.Hostname), nil
}

func (key HostIPKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key HostIPKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key HostIPKey) valueType() (reflect.Type, error) {
	return typeHostIp, nil
}

func (key HostIPKey) String() string {
	return fmt.Sprintf("Node(name=%s)", key.Hostname)
}

type OrchRefKey struct {
	Hostname string
}

func (key OrchRefKey) defaultPath() (string, error) {
	return fmt.Sprintf("/calico/v1/host/%s/orchestrator_refs",
		key.Hostname), nil
}

func (key OrchRefKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key OrchRefKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key OrchRefKey) valueType() (reflect.Type, error) {
	return typeOrchRefs, nil
}

func (key OrchRefKey) String() string {
	return fmt.Sprintf("OrchRefs(nodename=%s)", key.Hostname)
}

type OrchRefListOptions struct {
	Hostname string
}

func (options OrchRefListOptions) defaultPathRoot() string {
	return fmt.Sprintf("/calico/v1/host/%s/orchestrator_refs", options.Hostname)
}

func (options OrchRefListOptions) KeyFromDefaultPath(path string) Key {
	return OrchRefKey{Hostname: options.Hostname}
}

// The Felix Wireguard Key.
type WireguardKey struct {
	NodeName string
}

func (key WireguardKey) defaultPath() (string, error) {
	if key.NodeName == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	return fmt.Sprintf("/calico/v1/host/%s/wireguard",
		key.NodeName), nil
}

func (key WireguardKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key WireguardKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key WireguardKey) valueType() (reflect.Type, error) {
	return typeWireguard, nil
}

func (key WireguardKey) String() string {
	return fmt.Sprintf("Node(nodename=%s)", key.NodeName)
}

type WireguardListOptions struct {
	NodeName string
}

func (options WireguardListOptions) defaultPathRoot() string {
	if options.NodeName == "" {
		return "/calico/v1/host"
	} else {
		return fmt.Sprintf("/calico/v1/host/%s/wireguard", options.NodeName)
	}
}

func (options WireguardListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get Node key from %s", path)
	if r := matchWireguard.FindAllStringSubmatch(path, -1); len(r) == 1 {
		return WireguardKey{NodeName: r[0][1]}
	} else {
		log.Debugf("%s didn't match regex", path)
		return nil
	}
}
