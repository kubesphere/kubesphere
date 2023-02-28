// Copyright (c) 2016-2021 Tigera, Inc. All rights reserved.

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
	"math/big"
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
	"github.com/projectcalico/calico/libcalico-go/lib/net"
)

const (
	// Common attributes which may be set on allocations by clients.
	IPAMBlockAttributePod             = "pod"
	IPAMBlockAttributeNamespace       = "namespace"
	IPAMBlockAttributeNode            = "node"
	IPAMBlockAttributeType            = "type"
	IPAMBlockAttributeTypeIPIP        = "ipipTunnelAddress"
	IPAMBlockAttributeTypeVXLAN       = "vxlanTunnelAddress"
	IPAMBlockAttributeTypeVXLANV6     = "vxlanV6TunnelAddress"
	IPAMBlockAttributeTypeWireguard   = "wireguardTunnelAddress"
	IPAMBlockAttributeTypeWireguardV6 = "wireguardV6TunnelAddress"
	IPAMBlockAttributeTimestamp       = "timestamp"
)

var (
	matchBlock = regexp.MustCompile("^/?calico/ipam/v2/assignment/ipv./block/([^/]+)$")
	typeBlock  = reflect.TypeOf(AllocationBlock{})
)

type BlockKey struct {
	CIDR net.IPNet `json:"-" validate:"required,name"`
}

func (key BlockKey) defaultPath() (string, error) {
	if key.CIDR.IP == nil {
		return "", errors.ErrorInsufficientIdentifiers{}
	}
	c := strings.Replace(key.CIDR.String(), "/", "-", 1)
	e := fmt.Sprintf("/calico/ipam/v2/assignment/ipv%d/block/%s", key.CIDR.Version(), c)
	return e, nil
}

func (key BlockKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key BlockKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key BlockKey) valueType() (reflect.Type, error) {
	return typeBlock, nil
}

func (key BlockKey) String() string {
	return fmt.Sprintf("BlockKey(cidr=%s)", key.CIDR.String())
}

type BlockListOptions struct {
	IPVersion int `json:"-"`
}

func (options BlockListOptions) defaultPathRoot() string {
	k := "/calico/ipam/v2/assignment/"
	if options.IPVersion != 0 {
		k = k + fmt.Sprintf("ipv%d/", options.IPVersion)
	}
	return k
}

func (options BlockListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get Block key from %s", path)
	r := matchBlock.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("%s didn't match regex", path)
		return nil
	}
	cidrStr := strings.Replace(r[0][1], "-", "/", 1)
	_, cidr, err := net.ParseCIDR(cidrStr)
	if err != nil {
		log.Debugf("find an invalid cidr %s for path=%v , info=%v ", r[0][1], path, err)
		return nil
	}
	return BlockKey{CIDR: *cidr}
}

type AllocationBlock struct {
	// The block's CIDR.
	CIDR net.IPNet `json:"cidr"`

	// Affinity of the block, if this block has one. If set, it will be of the form
	// "host:<hostname>". If not set, this block is not affine to a host.
	Affinity *string `json:"affinity"`

	// Array of allocations in-use within this block. nil entries mean the allocation is free.
	// For non-nil entries at index i, the index is the ordinal of the allocation within this block
	// and the value is the index of the associated attributes in the Attributes array.
	Allocations []*int `json:"allocations"`

	// Unallocated is an ordered list of allocations which are free in the block.
	Unallocated []int `json:"unallocated"`

	// Attributes is an array of arbitrary metadata associated with allocations in the block. To find
	// attributes for a given allocation, use the value of the allocation's entry in the Allocations array
	// as the index of the element in this array.
	Attributes []AllocationAttribute `json:"attributes"`

	// We store a sequence number that is updated each time the block is written.
	// Each allocation will also store the sequence number of the block at the time of its creation.
	// When releasing an IP, passing the sequence number associated with the allocation allows us
	// to protect against a race condition and ensure the IP hasn't been released and re-allocated
	// since the release request.
	SequenceNumber uint64 `json:"sequenceNumber"`

	// Map of allocated ordinal within the block to sequence number of the block at
	// the time of allocation. Kubernetes does not allow numerical keys for maps, so
	// the key is cast to a string.
	SequenceNumberForAllocation map[string]uint64 `json:"sequenceNumberForAllocation"`

	// Deleted is an internal boolean used to workaround a limitation in the Kubernetes API whereby
	// deletion will not return a conflict error if the block has been updated.
	Deleted bool `json:"deleted"`

	// HostAffinity is deprecated in favor of Affinity.
	// This is only to keep compatibility with existing deployments.
	// The data format should be `Affinity: host:hostname` (not `hostAffinity: hostname`).
	HostAffinity *string `json:"hostAffinity,omitempty"`
}

