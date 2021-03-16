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

package ippool

import (
	"context"
	"fmt"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sinformers "k8s.io/client-go/informers"
	coreinfomers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	networkv1alpha1 "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	networkInformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/network/v1alpha1"
	tenantv1alpha1informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/network/utils"
	"kubesphere.io/kubesphere/pkg/controller/network/webhooks"
	"kubesphere.io/kubesphere/pkg/simple/client/network/ippool"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var (
	ErrCIDROverlap = fmt.Errorf("CIDR is overlap")
)

type IPPoolController struct {
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	provider ippool.Provider

	ippoolInformer networkInformer.IPPoolInformer
	ippoolSynced   cache.InformerSynced
	ippoolQueue    workqueue.RateLimitingInterface

	wsInformer tenantv1alpha1informers.WorkspaceInformer
	wsSynced   cache.InformerSynced

	nsInformer coreinfomers.NamespaceInformer
	nsSynced   cache.InformerSynced
	nsQueue    workqueue.RateLimitingInterface

	ipamblockInformer networkInformer.IPAMBlockInformer
	ipamblockSynced   cache.InformerSynced

	client           clientset.Interface
	kubesphereClient kubesphereclient.Interface
}

func (c *IPPoolController) enqueueIPPools(obj interface{}) {
	pool, ok := obj.(*networkv1alpha1.IPPool)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("IPPool informer returned non-ippool object: %#v", obj))
		return
	}

	c.ippoolQueue.Add(pool.Name)
}

func (c *IPPoolController) addFinalizer(pool *networkv1alpha1.IPPool) error {
	clone := pool.DeepCopy()
	controllerutil.AddFinalizer(clone, networkv1alpha1.IPPoolFinalizer)
	if clone.Labels == nil {
		clone.Labels = make(map[string]string)
	}
	clone.Labels[networkv1alpha1.IPPoolNameLabel] = clone.Name
	clone.Labels[networkv1alpha1.IPPoolTypeLabel] = clone.Spec.Type
	clone.Labels[networkv1alpha1.IPPoolIDLabel] = fmt.Sprintf("%d", clone.ID())
	pool, err := c.kubesphereClient.NetworkV1alpha1().IPPools().Update(context.TODO(), clone, metav1.UpdateOptions{})
	if err != nil {
		klog.V(3).Infof("Error adding  finalizer to pool %s: %v", pool.Name, err)
		return err
	}
	klog.V(3).Infof("Added finalizer to pool %s", pool.Name)
	return nil
}

func (c *IPPoolController) removeFinalizer(pool *networkv1alpha1.IPPool) error {
	clone := pool.DeepCopy()
	controllerutil.RemoveFinalizer(clone, networkv1alpha1.IPPoolFinalizer)
	pool, err := c.kubesphereClient.NetworkV1alpha1().IPPools().Update(context.TODO(), clone, metav1.UpdateOptions{})
	if err != nil {
		klog.V(3).Infof("Error removing  finalizer from pool %s: %v", pool.Name, err)
		return err
	}
	klog.V(3).Infof("Removed protection finalizer from pool %s", pool.Name)
	return nil
}

func (c *IPPoolController) ValidateCreate(obj runtime.Object) error {
	b := obj.(*networkv1alpha1.IPPool)
	ip, cidr, err := cnet.ParseCIDR(b.Spec.CIDR)
	if err != nil {
		return fmt.Errorf("invalid cidr")
	}

	size, _ := cidr.Mask.Size()
	if ip.IP.To4() != nil && size == 32 {
		return fmt.Errorf("the cidr mask must be less than 32")
	}
	if b.Spec.BlockSize > 0 && b.Spec.BlockSize < size {
		return fmt.Errorf("the blocksize should be larger than the cidr mask")
	}

	if b.Spec.RangeStart != "" || b.Spec.RangeEnd != "" {
		iStart := cnet.ParseIP(b.Spec.RangeStart)
		iEnd := cnet.ParseIP(b.Spec.RangeEnd)
		if iStart == nil || iEnd == nil {
			return fmt.Errorf("invalid rangeStart or rangeEnd")
		}
		offsetStart, err := b.IPToOrdinal(*iStart)
		if err != nil {
			return err
		}
		offsetEnd, err := b.IPToOrdinal(*iEnd)
		if err != nil {
			return err
		}
		if offsetEnd < offsetStart {
			return fmt.Errorf("rangeStart should not big than rangeEnd")
		}
	}

	pools, err := c.kubesphereClient.NetworkV1alpha1().IPPools().List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			networkv1alpha1.IPPoolIDLabel: fmt.Sprintf("%d", b.ID()),
		}).String(),
	})
	if err != nil {
		return err
	}

	for _, p := range pools.Items {
		if b.Overlapped(p) {
			return fmt.Errorf("ippool cidr is overlapped with %s", p.Name)
		}
	}

	return nil
}

