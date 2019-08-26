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
	"net"

	cnet "github.com/projectcalico/libcalico-go/lib/net"
)

// AssignIPArgs defines the set of arguments for assigning a specific IP address.
type AssignIPArgs struct {
	// The IP address to assign.
	IP cnet.IP

	// If specified, a handle which can be used to retrieve / release
	// the allocated IP addresses in the future.
	HandleID *string

	// A key/value mapping of metadata to store with the allocations.
	Attrs map[string]string

	// If specified, the hostname of the host on which IP addresses
	// will be allocated.  If not specified, this will default
	// to the value provided by os.Hostname.
	Hostname string
}

// AutoAssignArgs defines the set of arguments for assigning one or more
// IP addresses.
type AutoAssignArgs struct {
	// The number of IPv4 addresses to automatically assign.
	Num4 int

	// The number of IPv6 addresses to automatically assign.
	Num6 int

	// If specified, a handle which can be used to retrieve / release
	// the allocated IP addresses in the future.
	HandleID *string

	// A key/value mapping of metadata to store with the allocations.
	Attrs map[string]string

	// If specified, the hostname of the host on which IP addresses
	// will be allocated.  If not specified, this will default
	// to the value provided by os.Hostname.
	Hostname string

	// If specified, the previously configured IPv4 pools from which
	// to assign IPv4 addresses.  If not specified, this defaults to all IPv4 pools.
	IPv4Pools []cnet.IPNet

	// If specified, the previously configured IPv6 pools from which
	// to assign IPv6 addresses.  If not specified, this defaults to all IPv6 pools.
	IPv6Pools []cnet.IPNet

	// If non-zero, limit on the number of affine blocks this host is allowed to claim
	// (per IP version).
	MaxBlocksPerHost int
}

// IPAMConfig contains global configuration options for Calico IPAM.
// This IPAM configuration is stored in the datastore and configures the behavior
// of Calico IPAM across an entire Calico cluster.
type IPAMConfig struct {
	// When StrictAffinity is true, addresses from a given block can only be
	// assigned by hosts with the blocks affinity.  If false, then AutoAllocateBlocks
	// must be true.  The default value is false.
	StrictAffinity bool

	// When AutoAllocateBlocks is true, Calico will automatically
	// allocate blocks of IP address to hosts as needed to assign addresses.
	// If false, then StrictAffinity must be true.  The default value is true.
	AutoAllocateBlocks bool
}

// GetUtilizationArgs defines the set of arguments for requesting IP utilization.
type GetUtilizationArgs struct {
	// If specified, the pools whose utilization should be reported.  Each string here
	// can be a pool name or CIDR.  If not specified, this defaults to all pools.
	Pools []string
}

// BlockUtilization reports IP utilization for a single allocation block.
type BlockUtilization struct {
	// This block's CIDR.
	CIDR net.IPNet

	// Number of possible IPs in this block.
	Capacity int

	// Number of available IPs in this block.
	Available int
}

// PoolUtilization reports IP utilization for a single IP pool.
type PoolUtilization struct {
	// This pool's name.
	Name string

	// This pool's CIDR.
	CIDR net.IPNet

	// Utilization for each of this pool's blocks.
	Blocks []BlockUtilization
}
