package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	ResourceKindCluster      = "Cluster"
	ResourcesSingularCluster = "cluster"
	ResourcesPluralCluster   = "clusters"

	HostCluster = "cluster-role.kubesphere.io/host"
	// ClusterRegion is the description of which region the cluster been placed
	ClusterRegion = "cluster.kubesphere.io/region"
	// ClusterGroup is the name of the cluster group
	ClusterGroup            = "cluster.kubesphere.io/group"
	ClusterVisibilityLabel  = "cluster.kubesphere.io/visibility"
	ClusterVisibilityPublic = "public"
	Finalizer               = "finalizer.cluster.kubesphere.io"

	AnnotationClusterName     = "cluster.kubesphere.io/name"
	AnnotationHostClusterName = "cluster.kubesphere.io/host-cluster"

	ClusterRoleHost   ClusterRole = "host"
	ClusterRoleMember ClusterRole = "member"
)

type ClusterRole string

type ClusterSpec struct {
	// Join cluster as a kubefed cluster
	// Deprecated: will be removed in the next version.
	JoinFederation bool `json:"joinFederation,omitempty"`

	// Desired state of the cluster
	// Deprecated: will be removed in the next version.
	Enable bool `json:"enable,omitempty"`

	// Provider of the cluster, this field is just for description
	Provider string `json:"provider,omitempty"`

	// Connection holds info to connect to the member cluster
	Connection Connection `json:"connection,omitempty"`

	// Config represents the custom helm chart values used when installing the cluster
	Config []byte `json:"config,omitempty"`

	// ExternalKubeAPIEnabled export kube-apiserver to public use a lb type service if connection type is proxy
	ExternalKubeAPIEnabled bool `json:"externalKubeAPIEnabled,omitempty"`
}

type ConnectionType string

const (
	ConnectionTypeDirect ConnectionType = "direct"
	ConnectionTypeProxy  ConnectionType = "proxy"
)

type Connection struct {
	// type defines how host cluster will connect to host cluster
	// ConnectionTypeDirect means direct connection, this requires
	//   kubeconfig and kubesphere apiserver endpoint provided
	// ConnectionTypeProxy means using kubesphere proxy, no kubeconfig
	//   or kubesphere apiserver endpoint required
	Type ConnectionType `json:"type,omitempty"`

	// KubeSphere API Server endpoint. Example: http://10.10.0.11:8080
	// Should provide this field explicitly if connection type is direct.
	// Will be populated by ks-apiserver if connection type is proxy.
	KubeSphereAPIEndpoint string `json:"kubesphereAPIEndpoint,omitempty"`

	// Kubernetes API Server endpoint. Example: https://10.10.0.1:6443
	// Should provide this field explicitly if connection type is direct.
	// Will be populated by ks-apiserver if connection type is proxy.
	KubernetesAPIEndpoint string `json:"kubernetesAPIEndpoint,omitempty"`

	// External Kubernetes API Server endpoint
	// Will be populated by ks-apiserver if connection type is proxy and ExternalKubeAPIEnabled is true.
	ExternalKubernetesAPIEndpoint string `json:"externalKubernetesAPIEndpoint,omitempty"`

	// KubeConfig content used to connect to cluster api server
	// Should provide this field explicitly if connection type is direct.
	// Will be populated by ks-proxy if connection type is proxy.
	KubeConfig []byte `json:"kubeconfig,omitempty"`

	// Token used by agents of member cluster to connect to host cluster proxy.
	// This field is populated by apiserver only if connection type is proxy.
	Token string `json:"token,omitempty"`

	// KubeAPIServerPort is the port which listens for forwarding kube-apiserver traffic
	// Only applicable when connection type is proxy.
	KubernetesAPIServerPort uint16 `json:"kubernetesAPIServerPort,omitempty"`

	// KubeSphereAPIServerPort is the port which listens for forwarding kubesphere apigateway traffic
	// Only applicable when connection type is proxy.
	KubeSphereAPIServerPort uint16 `json:"kubesphereAPIServerPort,omitempty"`
}

type ClusterConditionType string

const (
	// ClusterInitialized indicates the Cluster agent is initialized and waiting for connecting
	ClusterInitialized ClusterConditionType = "Initialized"

	// ClusterAgentAvailable indicates the Cluster agent is available
	ClusterAgentAvailable ClusterConditionType = "AgentAvailable"

	// ClusterFederated indicates the Cluster has been one of federated clusters
	// Deprecated: will be removed in the next version.
	ClusterFederated ClusterConditionType = "Federated"

	// ClusterExternalAccessReady indicates the Cluster external access ready
	ClusterExternalAccessReady ClusterConditionType = "ExternalAccessReady"

	// ClusterReady indicates the Cluster is all available for requests
	ClusterReady ClusterConditionType = "Ready"

	ClusterSchedulable ClusterConditionType = "Schedulable"

	// ClusterOpenPitrixRuntimeReady indicates the Openpitrix runtime is created
	ClusterOpenPitrixRuntimeReady ClusterConditionType = "OpenPitrixRuntimeReady"

	// ClusterKubeConfigCertExpiresInSevenDays indicates that the cluster certificate is about to expire.
	ClusterKubeConfigCertExpiresInSevenDays ClusterConditionType = "KubeConfigCertExpiresInSevenDays"

	ClusterKSCoreReady = "KSCoreReady"
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
	// A human-readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

type ClusterStatus struct {
	// Represents the latest available observations of a cluster's current state.
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// GitVersion of the kubernetes cluster, this field is populated by cluster controller
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// GitVersion of the /kapis/version api response, this field is populated by cluster controller
	KubeSphereVersion string `json:"kubeSphereVersion,omitempty"`

	// Count of the kubernetes cluster nodes
	// This field may not reflect the instant status of the cluster.
	NodeCount int `json:"nodeCount,omitempty"`

	// Zones are the names of availability zones in which the nodes of the cluster exist, e.g. 'us-east1-a'.
	// +optional
	Zones []string `json:"zones,omitempty"`

	// Region is the name of the region in which all of the nodes in the cluster exist.  e.g. 'us-east1'.
	// +optional
	Region *string `json:"region,omitempty"`

	// Configz is status of components enabled in the member cluster. This is synchronized with member cluster
	// every amount of time, like 5 minutes.
	// Deprecated: this field will be removed in the future version.
	// +optional
	Configz map[string]bool `json:"configz,omitempty"`

	// UID is the kube-system namespace UID of the cluster, which represents the unique ID of the cluster.
	UID types.UID `json:"uid,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.kubernetesVersion"
// +kubebuilder:resource:scope=Cluster

// Cluster is the schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
