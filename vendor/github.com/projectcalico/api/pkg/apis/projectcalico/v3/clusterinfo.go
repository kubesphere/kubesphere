// Copyright (c) 2017, 2020-2021 Tigera, Inc. All rights reserved.

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
	KindClusterInformation     = "ClusterInformation"
	KindClusterInformationList = "ClusterInformationList"
)

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterInformationList is a list of ClusterInformation objects.
type ClusterInformationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ClusterInformation `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterInformation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec ClusterInformationSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// ClusterInformationSpec contains the values of describing the cluster.
type ClusterInformationSpec struct {
	// ClusterGUID is the GUID of the cluster
	ClusterGUID string `json:"clusterGUID,omitempty" validate:"omitempty"`
	// ClusterType describes the type of the cluster
	ClusterType string `json:"clusterType,omitempty" validate:"omitempty"`
	// CalicoVersion is the version of Calico that the cluster is running
	CalicoVersion string `json:"calicoVersion,omitempty" validate:"omitempty"`
	// DatastoreReady is used during significant datastore migrations to signal to components
	// such as Felix that it should wait before accessing the datastore.
	DatastoreReady *bool `json:"datastoreReady,omitempty"`
	// Variant declares which variant of Calico should be active.
	Variant string `json:"variant,omitempty"`
}

// New ClusterInformation creates a new (zeroed) ClusterInformation struct with the TypeMetadata
// initialized to the current version.
func NewClusterInformation() *ClusterInformation {
	return &ClusterInformation{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindClusterInformation,
			APIVersion: GroupVersionCurrent,
		},
	}
}
