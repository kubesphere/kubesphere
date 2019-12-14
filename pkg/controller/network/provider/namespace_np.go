/*
Copyright 2019 The KubeSphere Authors.

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
