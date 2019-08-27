// Copyright (c) 2017-2019 Tigera, Inc. All rights reserved.
//
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
	"context"

	cnet "github.com/projectcalico/libcalico-go/lib/net"
)

// ipam.Interface has methods to perform IP address management.
type Interface interface {
	// AssignIP assigns the provided IP address to the provided host.  The IP address
	// must fall within a configured pool.  AssignIP will claim block affinity as needed
	// in order to satisfy the assignment.  An error will be returned if the IP address
	// is already assigned, or if StrictAffinity is enabled and the address is within
	// a block that does not have affinity for the given host.
	AssignIP(ctx context.Context, args AssignIPArgs) error

	// AutoAssign automatically assigns one or more IP addresses as specified by the
	// provided AutoAssignArgs.  AutoAssign returns the list of the assigned IPv4 addresses,
	// and the list of the assigned IPv6 addresses in IPNet format.
	//
	// In case of error, returns the IPs allocated so far along with the error.
	AutoAssign(ctx context.Context, args AutoAssignArgs) ([]cnet.IPNet, []cnet.IPNet, error)

	// ReleaseIPs releases any of the given IP addresses that are currently assigned,
	// so that they are available to be used in another assignment.
	ReleaseIPs(ctx context.Context, ips []cnet.IP) ([]cnet.IP, error)

	// GetAssignmentAttributes returns the attributes stored with the given IP address
	// upon assignment.
	GetAssignmentAttributes(ctx context.Context, addr cnet.IP) (map[string]string, error)

	// IPsByHandle returns a list of all IP addresses that have been
	// assigned using the provided handle.
	IPsByHandle(ctx context.Context, handleID string) ([]cnet.IP, error)

	// ReleaseByHandle releases all IP addresses that have been assigned
	// using the provided handle.  Returns an error if no addresses
	// are assigned with the given handle.
	ReleaseByHandle(ctx context.Context, handleID string) error

	// ClaimAffinity claims affinity to the given host for all blocks
	// within the given CIDR.  The given CIDR must fall within a configured
	// pool. If an empty string is passed as the host, then the value returned by os.Hostname is used.
	ClaimAffinity(ctx context.Context, cidr cnet.IPNet, host string) ([]cnet.IPNet, []cnet.IPNet, error)

	// ReleaseAffinity releases affinity for all blocks within the given CIDR
	// on the given host.  If an empty string is passed as the host, then the
	// value returned by os.Hostname will be used. If mustBeEmpty is true, then an error
	// will be returned if any blocks within the CIDR are not empty - in this case, this
	// function may release some but not all blocks within the given CIDR.
	ReleaseAffinity(ctx context.Context, cidr cnet.IPNet, host string, mustBeEmpty bool) error

	// ReleaseHostAffinities releases affinity for all blocks that are affine
	// to the given host.  If an empty string is passed as the host, the value returned by
	// os.Hostname will be used. If mustBeEmpty is true, then an error
	// will be returned if any blocks within the CIDR are not empty - in this case, this
	// function may release some but not all blocks attached to this host.
	ReleaseHostAffinities(ctx context.Context, host string, mustBeEmpty bool) error

	// ReleasePoolAffinities releases affinity for all blocks within
	// the specified pool across all hosts.
	ReleasePoolAffinities(ctx context.Context, pool cnet.IPNet) error

	// GetIPAMConfig returns the global IPAM configuration.  If no IPAM configuration
	// has been set, returns a default configuration with StrictAffinity disabled
	// and AutoAllocateBlocks enabled.
	GetIPAMConfig(ctx context.Context) (*IPAMConfig, error)

	// SetIPAMConfig sets global IPAM configuration.  This can only
	// be done when there are no allocated blocks and IP addresses.
	SetIPAMConfig(ctx context.Context, cfg IPAMConfig) error

	// RemoveIPAMHost releases affinity for all blocks on the given host,
	// and removes all host-specific IPAM data from the datastore.
	// RemoveIPAMHost does not release any IP addresses claimed on the given host.
	// If an empty string is passed as the host then the value returned by os.Hostname is used.
	RemoveIPAMHost(ctx context.Context, host string) error

	// GetUtilization returns IP utilization info for the specified pools, or for all pools.
	GetUtilization(ctx context.Context, args GetUtilizationArgs) ([]*PoolUtilization, error)
}
