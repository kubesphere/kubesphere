/*
Copyright 2018 The Kubernetes Authors.

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

package v1beta1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/kubefed/pkg/apis/core/common"
)

type TLSValidation string

const (
	TLSAll            TLSValidation = "*"
	TLSSubjectName    TLSValidation = "SubjectName"
	TLSValidityPeriod TLSValidation = "ValidityPeriod"
)

// KubeFedClusterSpec defines the desired state of KubeFedCluster
type KubeFedClusterSpec struct {
	// The API endpoint of the member cluster. This can be a hostname,
	// hostname:port, IP or IP:port.
	APIEndpoint string `json:"apiEndpoint"`

	// CABundle contains the certificate authority information.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`

	// Name of the secret containing the token required to access the
	// member cluster. The secret needs to exist in the same namespace
	// as the control plane and should have a "token" key.
	SecretRef LocalSecretReference `json:"secretRef"`

	// DisabledTLSValidations defines a list of checks to ignore when validating
	// the TLS connection to the member cluster.  This can be any of *, SubjectName, or ValidityPeriod.
	// If * is specified, it is expected to be the only option in list.
	// +optional
	DisabledTLSValidations []TLSValidation `json:"disabledTLSValidations,omitempty"`

	// ProxyURL allows to set proxy URL for the cluster.
	// +optional
	ProxyURL string `json:"proxyURL"`
}

// LocalSecretReference is a reference to a secret within the enclosing
// namespace.
type LocalSecretReference struct {
	// Name of a secret within the enclosing
	// namespace
	Name string `json:"name"`
}

// KubeFedClusterStatus contains information about the current status of a
// cluster updated periodically by cluster controller.
type KubeFedClusterStatus struct {
	// Conditions is an array of current cluster conditions.
	Conditions []ClusterCondition `json:"conditions"`
	// KubernetesVersion is the Kubernetes git version of the cluster.
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
	// Zones are the names of availability zones in which the nodes of the cluster exist, e.g. 'us-east1-a'.
	// +optional
	Zones []string `json:"zones,omitempty"`
	// Region is the name of the region in which all of the nodes in the cluster exist.  e.g. 'us-east1'.
	// +optional
	Region *string `json:"region,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name=age,type=date,JSONPath=.metadata.creationTimestamp
// +kubebuilder:printcolumn:name=ready,type=string,JSONPath=.status.conditions[?(@.type=='Ready')].status
// +kubebuilder:printcolumn:name=kubernetes-version,type=string,JSONPath=.status.kubernetesVersion
// +kubebuilder:resource:path=kubefedclusters,shortName=kfc
// +kubebuilder:subresource:status

// KubeFedCluster configures KubeFed to be aware of a Kubernetes
// cluster and encapsulates the details necessary to communicate with
// the cluster.
type KubeFedCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KubeFedClusterSpec `json:"spec"`
	// +optional
	Status KubeFedClusterStatus `json:"status,omitempty"`
}

// ClusterCondition describes current state of a cluster.
type ClusterCondition struct {
	// Type of cluster condition, Ready or Offline.
	Type common.ClusterConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status apiv1.ConditionStatus `json:"status"`
	// Last time the condition was checked.
	LastProbeTime metav1.Time `json:"lastProbeTime"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason *string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	// +optional
	Message *string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true

// KubeFedClusterList contains a list of KubeFedCluster
type KubeFedClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeFedCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeFedCluster{}, &KubeFedClusterList{})
}
