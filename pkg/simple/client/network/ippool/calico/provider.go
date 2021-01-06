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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	v3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/network/calicov3"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	calicoset "kubesphere.io/kubesphere/pkg/simple/client/network/ippool/calico/client/clientset/versioned"
	calicoInformer "kubesphere.io/kubesphere/pkg/simple/client/network/ippool/calico/client/informers/externalversions"
	blockInformer "kubesphere.io/kubesphere/pkg/simple/client/network/ippool/calico/client/informers/externalversions/network/calicov3"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	CalicoAnnotationIPPoolV4  = "cni.projectcalico.org/ipv4pools"
	CalicoAnnotationIPPoolV6  = "cni.projectcalico.org/ipv6pools"
	CalicoPodAnnotationIPAddr = "cni.projectcalico.org/ipAddrs"
	CalicoPodAnnotationPodIP  = "cni.projectcalico.org/podIP"

	// Common attributes which may be set on allocations by clients.
	IPAMBlockAttributePod       = "pod"
	IPAMBlockAttributeNamespace = "namespace"
	IPAMBlockAttributeNode      = "node"
	IPAMBlockAttributeType      = "type"
	IPAMBlockAttributeTypeIPIP  = "ipipTunnelAddress"
	IPAMBlockAttributeTypeVXLAN = "vxlanTunnelAddress"

	CALICO_IPV4POOL_IPIP         = "CALICO_IPV4POOL_IPIP"
	CALICO_IPV4POOL_VXLAN        = "CALICO_IPV4POOL_VXLAN"
	CALICO_IPV4POOL_NAT_OUTGOING = "CALICO_IPV4POOL_NAT_OUTGOING"
	CalicoNodeDaemonset          = "calico-node"
	CalicoNodeNamespace          = "kube-system"

	DefaultBlockSize = 25
	// default re-sync period for all informer factories
	defaultResync = 600 * time.Second
)

var (
	ErrBlockInuse = errors.New("ipamblock in using")
)

