/*
Copyright 2018 The Kubernetes Authors.

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

// ClusterPropagatedVersionSpec defines the desired state of ClusterPropagatedVersion
type ClusterPropagatedVersionSpec struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clusterpropagatedversions,scope=Cluster
// +kubebuilder:subresource:status

// ClusterPropagatedVersion holds version information about the state
// propagated from KubeFed APIs (configured by FederatedTypeConfig
// resources) to member clusters. The name of a ClusterPropagatedVersion
// encodes the kind and name of the resource it stores information for
// (i.e. <lower-case kind>-<resource name>). If a target resource has
// a populated metadata.Generation field, the generation will be
// stored with a prefix of `gen:` as the version for the cluster.  If
// metadata.Generation is not available, metadata.ResourceVersion will
// be stored with a prefix of `rv:` as the version for the cluster.
type ClusterPropagatedVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status PropagatedVersionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterPropagatedVersionList contains a list of ClusterPropagatedVersion
type ClusterPropagatedVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPropagatedVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterPropagatedVersion{}, &ClusterPropagatedVersionList{})
}
