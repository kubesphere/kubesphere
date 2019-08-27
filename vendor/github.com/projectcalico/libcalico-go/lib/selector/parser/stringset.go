// Copyright (c) 2016 Tigera, Inc. All rights reserved.
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

package parser

import "sort"

type StringSet []string

// Contains returns true if a given string in current set
func (ss StringSet) Contains(s string) bool {
	if len(ss) == 0 {
		// Empty set or nil.
		return false
	}
	// There's at least one item, do a binary chop to find the correct
	// entry.  [minIdx, maxIdx) defines the (half-open) search interval.
	minIdx := 0
	maxIdx := len(ss)
	for minIdx < (maxIdx - 1) {
		// Select the partition index. The loop condition ensures that
		// minIdx < partitionIdx < maxIdx so we'll always shrink the
		// search interval on each iteration.
		partitionIdx := (minIdx + maxIdx) / 2
		partition := ss[partitionIdx]
		if s < partition {
			// target is strictly less than the partition, we can
			// move maxIdx down.
			maxIdx = partitionIdx
		} else {
			// Target is >= the partition, move minIdx up.
			minIdx = partitionIdx
		}
	}
	// When we exit the loop, minIdx == (maxIdx - 1).  Since the interval
	// is half-open that means that, if the value is present, it must be at
	// minIdx.  (minIdx cannot equal maxIdx due to the empty list check
	// above and the loop condition.)
	return ss[minIdx] == s
}

func ConvertToStringSetInPlace(s []string) StringSet {
	if s != nil {
		sort.Strings(s)
	}
	j := 0
	var last string
	for _, v := range s {
		if j != 0 && last == v {
			// Same as last value, skip.
			continue
		}
		s[j] = v
		j++
		last = v
	}
	s = s[:j]
	return StringSet(s)
}
