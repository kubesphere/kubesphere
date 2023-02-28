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

package net

import (
	"math/big"
	"net"

	"github.com/projectcalico/calico/libcalico-go/lib/json"
)

// Sub class net.IP so that we can add JSON marshalling and unmarshalling.
type IP struct {
	net.IP
}

// MarshalJSON interface for an IP
func (i IP) MarshalJSON() ([]byte, error) {
	s, err := i.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(s))
}

// UnmarshalJSON interface for an IP
func (i *IP) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if err := i.UnmarshalText([]byte(s)); err != nil {
		return err
	}
	// Always return IPv4 values as 4-bytes to be consistent with IPv4 IPNet
	// representations.
	if ipv4 := i.To4(); ipv4 != nil {
		i.IP = ipv4
	}

	return nil
}

// ParseIP returns an IP from a string
func ParseIP(ip string) *IP {
	addr := net.ParseIP(ip)
	if addr == nil {
		return nil
	}
	// Always return IPv4 values as 4-bytes to be consistent with IPv4 IPNet
	// representations.
	if addr4 := addr.To4(); addr4 != nil {
		addr = addr4
	}
	return &IP{addr}
}

// Version returns the IP version for an IP, or 0 if the IP is not valid.
func (i IP) Version() int {
	if i.To4() != nil {
		return 4
	} else if len(i.IP) == net.IPv6len {
		return 6
	}
	return 0
}

// Network returns the IP address as a fully masked IPNet type.
func (i *IP) Network() *IPNet {
	// Unmarshaling an IPv4 address returns a 16-byte format of the
	// address, so convert to 4-byte format to match the mask.
	n := &IPNet{}
	if ip4 := i.IP.To4(); ip4 != nil {
		n.IP = ip4
		n.Mask = net.CIDRMask(net.IPv4len*8, net.IPv4len*8)
	} else {
		n.IP = i.IP
		n.Mask = net.CIDRMask(net.IPv6len*8, net.IPv6len*8)
	}
	return n
}

// MustParseIP parses the string into an IP.
func MustParseIP(i string) IP {
	var ip IP
	err := ip.UnmarshalText([]byte(i))
	if err != nil {
		panic(err)
	}
	// Always return IPv4 values as 4-bytes to be consistent with IPv4 IPNet
	// representations.
	if ip4 := ip.To4(); ip4 != nil {
		ip.IP = ip4
	}
	return ip
}

func IPToBigInt(ip IP) *big.Int {
	if ip.To4() != nil {
		return big.NewInt(0).SetBytes(ip.To4())
	} else {
		return big.NewInt(0).SetBytes(ip.To16())
	}
}

func BigIntToIP(ipInt *big.Int, v6 bool) IP {
	var netIP net.IP
	// Older versions of this code tried to guess v4/v6 based on the length of the big.Int
	// but then we can't tell the difference between 0.0.0.0/0 and ::/0.
	if v6 {
		netIP = make(net.IP, 16)
	} else {
		netIP = make(net.IP, 4)
	}
	ipInt.FillBytes(netIP)
	return IP{netIP}
}

func IncrementIP(ip IP, increment *big.Int) IP {
	expectingV6 := ip.To4() == nil
	sum := big.NewInt(0).Add(IPToBigInt(ip), increment)
	return BigIntToIP(sum, expectingV6)
}
