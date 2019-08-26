package provider

import (
	k8snetworkinformer "k8s.io/client-go/informers/networking/v1"
	k8snetworklister "k8s.io/client-go/listers/networking/v1"
	api "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
)

// NsNetworkPolicyProvider is a interface to let different cnis to implement our api
type NsNetworkPolicyProvider interface {
	Add(*api.NamespaceNetworkPolicy) error
	CheckExist(*api.NamespaceNetworkPolicy) (bool, error)
	NeedUpdate(*api.NamespaceNetworkPolicy) (bool, error)
	Update(*api.NamespaceNetworkPolicy) error
	Delete(*api.NamespaceNetworkPolicy) error
	Get(*api.NamespaceNetworkPolicy) (interface{}, error)
}

// TODO: support no-calico CNI
type k8sNetworkProvider struct {
	networkPolicyInformer k8snetworkinformer.NetworkPolicyInformer
	networkPolicyLister   k8snetworklister.NetworkPolicyLister
}

func (k *k8sNetworkProvider) Add(o *api.NamespaceNetworkPolicy) error {
	return nil
}

func (k *k8sNetworkProvider) CheckExist(o *api.NamespaceNetworkPolicy) (bool, error) {
	return false, nil
}

func (k *k8sNetworkProvider) Delete(o *api.NamespaceNetworkPolicy) error {
	return nil
}
