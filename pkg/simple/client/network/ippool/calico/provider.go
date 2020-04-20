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

package calico

import (
	"errors"
	"fmt"

	v3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/network/calicov3"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	calicoset "kubesphere.io/kubesphere/pkg/simple/client/network/ippool/calico/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	CalicoNamespaceAnnotationIPPoolV4 = "cni.projectcalico.org/ipv4pools"
	CalicoNamespaceAnnotationIPPoolV6 = "cni.projectcalico.org/ipv6pools"
	CalicoPodAnnotationIPAddr         = "cni.projectcalico.org/ipAddrs"
)

var (
	ErrBlockInuse = errors.New("ipamblock in using")
)

type provider struct {
	client   calicoset.Interface
	ksclient kubesphereclient.Interface
	options  Options
}

func (c provider) CreateIPPool(pool *v1alpha1.IPPool) error {
	calicoPool := &calicov3.IPPool{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name: pool.Name,
		},
		Spec: v3.IPPoolSpec{
			CIDR:         pool.Spec.CIDR,
			Disabled:     pool.Spec.Disabled,
			NodeSelector: "all()",
			VXLANMode:    v3.VXLANMode(c.options.VXLANMode),
			IPIPMode:     v3.IPIPMode(c.options.IPIPMode),
			NATOutgoing:  c.options.NATOutgoing,
		},
	}

	err := controllerutil.SetControllerReference(pool, calicoPool, scheme.Scheme)
	if err != nil {
		klog.Warningf("cannot set reference for calico ippool %s, err=%v", pool.Name, err)
	}

	_, err = c.client.CrdCalicov3().IPPools().Create(calicoPool)
	if k8serrors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

func (c provider) UpdateIPPool(pool *v1alpha1.IPPool) error {
	return nil
}

