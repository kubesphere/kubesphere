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

// ListOptions is the query options a List or Watch operation in the Calico API.
type ListOptions struct {
	// The namespace of the resource to List or Watch.  If blank, the list or watch wildcards
	// the namespace.  Only used for namespaced resource types.
	Namespace string

	// The name of the resource to List or Watch.  If blank, the list or watch wildcards
	// the name.
	Name string

	// The resource version to List or Watch from.
	// When specified for list:
	// - if unset, then the result is returned from remote storage based on quorum-read flag;
	// - if set to non zero, then the result is at least as fresh as given rv.
	// +optional
	ResourceVersion string

	// Whether the Name specified is a prefix rather than the full name.  This is fully supported
	// for etcdv3, and is supported in a very limited fashion in KDD for WorkloadEndpoints only
	// as a mechanism for enumerating endpoints within a Pod (since the name construction for a
	// Workload endpoint is hierarchically constructed).
	Prefix bool
}
