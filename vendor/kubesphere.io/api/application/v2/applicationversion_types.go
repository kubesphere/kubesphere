package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/constants"
)

// ApplicationVersionSpec defines the desired state of ApplicationVersion
type ApplicationVersionSpec struct {
	VersionName string       `json:"versionName"`
	AppHome     string       `json:"appHome,omitempty"`
	Icon        string       `json:"icon,omitempty"`
	Created     *metav1.Time `json:"created,omitempty"`
	Digest      string       `json:"digest,omitempty"`
	AppType     string       `json:"appType,omitempty"`
	Maintainer  []Maintainer `json:"maintainer,omitempty"`
	PullUrl     string       `json:"pullUrl,omitempty"`
}

// ApplicationVersionStatus defines the observed state of ApplicationVersion
type ApplicationVersionStatus struct {
	State    string       `json:"state,omitempty"`
	Message  string       `json:"message,omitempty"`
	UserName string       `json:"userName,omitempty"`
	Updated  *metav1.Time `json:"updated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=appver
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="repo",type="string",JSONPath=".metadata.labels.application\\.kubesphere\\.io/repo-name"
// +kubebuilder:printcolumn:name="workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="app",type="string",JSONPath=".metadata.labels.application\\.kubesphere\\.io/app-id"
// +kubebuilder:printcolumn:name="appType",type="string",JSONPath=".spec.appType"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ApplicationVersion is the Schema for the applicationversions API
type ApplicationVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationVersionSpec   `json:"spec,omitempty"`
	Status ApplicationVersionStatus `json:"status,omitempty"`
}

// Maintainer describes a Chart maintainer.
type Maintainer struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Metadata for a Application detail.
type Metadata struct {
	Version string   `json:"version"`
	Home    string   `json:"home,omitempty"`
	Icon    string   `json:"icon,omitempty"`
	Sources []string `json:"sources,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationVersionList contains a list of ApplicationVersion
type ApplicationVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationVersion `json:"items"`
}

func (in *ApplicationVersion) GetCreator() string {
	return getValue(in.Annotations, constants.CreatorAnnotationKey)
}

func (in *ApplicationVersion) GetWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

func (in *ApplicationVersion) GetAppID() string {
	return getValue(in.Labels, AppIDLabelKey)
}
