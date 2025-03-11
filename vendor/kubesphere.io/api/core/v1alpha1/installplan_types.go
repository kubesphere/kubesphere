package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindInstallPlan = "InstallPlan"

	Automatic UpgradeStrategy = "Automatic"
	Manual    UpgradeStrategy = "Manual"
)

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
