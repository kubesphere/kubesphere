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
	"strconv"
	"time"

	snapinformers "github.com/kubernetes-csi/external-snapshotter/client/v3/informers/externalversions/volumesnapshot/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/discovery"
	storageinformersv1beta1 "k8s.io/client-go/informers/storage/v1beta1"
	storagelistersv1beta1 "k8s.io/client-go/listers/storage/v1beta1"

	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/client/v3/apis/volumesnapshot/v1beta1"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v3/clientset/versioned/typed/volumesnapshot/v1beta1"
	snapshotlisters "github.com/kubernetes-csi/external-snapshotter/client/v3/listers/volumesnapshot/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	storageinformersv1 "k8s.io/client-go/informers/storage/v1"
	"k8s.io/client-go/kubernetes/scheme"
	storageclient "k8s.io/client-go/kubernetes/typed/storage/v1"
	storagelistersv1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	crdscheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
)

const (
	minSnapshotSupportedVersion = "v1.17.0"
	annotationSupportSnapshot   = "storageclass.kubesphere.io/allow-snapshot"
	annotationSupportClone      = "storageclass.kubesphere.io/allow-clone"
)

type StorageCapabilityController struct {
	storageClassClient storageclient.StorageClassInterface
	storageClassLister storagelistersv1.StorageClassLister
	storageClassSynced cache.InformerSynced

	csiDriverLister storagelistersv1beta1.CSIDriverLister
	csiDriverSynced cache.InformerSynced

	snapshotSupported   bool
	snapshotClassClient snapshotclient.VolumeSnapshotClassInterface
	snapshotClassLister snapshotlisters.VolumeSnapshotClassLister
	snapshotClassSynced cache.InformerSynced

	workQueue    workqueue.RateLimitingInterface
	csiWorkQueue workqueue.RateLimitingInterface
}

// This controller is responsible to watch StorageClass/ProvisionerCapability.
// And then update StorageClassCapability CRD resource object to the newest status.
func NewController(
	storageClassClient storageclient.StorageClassInterface,
	storageClassInformer storageinformersv1.StorageClassInformer,
	csiDriverInformer storageinformersv1beta1.CSIDriverInformer,
	snapshotSupported bool,
	snapshotClassClient snapshotclient.VolumeSnapshotClassInterface,
	snapshotClassInformer snapinformers.VolumeSnapshotClassInformer,
) *StorageCapabilityController {

	utilruntime.Must(crdscheme.AddToScheme(scheme.Scheme))

	controller := &StorageCapabilityController{
		storageClassClient: storageClassClient,
		storageClassLister: storageClassInformer.Lister(),
		storageClassSynced: storageClassInformer.Informer().HasSynced,
		csiDriverLister:    csiDriverInformer.Lister(),
		csiDriverSynced:    csiDriverInformer.Informer().HasSynced,
		snapshotSupported:  snapshotSupported,
		workQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "StorageClasses"),
		csiWorkQueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "csiDriver"),
	}

	if snapshotSupported {
		controller.snapshotClassClient = snapshotClassClient
		controller.snapshotClassLister = snapshotClassInformer.Lister()
		controller.snapshotClassSynced = snapshotClassInformer.Informer().HasSynced
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

	csiDriverInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.enqueueStorageClassByCSI,
		UpdateFunc: nil,
		DeleteFunc: controller.enqueueStorageClassByCSI,
	})

	return controller
}

func (c *StorageCapabilityController) Start(ctx context.Context) error {
	return c.Run(5, ctx.Done())
}

