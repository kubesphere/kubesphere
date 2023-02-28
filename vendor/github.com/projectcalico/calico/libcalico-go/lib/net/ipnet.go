// Copyright (c) 2016-2017,2021 Tigera, Inc. All rights reserved.

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

package net

import (
	"math/big"
	"net"

	"github.com/projectcalico/calico/libcalico-go/lib/json"
)

// Sub class net.IPNet so that we can add JSON marshalling and unmarshalling.
type IPNet struct {
	net.IPNet
}

// MarshalJSON interface for an IPNet
func (i IPNet) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON interface for an IPNet
func (i *IPNet) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Decode and ensure we maintain the full IP address in the IPNet that we return.
	ip, ipnet, err := ParseCIDROrIP(s)
	if err != nil {
		return err
	}
	i.IP = ip.IP
	i.Mask = ipnet.Mask
	return nil
}

// Version returns the IP version for an IPNet, or 0 if not a valid IP net.
func (i *IPNet) Version() int {
	if i.IP.To4() != nil {
		return 4
	} else if len(i.IP) == net.IPv6len {
		return 6
	}
	return 0
}

// IsNetOverlap is a utility function that returns true if the two subnet have an overlap.
func (i IPNet) IsNetOverlap(n net.IPNet) bool {
	return n.Contains(i.IP) || i.Contains(n.IP)
}

// Covers returns true if the whole of n is covered by this CIDR.
func (i IPNet) Covers(n net.IPNet) bool {
	if !i.Contains(n.IP) {
		return false
	} // else start of n is within our bounds, what about the end...
	nPrefixLen, _ := n.Mask.Size()
	iPrefixLen, _ := i.Mask.Size()
	return iPrefixLen <= nPrefixLen
}

func (i IPNet) NthIP(n int) IP {
	bigN := big.NewInt(int64(n))
	return IncrementIP(IP{i.IP}, bigN)
}

// Network returns the masked IP network.
func (i *IPNet) Network() *IPNet {
	_, n, _ := ParseCIDR(i.String())
	return n
}

func ParseCIDR(c string) (*IP, *IPNet, error) {
	netIP, netIPNet, e := net.ParseCIDR(c)
	if netIPNet == nil || e != nil {
		return nil, nil, e
	}
	ip := &IP{netIP}
	ipnet := &IPNet{*netIPNet}

	// The base golang net library always uses a 4-byte IPv4 address in an
	// IPv4 IPNet, so for uniformity in the returned types, make sure the
	// IP address is also 4-bytes - this allows the user to safely assume
	// all IP addresses returned by this function use the same encoding
	// mechanism (not strictly required but better for testing and debugging).
	if ip4 := ip.IP.To4(); ip4 != nil {
		ip.IP = ip4
	}

	return ip, ipnet, nil
}

// Parse a CIDR or an IP address and return the IP, CIDR or error.  If an IP address
// string is supplied, then the CIDR returned is the fully masked IP address (i.e /32 or /128)
func ParseCIDROrIP(c string) (*IP, *IPNet, error) {
	// First try parsing as a CIDR.
	ip, cidr, err := ParseCIDR(c)
	if err == nil {
		return ip, cidr, nil
	}

	// That failed, so try parsing as an IP.
	ip = &IP{}
	if err2 := ip.UnmarshalText([]byte(c)); err2 == nil {
		if ip4 := ip.IP.To4(); ip4 != nil {
			ip.IP = ip4
		}
		n := ip.Network()
		return ip, n, nil
	}

	// That failed too, return the original error.
	return nil, nil, err
}

// String returns a friendly name for the network.  The standard net package
// implements String() on the pointer, which means it will not be invoked on a
// struct type, so we re-implement on the struct type.
func (i IPNet) String() string {
	ip := &i.IPNet
	return ip.String()
}

func (i IPNet) NumAddrs() *big.Int {
	ones, bits := i.Mask.Size()
	zeros := bits - ones
	numAddrs := big.NewInt(1)
	return numAddrs.Lsh(numAddrs, uint(zeros))
}

// MustParseNetwork parses the string into an IPNet.  The IP address in the
// IPNet is masked.
func MustParseNetwork(c string) IPNet {
	_, cidr, err := ParseCIDR(c)
	if err != nil {
		panic(err)
	}
	return *cidr
}

// MustParseCIDR parses the string into an IPNet.  The IP address in the
// IPNet is not masked.
func MustParseCIDR(c string) IPNet {
	ip, cidr, err := ParseCIDR(c)
	if err != nil {
		panic(err)
	}
	n := IPNet{}
	n.IP = ip.IP
	n.Mask = cidr.Mask
	return n
}
