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

	cnet "github.com/projectcalico/calico/libcalico-go/lib/net"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindIPPool     = "IPPool"
	ResourceSingularIPPool = "ippool"
	ResourcePluralIPPool   = "ippools"

	// scope type > id > name
	// id used to detect cidr overlap
	IPPoolTypeLabel    = "ippool.network.kubesphere.io/type"
	IPPoolNameLabel    = "ippool.network.kubesphere.io/name"
	IPPoolIDLabel      = "ippool.network.kubesphere.io/id"
	IPPoolDefaultLabel = "ippool.network.kubesphere.io/default"

	IPPoolTypeNone   = "none"
	IPPoolTypeLocal  = "local"
	IPPoolTypeCalico = "calico"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type IPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec IPPoolSpec `json:"spec,omitempty"`
	// +optional
	Status IPPoolStatus `json:"status,omitempty"`
}

type VLANConfig struct {
	VlanId uint32 `json:"vlanId"`
	Master string `json:"master"`
}

type Route struct {
	Dst string `json:"dst,omitempty"`
	GW  string `json:"gateway,omitempty"`
}

// DNS contains values interesting for DNS resolvers
type DNS struct {
	Nameservers []string `json:"nameservers,omitempty"`
	Domain      string   `json:"domain,omitempty"`
	Search      []string `json:"search,omitempty"`
	Options     []string `json:"options,omitempty"`
}

type WorkspaceStatus struct {
	Allocations int `json:"allocations"`
}

type IPPoolStatus struct {
	Unallocated int                        `json:"unallocated"`
	Allocations int                        `json:"allocations"`
	Capacity    int                        `json:"capacity"`
	Reserved    int                        `json:"reserved,omitempty"`
	Synced      bool                       `json:"synced,omitempty"`
	Workspaces  map[string]WorkspaceStatus `json:"workspaces,omitempty"`
}

type IPPoolSpec struct {
	Type string `json:"type"`

	// The pool CIDR.
	CIDR string `json:"cidr"`

	// The first ip, inclusive
	RangeStart string `json:"rangeStart,omitempty"`

	// The last ip, inclusive
	RangeEnd string `json:"rangeEnd,omitempty"`

	// When disabled is true, IPAM will not assign addresses from this pool.
	Disabled bool `json:"disabled,omitempty"`

	// The block size to use for IP address assignments from this pool. Defaults to 26 for IPv4 and 112 for IPv6.
	BlockSize int `json:"blockSize,omitempty"`

	VLAN VLANConfig `json:"vlanConfig,omitempty"`

	Gateway string  `json:"gateway,omitempty"`
	Routes  []Route `json:"routes,omitempty"`
	DNS     DNS     `json:"dns,omitempty"`
}

// +kubebuilder:object:root=true
// +genclient:nonNamespaced
type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPPool `json:"items"`
}

const (
	VLAN        = "vlan"
	Calico      = "calico"
	Porter      = "porter"
	Pod         = "pod"
	VLANIDStart = 1
	VLANIDEnd   = 4097
	PorterID    = 4098
	CalicoID    = 4099
	PodID       = 0
)

// Find the ordinal (i.e. how far into the block) a given IP lies.  Returns an error if the IP is outside the block.
func (b IPPool) IPToOrdinal(ip cnet.IP) (int, error) {
	_, cidr, _ := cnet.ParseCIDR(b.Spec.CIDR)
	ipAsInt := cnet.IPToBigInt(ip)
	baseInt := cnet.IPToBigInt(cnet.IP{IP: cidr.IP})
	ord := big.NewInt(0).Sub(ipAsInt, baseInt).Int64()
	if ord < 0 || ord >= int64(b.NumAddresses()) {
		return 0, fmt.Errorf("IP %s not in pool %s", ip, b.Spec.CIDR)
	}
	return int(ord), nil
}

// Get number of addresses covered by the block
func (b IPPool) NumAddresses() int {
	_, cidr, _ := cnet.ParseCIDR(b.Spec.CIDR)
	ones, size := cidr.Mask.Size()
	numAddresses := 1 << uint(size-ones)
	return numAddresses
}

func (b IPPool) Type() string {
	if b.Spec.Type == VLAN {
		return IPPoolTypeLocal
	}

	return b.Spec.Type
}

func (b IPPool) NumReservedAddresses() int {
	return b.StartReservedAddressed() + b.EndReservedAddressed()
}

func (b IPPool) StartReservedAddressed() int {
	if b.Spec.RangeStart == "" {
		return 0
	}
	start, _ := b.IPToOrdinal(*cnet.ParseIP(b.Spec.RangeStart))
	return start
}

func (b IPPool) EndReservedAddressed() int {
	if b.Spec.RangeEnd == "" {
		return 0
	}
	total := b.NumAddresses()
	end, _ := b.IPToOrdinal(*cnet.ParseIP(b.Spec.RangeEnd))
	return total - end - 1
}

func (b IPPool) Overlapped(dst IPPool) bool {
	if b.ID() != dst.ID() {
		return false
	}

	_, cidr, _ := cnet.ParseCIDR(b.Spec.CIDR)
	_, cidrDst, _ := cnet.ParseCIDR(dst.Spec.CIDR)

	return cidr.IsNetOverlap(cidrDst.IPNet)
}

func (pool IPPool) ID() uint32 {
	switch pool.Spec.Type {
	case VLAN:
		return pool.Spec.VLAN.VlanId + VLANIDStart
	case Porter:
		return PorterID
	case Calico:
		return CalicoID
	}

	return PodID
}

func (p IPPool) TypeInvalid() bool {
	typeStr := p.Spec.Type
	if typeStr == VLAN || typeStr == Porter || typeStr == Pod {
		return false
	}

	return true
}

func (p IPPool) Disabled() bool {
	return p.Spec.Disabled
}

func (p IPPool) V4() bool {
	ip, _, _ := cnet.ParseCIDR(p.Spec.CIDR)
	if ip.To4() != nil {
		return true
	}
	return false
}

const IPPoolFinalizer = "finalizers.network.kubesphere.io/ippool"
