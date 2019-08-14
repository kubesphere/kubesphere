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

package ipam

import (
	"errors"
	"fmt"
	"math/big"
	"net"
	"reflect"
	"strings"

	"github.com/projectcalico/libcalico-go/lib/apis/v3"
	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
)

// Wrap the backend AllocationBlock struct so that we can
// attach methods to it.
type allocationBlock struct {
	*model.AllocationBlock
}

func newBlock(cidr cnet.IPNet) allocationBlock {
	ones, size := cidr.Mask.Size()
	numAddresses := 1 << uint(size-ones)
	b := model.AllocationBlock{}
	b.Allocations = make([]*int, numAddresses)
	b.Unallocated = make([]int, numAddresses)
	b.StrictAffinity = false
	b.CIDR = cidr

	// Initialize unallocated ordinals.
	for i := 0; i < numAddresses; i++ {
		b.Unallocated[i] = i
	}

	return allocationBlock{&b}
}

func (b *allocationBlock) autoAssign(
	num int, handleID *string, host string, attrs map[string]string, affinityCheck bool) ([]cnet.IPNet, error) {

	// Determine if we need to check for affinity.
	checkAffinity := b.StrictAffinity || affinityCheck
	if checkAffinity && b.Affinity != nil && !hostAffinityMatches(host, b.AllocationBlock) {
		// Affinity check is enabled but the host does not match - error.
		s := fmt.Sprintf("Block affinity (%s) does not match provided (%s)", *b.Affinity, host)
		return nil, errors.New(s)
	} else if b.Affinity == nil {
		log.Warnf("Attempting to assign IPs from block with no affinity: %v", b)
		if checkAffinity {
			// If we're checking strict affinity, we can't assign from a block with no affinity.
			return nil, fmt.Errorf("Attempt to assign from block %v with no affinity", b.CIDR)
		}
	}

	// Walk the allocations until we find enough addresses.
	ordinals := []int{}
	for len(b.Unallocated) > 0 && len(ordinals) < num {
		ordinals = append(ordinals, b.Unallocated[0])
		b.Unallocated = b.Unallocated[1:]
	}

	// Create slice of IPs and perform the allocations.
	ips := []cnet.IPNet{}
	_, mask, _ := cnet.ParseCIDR(b.CIDR.String())
	for _, o := range ordinals {
		attrIndex := b.findOrAddAttribute(handleID, attrs)
		b.Allocations[o] = &attrIndex
		ipNets := cnet.IPNet(*mask)
		ipNets.IP = cnet.IncrementIP(cnet.IP{b.CIDR.IP}, big.NewInt(int64(o))).IP
		ips = append(ips, ipNets)
	}

	log.Debugf("Block %s returned ips: %v", b.CIDR.String(), ips)
	return ips, nil
}

func (b *allocationBlock) assign(address cnet.IP, handleID *string, attrs map[string]string, host string) error {
	if b.StrictAffinity && b.Affinity != nil && !hostAffinityMatches(host, b.AllocationBlock) {
		// Affinity check is enabled but the host does not match - error.
		return errors.New("Block host affinity does not match")
	} else if b.Affinity == nil {
		log.Warnf("Attempting to assign IP from block with no affinity: %v", b)
		if b.StrictAffinity {
			// If we're checking strict affinity, we can't assign from a block with no affinity.
			return fmt.Errorf("Attempt to assign from block %v with no affinity", b.CIDR)
		}
	}

	// Convert to an ordinal.
	ordinal, err := b.IPToOrdinal(address)
	if err != nil {
		return err
	}

	// Check if already allocated.
	if b.Allocations[ordinal] != nil {
		return errors.New("Address already assigned in block")
	}

	// Set up attributes.
	attrIndex := b.findOrAddAttribute(handleID, attrs)
	b.Allocations[ordinal] = &attrIndex

	// Remove from unallocated.
	for i, unallocated := range b.Unallocated {
		if unallocated == ordinal {
			b.Unallocated = append(b.Unallocated[:i], b.Unallocated[i+1:]...)
			break
		}
	}
	return nil
}

// hostAffinityMatches checks if the provided host matches the provided affinity.
func hostAffinityMatches(host string, block *model.AllocationBlock) bool {
	return *block.Affinity == "host:"+host
}

func getHostAffinity(block *model.AllocationBlock) string {
	if block.Affinity != nil && strings.HasPrefix(*block.Affinity, "host:") {
		return strings.TrimPrefix(*block.Affinity, "host:")
	}
	return ""
}

