/*


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
	"github.com/go-openapi/strfmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/constants"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HelmReleaseSpec defines the desired state of HelmRelease
type HelmReleaseSpec struct {
	// release workspace
	Workspace string `json:"workspace,omitempty"`
	// release name
	Name string `json:"name"`
	// release version
	Version int `json:"version"`
	// chart data
	ChartData strfmt.Base64 `json:"chartData,omitempty"`
	// helm release description msg
	Description string `json:"description,omitempty"`
	// release values.yaml, base64 encoded
	Values strfmt.Base64 `json:"values,omitempty"`
	// chart name
	ChartName string `json:"chartName"`
	// specify the exact chart version to install. If this is not specified, the latest version is installed
	ChartVersion string `json:"chartVersion"`
	// appVersion from Chart.yaml
	ChartAppVersion string `json:"chartAppVer,omitempty"`
	RepoId          string `json:"repoId,omitempty"`
	// use chart from helmapplication
	ApplicationId        string `json:"appId,omitempty"`
	ApplicationVersionId string `json:"appVerId,omitempty"`
}

type HelmReleaseDeployStatus struct {
	Message string `json:"message,omitempty"`
	//last deploy state
	State string      `json:"state"`
	Time  metav1.Time `json:"deployTime"`
}

// HelmReleaseStatus defines the observed state of HelmRelease
type HelmReleaseStatus struct {
	//current state
	State   string `json:"state"`
	Message string `json:"message,omitempty"`
	// current release version
	Version int `json:"version,omitempty"`
	// md5(Values)
	Hash         string                    `json:"hash,omitempty"`
	DeployStatus []HelmReleaseDeployStatus `json:"deployStatus,omitempty"`
	// last update time
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
	// last successful deploy time
	LastDeployed *metav1.Time `json:"lastDeployed,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Release Name",type=string,JSONPath=".spec.name"
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".spec.workspace"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +genclient

// HelmRelease is the Schema for the helmreleases API
type HelmRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmReleaseSpec   `json:"spec,omitempty"`
	Status HelmReleaseStatus `json:"status,omitempty"`
}

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
	if in == nil || in.Annotations == nil {
		return ""
	}
	return in.Annotations[constants.CreatorAnnotationKey]
}
