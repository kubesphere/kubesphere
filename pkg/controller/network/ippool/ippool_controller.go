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
	"reflect"
	"time"

	cnet "github.com/projectcalico/libcalico-go/lib/net"
	podv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	networkv1alpha1 "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	networkInformer "kubesphere.io/kubesphere/pkg/client/informers/externalversions/network/v1alpha1"
	"kubesphere.io/kubesphere/pkg/controller/network/utils"
	"kubesphere.io/kubesphere/pkg/controller/network/webhooks"
	"kubesphere.io/kubesphere/pkg/simple/client/network/ippool"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	ipamblockInformer networkInformer.IPAMBlockInformer
	ipamblockSynced   cache.InformerSynced

	client           clientset.Interface
	kubesphereClient kubesphereclient.Interface
}

func (c *IPPoolController) ippoolHandle(obj interface{}) {
	pool, ok := obj.(*networkv1alpha1.IPPool)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("IPPool informer returned non-ippool object: %#v", obj))
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(pool)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for ippool %#v: %v", pool, err))
		return
	}

	if utils.NeedToAddFinalizer(pool, networkv1alpha1.IPPoolFinalizer) || utils.IsDeletionCandidate(pool, networkv1alpha1.IPPoolFinalizer) {
		c.ippoolQueue.Add(key)
	}
}

func (c *IPPoolController) addFinalizer(pool *networkv1alpha1.IPPool) error {
	clone := pool.DeepCopy()
	controllerutil.AddFinalizer(clone, networkv1alpha1.IPPoolFinalizer)
	clone.Labels = map[string]string{
		networkv1alpha1.IPPoolNameLabel: clone.Name,
		networkv1alpha1.IPPoolTypeLabel: clone.Spec.Type,
		networkv1alpha1.IPPoolIDLabel:   fmt.Sprintf("%d", clone.ID()),
	}
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
	_, cidr, err := cnet.ParseCIDR(b.Spec.CIDR)
	if err != nil {
		return fmt.Errorf("invalid cidr")
	}

	size, _ := cidr.Mask.Size()
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

	return nil
}

func (c *IPPoolController) ValidateDelete(obj runtime.Object) error {
	p := obj.(*networkv1alpha1.IPPool)

	if p.Status.Allocations > 0 {
		return fmt.Errorf("ippool is in use, please remove the workload before deleting")
	}

	return nil
}

func (c *IPPoolController) disableIPPool(old *networkv1alpha1.IPPool) error {
	if old.Spec.Disabled {
		return nil
	}

	clone := old.DeepCopy()
	clone.Spec.Disabled = true

	old, err := c.kubesphereClient.NetworkV1alpha1().IPPools().Update(context.TODO(), clone, metav1.UpdateOptions{})

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

	if !cache.WaitForCacheSync(stopCh, c.ippoolSynced, c.ipamblockSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	return nil
}

func (c *IPPoolController) runWorker() {
	for c.processIPPoolItem() {
	}
}

func (c *IPPoolController) processIPPoolItem() bool {
	key, quit := c.ippoolQueue.Get()
	if quit {
		return false
	}
	defer c.ippoolQueue.Done(key)

	_, name, err := cache.SplitMetaNamespaceKey(key.(string))
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error parsing ippool key %q: %v", key, err))
		return true
	}

	delay, err := c.processIPPool(name)
	if err == nil {
		c.ippoolQueue.Forget(key)
		return true
	} else if delay != nil {
		c.ippoolQueue.AddAfter(key, *delay)
	}

	utilruntime.HandleError(fmt.Errorf("error processing ippool %v (will retry): %v", key, err))
	c.ippoolQueue.AddRateLimited(key)
	return true
}

func (c *IPPoolController) ipamblockHandle(obj interface{}) {
	block, ok := obj.(*networkv1alpha1.IPAMBlock)
	if !ok {
		return
	}

	poolName := block.Labels[networkv1alpha1.IPPoolNameLabel]
	c.ippoolQueue.Add(poolName)
}

func NewIPPoolController(
	ippoolInformer networkInformer.IPPoolInformer,
	ipamblockInformer networkInformer.IPAMBlockInformer,
	client clientset.Interface,
	kubesphereClient kubesphereclient.Interface,
	provider ippool.Provider) *IPPoolController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "ippool-controller"})

	c := &IPPoolController{
		eventBroadcaster:  broadcaster,
		eventRecorder:     recorder,
		ippoolInformer:    ippoolInformer,
		ippoolSynced:      ippoolInformer.Informer().HasSynced,
		ippoolQueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ippool"),
		ipamblockInformer: ipamblockInformer,
		ipamblockSynced:   ipamblockInformer.Informer().HasSynced,
		client:            client,
		kubesphereClient:  kubesphereClient,
		provider:          provider,
	}

	ippoolInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.ippoolHandle,
		UpdateFunc: func(old, new interface{}) {
			c.ippoolHandle(new)
		},
	})

	//just for update ippool status
	ipamblockInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.ipamblockHandle,
		UpdateFunc: func(old, new interface{}) {
			c.ipamblockHandle(new)
		},
		DeleteFunc: c.ipamblockHandle,
	})

	//register ippool webhook
	webhooks.RegisterValidator(networkv1alpha1.SchemeGroupVersion.WithKind(networkv1alpha1.ResourceKindIPPool).String(),
		&webhooks.ValidatorWrap{Obj: &networkv1alpha1.IPPool{}, Helper: c})
	webhooks.RegisterDefaulter(podv1.SchemeGroupVersion.WithKind("Pod").String(),
		&webhooks.DefaulterWrap{Obj: &podv1.Pod{}, Helper: provider})

	return c
}