func (b *AllocationBlock) SetSequenceNumberForOrdinal(ordinal int) {
	if b.SequenceNumberForAllocation == nil {
		b.SequenceNumberForAllocation = map[string]uint64{}
	}
	b.SequenceNumberForAllocation[fmt.Sprintf("%d", ordinal)] = b.SequenceNumber
}

func (b *AllocationBlock) GetSequenceNumberForOrdinal(ordinal int) uint64 {
	return b.SequenceNumberForAllocation[fmt.Sprintf("%d", ordinal)]
}

func (b *AllocationBlock) ClearSequenceNumberForOrdinal(ordinal int) {
	delete(b.SequenceNumberForAllocation, fmt.Sprintf("%d", ordinal))
}

func (b *AllocationBlock) MarkDeleted() {
	b.Deleted = true
}

func (b *AllocationBlock) IsDeleted() bool {
	return b.Deleted
}

func (b *AllocationBlock) Host() string {
	if b.Affinity != nil && strings.HasPrefix(*b.Affinity, "host:") {
		return strings.TrimPrefix(*b.Affinity, "host:")
	}
	return ""
}

type Allocation struct {
	Addr net.IP
	Host string
}

func (b *AllocationBlock) NonAffineAllocations() []Allocation {
	var allocs []Allocation
	myHost := b.Host()
	for ordinal, attrIdx := range b.Allocations {
		if attrIdx == nil {
			continue // Skip unallocated IPs.
		}
		if *attrIdx >= len(b.Attributes) {
			log.WithField("block", b).Warnf("Missing attributes for IP with ordinal %d", ordinal)
			continue
		}
		attrs := b.Attributes[*attrIdx]
		host := attrs.AttrSecondary[IPAMBlockAttributeNode]
		if myHost != "" && host == myHost {
			continue // Skip allocations that are affine to this block.
		}
		a := Allocation{
			Addr: b.OrdinalToIP(ordinal),
			Host: host,
		}
		allocs = append(allocs, a)
	}
	return allocs
}

// Get number of addresses covered by the block
func (b *AllocationBlock) NumAddresses() int {
	ones, size := b.CIDR.Mask.Size()
	numAddresses := 1 << uint(size-ones)
	return numAddresses
}

// Find the ordinal (i.e. how far into the block) a given IP lies.  Returns an error if the IP is outside the block.
func (b *AllocationBlock) IPToOrdinal(ip net.IP) (int, error) {
	ipAsInt := net.IPToBigInt(ip)
	baseInt := net.IPToBigInt(net.IP{b.CIDR.IP})
	ord := big.NewInt(0).Sub(ipAsInt, baseInt).Int64()
	if ord < 0 || ord >= int64(b.NumAddresses()) {
		return 0, fmt.Errorf("IP %s not in block %s", ip, b.CIDR)
	}
	return int(ord), nil
}

// Calculates the IP at the given position within the block.  ord=0 gives the first IP in the block.
func (b *AllocationBlock) OrdinalToIP(ord int) net.IP {
	return b.CIDR.NthIP(ord)
}

type AllocationAttribute struct {
	AttrPrimary   *string           `json:"handle_id"`
	AttrSecondary map[string]string `json:"secondary"`
}
