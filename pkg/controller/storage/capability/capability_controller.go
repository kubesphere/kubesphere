/*

 Copyright 2020 The KubeSphere Authors.

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

package capability

import (
	"fmt"
	"os"
	"reflect"
	"time"

	snapapi "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	snapinformers "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions/volumesnapshot/v1beta1"
	snaplisters "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/listers/volumesnapshot/v1beta1"
	v1strorage "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/util/wait"
	scinformers "k8s.io/client-go/informers/storage/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	sclisters "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	crdapi "kubesphere.io/kubesphere/pkg/apis/storage/v1alpha1"
	clientset "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	crdscheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	storageinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/storage/v1alpha1"
	crdlisters "kubesphere.io/kubesphere/pkg/client/listers/storage/v1alpha1"
)

const (
	minKubernetesVersion = "v1.17.0"
	CSIAddressFormat     = "/var/lib/kubelet/plugins/%s/csi.sock"
)

type csiAddressGetter func(storageClassProvisioner string) string

type StorageCapabilityController struct {
	k8sClient                    kubernetes.Interface
	storageClassCapabilityClient clientset.Interface
	storageClassLister           sclisters.StorageClassLister
	storageClassSynced           cache.InformerSynced
	snapshotClassLister          snaplisters.VolumeSnapshotClassLister
	snapshotClassSynced          cache.InformerSynced
	storageClassCapabilityLister crdlisters.StorageClassCapabilityLister
	storageClassCapabilitySynced cache.InformerSynced
	workQueue                    workqueue.RateLimitingInterface
	csiAddressGetter             csiAddressGetter
}

// This controller is responsible to watch StorageClass, SnapshotClass.
// And then update StorageClassCapability CRD resource object to the newest status.
func NewController(
	k8sClient kubernetes.Interface,
	storageClassCapabilityClient clientset.Interface,
	storageClassInformer scinformers.StorageClassInformer,
	snapshotClassInformer snapinformers.VolumeSnapshotClassInformer,
	storageClassCapabilityInformer storageinformers.StorageClassCapabilityInformer,
	csiAddressGetter csiAddressGetter,
) *StorageCapabilityController {

	utilruntime.Must(crdscheme.AddToScheme(scheme.Scheme))
	controller := &StorageCapabilityController{
		k8sClient:                    k8sClient,
		storageClassCapabilityClient: storageClassCapabilityClient,
		storageClassLister:           storageClassInformer.Lister(),
		storageClassSynced:           storageClassInformer.Informer().HasSynced,
		snapshotClassLister:          snapshotClassInformer.Lister(),
		snapshotClassSynced:          snapshotClassInformer.Informer().HasSynced,
		storageClassCapabilityLister: storageClassCapabilityInformer.Lister(),
		storageClassCapabilitySynced: storageClassCapabilityInformer.Informer().HasSynced,
		workQueue:                    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "StorageClasses"),
		csiAddressGetter:             csiAddressGetter,
	}
	storageClassInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueStorageClass,
		UpdateFunc: func(old, new interface{}) {
			newStorageClass := new.(*v1strorage.StorageClass)
			oldStorageClass := old.(*v1strorage.StorageClass)
			if newStorageClass.ResourceVersion == oldStorageClass.ResourceVersion {
				return
			}
			controller.enqueueStorageClass(newStorageClass)
		},
		DeleteFunc: controller.enqueueStorageClass,
	})
	snapshotClassInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSnapshotClass,
		UpdateFunc: func(old, new interface{}) {
			return
		},
		DeleteFunc: controller.enqueueSnapshotClass,
	})
	return controller
}

func (c *StorageCapabilityController) Start(stopCh <-chan struct{}) error {
	return c.Run(5, stopCh)
}

func (c *StorageCapabilityController) Run(threadCnt int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.storageClassSynced, c.snapshotClassSynced, c.storageClassCapabilitySynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < threadCnt; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
	return nil
}

func (c *StorageCapabilityController) enqueueStorageClass(obj interface{}) {
	storageClass := obj.(*v1strorage.StorageClass)
	if !fileExist(c.csiAddressGetter(storageClass.Provisioner)) {
		klog.V(4).Infof("CSI address of storage class: %s, provisioner :%s not exist", storageClass.Name, storageClass.Provisioner)
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workQueue.Add(key)
}

func (c *StorageCapabilityController) enqueueSnapshotClass(obj interface{}) {
	snapshotClass := obj.(*snapapi.VolumeSnapshotClass)
	if !fileExist(c.csiAddressGetter(snapshotClass.Driver)) {
		klog.V(4).Infof("CSI address of snapshot class: %s, driver:%s not exist", snapshotClass.Name, snapshotClass.Driver)
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workQueue.Add(key)
}

func (c *StorageCapabilityController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *StorageCapabilityController) processNextWorkItem() bool {
	obj, shutdown := c.workQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workQueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.workQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workQueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

// When creating a new storage class, the controller will create a new storage capability object.
// When updating storage class, the controller will update or create the storage capability object.
// When deleting storage class, the controller will delete storage capability object.
func (c *StorageCapabilityController) syncHandler(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get StorageClass
	storageClass, err := c.storageClassLister.Get(name)
	klog.V(4).Infof("Get storageClass %s: entity %v", name, storageClass)
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.storageClassCapabilityLister.Get(name)
			if err != nil {
				if errors.IsNotFound(err) {
					return nil
				}
				return err
			}
			return c.storageClassCapabilityClient.StorageV1alpha1().StorageClassCapabilities().Delete(name, &metav1.DeleteOptions{})
		}
		return err
	}

	// Get SnapshotClass
	snapshotClassCreated := true
	_, err = c.snapshotClassLister.Get(storageClass.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			snapshotClassCreated = false
		} else {
			return err
		}
	}

	// Get exist StorageClassCapability
	storageClassCapabilityExist, err := c.storageClassCapabilityLister.Get(storageClass.Name)
	if errors.IsNotFound(err) {
		// If the resource doesn't exist, we'll create it
		klog.V(4).Infof("Create StorageClassProvisioner %s", storageClass.GetName())
		storageClassCapabilityCreate := &crdapi.StorageClassCapability{ObjectMeta: metav1.ObjectMeta{Name: storageClass.Name}}
		err = c.addSpec(&storageClassCapabilityCreate.Spec, storageClass, snapshotClassCreated)
		if err != nil {
			return err
		}
		klog.V(4).Info("Create StorageClassCapability: ", storageClassCapabilityCreate)
		_, err = c.storageClassCapabilityClient.StorageV1alpha1().StorageClassCapabilities().Create(storageClassCapabilityCreate)
		return err
	}
	if err != nil {
		return err
	}

	// If the resource exist, we can update it.
	storageClassCapabilityUpdate := storageClassCapabilityExist.DeepCopy()
	err = c.addSpec(&storageClassCapabilityUpdate.Spec, storageClass, snapshotClassCreated)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(storageClassCapabilityExist, storageClassCapabilityUpdate) {
		klog.V(4).Info("Update StorageClassCapability: ", storageClassCapabilityUpdate)
		_, err = c.storageClassCapabilityClient.StorageV1alpha1().StorageClassCapabilities().Update(storageClassCapabilityUpdate)
		return err
	}
	return nil
}

func (c *StorageCapabilityController) IsValidKubernetesVersion() bool {
	minVer := version.MustParseGeneric(minKubernetesVersion)
	rawVer, err := c.k8sClient.Discovery().ServerVersion()
	if err != nil {
		return false
	}
	ver, err := version.ParseSemantic(rawVer.String())
	if err != nil {
		return false
	}
	return ver.AtLeast(minVer)
}

func (c *StorageCapabilityController) addSpec(spec *crdapi.StorageClassCapabilitySpec, storageClass *v1strorage.StorageClass, snapshotClassCreated bool) error {
	csiCapability, err := csiCapability(c.csiAddressGetter(storageClass.Provisioner))
	if err != nil {
		return err
	}
	spec.Provisioner = storageClass.Provisioner
	spec.Features.Volume = csiCapability.Features.Volume
	spec.Features.Topology = csiCapability.Features.Topology
	if *storageClass.AllowVolumeExpansion {
		spec.Features.Volume.Expand = csiCapability.Features.Volume.Expand
	} else {
		spec.Features.Volume.Expand = crdapi.ExpandModeUnknown
	}
	if snapshotClassCreated {
		spec.Features.Snapshot = csiCapability.Features.Snapshot
	} else {
		spec.Features.Snapshot.Create = false
		spec.Features.Snapshot.List = false
	}
	return nil
}

func fileExist(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
