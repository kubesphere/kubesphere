/*
Copyright 2020 The KubeSphere Authors.

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

	"kubesphere.io/api/constants"
)

const (
	ResourceKindHelmRepo     = "HelmRepo"
	ResourceSingularHelmRepo = "helmrepo"
	ResourcePluralHelmRepo   = "helmrepos"
)

type HelmRepoCredential struct {
	// chart repository username
	Username string `json:"username,omitempty"`
	// chart repository password
	Password string `json:"password,omitempty"`
	// identify HTTPS client using this SSL certificate file
	CertFile string `json:"certFile,omitempty"`
	// identify HTTPS client using this SSL key file
	KeyFile string `json:"keyFile,omitempty"`
	// verify certificates of HTTPS-enabled servers using this CA bundle
	CAFile string `json:"caFile,omitempty"`
	// skip tls certificate checks for the repository, default is ture
	InsecureSkipTLSVerify *bool `json:"insecureSkipTLSVerify,omitempty"`

	S3Config `json:",inline"`
}

type S3Config struct {
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
}

// HelmRepoSpec defines the desired state of HelmRepo
type HelmRepoSpec struct {
	// name of the repo
	Name string `json:"name"`
	// helm repo url
	Url string `json:"url"`
	// helm repo credential
	Credential HelmRepoCredential `json:"credential,omitempty"`
	// chart repo description from frontend
	Description string `json:"description,omitempty"`
	// sync period in seconds, no sync when SyncPeriod=0, the minimum SyncPeriod is 180s
	SyncPeriod int `json:"syncPeriod,omitempty"`
	// expected repo version, when this version is not equal status.version, the repo need upgrade
	// this filed should be modified when any filed of the spec modified.
	Version int `json:"version,omitempty"`
}

type HelmRepoSyncState struct {
	// last sync state, valid state are: "failed", "success", and ""
	State string `json:"state,omitempty"`
	// A human readable message indicating details about why the repo is in this state.
	Message  string       `json:"message,omitempty"`
	SyncTime *metav1.Time `json:"syncTime"`
}

// HelmRepoStatus defines the observed state of HelmRepo
type HelmRepoStatus struct {
	// repo index
	Data string `json:"data,omitempty"`
	// status last update time
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
	// current state of the repo, successful, failed or syncing
	State string `json:"state,omitempty"`
	// sync state list of history, which will store at most 10 state
	SyncState []HelmRepoSyncState `json:"syncState,omitempty"`
	// if status.version!=spec.Version, we need sync the repo now
	Version int `json:"version,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,path=helmrepos,shortName=hrepo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="url",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true

// HelmRepo is the Schema for the helmrepoes API
type HelmRepo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmRepoSpec   `json:"spec,omitempty"`
	Status HelmRepoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:root=true

// HelmRepoList contains a list of HelmRepo
type HelmRepoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmRepo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmRepo{}, &HelmRepoList{})
}

func (in *HelmRepo) GetTrueName() string {
	return in.Spec.Name
}

func (in *HelmRepo) GetHelmRepoId() string {
	return in.Name
}

func (in *HelmRepo) GetWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

func (in *HelmRepo) GetCreator() string {
	return getValue(in.Annotations, constants.CreatorAnnotationKey)
}
