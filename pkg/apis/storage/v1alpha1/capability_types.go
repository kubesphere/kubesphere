/*
Copyright 2019 The KubeSphere authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExpandMode string

const (
	ExpandModeUnknown ExpandMode = "UNKNOWN"
	ExpandModeOffline ExpandMode = "OFFLINE"
	ExpandModeOnline  ExpandMode = "ONLINE"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:noStatus
// +genclient:nonNamespaced

// StorageClassCapability is the Schema for the storage class capability API
// +k8s:openapi-gen=true
type StorageClassCapability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StorageClassCapabilitySpec `json:"spec"`
}

// StorageClassCapabilitySpec defines the desired state of StorageClassCapability
type StorageClassCapabilitySpec struct {
	Provisioner string                             `json:"provisioner"`
	Features    StorageClassCapabilitySpecFeatures `json:"features"`
}

// StorageClassCapabilitySpecFeatures describe storage class features
type StorageClassCapabilitySpecFeatures struct {
	Topology bool                                       `json:"topology"`
	Volume   StorageClassCapabilitySpecFeaturesVolume   `json:"volume"`
	Snapshot StorageClassCapabilitySpecFeaturesSnapshot `json:"snapshot"`
}

// StorageClassCapabilitySpecFeaturesVolume describe volume features
type StorageClassCapabilitySpecFeaturesVolume struct {
	Create bool       `json:"create"`
	Attach bool       `json:"attach"`
	List   bool       `json:"list"`
	Clone  bool       `json:"clone"`
	Stats  bool       `json:"stats"`
	Expand ExpandMode `json:"expandMode"`
}

// StorageClassCapabilitySpecFeaturesSnapshot describe snapshot features
type StorageClassCapabilitySpecFeaturesSnapshot struct {
	Create bool `json:"create"`
	List   bool `json:"list"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// StorageClassCapabilityList contains a list of StorageClassCapability
type StorageClassCapabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []StorageClassCapability `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&StorageClassCapability{},
		&StorageClassCapabilityList{})
}
