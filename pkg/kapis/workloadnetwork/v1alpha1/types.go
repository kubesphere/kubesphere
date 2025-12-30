/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// NetworkConfig represents the network configuration options for a workload.
// This mirrors the network-related fields in Kubernetes PodSpec.
type NetworkConfig struct {
	// DNSPolicy specifies the DNS policy for the pod.
	// Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// +optional
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// DNSConfig specifies the DNS parameters of a pod.
	// Parameters specified here will be merged to the generated DNS configuration based on DNSPolicy.
	// +optional
	DNSConfig *corev1.PodDNSConfig `json:"dnsConfig,omitempty"`

	// HostNetwork indicates whether the pod should use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// Default to false.
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// HostPID indicates whether the pod should use the host's pid namespace.
	// Default to false.
	// +optional
	HostPID bool `json:"hostPID,omitempty"`

	// HostIPC indicates whether the pod should use the host's ipc namespace.
	// Default to false.
	// +optional
	HostIPC bool `json:"hostIPC,omitempty"`

	// Hostname specifies the hostname of the pod.
	// If not specified, the pod's hostname will be set to a system-defined value.
	// +optional
	Hostname string `json:"hostname,omitempty"`

	// Subdomain specifies the subdomain of the pod.
	// If specified, the fully qualified Pod hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>".
	// +optional
	Subdomain string `json:"subdomain,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts file.
	// This is only valid for non-hostNetwork pods.
	// +optional
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty"`

	// ShareProcessNamespace enables sharing a single process namespace between all containers in a pod.
	// When enabled, containers can see and signal processes from other containers in the pod.
	// +optional
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty"`
}

// WorkloadNetworkConfigRequest is the request body for updating network configuration
// on a workload resource.
type WorkloadNetworkConfigRequest struct {
	// Kind is the workload kind (Deployment, StatefulSet, DaemonSet, Job)
	Kind string `json:"kind"`

	// NetworkConfig contains the network configuration to apply
	NetworkConfig NetworkConfig `json:"networkConfig"`
}

// WorkloadNetworkConfigResponse is the response body containing the current
// network configuration of a workload.
type WorkloadNetworkConfigResponse struct {
	// Kind is the workload kind
	Kind string `json:"kind"`

	// Name is the workload name
	Name string `json:"name"`

	// Namespace is the workload namespace
	Namespace string `json:"namespace"`

	// NetworkConfig contains the current network configuration
	NetworkConfig NetworkConfig `json:"networkConfig"`
}
