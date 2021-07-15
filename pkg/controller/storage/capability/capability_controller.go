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
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/discovery"

	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/client/v3/apis/volumesnapshot/v1beta1"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v3/clientset/versioned/typed/volumesnapshot/v1beta1"
	snapinformers "github.com/kubernetes-csi/external-snapshotter/client/v3/informers/externalversions/volumesnapshot/v1beta1"
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

	ksstorage "kubesphere.io/api/storage/v1alpha1"

	crdscheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	ksstorageclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/storage/v1alpha1"
	ksstorageinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/storage/v1alpha1"
	ksstoragelisters "kubesphere.io/kubesphere/pkg/client/listers/storage/v1alpha1"
)

const (
	minSnapshotSupportedVersion = "v1.17.0"
	annotationSupportSnapshot   = "storageclass.kubesphere.io/support-snapshot"
)

type StorageCapabilityController struct {
	storageClassCapabilityClient ksstorageclient.StorageClassCapabilityInterface
	storageClassCapabilityLister ksstoragelisters.StorageClassCapabilityLister
	storageClassCapabilitySynced cache.InformerSynced

	provisionerCapabilityLister ksstoragelisters.ProvisionerCapabilityLister
	provisionerCapabilitySynced cache.InformerSynced

	storageClassClient storageclient.StorageClassInterface
	storageClassLister storagelistersv1.StorageClassLister
	storageClassSynced cache.InformerSynced

	snapshotSupported   bool
	snapshotClassClient snapshotclient.VolumeSnapshotClassInterface
	snapshotClassLister snapshotlisters.VolumeSnapshotClassLister
	snapshotClassSynced cache.InformerSynced

	workQueue workqueue.RateLimitingInterface
}

// This controller is responsible to watch StorageClass/ProvisionerCapability.
// And then update StorageClassCapability CRD resource object to the newest status.
func NewController(
	storageClassCapabilityClient ksstorageclient.StorageClassCapabilityInterface,
	ksStorageInformer ksstorageinformers.Interface,
	storageClassClient storageclient.StorageClassInterface,
	storageClassInformer storageinformersv1.StorageClassInformer,
	snapshotSupported bool,
	snapshotClassClient snapshotclient.VolumeSnapshotClassInterface,
	snapshotClassInformer snapinformers.VolumeSnapshotClassInformer,
) *StorageCapabilityController {

	utilruntime.Must(crdscheme.AddToScheme(scheme.Scheme))

	controller := &StorageCapabilityController{
		storageClassCapabilityClient: storageClassCapabilityClient,
		storageClassCapabilityLister: ksStorageInformer.StorageClassCapabilities().Lister(),
		storageClassCapabilitySynced: ksStorageInformer.StorageClassCapabilities().Informer().HasSynced,
		provisionerCapabilityLister:  ksStorageInformer.ProvisionerCapabilities().Lister(),
		provisionerCapabilitySynced:  ksStorageInformer.ProvisionerCapabilities().Informer().HasSynced,
		storageClassClient:           storageClassClient,
		storageClassLister:           storageClassInformer.Lister(),
		storageClassSynced:           storageClassInformer.Informer().HasSynced,
		snapshotSupported:            snapshotSupported,
		workQueue:                    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "StorageClasses"),
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

	// ProvisionerCapability acts as a value source of its relevant StorageClassCapabilities
	// so when a PC is created/updated, the corresponding SCCs should be created(if not exists)/updated
	// we achieve this by simply enqueueing the StorageClasses of the same provisioner
	// but don't overdo by cascade deleting the SCCs when a PC is deleted
	// since the role of PCs is more like a template rather than owner to SCCs

	// This is a backward compatible fix to remove the useless auto detection of SCCs
	// in the future, we will only keep ProvisionerCapability and remove the StorageClassCapability CRD entirely
	ksStorageInformer.ProvisionerCapabilities().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleProvisionerCapability,
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPC := newObj.(*ksstorage.ProvisionerCapability)
			oldPC := oldObj.(*ksstorage.ProvisionerCapability)
			if newPC.ResourceVersion == oldPC.ResourceVersion {
				return
			}
			controller.handleProvisionerCapability(newObj)
		},
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
	cacheSyncs := []cache.InformerSynced{
		c.storageClassCapabilitySynced,
		c.provisionerCapabilitySynced,
		c.storageClassSynced,
	}

	if c.snapshotAllowed() {
		cacheSyncs = append(cacheSyncs, c.snapshotClassSynced)
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

func (c *StorageCapabilityController) handleProvisionerCapability(obj interface{}) {
	provisionerCapability := obj.(*ksstorage.ProvisionerCapability)
	storageClasses, err := c.storageClassLister.List(labels.Everything())
	if err != nil {
		klog.Error("list StorageClass error when handle provisionerCapability", err)
		return
	}
	for _, storageClass := range storageClasses {
		if getProvisionerCapabilityName(storageClass.Provisioner) == provisionerCapability.Name {
			klog.V(4).Infof("enqueue StorageClass %s while handling provisionerCapability", storageClass.Name)
			c.enqueueStorageClass(storageClass)
		}
	}
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
		if errors.IsNotFound(err) {
			if c.snapshotAllowed() {
				err = c.deleteSnapshotClass(name)
				if err != nil {
					return err
				}
			}
			return c.deleteStorageCapability(name)
		}
		return err
	}

	// Get capability spec
	capabilitySpec, err := c.getCapabilitySpec(storageClass)
	if err != nil {
		return err
	}
	// The corresponding ProvisionerCapability Object does not exist
	if capabilitySpec == nil {
		klog.Infof("Can't get StorageClass %s's capability", name)
		err = c.updateStorageClassSnapshotSupported(storageClass, false)
		if err != nil {
			return err
		}
		// Don't delete the already created SCC
		// as it might be created manually by user
		return nil
	}
	klog.Infof("StorageClass %s has capability %v", name, capabilitySpec)

	// Handle VolumeSnapshotClass with same name of StorageClass
	// annotate "support-snapshot" of StorageClass
	withSnapshotCapability := false
	if c.snapshotAllowed() && capabilitySpec.Features.Snapshot.Create {
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
		withSnapshotCapability = true
	}
	err = c.updateStorageClassSnapshotSupported(storageClass, withSnapshotCapability)
	if err != nil {
		return err
	}

	// Handle StorageClassCapability with the same name of StorageClass
	storageClassCapabilityExist, err := c.storageClassCapabilityLister.Get(storageClass.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// If StorageClassCapability doesn't exist, create it
			storageClassCapabilityCreate := &ksstorage.StorageClassCapability{ObjectMeta: metav1.ObjectMeta{Name: storageClass.Name}}
			storageClassCapabilityCreate.Spec = *capabilitySpec
			klog.Info("Create StorageClassCapability: ", storageClassCapabilityCreate)
			_, err = c.storageClassCapabilityClient.Create(context.Background(), storageClassCapabilityCreate, metav1.CreateOptions{})
			return err
		}
		return err
	}
	// If StorageClassCapability exist, update it.
	storageClassCapabilityUpdate := storageClassCapabilityExist.DeepCopy()
	storageClassCapabilityUpdate.Spec = *capabilitySpec
	if !reflect.DeepEqual(storageClassCapabilityExist, storageClassCapabilityUpdate) {
		klog.Info("Update StorageClassCapability: ", storageClassCapabilityUpdate)
		_, err = c.storageClassCapabilityClient.Update(context.Background(), storageClassCapabilityUpdate, metav1.UpdateOptions{})
		return err
	}
	return nil
}

