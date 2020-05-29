package provider

import netv1 "k8s.io/api/networking/v1"

// NsNetworkPolicyProvider is a interface to let different cnis to implement our api
type NsNetworkPolicyProvider interface {
	Delete(key string)
	Set(policy *netv1.NetworkPolicy) error
	Start(stopCh <-chan struct{})
	GetKey(name, nsname string) string
}