func (b allocationBlock) numFreeAddresses() int {
	return len(b.Unallocated)
}

func (b allocationBlock) empty() bool {
	return b.numFreeAddresses() == b.NumAddresses()
}

func (b *allocationBlock) release(addresses []cnet.IP) ([]cnet.IP, map[string]int, error) {
	// Store return values.
	unallocated := []cnet.IP{}
	countByHandle := map[string]int{}

	// Used internally.
	var ordinals []int
	delRefCounts := map[int]int{}
	attrsToDelete := []int{}

	// De-duplicate addresses to ensure reference counting is correcet
	uniqueAddresses := make(map[string]struct{})
	for _, ip := range addresses {
		uniqueAddresses[ip.IP.String()] = struct{}{}
	}

	// Determine the ordinals that need to be released and the
	// attributes that need to be cleaned up.
	log.Debugf("Releasing addresses from block: %v", uniqueAddresses)
	for ipStr := range uniqueAddresses {
		ip := cnet.MustParseIP(ipStr)
		// Convert to an ordinal.
		ordinal, err := b.IPToOrdinal(ip)
		if err != nil {
			return nil, nil, err
		}
		log.Debugf("Address %s is ordinal %d", ip, ordinal)

		// Check if allocated.
		log.Debugf("Checking if allocated: %v", b.Allocations)
		attrIdx := b.Allocations[ordinal]
		if attrIdx == nil {
			log.Debugf("Asked to release address that was not allocated")
			unallocated = append(unallocated, ip)
			continue
		}
		ordinals = append(ordinals, ordinal)
		log.Debugf("%s is allocated, ordinals to release are now %v", ip, ordinals)

		// Increment reference counting for attributes.
		cnt := 1
		if cur, exists := delRefCounts[*attrIdx]; exists {
			cnt = cur + 1
		}
		delRefCounts[*attrIdx] = cnt
		log.Debugf("delRefCounts: %v", delRefCounts)

		// Increment count of addresses by handle if a handle
		// exists.
		log.Debugf("Looking up attribute with index %d", *attrIdx)
		handleID := b.Attributes[*attrIdx].AttrPrimary
		if handleID != nil {
			log.Debugf("HandleID is %s", *handleID)
			handleCount := 0
			if count, ok := countByHandle[*handleID]; !ok {
				handleCount = count
			}
			log.Debugf("Handle ref count is %d, incrementing", handleCount)
			handleCount += 1
			countByHandle[*handleID] = handleCount
			log.Debugf("countByHandle %v", countByHandle)
		}
	}

	// Handle cleaning up of attributes.  We do this by
	// reference counting.  If we're deleting the last reference to
	// a given attribute, then it needs to be cleaned up.
	refCounts := b.attributeRefCounts()
	log.Debugf("Cleaning up attributes, refCounts: %v", refCounts)
	for idx, refs := range delRefCounts {
		log.Debugf("Checking ref count index %d", idx)
		if refCounts[idx] == refs {
			attrsToDelete = append(attrsToDelete, idx)
		}
	}
	if len(attrsToDelete) != 0 {
		log.Debugf("Deleting attributes: %v", attrsToDelete)
		b.deleteAttributes(attrsToDelete, ordinals)
	}

	// Release requested addresses.
	log.Debugf("Allocations: %v", b.Allocations)
	log.Debugf("Releasing ordinals: %v", ordinals)
	for _, ordinal := range ordinals {
		log.Debugf("Releasing ordinal %d", ordinal)
		b.Allocations[ordinal] = nil
		b.Unallocated = append(b.Unallocated, ordinal)
	}
	return unallocated, countByHandle, nil
}

func (b *allocationBlock) deleteAttributes(delIndexes, ordinals []int) {
	newIndexes := make([]*int, len(b.Attributes))
	newAttrs := []model.AllocationAttribute{}
	y := 0 // Next free slot in the new attributes list.
	for x := range b.Attributes {
		if !intInSlice(x, delIndexes) {
			// Attribute at x is not being deleted.  Build a mapping
			// of old attribute index (x) to new attribute index (y).
			log.Debugf("%d in %v", x, delIndexes)
			newIndex := y
			newIndexes[x] = &newIndex
			y += 1
			newAttrs = append(newAttrs, b.Attributes[x])
		}
	}
	b.Attributes = newAttrs

	// Update attribute indexes for all allocations in this block.
	for i := 0; i < b.NumAddresses(); i++ {
		if b.Allocations[i] != nil {
			// Get the new index that corresponds to the old index
			// and update the allocation.
			newIndex := newIndexes[*b.Allocations[i]]
			b.Allocations[i] = newIndex
		}
	}
}

