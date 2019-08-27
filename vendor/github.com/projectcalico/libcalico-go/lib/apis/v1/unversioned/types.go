// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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

package unversioned

// All resources (and resource lists) implement the Resource interface.
type Resource interface {
	GetTypeMetadata() TypeMetadata
}

// All singular resources (all resources not including lists) implement the ResourceObject interface.
type ResourceObject interface {
	Resource

	// GetResourceMetadata returns the ResourceMetadata for each Resource Object.
	GetResourceMetadata() ResourceMetadata

	// String returns a human-readable string representation of a ResourceObject which
	// includes the important ID fields for a ResourceObject.
	String() string
}

// Define available versions.
var (
	// `apiVersion` in the config yaml files
	VersionV1      = "v1"
	VersionCurrent = VersionV1
)

// ---- Type metadata ----
//
// All resource and resource lists embed a TypeMetadata as an anonymous field.
type TypeMetadata struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}

func (md TypeMetadata) GetTypeMetadata() TypeMetadata {
	return md
}

// All resource Metadata (not lists) implement the ResourceMetadata interface.
type ResourceMetadata interface {
	// GetObjectMetadata returns the ObjectMetadata instance of the ResourceMetadata.
	GetObjectMetadata() ObjectMetadata
}

// ---- Metadata common to all resources ----
type ObjectMetadata struct {
	// Object revision used to perform atomic updates and deletes.  Currently
	// only supported on Get and Delete operations of the WorkloadEndpoint
	// resource type.
	Revision string `json:"-"`
}

func (md ObjectMetadata) GetObjectMetadata() ObjectMetadata {
	return md
}

// ---- Metadata common to all lists ----
type ListMetadata struct {
}
