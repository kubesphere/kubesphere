/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestExtractNetworkConfig(t *testing.T) {
	tests := []struct {
		name     string
		podSpec  *corev1.PodSpec
		expected *NetworkConfig
	}{
		{
			name:    "empty pod spec",
			podSpec: &corev1.PodSpec{},
			expected: &NetworkConfig{
				DNSPolicy:             "",
				DNSConfig:             nil,
				HostNetwork:           false,
				HostPID:               false,
				HostIPC:               false,
				Hostname:              "",
				Subdomain:             "",
				HostAliases:           nil,
				ShareProcessNamespace: nil,
			},
		},
		{
			name: "full network config",
			podSpec: &corev1.PodSpec{
				DNSPolicy:   corev1.DNSClusterFirst,
				HostNetwork: true,
				HostPID:     true,
				HostIPC:     true,
				Hostname:    "test-host",
				Subdomain:   "test-subdomain",
				DNSConfig: &corev1.PodDNSConfig{
					Nameservers: []string{"8.8.8.8"},
					Searches:    []string{"example.com"},
				},
				HostAliases: []corev1.HostAlias{
					{IP: "127.0.0.1", Hostnames: []string{"localhost"}},
				},
				ShareProcessNamespace: ptr.To(true),
			},
			expected: &NetworkConfig{
				DNSPolicy:   corev1.DNSClusterFirst,
				HostNetwork: true,
				HostPID:     true,
				HostIPC:     true,
				Hostname:    "test-host",
				Subdomain:   "test-subdomain",
				DNSConfig: &corev1.PodDNSConfig{
					Nameservers: []string{"8.8.8.8"},
					Searches:    []string{"example.com"},
				},
				HostAliases: []corev1.HostAlias{
					{IP: "127.0.0.1", Hostnames: []string{"localhost"}},
				},
				ShareProcessNamespace: ptr.To(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractNetworkConfig(tt.podSpec)

			if result.DNSPolicy != tt.expected.DNSPolicy {
				t.Errorf("DNSPolicy = %v, want %v", result.DNSPolicy, tt.expected.DNSPolicy)
			}
			if result.HostNetwork != tt.expected.HostNetwork {
				t.Errorf("HostNetwork = %v, want %v", result.HostNetwork, tt.expected.HostNetwork)
			}
			if result.HostPID != tt.expected.HostPID {
				t.Errorf("HostPID = %v, want %v", result.HostPID, tt.expected.HostPID)
			}
			if result.HostIPC != tt.expected.HostIPC {
				t.Errorf("HostIPC = %v, want %v", result.HostIPC, tt.expected.HostIPC)
			}
			if result.Hostname != tt.expected.Hostname {
				t.Errorf("Hostname = %v, want %v", result.Hostname, tt.expected.Hostname)
			}
			if result.Subdomain != tt.expected.Subdomain {
				t.Errorf("Subdomain = %v, want %v", result.Subdomain, tt.expected.Subdomain)
			}
		})
	}
}

func TestApplyNetworkConfig(t *testing.T) {
	tests := []struct {
		name        string
		initialSpec *corev1.PodSpec
		config      *NetworkConfig
		expected    *corev1.PodSpec
	}{
		{
			name:        "apply empty config",
			initialSpec: &corev1.PodSpec{},
			config:      &NetworkConfig{},
			expected: &corev1.PodSpec{
				HostNetwork: false,
				HostPID:     false,
				HostIPC:     false,
			},
		},
		{
			name:        "apply full config",
			initialSpec: &corev1.PodSpec{},
			config: &NetworkConfig{
				DNSPolicy:   corev1.DNSClusterFirst,
				HostNetwork: true,
				HostPID:     true,
				HostIPC:     true,
				Hostname:    "new-host",
				Subdomain:   "new-subdomain",
				DNSConfig: &corev1.PodDNSConfig{
					Nameservers: []string{"1.1.1.1"},
				},
				HostAliases: []corev1.HostAlias{
					{IP: "192.168.1.1", Hostnames: []string{"myhost"}},
				},
				ShareProcessNamespace: ptr.To(true),
			},
			expected: &corev1.PodSpec{
				DNSPolicy:   corev1.DNSClusterFirst,
				HostNetwork: true,
				HostPID:     true,
				HostIPC:     true,
				Hostname:    "new-host",
				Subdomain:   "new-subdomain",
				DNSConfig: &corev1.PodDNSConfig{
					Nameservers: []string{"1.1.1.1"},
				},
				HostAliases: []corev1.HostAlias{
					{IP: "192.168.1.1", Hostnames: []string{"myhost"}},
				},
				ShareProcessNamespace: ptr.To(true),
			},
		},
		{
			name: "override existing config",
			initialSpec: &corev1.PodSpec{
				DNSPolicy:   corev1.DNSDefault,
				HostNetwork: true,
				Hostname:    "old-host",
			},
			config: &NetworkConfig{
				DNSPolicy:   corev1.DNSClusterFirst,
				HostNetwork: false,
				Hostname:    "new-host",
			},
			expected: &corev1.PodSpec{
				DNSPolicy:   corev1.DNSClusterFirst,
				HostNetwork: false,
				HostPID:     false,
				HostIPC:     false,
				Hostname:    "new-host",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyNetworkConfig(tt.initialSpec, tt.config)

			if tt.initialSpec.DNSPolicy != tt.expected.DNSPolicy {
				t.Errorf("DNSPolicy = %v, want %v", tt.initialSpec.DNSPolicy, tt.expected.DNSPolicy)
			}
			if tt.initialSpec.HostNetwork != tt.expected.HostNetwork {
				t.Errorf("HostNetwork = %v, want %v", tt.initialSpec.HostNetwork, tt.expected.HostNetwork)
			}
			if tt.initialSpec.HostPID != tt.expected.HostPID {
				t.Errorf("HostPID = %v, want %v", tt.initialSpec.HostPID, tt.expected.HostPID)
			}
			if tt.initialSpec.HostIPC != tt.expected.HostIPC {
				t.Errorf("HostIPC = %v, want %v", tt.initialSpec.HostIPC, tt.expected.HostIPC)
			}
			if tt.initialSpec.Hostname != tt.expected.Hostname {
				t.Errorf("Hostname = %v, want %v", tt.initialSpec.Hostname, tt.expected.Hostname)
			}
			if tt.initialSpec.Subdomain != tt.expected.Subdomain {
				t.Errorf("Subdomain = %v, want %v", tt.initialSpec.Subdomain, tt.expected.Subdomain)
			}
		})
	}
}
