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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/constants"
)

const (
	ResourceKindHelmRelease     = "HelmRelease"
	ResourceSingularHelmRelease = "helmrelease"
	ResourcePluralHelmRelease   = "helmreleases"
)

// HelmReleaseSpec defines the desired state of HelmRelease
type HelmReleaseSpec struct {
	// Name of the release
	Name string `json:"name"`
	// Message got from frontend
	Description string `json:"description,omitempty"`
	// helm release values.yaml
	Values []byte `json:"values,omitempty"`
	// The name of the chart which will be installed.
	ChartName string `json:"chartName"`
	// Specify the exact chart version to install. If this is not specified, the latest version is installed
	ChartVersion string `json:"chartVersion"`
	// appVersion from Chart.yaml
	ChartAppVersion string `json:"chartAppVer,omitempty"`
	// id of  the repo
	RepoId string `json:"repoId,omitempty"`
	// id of the helmapplication
	ApplicationId string `json:"appId,omitempty"`
	// application version id
	ApplicationVersionId string `json:"appVerId,omitempty"`
	// expected release version, when this version is not equal status.version, the release need upgrade
	// this filed should be modified when any filed of the spec modified.
	Version int `json:"version"`
}

type HelmReleaseDeployStatus struct {
	// A human readable message indicating details about why the release is in this state.
	Message string `json:"message,omitempty"`
	// current state of the release
	State string `json:"state"`
	// deploy time, upgrade time or check status time
	Time metav1.Time `json:"deployTime"`
}

// HelmReleaseStatus defines the observed state of HelmRelease
type HelmReleaseStatus struct {
	// current state
	State string `json:"state"`
	// A human readable message indicating details about why the release is in this state.
	Message string `json:"message,omitempty"`
	// current release version
	Version int `json:"version,omitempty"`
	// deploy status list of history, which will store at most 10 state
	DeployStatus []HelmReleaseDeployStatus `json:"deployStatus,omitempty"`
	// last update time
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
	// last deploy time or upgrade time
	LastDeployed *metav1.Time `json:"lastDeployed,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=hrls
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Release Name",type=string,JSONPath=".spec.name"
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.kubesphere\\.io/cluster"
// +kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/namespace"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true

// HelmRelease is the Schema for the helmreleases API
type HelmRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmReleaseSpec   `json:"spec,omitempty"`
	Status HelmReleaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:root=true

// HelmReleaseList contains a list of HelmRelease
type HelmReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmRelease{}, &HelmReleaseList{})
}

func (in *HelmRelease) GetCreator() string {
	return getValue(in.Annotations, constants.CreatorAnnotationKey)
}

func (in *HelmRelease) GetTrueName() string {
	return in.Spec.Name
}

func (in *HelmRelease) GetChartVersionName() string {
	appV := in.GetChartAppVersion()
	if appV != "" {
		return fmt.Sprintf("%s [%s]", in.GetChartVersion(), appV)
	} else {
		return in.GetChartVersion()
	}
}

func (in *HelmRelease) GetChartAppVersion() string {
	return in.Spec.ChartAppVersion
}

func (in *HelmRelease) GetChartVersion() string {
	return in.Spec.ChartVersion
}

func (in *HelmRelease) GetRlsCluster() string {
	return getValue(in.Labels, constants.ClusterNameLabelKey)
}

func (in *HelmRelease) GetWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

func (in *HelmRelease) GetRlsNamespace() string {
	return getValue(in.Labels, constants.NamespaceLabelKey)
}
