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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressDNSRecordSpec defines the desired state of IngressDNSRecord
type IngressDNSRecordSpec struct {
	// Host from the IngressRule in Cluster Ingress Spec
	Hosts []string `json:"hosts,omitempty"`
	// RecordTTL is the TTL in seconds for DNS records created for the Ingress, if omitted a default would be used
	RecordTTL TTL `json:"recordTTL,omitempty"`
}

// IngressDNSRecordStatus defines the observed state of IngressDNSRecord
type IngressDNSRecordStatus struct {
	// Array of Ingress Controller LoadBalancers
	DNS []ClusterIngressDNS `json:"dns,omitempty"`
}

// ClusterIngressDNS defines the observed status of Ingress within a cluster.
type ClusterIngressDNS struct {
	// Cluster name
	Cluster string `json:"cluster,omitempty"`
	// LoadBalancer for the corresponding ingress controller
	LoadBalancer corev1.LoadBalancerStatus `json:"loadBalancer,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=ingressdnsrecords
// +kubebuilder:subresource:status

type IngressDNSRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IngressDNSRecordSpec   `json:"spec,omitempty"`
	Status IngressDNSRecordStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IngressDNSRecordList contains a list of IngressDNSRecord
type IngressDNSRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IngressDNSRecord `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IngressDNSRecord{}, &IngressDNSRecordList{})
}
