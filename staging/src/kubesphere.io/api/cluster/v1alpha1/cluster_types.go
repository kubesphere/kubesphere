/*
Copyright 2020 KubeSphere Authors

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	ResourceKindCluster      = "Cluster"
	ResourcesSingularCluster = "cluster"
	ResourcesPluralCluster   = "clusters"

	HostCluster = "cluster-role.kubesphere.io/host"
	// Description of which region the cluster been placed
	ClusterRegion = "cluster.kubesphere.io/region"
	// Name of the cluster group
	ClusterGroup = "cluster.kubesphere.io/group"

	Finalizer = "finalizer.cluster.kubesphere.io"
)

type ClusterSpec struct {

	// Join cluster as a kubefed cluster
	JoinFederation bool `json:"joinFederation,omitempty"`

	// Desired state of the cluster
	Enable bool `json:"enable,omitempty"`

	// Provider of the cluster, this field is just for description
	Provider string `json:"provider,omitempty"`

	// Connection holds info to connect to the member cluster
	Connection Connection `json:"connection,omitempty"`

	// ExternalKubeAPIEnabled export kubeapiserver to public use a lb type service if connection type is proxy
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
	// Cluster agent is initialized and waiting for connecting
	ClusterInitialized ClusterConditionType = "Initialized"

	// Cluster agent is available
	ClusterAgentAvailable ClusterConditionType = "AgentAvailable"

	// Cluster has been one of federated clusters
	ClusterFederated ClusterConditionType = "Federated"

	// Cluster external access ready
	ClusterExternalAccessReady ClusterConditionType = "ExternalAccessReady"

	// Cluster is all available for requests
	ClusterReady ClusterConditionType = "Ready"

	// Openpitrix runtime is created
	ClusterOpenPitrixRuntimeReady ClusterConditionType = "OpenPitrixRuntimeReady"

	// ClusterKubeConfigCertExpiresInSevenDays indicates that the cluster certificate is about to expire.
	ClusterKubeConfigCertExpiresInSevenDays ClusterConditionType = "KubeConfigCertExpiresInSevenDays"
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
	// +optional
	Configz map[string]bool `json:"configz,omitempty"`

	// UID is the kube-system namespace UID of the cluster, which represents the unique ID of the cluster.
	UID types.UID `json:"uid,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +genclient:nonNamespaced
// +kubebuilder:printcolumn:name="Federated",type="boolean",JSONPath=".spec.joinFederation"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".spec.provider"
// +kubebuilder:printcolumn:name="Active",type="boolean",JSONPath=".spec.enable"
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