type provider struct {
	client    calicoset.Interface
	ksclient  kubesphereclient.Interface
	k8sclient clientset.Interface
	pods      informercorev1.PodInformer
	block     blockInformer.IPAMBlockInformer
	queue     workqueue.RateLimitingInterface
	poolQueue workqueue.RateLimitingInterface

	options Options
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

	_, cidr, _ := net.ParseCIDR(pool.Spec.CIDR)
	size, _ := cidr.Mask.Size()
	if size > DefaultBlockSize {
		calicoPool.Spec.BlockSize = size
	}

	err := controllerutil.SetControllerReference(pool, calicoPool, scheme.Scheme)
	if err != nil {
		klog.Warningf("cannot set reference for calico ippool %s, err=%v", pool.Name, err)
	}

	_, err = c.client.CrdCalicov3().IPPools().Create(context.TODO(), calicoPool, v1.CreateOptions{})
	if k8serrors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

func (c provider) UpdateIPPool(pool *v1alpha1.IPPool) error {
	return nil
}

func (c provider) GetIPPoolStats(pool *v1alpha1.IPPool) (*v1alpha1.IPPool, error) {
	stats := pool.DeepCopy()

	calicoPool, err := c.client.CrdCalicov3().IPPools().Get(context.TODO(), pool.Name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	blocks, err := c.listBlocks(calicoPool)
	if err != nil {
		return nil, err
	}

	if stats.Status.Capacity == 0 {
		stats.Status.Capacity = pool.NumAddresses()
	}
	stats.Status.Synced = true
	stats.Status.Allocations = 0
	stats.Status.Reserved = 0
	if stats.Status.Workspaces == nil {
		stats.Status.Workspaces = make(map[string]v1alpha1.WorkspaceStatus)
	}

	if len(blocks) <= 0 {
		stats.Status.Unallocated = pool.NumAddresses()
		stats.Status.Allocations = 0
	} else {
		for _, block := range blocks {
			stats.Status.Allocations += block.NumAddresses() - block.NumFreeAddresses() - block.NumReservedAddresses()
			stats.Status.Reserved += block.NumReservedAddresses()
		}

		stats.Status.Unallocated = stats.Status.Capacity - stats.Status.Allocations - stats.Status.Reserved
	}

	wks, err := c.getAssociatedWorkspaces(pool)
	if err != nil {
		return nil, err
	}

	for _, wk := range wks {
		status, err := c.getWorkspaceStatus(wk, pool.GetName())
		if err != nil {
			return nil, err
		}
		stats.Status.Workspaces[wk] = *status
	}

	for name, wk := range stats.Status.Workspaces {
		if wk.Allocations == 0 {
			delete(stats.Status.Workspaces, name)
		}
	}

	return stats, nil
}

func setBlockAffiDeletion(c calicoset.Interface, blockAffi *calicov3.BlockAffinity) error {
	if blockAffi.Spec.State == string(model.StatePendingDeletion) {
		return nil
	}

	blockAffi.Spec.State = string(model.StatePendingDeletion)
	_, err := c.CrdCalicov3().BlockAffinities().Update(context.TODO(), blockAffi, v1.UpdateOptions{})
	return err
}

func deleteBlockAffi(c calicoset.Interface, blockAffi *calicov3.BlockAffinity) error {
	trueStr := fmt.Sprintf("%t", true)
	if blockAffi.Spec.Deleted != trueStr {
		blockAffi.Spec.Deleted = trueStr
		_, err := c.CrdCalicov3().BlockAffinities().Update(context.TODO(), blockAffi, v1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	err := c.CrdCalicov3().BlockAffinities().Delete(context.TODO(), blockAffi.Name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (c provider) doBlockAffis(pool *calicov3.IPPool, do func(calicoset.Interface, *calicov3.BlockAffinity) error) error {
	_, cidrNet, _ := cnet.ParseCIDR(pool.Spec.CIDR)

	blockAffis, err := c.client.CrdCalicov3().BlockAffinities().List(context.TODO(), v1.ListOptions{})
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

	blocks, err := c.client.CrdCalicov3().IPAMBlocks().List(context.TODO(), v1.ListOptions{})
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
			_, err := c.CrdCalicov3().IPAMBlocks().Update(context.TODO(), block, v1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	} else {
		return ErrBlockInuse
	}
	err := c.CrdCalicov3().IPAMBlocks().Delete(context.TODO(), block.Name, v1.DeleteOptions{})
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
	calicoPool, err := c.client.CrdCalicov3().IPPools().Get(context.TODO(), pool.Name, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	// If the pool is active, set the disabled flag to ensure we stop allocating from this pool.
	if !calicoPool.Spec.Disabled {
		calicoPool.Spec.Disabled = true

		calicoPool, err = c.client.CrdCalicov3().IPPools().Update(context.TODO(), calicoPool, v1.UpdateOptions{})
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
	err = c.client.CrdCalicov3().IPPools().Delete(context.TODO(), calicoPool.Name, v1.DeleteOptions{})
	if err != nil {
		return false, err
	}

	//Congratulations, the ippool has been completely cleared.
	return true, nil
}

//Synchronizing address pools at boot time
func (c provider) syncIPPools() error {
	calicoPools, err := c.client.CrdCalicov3().IPPools().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		klog.V(4).Infof("syncIPPools: cannot list calico ippools, err=%v", err)
		return err
	}

	pools, err := c.ksclient.NetworkV1alpha1().IPPools().List(context.TODO(), v1.ListOptions{})
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
					Labels: map[string]string{
						v1alpha1.IPPoolDefaultLabel: "",
					},
				},
				Spec: v1alpha1.IPPoolSpec{
					Type:      v1alpha1.Calico,
					CIDR:      calicoPool.Spec.CIDR,
					Disabled:  calicoPool.Spec.Disabled,
					BlockSize: calicoPool.Spec.BlockSize,
				},
				Status: v1alpha1.IPPoolStatus{},
			}

			_, err = c.ksclient.NetworkV1alpha1().IPPools().Create(context.TODO(), pool, v1.CreateOptions{})
			if err != nil {
				klog.V(4).Infof("syncIPPools: cannot create kubesphere ippools, err=%v", err)
				return err
			}
		}
	}

	return nil
}

func (p provider) getAssociatedWorkspaces(pool *v1alpha1.IPPool) ([]string, error) {
	var result []string

	poolLabel := constants.WorkspaceLabelKey
	if pool.GetLabels() == nil || pool.GetLabels()[poolLabel] == "" {
		wks, err := p.ksclient.TenantV1alpha1().Workspaces().List(context.TODO(), v1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, wk := range wks.Items {
			result = append(result, wk.GetName())
		}

		return result, nil
	}

	return append(result, pool.GetLabels()[poolLabel]), nil
}

func (p provider) getWorkspaceStatus(name string, poolName string) (*v1alpha1.WorkspaceStatus, error) {
	var result v1alpha1.WorkspaceStatus

	namespaces, err := p.k8sclient.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			map[string]string{
				constants.WorkspaceLabelKey: name,
			},
		).String(),
	})
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces.Items {
		pods, err := p.k8sclient.CoreV1().Pods(ns.GetName()).List(context.TODO(), v1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			if pod.GetLabels() != nil && pod.GetLabels()[v1alpha1.IPPoolNameLabel] == poolName {
				result.Allocations++
			}
		}
	}

	return &result, nil
}

func (p provider) Type() string {
	return v1alpha1.IPPoolTypeCalico
}

func (p provider) SyncStatus(stopCh <-chan struct{}, q workqueue.RateLimitingInterface) error {
	defer utilruntime.HandleCrash()
	defer p.queue.ShutDown()

	klog.Info("starting calico block controller")
	defer klog.Info("shutting down calico block controller")

	p.poolQueue = q
	go p.block.Informer().Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, p.pods.Informer().HasSynced, p.block.Informer().HasSynced) {
		klog.Fatal("failed to wait for caches to sync")
	}

	for i := 0; i < 5; i++ {
		go wait.Until(p.runWorker, time.Second, stopCh)
	}

	<-stopCh
	return nil
}

