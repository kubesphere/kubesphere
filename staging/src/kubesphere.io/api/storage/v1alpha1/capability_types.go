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

// VolumeFeature describe volume features
type VolumeFeature struct {
	Create bool       `json:"create"`
	Attach bool       `json:"attach"`
	List   bool       `json:"list"`
	Clone  bool       `json:"clone"`
	Stats  bool       `json:"stats"`
	Expand ExpandMode `json:"expandMode"`
}

// SnapshotFeature describe snapshot features
type SnapshotFeature struct {
	Create bool `json:"create"`
	List   bool `json:"list"`
}

// CapabilityFeatures describe storage features
type CapabilityFeatures struct {
	Topology bool            `json:"topology"`
	Volume   VolumeFeature   `json:"volume"`
	Snapshot SnapshotFeature `json:"snapshot"`
}

// PluginInfo describes plugin info
type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// +genclient
// +kubebuilder:object:root=true
// +genclient:noStatus
// +genclient:nonNamespaced
// +kubebuilder:printcolumn:name="Provisioner",type="string",JSONPath=".spec.provisioner"
// +kubebuilder:printcolumn:name="Volume",type="boolean",JSONPath=".spec.features.volume.create"
// +kubebuilder:printcolumn:name="Expand",type="string",JSONPath=".spec.features.volume.expandMode"
// +kubebuilder:printcolumn:name="Clone",type="boolean",JSONPath=".spec.features.volume.clone"
// +kubebuilder:printcolumn:name="Snapshot",type="boolean",JSONPath=".spec.features.snapshot.create"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope="Cluster"

// StorageClassCapability is the Schema for the storage class capability API
// +k8s:openapi-gen=true
type StorageClassCapability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StorageClassCapabilitySpec `json:"spec"`
}

// StorageClassCapabilitySpec defines the desired state of StorageClassCapability
type StorageClassCapabilitySpec struct {
	Provisioner string             `json:"provisioner"`
	Features    CapabilityFeatures `json:"features"`
}

// +kubebuilder:object:root=true
// +genclient:nonNamespaced

// StorageClassCapabilityList contains a list of StorageClassCapability
type StorageClassCapabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []StorageClassCapability `json:"items"`
}

// +genclient
// +kubebuilder:object:root=true
// +genclient:noStatus
// +genclient:nonNamespaced
// +kubebuilder:printcolumn:name="Provisioner",type="string",JSONPath=".spec.pluginInfo.name"
// +kubebuilder:printcolumn:name="Expand",type="string",JSONPath=".spec.features.volume.expandMode"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope="Cluster"

// ProvisionerCapability is the schema for the provisionercapability API
// +k8s:openapi-gen=true
type ProvisionerCapability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProvisionerCapabilitySpec `json:"spec"`
}

// ProvisionerCapabilitySpec defines the desired state of ProvisionerCapability
type ProvisionerCapabilitySpec struct {
	PluginInfo PluginInfo         `json:"pluginInfo"`
	Features   CapabilityFeatures `json:"features"`
}

// +kubebuilder:object:root=true
type ProvisionerCapabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ProvisionerCapability `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&StorageClassCapability{},
		&StorageClassCapabilityList{},
		&ProvisionerCapability{},
		&ProvisionerCapabilityList{})
}