func (c *IPPoolController) validateDefaultIPPool(p *networkv1alpha1.IPPool) error {
	pools, err := c.kubesphereClient.NetworkV1alpha1().IPPools().List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			labels.Set{
				networkv1alpha1.IPPoolDefaultLabel: "",
			}).String(),
	})
	if err != nil {
		return err
	}

	poolLen := len(pools.Items)
	if poolLen != 1 || pools.Items[0].Name != p.Name {
		return nil
	}

	return fmt.Errorf("Must ensure that there is at least one default ippool")
}

func (c *IPPoolController) ValidateUpdate(old runtime.Object, new runtime.Object) error {
	oldP := old.(*networkv1alpha1.IPPool)
	newP := new.(*networkv1alpha1.IPPool)

	if newP.Spec.CIDR != oldP.Spec.CIDR {
		return fmt.Errorf("cidr cannot be modified")
	}

	if newP.Spec.Type != oldP.Spec.Type {
		return fmt.Errorf("ippool type cannot be modified")
	}

	if newP.Spec.BlockSize != oldP.Spec.BlockSize {
		return fmt.Errorf("ippool blockSize cannot be modified")
	}

	if newP.Spec.RangeEnd != oldP.Spec.RangeEnd || newP.Spec.RangeStart != oldP.Spec.RangeStart {
		return fmt.Errorf("ippool rangeEnd/rangeStart cannot be modified")
	}

	_, defaultOld := oldP.Labels[networkv1alpha1.IPPoolDefaultLabel]
	_, defaultNew := newP.Labels[networkv1alpha1.IPPoolDefaultLabel]
	if !defaultNew && defaultOld != defaultNew {
		err := c.validateDefaultIPPool(newP)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *IPPoolController) ValidateDelete(obj runtime.Object) error {
	p := obj.(*networkv1alpha1.IPPool)

	if p.Status.Allocations > 0 {
		return fmt.Errorf("ippool is in use, please remove the workload before deleting")
	}

	return c.validateDefaultIPPool(p)
}

func (c *IPPoolController) disableIPPool(old *networkv1alpha1.IPPool) error {
	if old.Spec.Disabled {
		return nil
	}

	clone := old.DeepCopy()
	clone.Spec.Disabled = true

	_, err := c.kubesphereClient.NetworkV1alpha1().IPPools().Update(context.TODO(), clone, metav1.UpdateOptions{})

	return err
}

func (c *IPPoolController) updateIPPoolStatus(old *networkv1alpha1.IPPool) error {
	new, err := c.provider.GetIPPoolStats(old)
	if err != nil {
		return fmt.Errorf("failed to get ippool %s status %v", old.Name, err)
	}

	if reflect.DeepEqual(old.Status, new.Status) {
		return nil
	}

	_, err = c.kubesphereClient.NetworkV1alpha1().IPPools().UpdateStatus(context.TODO(), new, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ippool %s status  %v", old.Name, err)
	}

	return nil
}

func (c *IPPoolController) processIPPool(name string) (*time.Duration, error) {
	klog.V(4).Infof("Processing IPPool %s", name)
	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished processing IPPool %s (%v)", name, time.Since(startTime))
	}()

	pool, err := c.ippoolInformer.Lister().Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ippool %s: %v", name, err)
	}

	if pool.Type() != c.provider.Type() {
		klog.V(4).Infof("pool %s type not match, ignored", pool.Name)
		return nil, nil
	}

	if utils.IsDeletionCandidate(pool, networkv1alpha1.IPPoolFinalizer) {
		err = c.disableIPPool(pool)
		if err != nil {
			return nil, err
		}

		// Pool should be deleted. Check if it's used and remove finalizer if
		// it's not.
		canDelete, err := c.provider.DeleteIPPool(pool)
		if err != nil {
			return nil, err
		}

		if canDelete {
			return nil, c.removeFinalizer(pool)
		}

		//The  ippool is being used, update status and try again later.
		delay := time.Second * 3
		return &delay, c.updateIPPoolStatus(pool)
	}

	if utils.NeedToAddFinalizer(pool, networkv1alpha1.IPPoolFinalizer) {
		err = c.addFinalizer(pool)
		if err != nil {
			return nil, err
		}

		err = c.provider.CreateIPPool(pool)
		if err != nil {
			klog.V(4).Infof("Provider failed to create IPPool %s, err=%v", pool.Name, err)
			return nil, err
		}

		return nil, c.updateIPPoolStatus(pool)
	}

	err = c.provider.UpdateIPPool(pool)
	if err != nil {
		klog.V(4).Infof("Provider failed to update IPPool %s, err=%v", pool.Name, err)
		return nil, err
	}

	return nil, c.updateIPPoolStatus(pool)
}

