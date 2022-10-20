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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/constants"
)

const (
	ResourceKindHelmApplicationVersion     = "HelmApplicationVersion"
	ResourceSingularHelmApplicationVersion = "helmapplicationversion"
	ResourcePluralHelmApplicationVersion   = "helmapplicationversions"
)

// HelmApplicationVersionSpec defines the desired state of HelmApplicationVersion
type HelmApplicationVersionSpec struct {
	// metadata from chart
	*Metadata `json:",inline"`
	// chart url
	URLs []string `json:"urls,omitempty"`
	// raw data of chart, it will !!!NOT!!! be save to etcd
	Data []byte `json:"data,omitempty"`

	// dataKey in the storage
	DataKey string `json:"dataKey,omitempty"`

	// chart create time
	Created *metav1.Time `json:"created,omitempty"`

	// chart digest
	Digest string `json:"digest,omitempty"`
}

// HelmApplicationVersionStatus defines the observed state of HelmApplicationVersion
type HelmApplicationVersionStatus struct {
	State string  `json:"state,omitempty"`
	Audit []Audit `json:"audit,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=happver
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="application name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true

// HelmApplicationVersion is the Schema for the helmapplicationversions API
type HelmApplicationVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmApplicationVersionSpec   `json:"spec,omitempty"`
	Status HelmApplicationVersionStatus `json:"status,omitempty"`
}

// Maintainer describes a Chart maintainer.
type Maintainer struct {
	// Name is a user name or organization name
	Name string `json:"name,omitempty"`
	// Email is an optional email address to contact the named maintainer
	Email string `json:"email,omitempty"`
	// URL is an optional URL to an address for the named maintainer
	URL string `json:"url,omitempty"`
}

// Metadata for a Chart file. This models the structure of a Chart.yaml file.
type Metadata struct {
	// The name of the chart
	Name string `json:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `json:"home,omitempty"`
	// Source is the URL to the source code of this chart
	Sources []string `json:"sources,omitempty"`
	// A SemVer 2 conformant version string of the chart
	Version string `json:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `json:"description,omitempty"`
	// A list of string keywords
	Keywords []string `json:"keywords,omitempty"`
	// A list of name and URL/email address combinations for the maintainer(s)
	Maintainers []*Maintainer `json:"maintainers,omitempty"`
	// The URL to an icon file.
	Icon string `json:"icon,omitempty"`
	// The API Version of this chart.
	APIVersion string `json:"apiVersion,omitempty"`
	// The condition to check to enable chart
	Condition string `json:"condition,omitempty"`
	// The tags to check to enable chart
	Tags string `json:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `json:"deprecated,omitempty"`
	// Annotations are additional mappings uninterpreted by Helm,
	// made available for inspection by other applications.
	Annotations map[string]string `json:"annotations,omitempty"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string `json:"kubeVersion,omitempty"`
	// Dependencies are a list of dependencies for a chart.
	Dependencies []*Dependency `json:"dependencies,omitempty"`
	// Specifies the chart type: application or library
	Type string `json:"type,omitempty"`
}

type Audit struct {
	// audit message
	Message string `json:"message,omitempty"`
	// audit state: submitted, passed, draft, active, rejected, suspended
	State string `json:"state,omitempty"`
	// audit time
	Time metav1.Time `json:"time"`
	// audit operator
	Operator     string `json:"operator,omitempty"`
	OperatorType string `json:"operatorType,omitempty"`
}

// Dependency describes a chart upon which another chart depends.
// Dependencies can be used to express developer intent, or to capture the state
// of a chart.
type Dependency struct {
	// Name is the name of the dependency.
	// This must mach the name in the dependency's Chart.yaml.
	Name string `json:"name"`
	// Version is the version (range) of this chart.
	// A lock file will always produce a single version, while a dependency
	// may contain a semantic version range.
	Version string `json:"version,omitempty"`
	// The URL to the repository.
	// Appending `index.yaml` to this string should result in a URL that can be
	// used to fetch the repository index.
	Repository string `json:"repository"`
	// A yaml path that resolves to a boolean, used for enabling/disabling charts (e.g. subchart1.enabled )
	Condition string `json:"condition,omitempty"`
	// Tags can be used to group charts for enabling/disabling together
	Tags []string `json:"tags,omitempty"`
	// Enabled bool determines if chart should be loaded
	Enabled bool `json:"enabled,omitempty"`
	// ImportValues holds the mapping of source values to parent key to be imported. Each item can be a
	// string or pair of child/parent sublist items.
	// ImportValues []interface{} `json:"import_values,omitempty"`

	// Alias usable alias to be used for the chart
	Alias string `json:"alias,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:root=true

// HelmApplicationVersionList contains a list of HelmApplicationVersion
type HelmApplicationVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmApplicationVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmApplicationVersion{}, &HelmApplicationVersionList{})
}

func (in *HelmApplicationVersion) GetCreator() string {
	return getValue(in.Annotations, constants.CreatorAnnotationKey)
}

func (in *HelmApplicationVersion) GetHelmApplicationVersionId() string {
	return in.Name
}

func (in *HelmApplicationVersion) GetWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

func (in *HelmApplicationVersion) GetVersionName() string {
	appV := in.GetChartAppVersion()
	if appV != "" {
		return fmt.Sprintf("%s [%s]", in.GetChartVersion(), appV)
	} else {
		return in.GetChartVersion()
	}
}

func (in *HelmApplicationVersion) GetHelmApplicationId() string {
	return getValue(in.Labels, constants.ChartApplicationIdLabelKey)
}

func (in *HelmApplicationVersion) GetSemver() string {
	return strings.Split(in.GetVersionName(), " ")[0]
}

func (in *HelmApplicationVersion) GetTrueName() string {
	return in.Spec.Name
}

func (in *HelmApplicationVersion) GetChartVersion() string {
	return in.Spec.Version
}

func (in *HelmApplicationVersion) GetChartAppVersion() string {
	return in.Spec.AppVersion
}

func (in *HelmApplicationVersion) GetHelmRepoId() string {
	return getValue(in.Labels, constants.ChartRepoIdLabelKey)
}

func (in *HelmApplicationVersion) State() string {
	if in.Status.State == "" {
		return StateDraft
	}

	return in.Status.State
}