func (c *StorageCapabilityController) updateStorageClassSnapshotSupported(storageClass *storagev1.StorageClass, snapshotSupported bool) error {
	if storageClass.Annotations == nil {
		storageClass.Annotations = make(map[string]string)
	}
	snapshotSupportedAnnotated, err := strconv.ParseBool(storageClass.Annotations[annotationSupportSnapshot])
	// err != nil means annotationSupportSnapshot is not illegal, include empty
	if err != nil || snapshotSupported != snapshotSupportedAnnotated {
		storageClass.Annotations[annotationSupportSnapshot] = strconv.FormatBool(snapshotSupported)
		_, err = c.storageClassClient.Update(context.Background(), storageClass, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *StorageCapabilityController) deleteStorageCapability(name string) error {
	_, err := c.storageClassCapabilityLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	klog.Infof("Delete StorageClassCapability %s", name)
	return c.storageClassCapabilityClient.Delete(context.Background(), name, metav1.DeleteOptions{})
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

func (c *StorageCapabilityController) capabilityFromProvisioner(provisioner string) (*ksstorage.StorageClassCapabilitySpec, error) {
	provisionerCapability, err := c.provisionerCapabilityLister.Get(getProvisionerCapabilityName(provisioner))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	klog.V(4).Infof("get provisioner capability:%s %s", provisioner, provisionerCapability.Name)
	capabilitySpec := &ksstorage.StorageClassCapabilitySpec{
		Features: provisionerCapability.Spec.Features,
	}
	return capabilitySpec, nil
}

func (c *StorageCapabilityController) getCapabilitySpec(storageClass *storagev1.StorageClass) (*ksstorage.StorageClassCapabilitySpec, error) {
	// get from provisioner capability first
	klog.V(4).Info("get cap ", storageClass.Provisioner)
	capabilitySpec, err := c.capabilityFromProvisioner(storageClass.Provisioner)
	if err != nil {
		return nil, err
	}

	if capabilitySpec != nil {
		capabilitySpec.Provisioner = storageClass.Provisioner
		if storageClass.AllowVolumeExpansion == nil || !*storageClass.AllowVolumeExpansion {
			capabilitySpec.Features.Volume.Expand = ksstorage.ExpandModeUnknown
		}
		if !c.snapshotSupported {
			capabilitySpec.Features.Snapshot.Create = false
			capabilitySpec.Features.Snapshot.List = false
		}
	}
	return capabilitySpec, nil
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

func getProvisionerCapabilityName(provisioner string) string {
	return strings.NewReplacer(".", "-", "/", "-").Replace(provisioner)
}
