// Copyright (c) 2018 Tigera, Inc. All rights reserved.
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

package model

import "fmt"

const (
	NoRegion     string = "no-region"
	RegionPrefix string = "region-"
)

func RegionString(region string) string {
	if region != "" {
		return RegionPrefix + region
	} else {
		return NoRegion
	}
}

func ErrorSlashInRegionString(regionString string) error {
	return fmt.Errorf("RegionString %v is invalid because it includes a slash", regionString)
}
