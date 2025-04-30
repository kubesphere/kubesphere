/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InstallationMode string

const (
	InstallationModeHostOnly InstallationMode = "HostOnly"
	InstallationMulticluster InstallationMode = "Multicluster"

	ResourceKindInstallPlan                 = "InstallPlan"
	Automatic               UpgradeStrategy = "Automatic"
	Manual                  UpgradeStrategy = "Manual"

	ServiceAccountName            = "kubesphere.io/service-account.name"
	ServiceAccountUID             = "kubesphere.io/service-account.uid"
	ServiceAccountToken           = "token"
	SecretTypeServiceAccountToken = "kubesphere.io/service-account-token"

	ServiceAccountGroup                     = "kubesphere:serviceaccount"
	ServiceAccountTokenPrefix               = ServiceAccountGroup + ":"
	ServiceAccountTokenSubFormat            = ServiceAccountTokenPrefix + "%s:%s"
	ServiceAccountTokenExtraSecretNamespace = "secret-namespace"
	ServiceAccountTokenExtraSecretName      = "secret-name"
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
	Name          string   `json:"-"`
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

type Placement struct {
	// +listType=set
	// +optional
	Clusters        []string              `json:"clusters,omitempty"`
	ClusterSelector *metav1.LabelSelector `json:"clusterSelector,omitempty"`
}

type ClusterScheduling struct {
	Placement *Placement        `json:"placement,omitempty"`
	Overrides map[string]string `json:"overrides,omitempty"`
}

type InstallPlanState struct {
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	State              string      `json:"state"`
}

type InstallationStatus struct {
	State           string             `json:"state,omitempty"`
	ConfigHash      string             `json:"configHash,omitempty"`
	TargetNamespace string             `json:"targetNamespace,omitempty"`
	ReleaseName     string             `json:"releaseName,omitempty"`
	Version         string             `json:"version,omitempty"`
	JobName         string             `json:"jobName,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
	StateHistory    []InstallPlanState `json:"stateHistory,omitempty"`
}

type ExtensionRef struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type UpgradeStrategy string

type InstallPlanSpec struct {
	Extension ExtensionRef `json:"extension"`
	Enabled   bool         `json:"enabled"`
	// +kubebuilder:default:=Manual
	UpgradeStrategy   UpgradeStrategy    `json:"upgradeStrategy,omitempty"`
	Config            string             `json:"config,omitempty"`
	ClusterScheduling *ClusterScheduling `json:"clusterScheduling,omitempty"`
}

type InstallPlanStatus struct {
	InstallationStatus `json:",inline"`
	Enabled            bool `json:"enabled,omitempty"`
	// ClusterSchedulingStatuses describes the subchart installation status of the extension
	ClusterSchedulingStatuses map[string]InstallationStatus `json:"clusterSchedulingStatuses,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="extensions",scope="Cluster"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// InstallPlan defines how to install an extension in the cluster.
type InstallPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              InstallPlanSpec   `json:"spec,omitempty"`
	Status            InstallPlanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type InstallPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InstallPlan `json:"items"`
}

type UpdateStrategy struct {
	RegistryPoll `json:"registryPoll,omitempty"`
	Timeout      metav1.Duration `json:"timeout"`
}

type RegistryPoll struct {
	Interval metav1.Duration `json:"interval"`
}

type BasicAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type RepositorySpec struct {
	URL            string          `json:"url,omitempty"`
	Description    string          `json:"description,omitempty"`
	BasicAuth      *BasicAuth      `json:"basicAuth,omitempty"`
	UpdateStrategy *UpdateStrategy `json:"updateStrategy,omitempty"`
	// The caBundle (base64 string) is used in helmExecutor to verify the helm server.
	// +optional
	CABundle string `json:"caBundle,omitempty"`
	// --insecure-skip-tls-verify. default false
	Insecure bool `json:"insecure,omitempty"`
	// The maximum number of synchronized versions for each extension. A value of 0 indicates that all versions will be synchronized. The default is 3.
	// +optional
	Depth *int `json:"depth,omitempty"`
}

type RepositoryStatus struct {
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty'"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="extensions",scope="Cluster"

// Repository declared a docker image containing the extension helm chart.
// The extension manager controller will deploy and synchronizes the extensions from the image repository.
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Namespaced"

type ServiceAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Secrets []corev1.ObjectReference `json:"secrets,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// +kubebuilder:object:root=true

type ServiceAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccount `json:"items"`
}
