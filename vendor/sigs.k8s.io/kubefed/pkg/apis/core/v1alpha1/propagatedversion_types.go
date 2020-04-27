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

// PropagatedVersionSpec defines the desired state of PropagatedVersion
type PropagatedVersionSpec struct {
}

// PropagatedVersionStatus defines the observed state of PropagatedVersion
type PropagatedVersionStatus struct {
	// The observed version of the template for this resource.
	TemplateVersion string `json:"templateVersion"`
	// The observed version of the overrides for this resource.
	OverrideVersion string `json:"overridesVersion"`
	// The last versions produced in each cluster for this resource.
	// +optional
	ClusterVersions []ClusterObjectVersion `json:"clusterVersions,omitempty"`
}

type ClusterObjectVersion struct {
	// The name of the cluster the version is for.
	ClusterName string `json:"clusterName"`
	// The last version produced for the resource by a KubeFed
	// operation.
	Version string `json:"version"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=propagatedversions
// +kubebuilder:subresource:status

// PropagatedVersion holds version information about the state
// propagated from KubeFed APIs (configured by FederatedTypeConfig
// resources) to member clusters. The name of a PropagatedVersion
// encodes the kind and name of the resource it stores information for
// (i.e. <lower-case kind>-<resource name>). If a target resource has
// a populated metadata.Generation field, the generation will be
// stored with a prefix of `gen:` as the version for the cluster.  If
// metadata.Generation is not available, metadata.ResourceVersion will
// be stored with a prefix of `rv:` as the version for the cluster.
type PropagatedVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status PropagatedVersionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PropagatedVersionList contains a list of PropagatedVersion
type PropagatedVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PropagatedVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PropagatedVersion{}, &PropagatedVersionList{})
}
