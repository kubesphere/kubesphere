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

package v1

import (
	"fmt"

	"github.com/projectcalico/libcalico-go/lib/apis/v1/unversioned"
)

// Profile contains the details a security profile resource.  A profile is set of security rules
// to apply on an endpoint.  An endpoint (either a host endpoint or an endpoint on a workload) can
// reference zero or more profiles.  The profile rules are applied directly to the endpoint *after*
// the selector-based security policy has been applied, and in the order the profiles are declared on the
// endpoint.
type Profile struct {
	unversioned.TypeMetadata
	Metadata ProfileMetadata `json:"metadata,omitempty"`
	Spec     ProfileSpec     `json:"spec,omitempty"`
}

func (t Profile) GetResourceMetadata() unversioned.ResourceMetadata {
	return t.Metadata
}

// String() returns the human-readable string representation of a Profile instance
// which is defined by its Name.
func (t Profile) String() string {
	return fmt.Sprintf("Profile(Name=%s)", t.Metadata.Name)
}

// ProfileMetadata contains the metadata for a security Profile resource.
type ProfileMetadata struct {
	unversioned.ObjectMetadata

	// The name of the endpoint.
	Name string `json:"name,omitempty" validate:"omitempty,namespacedName"`

	// A list of tags that are applied to each endpoint that references this profile.
	Tags []string `json:"tags,omitempty" validate:"omitempty,dive,tag"`

	// The labels to apply to each endpoint that references this profile.  It is expected
	// that many endpoints share the same labels. For example, they could be used to label all
	// “production” workloads with “deployment=prod” so that security policy can be applied
	// to production workloads.
	Labels map[string]string `json:"labels,omitempty" validate:"omitempty,labels"`
}

// ProfileSpec contains the specification for a security Profile resource.
type ProfileSpec struct {
	// The ordered set of ingress rules.  Each rule contains a set of packet match criteria and
	// a corresponding action to apply.
	IngressRules []Rule `json:"ingress,omitempty" validate:"omitempty,dive"`

	// The ordered set of egress rules.  Each rule contains a set of packet match criteria and
	// a corresponding action to apply.
	EgressRules []Rule `json:"egress,omitempty" validate:"omitempty,dive"`
}

// NewProfile creates a new (zeroed) Profile struct with the TypeMetadata initialised to the current
// version.
func NewProfile() *Profile {
	return &Profile{
		TypeMetadata: unversioned.TypeMetadata{
			Kind:       "profile",
			APIVersion: unversioned.VersionCurrent,
		},
	}
}

// A ProfileList contains a list of security Profile resources.  List types are returned from List()
// enumerations on the client interface.
type ProfileList struct {
	unversioned.TypeMetadata
	Metadata unversioned.ListMetadata `json:"metadata,omitempty"`
	Items    []Profile                `json:"items" validate:"dive,omitempty"`
}

// NewProfile creates a new (zeroed) Profile struct with the TypeMetadata initialised to the current
// version.
func NewProfileList() *ProfileList {
	return &ProfileList{
		TypeMetadata: unversioned.TypeMetadata{
			Kind:       "profileList",
			APIVersion: unversioned.VersionCurrent,
		},
	}
}