func (c *StorageCapabilityController) Run(threadCnt int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

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
	c.workQueue.Add(key)
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
			c.workQueue.Add(obj.Name)
		}
	}
	return
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
	if err != nil {
		// StorageClass has been deleted, delete StorageClassCapability and VolumeSnapshotClass
		if errors.IsNotFound(err) && c.snapshotAllowed() {
			err = c.deleteSnapshotClass(name)
			if err != nil {
				return err
			}
		}
		return err
	}

	//Cloning and volumeSnapshot support only available for CSI drivers.
	withCapability := c.supportCapability(storageClass)
	// Handle VolumeSnapshotClass with same name of StorageClass
	// annotate "support-snapshot" of StorageClass
	if c.snapshotAllowed() && withCapability {
		_, err = c.snapshotClassLister.Get(name)
		if err != nil {
			// If VolumeSnapshotClass not exist, create it
			if errors.IsNotFound(err) {
				volumeSnapshotClassCreate := &snapshotv1beta1.VolumeSnapshotClass{
					ObjectMeta:     metav1.ObjectMeta{Name: name},
					Driver:         storageClass.Provisioner,
					DeletionPolicy: snapshotv1beta1.VolumeSnapshotContentDelete,
				}
				_, err = c.snapshotClassClient.Create(context.Background(), volumeSnapshotClassCreate, metav1.CreateOptions{})
				if err != nil {
					return err
				}
			}
		}
	}

	err = c.addStorageClassSnapshotAnnotation(storageClass, withCapability)
	if err != nil {
		return err
	}
	err = c.addCloneVolumeAnnotation(storageClass, withCapability)
	if err != nil {
		return nil
	}
	_, err = c.storageClassClient.Update(context.Background(), storageClass, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *StorageCapabilityController) supportCapability(storageClass *storagev1.StorageClass) bool {
	driver := storageClass.Provisioner
	if driver != "" {
		if _, err := c.csiDriverLister.Get(driver); err != nil {
			return false
		}
		return true
	}
	return false
}

func (c *StorageCapabilityController) addStorageClassSnapshotAnnotation(storageClass *storagev1.StorageClass, snapshotSupported bool) error {
	if snapshotSupported || !c.snapshotSupported {
		if storageClass.Annotations == nil {
			storageClass.Annotations = make(map[string]string)
		}
		_, err := strconv.ParseBool(storageClass.Annotations[annotationSupportSnapshot])
		// err != nil means annotationSupportSnapshot is not illegal, include empty
		if err != nil {
			storageClass.Annotations[annotationSupportSnapshot] = strconv.FormatBool(c.snapshotSupported)
		}
	} else {
		if storageClass.Annotations != nil && c.snapshotSupported {
			if _, ok := storageClass.Annotations[annotationSupportSnapshot]; ok {
				delete(storageClass.Annotations, annotationSupportSnapshot)
			}
		}
	}
	return nil
}

func (c *StorageCapabilityController) addCloneVolumeAnnotation(storageClass *storagev1.StorageClass, cloneSupported bool) error {
	if cloneSupported {
		if storageClass.Annotations == nil {
			storageClass.Annotations = make(map[string]string)
		}
		_, err := strconv.ParseBool(storageClass.Annotations[annotationSupportClone])
		if err != nil {
			storageClass.Annotations[annotationSupportClone] = strconv.FormatBool(cloneSupported)
		}
	} else {
		if storageClass.Annotations != nil {
			if _, ok := storageClass.Annotations[annotationSupportClone]; ok {
				delete(storageClass.Annotations, annotationSupportClone)
			}
		}
	}
	return nil
}

func (c *StorageCapabilityController) deleteSnapshotClass(name string) error {
	if !c.snapshotAllowed() {
		return nil
	}
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

func (c *StorageCapabilityController) snapshotAllowed() bool {
	return c.snapshotSupported && c.snapshotClassClient != nil && c.snapshotClassLister != nil && c.snapshotClassSynced != nil
}

func SnapshotSupported(discoveryInterface discovery.DiscoveryInterface) bool {
	minVer := version.MustParseGeneric(minSnapshotSupportedVersion)
	rawVer, err := discoveryInterface.ServerVersion()
	if err != nil {
		return false
	}
	ver, err := version.ParseSemantic(rawVer.String())
	if err != nil {
		return false
	}
	return ver.AtLeast(minVer)
}
