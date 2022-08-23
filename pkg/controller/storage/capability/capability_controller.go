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
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	storageinformersv1 "k8s.io/client-go/informers/storage/v1"
	storageclient "k8s.io/client-go/kubernetes/typed/storage/v1"
	storagelistersv1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (
	annotationAllowSnapshot = "storageclass.kubesphere.io/allow-snapshot"
	annotationAllowClone    = "storageclass.kubesphere.io/allow-clone"
)

type StorageCapabilityController struct {
	storageClassClient storageclient.StorageClassInterface
	storageClassLister storagelistersv1.StorageClassLister
	storageClassSynced cache.InformerSynced

	csiDriverLister storagelistersv1.CSIDriverLister
	csiDriverSynced cache.InformerSynced

	storageClassWorkQueue workqueue.RateLimitingInterface
}

// This controller is responsible to watch StorageClass and CSIDriver.
// And then update StorageClass CRD resource object to the newest status.
func NewController(
	storageClassClient storageclient.StorageClassInterface,
	storageClassInformer storageinformersv1.StorageClassInformer,
	csiDriverInformer storageinformersv1.CSIDriverInformer,
) *StorageCapabilityController {
	controller := &StorageCapabilityController{
		storageClassClient:    storageClassClient,
		storageClassLister:    storageClassInformer.Lister(),
		storageClassSynced:    storageClassInformer.Informer().HasSynced,
		csiDriverLister:       csiDriverInformer.Lister(),
		csiDriverSynced:       csiDriverInformer.Informer().HasSynced,
		storageClassWorkQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "StorageClasses"),
	}

	storageClassInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueStorageClass,
		UpdateFunc: func(old, new interface{}) {
			newStorageClass := new.(*storagev1.StorageClass)
			oldStorageClass := old.(*storagev1.StorageClass)
			if newStorageClass.ResourceVersion == oldStorageClass.ResourceVersion {
				return
			}
			controller.enqueueStorageClass(newStorageClass)
		},
	})

	csiDriverInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.enqueueStorageClassByCSI,
		DeleteFunc: controller.enqueueStorageClassByCSI,
	})

	return controller
}

func (c *StorageCapabilityController) Start(ctx context.Context) error {
	return c.Run(5, ctx.Done())
}

func (c *StorageCapabilityController) Run(threadCnt int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.storageClassWorkQueue.ShutDown()

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	cacheSyncs := []cache.InformerSynced{
		c.storageClassSynced,
		c.csiDriverSynced,
	}

	if ok := cache.WaitForCacheSync(stopCh, cacheSyncs...); !ok {
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
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.storageClassWorkQueue.Add(key)
}

func (c *StorageCapabilityController) enqueueStorageClassByCSI(csi interface{}) {
	var objs []*storagev1.StorageClass
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(csi); err != nil {
		utilruntime.HandleError(err)
		return
	}
	objs, err = c.storageClassLister.List(labels.NewSelector())
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	for _, obj := range objs {
		if obj.Provisioner == key {
			c.enqueueStorageClass(obj)
		}
	}
}

func (c *StorageCapabilityController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *StorageCapabilityController) processNextWorkItem() bool {
	obj, shutdown := c.storageClassWorkQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.storageClassWorkQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.storageClassWorkQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workQueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.storageClassWorkQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.storageClassWorkQueue.Forget(obj)
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
	if err != nil {
		return err
	}

	// Cloning and volumeSnapshot support only available for CSI drivers.
	isCSIStorage := c.hasCSIDriver(storageClass)
	// Annotate storageClass
	storageClassUpdated := storageClass.DeepCopy()
	if isCSIStorage {
		c.updateSnapshotAnnotation(storageClassUpdated, isCSIStorage)
		c.updateCloneVolumeAnnotation(storageClassUpdated, isCSIStorage)
	} else {
		c.removeAnnotations(storageClassUpdated)
	}
	if !reflect.DeepEqual(storageClass, storageClassUpdated) {
		_, err = c.storageClassClient.Update(context.Background(), storageClassUpdated, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *StorageCapabilityController) hasCSIDriver(storageClass *storagev1.StorageClass) bool {
	driver := storageClass.Provisioner
	if driver != "" {
		if _, err := c.csiDriverLister.Get(driver); err != nil {
			return false
		}
		return true
	}
	return false
}

func (c *StorageCapabilityController) updateSnapshotAnnotation(storageClass *storagev1.StorageClass, snapshotAllow bool) {
	if storageClass.Annotations == nil {
		storageClass.Annotations = make(map[string]string)
	}
	if _, err := strconv.ParseBool(storageClass.Annotations[annotationAllowSnapshot]); err != nil {
		storageClass.Annotations[annotationAllowSnapshot] = strconv.FormatBool(snapshotAllow)
	}
}

func (c *StorageCapabilityController) updateCloneVolumeAnnotation(storageClass *storagev1.StorageClass, cloneAllow bool) {
	if storageClass.Annotations == nil {
		storageClass.Annotations = make(map[string]string)
	}
	if _, err := strconv.ParseBool(storageClass.Annotations[annotationAllowClone]); err != nil {
		storageClass.Annotations[annotationAllowClone] = strconv.FormatBool(cloneAllow)
	}
}

func (c *StorageCapabilityController) removeAnnotations(storageClass *storagev1.StorageClass) {
	delete(storageClass.Annotations, annotationAllowClone)
	delete(storageClass.Annotations, annotationAllowSnapshot)
}
