/*
Copyright 2020 The KubeSphere authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/projectcalico/libcalico-go/lib/names"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindIPAMBlock     = "IPAMBlock"
	ResourceSingularIPAMBlock = "ipamblock"
	ResourcePluralIPAMBlock   = "ipamblocks"

	IPAMBlockAttributePod          = "pod"
	IPAMBlockAttributeVm           = "vm"
	IPAMBlockAttributeWorkloadType = "workload-type"
	IPAMBlockAttributeNamespace    = "namespace"
	IPAMBlockAttributeWorkspace    = "workspace"
	IPAMBlockAttributeNode         = "node"
	IPAMBlockAttributePool         = "pool-name"
	IPAMBlockAttributeType         = "pool-type"

	ReservedHandle = "kubesphere-reserved-handle"
	ReservedNote   = "kubesphere reserved"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Cluster
type IPAMBlock struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the IPAMBlock.
	Spec IPAMBlockSpec `json:"spec,omitempty"`
}

// IPAMBlockSpec contains the specification for an IPAMBlock resource.
type IPAMBlockSpec struct {
	ID          uint32                `json:"id"`
	CIDR        string                `json:"cidr"`
	Allocations []*int                `json:"allocations"`
	Unallocated []int                 `json:"unallocated"`
	Attributes  []AllocationAttribute `json:"attributes"`
	Deleted     bool                  `json:"deleted"`
}

type AllocationAttribute struct {
	AttrPrimary   string            `json:"handle_id,omitempty"`
	AttrSecondary map[string]string `json:"secondary,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
type IPAMBlockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IPAMBlock `json:"items"`
}

// The caller needs to check that the returned slice length is correct.
func (b *IPAMBlock) AutoAssign(
	num int, handleID string, attrs map[string]string) []cnet.IPNet {

	// Walk the allocations until we find enough addresses.
	ordinals := []int{}
	for len(b.Spec.Unallocated) > 0 && len(ordinals) < num {
		ordinals = append(ordinals, b.Spec.Unallocated[0])
		b.Spec.Unallocated = b.Spec.Unallocated[1:]
	}

	// Create slice of IPs and perform the allocations.
	ips := []cnet.IPNet{}
	ip, mask, _ := cnet.ParseCIDR(b.Spec.CIDR)
	for _, o := range ordinals {
		attrIndex := b.findOrAddAttribute(handleID, attrs)
		b.Spec.Allocations[o] = &attrIndex
		ipNets := cnet.IPNet(*mask)
		ipNets.IP = cnet.IncrementIP(*ip, big.NewInt(int64(o))).IP
		ips = append(ips, ipNets)
	}

	return ips
}

func (b *IPAMBlock) String() string {
	return fmt.Sprintf("%d-%s", b.Spec.ID, b.Spec.CIDR)
}

func (b *IPAMBlock) ID() uint32 {
	return b.Spec.ID
}

func (b *IPAMBlock) BlockName() string {
	_, cidr, _ := cnet.ParseCIDR(b.Spec.CIDR)
	return fmt.Sprintf("%d-%s", b.ID(), names.CIDRToName(*cidr))
}

// Get number of addresses covered by the block
func (b *IPAMBlock) NumAddresses() int {
	_, cidr, _ := cnet.ParseCIDR(b.Spec.CIDR)
	ones, size := cidr.Mask.Size()
	numAddresses := 1 << uint(size-ones)
	return numAddresses
}

// Find the ordinal (i.e. how far into the block) a given IP lies.  Returns an error if the IP is outside the block.
func (b *IPAMBlock) IPToOrdinal(ip cnet.IP) (int, error) {
	netIP, _, _ := cnet.ParseCIDR(b.Spec.CIDR)
	ipAsInt := cnet.IPToBigInt(ip)
	baseInt := cnet.IPToBigInt(*netIP)
	ord := big.NewInt(0).Sub(ipAsInt, baseInt).Int64()
	if ord < 0 || ord >= int64(b.NumAddresses()) {
		return 0, fmt.Errorf("IP %s not in block %d-%s", ip, b.Spec.ID, b.Spec.CIDR)
	}
	return int(ord), nil
}

func (b *IPAMBlock) NumFreeAddresses() int {
	return len(b.Spec.Unallocated)
}

// empty returns true if the block has released all of its assignable addresses,
// and returns false if any assignable addresses are in use.
func (b *IPAMBlock) Empty() bool {
	return b.containsOnlyReservedIPs()
}

func (b *IPAMBlock) MarkDeleted() {
	b.Spec.Deleted = true
}

func (b *IPAMBlock) IsDeleted() bool {
	return b.Spec.Deleted
}

// containsOnlyReservedIPs returns true if the block is empty excepted for
// expected "reserved" IP addresses.
func (b *IPAMBlock) containsOnlyReservedIPs() bool {
	for _, attrIdx := range b.Spec.Allocations {
		if attrIdx == nil {
			continue
		}
		attrs := b.Spec.Attributes[*attrIdx]
		if strings.ToLower(attrs.AttrPrimary) != ReservedHandle {
			return false
		}
	}
	return true
}

func (b *IPAMBlock) NumReservedAddresses() int {
	sum := 0
	for _, attrIdx := range b.Spec.Allocations {
		if attrIdx == nil {
			continue
		}
		attrs := b.Spec.Attributes[*attrIdx]
		if strings.ToLower(attrs.AttrPrimary) == ReservedHandle {
			sum += 1
		}
	}
	return sum
}

func (b IPAMBlock) attributeIndexesByHandle(handleID string) []int {
	indexes := []int{}
	for i, attr := range b.Spec.Attributes {
		if attr.AttrPrimary == handleID {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (b *IPAMBlock) deleteAttributes(delIndexes, ordinals []int) {
	newIndexes := make([]*int, len(b.Spec.Attributes))
	newAttrs := []AllocationAttribute{}
	y := 0 // Next free slot in the new attributes list.
	for x := range b.Spec.Attributes {
		if !intInSlice(x, delIndexes) {
			// Attribute at x is not being deleted.  Build a mapping
			// of old attribute index (x) to new attribute index (y).
			newIndex := y
			newIndexes[x] = &newIndex
			y += 1
			newAttrs = append(newAttrs, b.Spec.Attributes[x])
		}
	}
	b.Spec.Attributes = newAttrs

	// Update attribute indexes for all allocations in this block.
	for i := 0; i < b.NumAddresses(); i++ {
		if b.Spec.Allocations[i] != nil {
			// Get the new index that corresponds to the old index
			// and update the allocation.
			newIndex := newIndexes[*b.Spec.Allocations[i]]
			b.Spec.Allocations[i] = newIndex
		}
	}
}

func (b *IPAMBlock) ReleaseByHandle(handleID string) int {
	attrIndexes := b.attributeIndexesByHandle(handleID)
	if len(attrIndexes) == 0 {
		// Nothing to release.
		return 0
	}

	// There are addresses to release.
	ordinals := []int{}
	var o int
	for o = 0; o < b.NumAddresses(); o++ {
		// Only check allocated ordinals.
		if b.Spec.Allocations[o] != nil && intInSlice(*b.Spec.Allocations[o], attrIndexes) {
			// Release this ordinal.
			ordinals = append(ordinals, o)
		}
	}

	// Clean and reorder attributes.
	b.deleteAttributes(attrIndexes, ordinals)

	// Release the addresses.
	for _, o := range ordinals {
		b.Spec.Allocations[o] = nil
		b.Spec.Unallocated = append(b.Spec.Unallocated, o)
	}
	return len(ordinals)
}

func (b *IPAMBlock) findOrAddAttribute(handleID string, attrs map[string]string) int {
	attr := AllocationAttribute{handleID, attrs}
	for idx, existing := range b.Spec.Attributes {
		if reflect.DeepEqual(attr, existing) {
			return idx
		}
	}

	// Does not exist - add it.
	attrIndex := len(b.Spec.Attributes)
	b.Spec.Attributes = append(b.Spec.Attributes, attr)
	return attrIndex
}

func intInSlice(searchInt int, slice []int) bool {
	for _, v := range slice {
		if v == searchInt {
			return true
		}
	}
	return false
}

//This just initializes the data structure and does not call the api to create
func NewBlock(pool *IPPool, cidr cnet.IPNet, rsvdAttr *ReservedAttr) *IPAMBlock {
	b := IPAMBlock{}

	b.Labels = map[string]string{
		IPPoolNameLabel: pool.Name,
	}
	b.Spec.CIDR = cidr.String()
	b.Spec.ID = pool.ID()
	b.Name = b.BlockName()

	numAddresses := b.NumAddresses()
	b.Spec.Allocations = make([]*int, numAddresses)
	b.Spec.Unallocated = make([]int, numAddresses)

	// Initialize unallocated ordinals.
	for i := 0; i < numAddresses; i++ {
		b.Spec.Unallocated[i] = i
	}

	if rsvdAttr != nil {
		// Reserve IPs based on host reserved attributes.
		// For example, with windows OS, the following IP addresses of the block are
		// reserved. This is done by pre-allocating them during initialization
		// time only.
		// IPs : x.0, x.1, x.2 and x.bcastAddr (e.g. x.255 for /24 subnet)

		// nil attributes
		attrs := make(map[string]string)
		attrs["note"] = rsvdAttr.Note
		handleID := rsvdAttr.Handle
		b.Spec.Unallocated = b.Spec.Unallocated[rsvdAttr.StartOfBlock : numAddresses-rsvdAttr.EndOfBlock]
		attrIndex := len(b.Spec.Attributes)
		for i := 0; i < rsvdAttr.StartOfBlock; i++ {
			b.Spec.Allocations[i] = &attrIndex
		}
		for i := 1; i <= rsvdAttr.EndOfBlock; i++ {
			b.Spec.Allocations[numAddresses-i] = &attrIndex
		}

		// Create slice of IPs and perform the allocations.
		attr := AllocationAttribute{
			AttrPrimary:   handleID,
			AttrSecondary: attrs,
		}
		b.Spec.Attributes = append(b.Spec.Attributes, attr)
	}

	return &b
}

type ReservedAttr struct {
	// Number of addresses reserved from start of the block.
	StartOfBlock int

	// Number of addresses reserved from end of the block.
	EndOfBlock int

	// Handle for reserved addresses.
	Handle string

	// A description about the reserves.
	Note string
}