func (p provider) processBlock(name string) error {
	block, err := p.block.Lister().Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	_, blockCIDR, _ := cnet.ParseCIDR(block.Spec.CIDR)

	poolName := block.Labels[v1alpha1.IPPoolNameLabel]
	if poolName == "" {
		pools, err := p.ksclient.NetworkV1alpha1().IPPools().List(context.TODO(), v1.ListOptions{})
		if err != nil {
			return err
		}

		for _, pool := range pools.Items {
			_, poolCIDR, _ := cnet.ParseCIDR(pool.Spec.CIDR)
			if poolCIDR.IsNetOverlap(blockCIDR.IPNet) {
				poolName = pool.Name

				block.Labels = map[string]string{
					v1alpha1.IPPoolNameLabel: pool.Name,
				}
				p.client.CrdCalicov3().IPAMBlocks().Update(context.TODO(), block, v1.UpdateOptions{})
				break
			}
		}
	}

	for _, podAttr := range block.Spec.Attributes {
		name := podAttr.AttrSecondary[IPAMBlockAttributePod]
		namespace := podAttr.AttrSecondary[IPAMBlockAttributeNamespace]

		if name == "" || namespace == "" {
			continue
		}

		pod, err := p.pods.Lister().Pods(namespace).Get(name)
		if err != nil {
			continue
		}

		labels := pod.GetLabels()
		if labels != nil {
			poolLabel := labels[v1alpha1.IPPoolNameLabel]
			if poolLabel != "" {
				continue
			}
		}

		retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			pod, err = p.k8sclient.CoreV1().Pods(namespace).Get(context.TODO(), name, v1.GetOptions{})
			if err != nil {
				return err
			}

			labels := pod.GetLabels()
			if labels != nil {
				poolLabel := labels[v1alpha1.IPPoolNameLabel]
				if poolLabel != "" {
					return nil
				}
			} else {
				pod.Labels = make(map[string]string)
			}

			if pod.GetAnnotations() == nil {
				pod.Annotations = make(map[string]string)
			}

			annostrs, _ := json.Marshal([]string{poolName})
			pod.GetAnnotations()[CalicoAnnotationIPPoolV4] = string(annostrs)
			pod.Labels[v1alpha1.IPPoolNameLabel] = poolName

			_, err = p.k8sclient.CoreV1().Pods(namespace).Update(context.TODO(), pod, v1.UpdateOptions{})

			return err
		})
	}

	p.poolQueue.Add(poolName)
	return nil
}