func (c *IPPoolController) Start(stopCh <-chan struct{}) error {
	go c.provider.SyncStatus(stopCh, c.ippoolQueue)
	return c.Run(5, stopCh)
}

func (c *IPPoolController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.ippoolQueue.ShutDown()

	klog.Info("starting ippool controller")
	defer klog.Info("shutting down ippool controller")

	if !cache.WaitForCacheSync(stopCh, c.ippoolSynced, c.ipamblockSynced, c.wsSynced, c.nsSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runIPPoolWorker, time.Second, stopCh)
		go wait.Until(c.runNSWorker, time.Second, stopCh)
	}

	<-stopCh
	return nil
}

func (c *IPPoolController) runIPPoolWorker() {
	for c.processIPPoolItem() {
	}
}

func (c *IPPoolController) processIPPoolItem() bool {
	key, quit := c.ippoolQueue.Get()
	if quit {
		return false
	}
	defer c.ippoolQueue.Done(key)

	delay, err := c.processIPPool(key.(string))
	if err == nil {
		c.ippoolQueue.Forget(key)
		return true
	}

	if delay != nil {
		c.ippoolQueue.AddAfter(key, *delay)
	} else {
		c.ippoolQueue.AddRateLimited(key)
	}
	utilruntime.HandleError(fmt.Errorf("error processing ippool %v (will retry): %v", key, err))
	return true
}

func (c *IPPoolController) runNSWorker() {
	for c.processNSItem() {
	}
}

func (c *IPPoolController) processNS(name string) error {
	ns, err := c.nsInformer.Lister().Get(name)
	if apierrors.IsNotFound(err) {
		return nil
	}

	var poolsName []string
	if ns.Labels != nil && ns.Labels[constants.WorkspaceLabelKey] != "" {
		pools, err := c.ippoolInformer.Lister().List(labels.SelectorFromSet(labels.Set{
			networkv1alpha1.IPPoolDefaultLabel: "",
		}))
		if err != nil {
			return err
		}

		for _, pool := range pools {
			if pool.Status.Synced {
				poolsName = append(poolsName, pool.Name)
			}
		}
	}

	clone := ns.DeepCopy()
	err = c.provider.UpdateNamespace(clone, poolsName)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(clone, ns) {
		return nil
	}

	_, err = c.client.CoreV1().Namespaces().Update(context.TODO(), clone, metav1.UpdateOptions{})
	return err
}

func (c *IPPoolController) processNSItem() bool {
	key, quit := c.nsQueue.Get()
	if quit {
		return false
	}
	defer c.nsQueue.Done(key)

	err := c.processNS(key.(string))
	if err == nil {
		c.nsQueue.Forget(key)
		return true
	}

	c.nsQueue.AddRateLimited(key)
	utilruntime.HandleError(fmt.Errorf("error processing ns %v (will retry): %v", key, err))
	return true
}

func (c *IPPoolController) enqueueIPAMBlocks(obj interface{}) {
	block, ok := obj.(*networkv1alpha1.IPAMBlock)
	if !ok {
		return
	}

	poolName := block.Labels[networkv1alpha1.IPPoolNameLabel]
	c.ippoolQueue.Add(poolName)
}

