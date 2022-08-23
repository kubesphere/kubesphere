/*

 Copyright 2021 The KubeSphere Authors.

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

package snapshotclass

import (
	"context"
	"fmt"
	"strconv"
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/typed/volumesnapshot/v1"
	snapinformers "github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions/volumesnapshot/v1"
	snapshotlisters "github.com/kubernetes-csi/external-snapshotter/client/v4/listers/volumesnapshot/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	storageinformersv1 "k8s.io/client-go/informers/storage/v1"
	storagelistersv1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const annotationAllowSnapshot = "storageclass.kubesphere.io/allow-snapshot"

type VolumeSnapshotClassController struct {
	storageClassLister  storagelistersv1.StorageClassLister
	storageClassSynced  cache.InformerSynced
	snapshotClassClient snapshotclient.VolumeSnapshotClassInterface
	snapshotClassLister snapshotlisters.VolumeSnapshotClassLister
	snapshotClassSynced cache.InformerSynced

	snapshotClassWorkQueue workqueue.RateLimitingInterface
}

// This controller is responsible to watch StorageClass
// When storageClass has created ,create snapshot class
func NewController(
	storageClassInformer storageinformersv1.StorageClassInformer,
	snapshotClassClient snapshotclient.VolumeSnapshotClassInterface,
	snapshotClassInformer snapinformers.VolumeSnapshotClassInformer,
) *VolumeSnapshotClassController {
	controller := &VolumeSnapshotClassController{
		storageClassLister:     storageClassInformer.Lister(),
		storageClassSynced:     storageClassInformer.Informer().HasSynced,
		snapshotClassClient:    snapshotClassClient,
		snapshotClassLister:    snapshotClassInformer.Lister(),
		snapshotClassSynced:    snapshotClassInformer.Informer().HasSynced,
		snapshotClassWorkQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SnapshotClass"),
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
		DeleteFunc: controller.enqueueStorageClass,
	})

	return controller
}

func (c *VolumeSnapshotClassController) Start(ctx context.Context) error {
	return c.Run(5, ctx.Done())
}

func (c *VolumeSnapshotClassController) Run(threadCnt int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.snapshotClassWorkQueue.ShutDown()

	klog.Info("Waiting for informer cache to sync.")
	cacheSyncs := []cache.InformerSynced{
		c.storageClassSynced,
		c.snapshotClassSynced,
	}

	if ok := cache.WaitForCacheSync(stopCh, cacheSyncs...); !ok {
		return fmt.Errorf("failed to wait for caches to syne")
	}

	for i := 0; i < threadCnt; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
	return nil
}

func (c *VolumeSnapshotClassController) enqueueStorageClass(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.snapshotClassWorkQueue.Add(key)
}

func (c *VolumeSnapshotClassController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *VolumeSnapshotClassController) processNextWorkItem() bool {
	obj, shutdown := c.snapshotClassWorkQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.snapshotClassWorkQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.snapshotClassWorkQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workQueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.snapshotClassWorkQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.snapshotClassWorkQueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *VolumeSnapshotClassController) syncHandler(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	storageClass, err := c.storageClassLister.Get(name)
	if err != nil {
		// StorageClass has been deleted, delete VolumeSnapshotClass
		if errors.IsNotFound(err) {
			err = c.deleteSnapshotClass(name)
		}
		return err
	}

	if storageClass.Annotations != nil {
		if annotationSnap, ok := storageClass.Annotations[annotationAllowSnapshot]; ok {
			allowSnapshot, err := strconv.ParseBool(annotationSnap)
			if err == nil && allowSnapshot {
				// If VolumeSnapshotClass not exist, create it
				_, err = c.snapshotClassLister.Get(name)
				if err != nil {
					if errors.IsNotFound(err) {
						volumeSnapshotClassCreate := &snapshotv1.VolumeSnapshotClass{
							ObjectMeta:     metav1.ObjectMeta{Name: name},
							Driver:         storageClass.Provisioner,
							DeletionPolicy: snapshotv1.VolumeSnapshotContentDelete,
						}
						_, err = c.snapshotClassClient.Create(context.Background(), volumeSnapshotClassCreate, metav1.CreateOptions{})
					}
				}
			}
			return err
		}
	}
	return nil
}

func (c *VolumeSnapshotClassController) deleteSnapshotClass(name string) error {
	_, err := c.snapshotClassLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	klog.Infof("Delete SnapshotClass %s", name)
	return c.snapshotClassClient.Delete(context.Background(), name, metav1.DeleteOptions{})
}
