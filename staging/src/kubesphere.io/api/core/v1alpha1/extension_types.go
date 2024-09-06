package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InstallationMode string

const (
	InstallationModeHostOnly InstallationMode = "HostOnly"
	InstallationMulticluster InstallationMode = "Multicluster"
)

// Provider describes an extension provider.
type Provider struct {
	// Name is a username or organization name
	Name string `json:"name,omitempty"`
	// URL is an optional URL to an address for the named provider
	URL string `json:"url,omitempty"`
	// Email is an optional email address to contact the named provider
	Email string `json:"email,omitempty"`
}

// ExtensionInfo describes an extension's basic information.
type ExtensionInfo struct {
	// +optional
	DisplayName Locales `json:"displayName,omitempty"`
	// +optional
	Description Locales `json:"description,omitempty"`
	// +optional
	Icon string `json:"icon,omitempty"`
	// +optional
	Provider map[LanguageCode]*Provider `json:"provider,omitempty"`
	// +optional
	Created metav1.Time `json:"created,omitempty"`
}

// ExtensionSpec only contains basic extension information copied from the latest ExtensionVersion.
type ExtensionSpec struct {
	ExtensionInfo `json:",inline"`
}

// ExtensionVersionSpec contains the details of a specific version extension.
type ExtensionVersionSpec struct {
	ExtensionInfo `json:",inline"`
	Version       string   `json:"version,omitempty"`
	Keywords      []string `json:"keywords,omitempty"`
	Sources       []string `json:"sources,omitempty"`
	Repository    string   `json:"repository,omitempty"`
	Category      string   `json:"category,omitempty"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	// eg: >= 1.2.0, see https://github.com/Masterminds/semver for more info.
	KubeVersion string `json:"kubeVersion,omitempty"`
	// KSVersion is a SemVer constraint specifying the version of KubeSphere required.
	// eg: >= 1.2.0, see https://github.com/Masterminds/semver for more info.
	KSVersion   string   `json:"ksVersion,omitempty"`
	Home        string   `json:"home,omitempty"`
	Docs        string   `json:"docs,omitempty"`
	Digest      string   `json:"digest,omitempty"`
	Screenshots []string `json:"screenshots,omitempty"`
	// ChartDataRef refers to a configMap which contains raw chart data.
	ChartDataRef *ConfigMapKeyRef `json:"chartDataRef,omitempty"`
	ChartURL     string           `json:"chartURL,omitempty"`
	// Namespace represents the namespace in which the extension is installed.
	// If empty, it will be installed in the namespace named extension-{name}.
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// +kubebuilder:default:=HostOnly
	// +kubebuilder:validation:Enum=HostOnly;Multicluster
	InstallationMode InstallationMode `json:"installationMode,omitempty"`
	// ExternalDependencies
	ExternalDependencies []ExternalDependency `json:"externalDependencies,omitempty"`
}

type ExternalDependency struct {
	// Name of the external dependency
	Name string `json:"name"`
	// Type of dependency, defaults to extension
	// +optional
	Type string `json:"type,omitempty"`
	// SemVer
	Version string `json:"version"`
	// Indicates if the dependency is required
	Required bool `json:"required"`
}

type ConfigMapKeyRef struct {
	corev1.ConfigMapKeySelector `json:",inline"`
	Namespace                   string `json:"namespace"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="extensions",scope="Cluster"

type ExtensionVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExtensionVersionSpec `json:"spec,omitempty"`
}

type CategorySpec struct {
	DisplayName Locales `json:"displayName,omitempty"`
	Description Locales `json:"description,omitempty"`
	Icon        string  `json:"icon,omitempty"`
}

// +kubebuilder:object:root=true
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
	State                 string                 `json:"state,omitempty"`
	Enabled               bool                   `json:"enabled,omitempty"`
	PlannedInstallVersion string                 `json:"plannedInstallVersion,omitempty"`
	InstalledVersion      string                 `json:"installedVersion,omitempty"`
	RecommendedVersion    string                 `json:"recommendedVersion,omitempty"`
	Versions              []ExtensionVersionInfo `json:"versions,omitempty"`
	Conditions            []metav1.Condition     `json:"conditions,omitempty"`
	// +optional
	// ClusterSchedulingStatuses describes the subchart installation status of the extension
	ClusterSchedulingStatuses map[string]InstallationStatus `json:"clusterSchedulingStatuses,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="extensions",scope="Cluster"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// Extension is synchronized from the Repository.
// An extension can contain multiple versions.
type Extension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExtensionSpec   `json:"spec,omitempty"`
	Status            ExtensionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ExtensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Extension `json:"items"`
}

// +kubebuilder:object:root=true

type ExtensionVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExtensionVersion `json:"items"`
}

// +kubebuilder:object:root=true

type CategoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Category `json:"items"`
}