func (b allocationBlock) attributeRefCounts() map[int]int {
	refCounts := map[int]int{}
	for _, a := range b.Allocations {
		if a == nil {
			continue
		}

		if count, ok := refCounts[*a]; !ok {
			// No entry for given attribute index.
			refCounts[*a] = 1
		} else {
			refCounts[*a] = count + 1
		}
	}
	return refCounts
}

func (b allocationBlock) attributeIndexesByHandle(handleID string) []int {
	indexes := []int{}
	for i, attr := range b.Attributes {
		if attr.AttrPrimary != nil && *attr.AttrPrimary == handleID {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (b *allocationBlock) releaseByHandle(handleID string) int {
	attrIndexes := b.attributeIndexesByHandle(handleID)
	log.Debugf("Attribute indexes to release: %v", attrIndexes)
	if len(attrIndexes) == 0 {
		// Nothing to release.
		log.Debugf("No addresses assigned to handle '%s'", handleID)
		return 0
	}

	// There are addresses to release.
	ordinals := []int{}
	var o int
	for o = 0; o < b.NumAddresses(); o++ {
		// Only check allocated ordinals.
		if b.Allocations[o] != nil && intInSlice(*b.Allocations[o], attrIndexes) {
			// Release this ordinal.
			ordinals = append(ordinals, o)
		}
	}

	// Clean and reorder attributes.
	b.deleteAttributes(attrIndexes, ordinals)

	// Release the addresses.
	for _, o := range ordinals {
		b.Allocations[o] = nil
		b.Unallocated = append(b.Unallocated, o)
	}
	return len(ordinals)
}

func (b allocationBlock) ipsByHandle(handleID string) []cnet.IP {
	ips := []cnet.IP{}
	attrIndexes := b.attributeIndexesByHandle(handleID)
	var o int
	for o = 0; o < b.NumAddresses(); o++ {
		if b.Allocations[o] != nil && intInSlice(*b.Allocations[o], attrIndexes) {
			ip := b.OrdinalToIP(o)
			ips = append(ips, ip)
		}
	}
	return ips
}

func (b allocationBlock) attributesForIP(ip cnet.IP) (map[string]string, error) {
	// Convert to an ordinal.
	ordinal, err := b.IPToOrdinal(ip)
	if err != nil {
		return nil, err
	}

	// Check if allocated.
	attrIndex := b.Allocations[ordinal]
	if attrIndex == nil {
		return nil, errors.New(fmt.Sprintf("IP %s is not currently assigned in block", ip))
	}
	return b.Attributes[*attrIndex].AttrSecondary, nil
}

func (b *allocationBlock) findOrAddAttribute(handleID *string, attrs map[string]string) int {
	logCtx := log.WithField("attrs", attrs)
	if handleID != nil {
		logCtx = log.WithField("handle", *handleID)
	}
	attr := model.AllocationAttribute{handleID, attrs}
	for idx, existing := range b.Attributes {
		if reflect.DeepEqual(attr, existing) {
			log.Debugf("Attribute '%+v' already exists", attr)
			return idx
		}
	}

	// Does not exist - add it.
	logCtx.Debugf("New allocation attribute: %#v", attr)
	attrIndex := len(b.Attributes)
	b.Attributes = append(b.Attributes, attr)
	return attrIndex
}

func getBlockCIDRForAddress(addr cnet.IP, pool *v3.IPPool) cnet.IPNet {
	var mask net.IPMask
	if addr.Version() == 6 {
		// This is an IPv6 address.
		mask = net.CIDRMask(pool.Spec.BlockSize, 128)
	} else {
		// This is an IPv4 address.
		mask = net.CIDRMask(pool.Spec.BlockSize, 32)
	}
	masked := addr.Mask(mask)
	return cnet.IPNet{IPNet: net.IPNet{IP: masked, Mask: mask}}
}

func getIPVersion(ip cnet.IP) int {
	if ip.To4() == nil {
		return 6
	}
	return 4
}

func largerThanOrEqualToBlock(blockCIDR cnet.IPNet, pool *v3.IPPool) bool {
	ones, _ := blockCIDR.Mask.Size()
	return ones <= pool.Spec.BlockSize
}

func intInSlice(searchInt int, slice []int) bool {
	for _, v := range slice {
		if v == searchInt {
			return true
		}
	}
	return false
}