func (p provider) processBlockItem() bool {
	key, quit := p.queue.Get()
	if quit {
		return false
	}
	defer p.queue.Done(key)

	err := p.processBlock(key.(string))
	if err == nil {
		p.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("error processing calico block %v (will retry): %v", key, err))
	p.queue.AddRateLimited(key)
	return true
}

func (p provider) runWorker() {
	for p.processBlockItem() {
	}
}

func (p provider) addBlock(obj interface{}) {
	block, ok := obj.(*calicov3.IPAMBlock)
	if !ok {
		return
	}

	p.queue.Add(block.Name)
}

func (p provider) Default(obj runtime.Object) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil
	}

	annos := pod.GetAnnotations()
	if annos == nil {
		pod.Annotations = make(map[string]string)
	}

	if annos[CalicoAnnotationIPPoolV4] == "" {
		pools, err := p.ksclient.NetworkV1alpha1().IPPools().List(context.TODO(), v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				v1alpha1.IPPoolDefaultLabel: "",
			}).String(),
		})
		if err != nil {
			return err
		}
		var poolNames []string
		for _, pool := range pools.Items {
			poolNames = append(poolNames, pool.Name)
		}
		if len(poolNames) > 0 {
			annostrs, _ := json.Marshal(poolNames)
			pod.Annotations[CalicoAnnotationIPPoolV4] = string(annostrs)
		}
	}

	return nil
}

func NewProvider(podInformer informercorev1.PodInformer, ksclient kubesphereclient.Interface, k8sClient clientset.Interface, k8sOptions *k8s.KubernetesOptions) provider {
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

	ds, err := k8sClient.AppsV1().DaemonSets(CalicoNodeNamespace).Get(context.TODO(), CalicoNodeDaemonset, v1.GetOptions{})
	if err != nil {
		klog.Fatalf("failed to get calico-node deployment in kube-system, err=%v", err)
	}
	opts := Options{
		IPIPMode:    "Always",
		VXLANMode:   "Never",
		NATOutgoing: true,
	}
	envs := ds.Spec.Template.Spec.Containers[0].Env
	for _, env := range envs {
		if env.Name == CALICO_IPV4POOL_IPIP {
			opts.IPIPMode = env.Value
		}

		if env.Name == CALICO_IPV4POOL_VXLAN {
			opts.VXLANMode = env.Value
		}

		if env.Name == CALICO_IPV4POOL_NAT_OUTGOING {
			if env.Value != "true" {
				opts.NATOutgoing = false
			}
		}
	}

	p := provider{
		client:    client,
		ksclient:  ksclient,
		k8sclient: k8sClient,
		pods:      podInformer,
		queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "calicoBlock"),
		options:   opts,
	}

	blockI := calicoInformer.NewSharedInformerFactory(client, defaultResync).Crd().Calicov3().IPAMBlocks()
	blockI.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: p.addBlock,
		UpdateFunc: func(old, new interface{}) {
			p.addBlock(new)
		},
	})
	p.block = blockI

	if err := p.syncIPPools(); err != nil {
		klog.Fatalf("failed to sync calico ippool to kubesphere ippool, err=%v", err)
	}

	return p
}
