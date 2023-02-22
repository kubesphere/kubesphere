// Copyright (c) 2017,2021 Tigera, Inc. All rights reserved.

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

package v3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KindProfile     = "Profile"
	KindProfileList = "ProfileList"
)

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProfileList is a list of Profile objects.
type ProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Profile `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Profile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec ProfileSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// ProfileSpec contains the specification for a security Profile resource.
type ProfileSpec struct {
	// The ordered set of ingress rules.  Each rule contains a set of packet match criteria and
	// a corresponding action to apply.
	Ingress []Rule `json:"ingress,omitempty" validate:"omitempty,dive"`
	// The ordered set of egress rules.  Each rule contains a set of packet match criteria and
	// a corresponding action to apply.
	Egress []Rule `json:"egress,omitempty" validate:"omitempty,dive"`
	// An option set of labels to apply to each endpoint (in addition to their own labels)
	// referencing this profile.  If labels configured on the endpoint have keys matching those
	// labels inherited from the profile, the endpoint label values take precedence.
	LabelsToApply map[string]string `json:"labelsToApply,omitempty" validate:"omitempty,labels"`
}

// NewProfile creates a new (zeroed) Profile struct with the TypeMetadata initialised to the current
// version.
func NewProfile() *Profile {
	return &Profile{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindProfile,
			APIVersion: GroupVersionCurrent,
		},
	}
}
