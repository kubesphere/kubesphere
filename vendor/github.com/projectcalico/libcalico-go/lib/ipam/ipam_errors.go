// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.

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
	"fmt"
)

// invalidSizeError indicates that the requested IP network size is not valid.
type invalidSizeError string

func (e invalidSizeError) Error() string {
	return string(e)
}

// ipamConfigConflictError indicates an attempt to change IPAM configuration
// that conflicts with existing allocations.
type ipamConfigConflictError string

func (e ipamConfigConflictError) Error() string {
	return string(e)
}

// noFreeBlocksError indicates an attempt to claim a block
// when there are none available.
type noFreeBlocksError string

func (e noFreeBlocksError) Error() string {
	return string(e)
}

// errBlockClaimConflict indicates that a given block has already
// been claimed by another host.
type errBlockClaimConflict struct {
	Block allocationBlock
}

func (e errBlockClaimConflict) Error() string {
	if e.Block.Affinity != nil {
		return fmt.Sprintf("%v already claimed by %v", e.Block.CIDR, *e.Block.Affinity)
	}
	return fmt.Sprintf("%v already claimed", e.Block.CIDR)
}

// errBlockNotEmpty indicates that a given block has already
// been claimed by another host.
type errBlockNotEmpty struct {
	Block allocationBlock
}

func (e errBlockNotEmpty) Error() string {
	return fmt.Sprintf("block '%v' is not empty", e.Block.CIDR)
}

// errStaleAffinity indicates to the calling code that the given affinity
// is not confirmed, and that the corresponding block belongs to another host.
type errStaleAffinity string

func (e errStaleAffinity) Error() string {
	return string(e)
}
