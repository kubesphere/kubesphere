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

// ServiceDNSRecordSpec defines the desired state of ServiceDNSRecord.
type ServiceDNSRecordSpec struct {
	// DomainRef is the name of the domain object to which the corresponding federated service belongs
	DomainRef string `json:"domainRef"`
	// RecordTTL is the TTL in seconds for DNS records created for this Service, if omitted a default would be used
	RecordTTL TTL `json:"recordTTL,omitempty"`
	// DNSPrefix when specified, an additional DNS record would be created with <DNSPrefix>.<KubeFedDomain>
	DNSPrefix string `json:"dnsPrefix,omitempty"`
	// ExternalName when specified, replaces the service name portion of a resource record
	// with the value of ExternalName.
	ExternalName string `json:"externalName,omitempty"`
	// AllowServiceWithoutEndpoints allows DNS records to be written for Service shards without endpoints
	AllowServiceWithoutEndpoints bool `json:"allowServiceWithoutEndpoints,omitempty"`
}

// ServiceDNSRecordStatus defines the observed state of ServiceDNSRecord.
type ServiceDNSRecordStatus struct {
	// Domain is the DNS domain of the KubeFed control plane as in Domain API
	Domain string       `json:"domain,omitempty"`
	DNS    []ClusterDNS `json:"dns,omitempty"`
}

// ClusterDNS defines the observed status of LoadBalancer within a cluster.
type ClusterDNS struct {
	// Cluster name
	Cluster string `json:"cluster,omitempty"`
	// LoadBalancer for the corresponding service
	LoadBalancer corev1.LoadBalancerStatus `json:"loadBalancer,omitempty"`
	// Zones to which the cluster belongs
	Zones []string `json:"zones,omitempty"`
	// Region to which the cluster belongs
	Region string `json:"region,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceDNSRecord defines a scheme of DNS name and subdomains that
// should be programmed with endpoint information about a Service deployed in
// multiple Kubernetes clusters. ServiceDNSRecord is name-associated
// with the Services it programs endpoint information for, meaning that a
// ServiceDNSRecord expresses the intent to program DNS with
// information about endpoints for the Kubernetes Service resources with the
// same name and namespace in different clusters.
//
// For the example, given the following values:
//
// metadata.name: test-service
// metadata.namespace: test-namespace
// spec.federationName: test-federation
//
// the following set of DNS names will be programmed:
//
// Global Level: test-service.test-namespace.test-federation.svc.<federation-domain>
// Region Level: test-service.test-namespace.test-federation.svc.(status.DNS[*].region).<federation-domain>
// Zone Level  : test-service.test-namespace.test-federation.svc.(status.DNS[*].zone).(status.DNS[*].region).<federation-domain>
//
// Optionally, when DNSPrefix is specified, another DNS name will be programmed
// which would be a CNAME record pointing to DNS name at global level as below:
// <dns-prefix>.<federation-domain> --> test-service.test-namespace.test-federation.svc.<federation-domain>
//
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=servicednsrecords
// +kubebuilder:subresource:status
type ServiceDNSRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceDNSRecordSpec   `json:"spec,omitempty"`
	Status ServiceDNSRecordStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceDNSRecordList contains a list of ServiceDNSRecord
type ServiceDNSRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceDNSRecord `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceDNSRecord{}, &ServiceDNSRecordList{})
}
