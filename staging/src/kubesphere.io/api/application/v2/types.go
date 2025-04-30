/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/api/constants"
)

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	AppHome     string                 `json:"appHome,omitempty"`
	AppType     string                 `json:"appType,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Abstraction string                 `json:"abstraction,omitempty"`
	Attachments []string               `json:"attachments,omitempty"`
	Resources   []GroupVersionResource `json:"resources,omitempty"`
}

type GroupVersionResource struct {
	Group      string `json:"Group,omitempty"`
	Version    string `json:"Version,omitempty"`
	Resource   string `json:"Resource,omitempty"`
	Name       string `json:"Name,omitempty"`
	Desc       string `json:"Desc,omitempty"`
	ParentNode string `json:"ParentNode,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// the state of the helm application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state,omitempty"`
	UpdateTime *metav1.Time `json:"updateTime"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=app
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="repo",type="string",JSONPath=".metadata.labels.application\\.kubesphere\\.io/repo-name"
// +kubebuilder:printcolumn:name="workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="appType",type="string",JSONPath=".spec.appType"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func getValue(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}

func (in *Application) GetWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

// ApplicationReleaseSpec defines the desired state of ApplicationRelease
type ApplicationReleaseSpec struct {
	AppID        string `json:"appID"`
	AppVersionID string `json:"appVersionID"`
	Values       []byte `json:"values,omitempty"`
	AppType      string `json:"appType,omitempty"`
	Icon         string `json:"icon,omitempty"`
}

// ApplicationReleaseStatus defines the observed state of ApplicationRelease
type ApplicationReleaseStatus struct {
	State             string            `json:"state"`
	Message           string            `json:"message,omitempty"`
	SpecHash          string            `json:"specHash,omitempty"`
	InstallJobName    string            `json:"installJobName,omitempty"`
	UninstallJobName  string            `json:"uninstallJobName,omitempty"`
	LastUpdate        metav1.Time       `json:"lastUpdate,omitempty"`
	RealTimeResources []json.RawMessage `json:"realTimeResources,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=apprls
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="app",type="string",JSONPath=".metadata.labels.application\\.kubesphere\\.io/app-id"
// +kubebuilder:printcolumn:name="appversion",type="string",JSONPath=".metadata.labels.application\\.kubesphere\\.io/appversion-id"
// +kubebuilder:printcolumn:name="appType",type="string",JSONPath=".spec.appType"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.kubesphere\\.io/cluster"
// +kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/namespace"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ApplicationRelease is the Schema for the applicationreleases API
type ApplicationRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationReleaseSpec   `json:"spec,omitempty"`
	Status ApplicationReleaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationReleaseList contains a list of ApplicationRelease
type ApplicationReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationRelease `json:"items"`
}

func (in *ApplicationRelease) GetCreator() string {
	return getValue(in.Annotations, constants.CreatorAnnotationKey)
}

func (in *ApplicationRelease) GetRlsCluster() string {
	name := getValue(in.Labels, constants.ClusterNameLabelKey)
	if name != "" {
		return name
	}
	//todo remove hardcode
	return "host"
}

func (in *ApplicationRelease) GetRlsNamespace() string {
	ns := getValue(in.Labels, constants.NamespaceLabelKey)
	if ns == "" {
		return "default"
	}
	return ns
}

func (in *ApplicationRelease) HashSpec() string {
	specJSON, _ := json.Marshal(in.Spec)
	return fmt.Sprintf("%x", md5.Sum(specJSON))
}

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

// CategorySpec defines the desired state of HelmRepo
type CategorySpec struct {
	Icon string `json:"icon,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=appctg
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="total",type=string,JSONPath=`.status.total`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Category is the Schema for the categories API
type Category struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CategorySpec   `json:"spec,omitempty"`
	Status CategoryStatus `json:"status,omitempty"`
}

type CategoryStatus struct {
	Total int `json:"total"`
}

// +kubebuilder:object:root=true

// CategoryList contains a list of Category
type CategoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Category `json:"items"`
}

type RepoCredential struct {
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
}

// RepoSpec defines the desired state of Repo
type RepoSpec struct {
	Url         string         `json:"url"`
	Credential  RepoCredential `json:"credential,omitempty"`
	Description string         `json:"description,omitempty"`
	SyncPeriod  *int           `json:"syncPeriod"`
}

// RepoStatus defines the observed state of Repo
type RepoStatus struct {
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	State          string      `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,path=repos,shortName=repo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Workspace",type="string",JSONPath=".metadata.labels.kubesphere\\.io/workspace"
// +kubebuilder:printcolumn:name="url",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Repo is the Schema for the repoes API
type Repo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepoSpec   `json:"spec,omitempty"`
	Status RepoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepoList contains a list of Repo
type RepoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repo `json:"items"`
}

func (in *Repo) GetWorkspace() string {
	return getValue(in.Labels, constants.WorkspaceLabelKey)
}

func (in *Repo) GetCreator() string {
	return getValue(in.Annotations, constants.CreatorAnnotationKey)
}
