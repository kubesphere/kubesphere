// Copyright (c) 2017 Tigera, Inc. All rights reserved.

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

package options

// GetOptions is the standard query options for getting a resource through the Calico API.
type GetOptions struct {
	// When specified:
	// - if unset, then the result is returned from remote storage based on quorum-read flag;
	// - if set to non zero, then the result is at least as fresh as given rv.
	// +optional
	ResourceVersion string
}
