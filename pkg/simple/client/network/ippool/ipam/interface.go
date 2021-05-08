/*
Copyright 2020 KubeSphere Authors

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

package ipam

import (
	"github.com/containernetworking/cni/pkg/types/current"

	"kubesphere.io/api/network/v1alpha1"
)

// ipam.Interface has methods to perform IP address management.
type Interface interface {
	// AutoAssign automatically assigns one or more IP addresses as specified by the
	// provided AutoAssignArgs.  AutoAssign returns the list of the assigned IPv4 addresses,
	// and the list of the assigned IPv6 addresses in IPNet format.
	// The returned IPNet represents the allocation block from which the IP was allocated,
	// which is useful for dataplanes that need to know the subnet (such as Windows).
	//
	// In case of error, returns the IPs allocated so far along with the error.
	AutoAssign(args AutoAssignArgs) (*current.Result, error)

	// ReleaseByHandle releases all IP addresses that have been assigned
	// using the provided handle.  Returns an error if no addresses
	// are assigned with the given handle.
	ReleaseByHandle(handleID string) error

	GetUtilization(args GetUtilizationArgs) ([]*PoolUtilization, error)
}

// Interface used to access the enabled IPPools.
type PoolAccessorInterface interface {
	// Returns a list of all pools sorted in alphanumeric name order.
	getAllPools() ([]v1alpha1.IPPool, error)
}
