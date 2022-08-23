/*

 Copyright 2022 The KubeSphere Authors.

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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Maintainer struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
}

type ExtensionInfo struct {
	DisplayName string       `json:"displayName,omitempty"`
	Description string       `json:"description,omitempty"`
	Icon        string       `json:"icon,omitempty"`
	Maintainers []Maintainer `json:"maintainers,omitempty"`
	Version     string       `json:"version,omitempty"`
}

type ExtensionSpec struct {
	*ExtensionInfo `json:",inline"`
}

type ExtensionVersionSpec struct {
	*ExtensionInfo `json:",inline"`
	Keywords       []string `json:"keywords,omitempty"`
	Sources        []string `json:"sources,omitempty"`
	Repo           string   `json:"repo,omitempty"`
	MinKubeVersion string   `json:"minKubeVersion,omitempty"`
	Home           string   `json:"home,omitempty"`
	Digest         string   `json:"digest,omitempty"`
	URLs           []string `json:"urls"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories="extensions",scope="Cluster"

type ExtensionVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExtensionVersionSpec `json:"spec,omitempty"`
}

type CategorySpec struct {
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories="extensions",scope="Cluster"

// Category can help us group the extensions.
type Category struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CategorySpec `json:"spec,omitempty"`
}

type ExtensionVersionInfo struct {
	Version           string      `json:"version"`
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty"`
}

type ExtensionStatus struct {
	State             string                 `json:"state,omitempty"`
	SubscribedVersion string                 `json:"subscribedVersion,omitempty"`
	RecommendVersion  string                 `json:"recommendVersion,omitempty"`
	Versions          []ExtensionVersionInfo `json:"versions,omitempty"`
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories="extensions",scope="Cluster"

// Extension is synchronized from the Repository.
// An extension can contain multiple versions.
type Extension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExtensionSpec   `json:"spec,omitempty"`
	Status            ExtensionStatus `json:"status,omitempty"`
}

type UpdateStrategy struct {
	*RegistryPoll `json:"registryPoll,omitempty"`
}

type RegistryPoll struct {
	Interval *metav1.Duration `json:"interval,omitempty"`
}

type RepositorySpec struct {
	Image          string         `json:"image"`
	DisplayName    string         `json:"displayName,omitempty"`
	Description    string         `json:"description,omitempty"`
	UpdateStrategy UpdateStrategy `json:"updateStrategy,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExtensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Extension `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExtensionVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExtensionVersion `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CategoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Category `json:"items"`
}
