/*
Copyright 2020 The KubeSphere authors.

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

package ippool

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sinformers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	networkv1alpha1 "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	calicoclient "kubesphere.io/kubesphere/pkg/simple/client/network/ippool/calico"
	"kubesphere.io/kubesphere/pkg/simple/client/network/ippool/ipam"
)

type Provider interface {
	// canDelete indicates whether the address pool is being used or not.
	DeleteIPPool(pool *networkv1alpha1.IPPool) (canDelete bool, err error)
	CreateIPPool(pool *networkv1alpha1.IPPool) error
	UpdateIPPool(pool *networkv1alpha1.IPPool) error
	GetIPPoolStats(pool *networkv1alpha1.IPPool) (*networkv1alpha1.IPPool, error)
	SyncStatus(stopCh <-chan struct{}, q workqueue.RateLimitingInterface) error
	UpdateNamespace(ns *corev1.Namespace, pools []string) error
	Type() string
	Default(obj runtime.Object) error
}

type provider struct {
	kubesphereClient kubesphereclient.Interface
	ipamclient       ipam.IPAMClient
}

func (p provider) Type() string {
	return networkv1alpha1.IPPoolTypeLocal
}

func (p provider) Default(obj runtime.Object) error {
	return nil
}

func (p provider) UpdateNamespace(ns *corev1.Namespace, pools []string) error {
	return nil
}

func (p provider) DeleteIPPool(pool *networkv1alpha1.IPPool) (bool, error) {
	blocks, err := p.ipamclient.ListBlocks(pool.Name)
	if err != nil {
		return false, err
	}

	for _, block := range blocks {
		if block.Empty() {
			if err = p.ipamclient.DeleteBlock(&block); err != nil {
				return false, err
			}
		} else {
			return false, nil
		}
	}

	return true, nil
}

func (p provider) CreateIPPool(pool *networkv1alpha1.IPPool) error {
	return nil
}

func (p provider) UpdateIPPool(pool *networkv1alpha1.IPPool) error {
	return nil
}

func (p provider) SyncStatus(stopCh <-chan struct{}, q workqueue.RateLimitingInterface) error {
	return nil
}

func (p provider) GetIPPoolStats(pool *networkv1alpha1.IPPool) (*networkv1alpha1.IPPool, error) {
	stats, err := p.ipamclient.GetUtilization(ipam.GetUtilizationArgs{
		Pools: []string{pool.Name},
	})
	if err != nil {
		return nil, err
	}

	stat := stats[0]
	clone := pool.DeepCopy()
	clone.Status = networkv1alpha1.IPPoolStatus{
		Allocations: stat.Allocate,
		Unallocated: stat.Unallocated,
		Reserved:    stat.Reserved,
		Capacity:    stat.Capacity,
		Synced:      true,
	}
	return clone, nil
}

func newProvider(clientset kubesphereclient.Interface) provider {
	return provider{
		kubesphereClient: clientset,
		ipamclient:       ipam.NewIPAMClient(clientset, networkv1alpha1.VLAN),
	}
}

func NewProvider(k8sInformer k8sinformers.SharedInformerFactory, clientset kubesphereclient.Interface, client clientset.Interface, pt string, k8sOptions *k8s.KubernetesOptions) Provider {
	var p Provider

	switch pt {
	case networkv1alpha1.IPPoolTypeLocal:
		p = provider{
			kubesphereClient: clientset,
			ipamclient:       ipam.NewIPAMClient(clientset, networkv1alpha1.VLAN),
		}
	case networkv1alpha1.IPPoolTypeCalico:
		p = calicoclient.NewProvider(k8sInformer, clientset, client, k8sOptions)
	}

	return p
}
