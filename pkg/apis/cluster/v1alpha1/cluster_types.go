package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindCluster      = "Cluster"
	ResourcesSingularCluster = "cluster"
	ResourcesPluralCluster   = "clusters"

	IsHostCluster = "cluster.kubesphere.io/is-host-cluster"
)

type ClusterSpec struct {

	// Join cluster as a kubefed cluster
	// +optional
	Federated bool `json:"federated,omitempty"`

	// Desired state of the cluster
	Active bool `json:"active,omitempty"`

	// Provider of the cluster, this field is just for description
	// +optional
	Provider string `json:"provider,omitempty"`
}

type ClusterConditionType string

const (
	// Cluster agent is initialized and waiting for connecting
	ClusterInitialized ClusterConditionType = "Initialized"

	// Cluster agent is available
	ClusterAgentAvailable ClusterConditionType = "AgentAvailable"

	// Cluster has been one of federated clusters
	ClusterFederated ClusterConditionType = "Federated"
)

type ClusterCondition struct {
	// Type of the condition
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

type ClusterStatus struct {

	// Represents the latest available observations of a cluster's current state.
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// GitVersion of the kubernetes cluster, this field is set by cluster controller
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// Count of the kubernetes cluster nodes
	// +optional
	NodeCount int `json:"nodeCount,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +genclient:nonNamespaced
// +kubebuilder:printcolumn:name="Federated",type="boolean",JSONPath=".spec.federated"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="Active",type="boolean",JSONPath=".spec.active"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.kubernetesVersion"
// +kubebuilder:resource:scope=Cluster

// Cluster is the schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
