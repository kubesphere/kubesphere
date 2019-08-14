// Copyright (c) 2017-2019 Tigera, Inc. All rights reserved.

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

import "github.com/projectcalico/libcalico-go/lib/apis/v3"

// Interface used to access the enabled IPPools.
type PoolAccessorInterface interface {
	// Returns a list of enabled pools sorted in alphanumeric name order.
	GetEnabledPools(ipVersion int) ([]v3.IPPool, error)
	// Returns a list of all pools sorted in alphanumeric name order.
	GetAllPools() ([]v3.IPPool, error)
}