func (c *IPPoolController) enqueueWorkspace(obj interface{}) {
	wk, ok := obj.(*tenantv1alpha1.Workspace)
	if !ok {
		return
	}

	pools, err := c.ippoolInformer.Lister().List(labels.SelectorFromSet(labels.Set{
		constants.WorkspaceLabelKey: wk.Name,
	}))
	if err != nil {
		klog.Errorf("failed to list ippools by worksapce %s, err=%v", wk.Name, err)
	}

	for _, pool := range pools {
		c.ippoolQueue.Add(pool.Name)
	}
}

func (c *IPPoolController) enqueueNamespace(old interface{}, new interface{}) {
	workspaceOld := ""
	if old != nil {
		nsOld := old.(*corev1.Namespace)
		if nsOld.Labels != nil {
			workspaceOld = nsOld.Labels[constants.WorkspaceLabelKey]
		}
	}

	nsNew := new.(*corev1.Namespace)
	workspaceNew := ""
	if nsNew.Labels != nil {
		workspaceNew = nsNew.Labels[constants.WorkspaceLabelKey]
	}

	if workspaceOld != workspaceNew {
		c.nsQueue.Add(nsNew.Name)
	}
}

func NewIPPoolController(
	kubesphereInformers ksinformers.SharedInformerFactory,
	kubernetesInformers k8sinformers.SharedInformerFactory,
	client clientset.Interface,
	kubesphereClient kubesphereclient.Interface,
	provider ippool.Provider) *IPPoolController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&clientcorev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "ippool-controller"})

	c := &IPPoolController{
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		ippoolQueue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ippool"),
		nsQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ippool-ns"),
		client:           client,
		kubesphereClient: kubesphereClient,
		provider:         provider,
	}
	c.ippoolInformer = kubesphereInformers.Network().V1alpha1().IPPools()
	c.ippoolSynced = c.ippoolInformer.Informer().HasSynced
	c.ipamblockInformer = kubesphereInformers.Network().V1alpha1().IPAMBlocks()
	c.ipamblockSynced = c.ipamblockInformer.Informer().HasSynced
	c.wsInformer = kubesphereInformers.Tenant().V1alpha1().Workspaces()
	c.wsSynced = c.wsInformer.Informer().HasSynced
	c.nsInformer = kubernetesInformers.Core().V1().Namespaces()
	c.nsSynced = c.nsInformer.Informer().HasSynced

	c.ippoolInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueIPPools,
		UpdateFunc: func(old, new interface{}) {
			_, defaultOld := old.(*networkv1alpha1.IPPool).Labels[networkv1alpha1.IPPoolDefaultLabel]
			_, defaultNew := new.(*networkv1alpha1.IPPool).Labels[networkv1alpha1.IPPoolDefaultLabel]
			if defaultOld != defaultNew {
				nss, err := c.nsInformer.Lister().List(labels.Everything())
				if err != nil {
					return
				}

				for _, ns := range nss {
					c.enqueueNamespace(nil, ns)
				}
			}
			c.enqueueIPPools(new)
		},
		DeleteFunc: func(new interface{}) {
			_, defaultNew := new.(*networkv1alpha1.IPPool).Labels[networkv1alpha1.IPPoolDefaultLabel]
			if defaultNew {
				nss, err := c.nsInformer.Lister().List(labels.Everything())
				if err != nil {
					return
				}

				for _, ns := range nss {
					c.enqueueNamespace(nil, ns)
				}
			}
		},
	})

	//just for update ippool status
	c.ipamblockInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueIPAMBlocks,
		UpdateFunc: func(old, new interface{}) {
			c.enqueueIPAMBlocks(new)
		},
		DeleteFunc: c.enqueueIPAMBlocks,
	})

	c.wsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.enqueueWorkspace,
	})

	c.nsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			c.enqueueNamespace(nil, new)
		},
		UpdateFunc: c.enqueueNamespace,
	})

	//register ippool webhook
	webhooks.RegisterValidator(networkv1alpha1.SchemeGroupVersion.WithKind(networkv1alpha1.ResourceKindIPPool).String(),
		&webhooks.ValidatorWrap{Obj: &networkv1alpha1.IPPool{}, Helper: c})
	webhooks.RegisterDefaulter(corev1.SchemeGroupVersion.WithKind("Pod").String(),
		&webhooks.DefaulterWrap{Obj: &corev1.Pod{}, Helper: provider})

	return c
}