func (c provider) GetIPPoolStats(pool *v1alpha1.IPPool) (*v1alpha1.IPPool, error) {
	stats := &v1alpha1.IPPool{}

	calicoPool, err := c.client.CrdCalicov3().IPPools().Get(pool.Name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	blocks, err := c.listBlocks(calicoPool)
	if err != nil {
		return nil, err
	}

	stats.Status.Capacity = pool.NumAddresses()
	stats.Status.Reserved = 0
	stats.Status.Unallocated = 0
	stats.Status.Synced = true
	stats.Status.Allocations = 0

	if len(blocks) <= 0 {
		stats.Status.Unallocated = pool.NumAddresses()
		stats.Status.Allocations = 0
		return stats, nil
	}

	for _, block := range blocks {
		stats.Status.Allocations += block.NumAddresses() - block.NumFreeAddresses() - block.NumReservedAddresses()
		stats.Status.Reserved += block.NumReservedAddresses()
	}

	stats.Status.Unallocated = stats.Status.Capacity - stats.Status.Allocations - stats.Status.Reserved

	return stats, nil
}

func setBlockAffiDeletion(c calicoset.Interface, blockAffi *calicov3.BlockAffinity) error {
	if blockAffi.Spec.State == string(model.StatePendingDeletion) {
		return nil
	}

	blockAffi.Spec.State = string(model.StatePendingDeletion)
	_, err := c.CrdCalicov3().BlockAffinities().Update(blockAffi)
	return err
}

func deleteBlockAffi(c calicoset.Interface, blockAffi *calicov3.BlockAffinity) error {
	trueStr := fmt.Sprintf("%t", true)
	if blockAffi.Spec.Deleted != trueStr {
		blockAffi.Spec.Deleted = trueStr
		_, err := c.CrdCalicov3().BlockAffinities().Update(blockAffi)
		if err != nil {
			return err
		}
	}

	err := c.CrdCalicov3().BlockAffinities().Delete(blockAffi.Name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c provider) doBlockAffis(pool *calicov3.IPPool, do func(calicoset.Interface, *calicov3.BlockAffinity) error) error {
	_, cidrNet, _ := cnet.ParseCIDR(pool.Spec.CIDR)

	blockAffis, err := c.client.CrdCalicov3().BlockAffinities().List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, blockAffi := range blockAffis.Items {
		_, blockCIDR, _ := cnet.ParseCIDR(blockAffi.Spec.CIDR)
		if !cidrNet.IsNetOverlap(blockCIDR.IPNet) {
			continue
		}

		err = do(c.client, &blockAffi)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c provider) listBlocks(pool *calicov3.IPPool) ([]calicov3.IPAMBlock, error) {
	_, cidrNet, _ := cnet.ParseCIDR(pool.Spec.CIDR)

	blocks, err := c.client.CrdCalicov3().IPAMBlocks().List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []calicov3.IPAMBlock
	for _, block := range blocks.Items {
		_, blockCIDR, _ := cnet.ParseCIDR(block.Spec.CIDR)
		if !cidrNet.IsNetOverlap(blockCIDR.IPNet) {
			continue
		}
		result = append(result, block)
	}

	return result, nil
}

func (c provider) doBlocks(pool *calicov3.IPPool, do func(calicoset.Interface, *calicov3.IPAMBlock) error) error {
	blocks, err := c.listBlocks(pool)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		err = do(c.client, &block)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteBlock(c calicoset.Interface, block *calicov3.IPAMBlock) error {
	if block.Empty() {
		if !block.Spec.Deleted {
			block.Spec.Deleted = true
			_, err := c.CrdCalicov3().IPAMBlocks().Update(block)
			if err != nil {
				return err
			}
		}
	} else {
		return ErrBlockInuse
	}
	err := c.CrdCalicov3().IPAMBlocks().Delete(block.Name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c provider) DeleteIPPool(pool *v1alpha1.IPPool) (bool, error) {
	// Deleting a pool requires a little care because of existing endpoints
	// using IP addresses allocated in the pool.  We do the deletion in
	// the following steps:
	// -  disable the pool so no more IPs are assigned from it
	// -  remove all affinities associated with the pool
	// -  delete the pool

	// Get the pool so that we can find the CIDR associated with it.
	calicoPool, err := c.client.CrdCalicov3().IPPools().Get(pool.Name, v1.GetOptions{})
	if err != nil {
		return false, err
	}

	// If the pool is active, set the disabled flag to ensure we stop allocating from this pool.
	if !calicoPool.Spec.Disabled {
		calicoPool.Spec.Disabled = true

		calicoPool, err = c.client.CrdCalicov3().IPPools().Update(calicoPool)
		if err != nil {
			return false, err
		}
	}

	//If the address pool is being used, we return, avoiding deletions that cause other problems.
	stat, err := c.GetIPPoolStats(pool)
	if err != nil {
		return false, err
	}
	if stat.Status.Allocations > 0 {
		return false, nil
	}

	//set blockaffi to pendingdelete
	err = c.doBlockAffis(calicoPool, setBlockAffiDeletion)
	if err != nil {
		return false, err
	}

	//delete block
	err = c.doBlocks(calicoPool, deleteBlock)
	if err != nil {
		if errors.Is(err, ErrBlockInuse) {
			return false, nil
		}
		return false, err
	}

	//delete blockaffi
	err = c.doBlockAffis(calicoPool, deleteBlockAffi)
	if err != nil {
		return false, err
	}

	//delete calico ippool
	err = c.client.CrdCalicov3().IPPools().Delete(calicoPool.Name, &v1.DeleteOptions{})
	if err != nil {
		return false, err
	}

	//Congratulations, the ippool has been completely cleared.
	return true, nil
}

//Synchronizing address pools at boot time
func (c provider) syncIPPools() error {
	calicoPools, err := c.client.CrdCalicov3().IPPools().List(v1.ListOptions{})
	if err != nil {
		klog.V(4).Infof("syncIPPools: cannot list calico ippools, err=%v", err)
		return err
	}

	pools, err := c.ksclient.NetworkV1alpha1().IPPools().List(v1.ListOptions{})
	if err != nil {
		klog.V(4).Infof("syncIPPools: cannot list kubesphere ippools, err=%v", err)
		return err
	}

	existPools := map[string]bool{}
	for _, pool := range pools.Items {
		existPools[pool.Name] = true
	}

	for _, calicoPool := range calicoPools.Items {
		if _, ok := existPools[calicoPool.Name]; !ok {
			pool := &v1alpha1.IPPool{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name: calicoPool.Name,
				},
				Spec: v1alpha1.IPPoolSpec{
					Type:      v1alpha1.Calico,
					CIDR:      calicoPool.Spec.CIDR,
					Disabled:  calicoPool.Spec.Disabled,
					BlockSize: calicoPool.Spec.BlockSize,
				},
				Status: v1alpha1.IPPoolStatus{},
			}

			_, err = c.ksclient.NetworkV1alpha1().IPPools().Create(pool)
			if err != nil {
				klog.V(4).Infof("syncIPPools: cannot create kubesphere ippools, err=%v", err)
				return err
			}
		}
	}

	return nil
}

func (c provider) SyncStatus(stopCh <-chan struct{}, q workqueue.RateLimitingInterface) error {
	blockWatch, err := c.client.CrdCalicov3().IPAMBlocks().Watch(v1.ListOptions{})
	if err != nil {
		return err
	}

	ch := blockWatch.ResultChan()
	defer blockWatch.Stop()

	for {
		select {
		case <-stopCh:
			return nil
		case event, ok := <-ch:
			if !ok {
				// End of results.
				return fmt.Errorf("calico ipamblock watch closed")
			}

			if event.Type == watch.Added || event.Type == watch.Deleted || event.Type == watch.Modified {
				block := event.Object.(*calicov3.IPAMBlock)
				_, blockCIDR, _ := cnet.ParseCIDR(block.Spec.CIDR)

				if block.Labels[v1alpha1.IPPoolNameLabel] != "" {
					q.Add(block.Labels[v1alpha1.IPPoolNameLabel])
					continue
				}

				pools, err := c.ksclient.NetworkV1alpha1().IPPools().List(v1.ListOptions{})
				if err != nil {
					continue
				}

				for _, pool := range pools.Items {
					_, poolCIDR, _ := cnet.ParseCIDR(pool.Spec.CIDR)
					if poolCIDR.IsNetOverlap(blockCIDR.IPNet) {
						q.Add(pool.Name)

						block.Labels = map[string]string{
							v1alpha1.IPPoolNameLabel: pool.Name,
						}
						c.client.CrdCalicov3().IPAMBlocks().Update(block)
						break
					}
				}
			}
		}
	}
}

func NewProvider(ksclient kubesphereclient.Interface, options Options, k8sOptions *k8s.KubernetesOptions) provider {
	config, err := clientcmd.BuildConfigFromFlags("", k8sOptions.KubeConfig)
	if err != nil {
		klog.Fatalf("failed to build k8s config , err=%v", err)
	}
	config.QPS = k8sOptions.QPS
	config.Burst = k8sOptions.Burst
	client, err := calicoset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("failed to new calico client , err=%v", err)
	}

	p := provider{
		client:   client,
		ksclient: ksclient,
		options:  options,
	}

	if err := p.syncIPPools(); err != nil {
		klog.Fatalf("failed to sync calico ippool to kubesphere ippool, err=%v", err)
	}

	return p
}
