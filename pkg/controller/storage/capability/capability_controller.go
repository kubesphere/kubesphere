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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/discovery"
	"reflect"
	"strconv"
	"strings"
	"time"

	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned/typed/volumesnapshot/v1beta1"
	snapinformers "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions/volumesnapshot/v1beta1"
	snapshotlisters "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/listers/volumesnapshot/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	storageinformersv1 "k8s.io/client-go/informers/storage/v1"
	storageinformersv1beta1 "k8s.io/client-go/informers/storage/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	storageclient "k8s.io/client-go/kubernetes/typed/storage/v1"
	storagelistersv1 "k8s.io/client-go/listers/storage/v1"
	storagelistersv1beta1 "k8s.io/client-go/listers/storage/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	capability "kubesphere.io/kubesphere/pkg/apis/storage/v1alpha1"
	crdscheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	capabilityclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned/typed/storage/v1alpha1"
	capabilityinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/storage/v1alpha1"
	capabilitylisters "kubesphere.io/kubesphere/pkg/client/listers/storage/v1alpha1"
)

const (
	minSnapshotSupportedVersion = "v1.17.0"
	csiAddressFormat            = "/var/lib/kubelet/plugins/%s/csi.sock"
	annotationSupportSnapshot   = "storageclass.kubesphere.io/support-snapshot"
)

type csiAddressGetter func(storageClassProvisioner string) string

type StorageCapabilityController struct {
	storageClassCapabilityClient capabilityclient.StorageClassCapabilityInterface
	storageCapabilityLister      capabilitylisters.StorageClassCapabilityLister
	storageClassCapabilitySynced cache.InformerSynced

	provisionerCapabilityLister capabilitylisters.ProvisionerCapabilityLister
	provisionerCapabilitySynced cache.InformerSynced

	storageClassClient storageclient.StorageClassInterface
	storageClassLister storagelistersv1.StorageClassLister
	storageClassSynced cache.InformerSynced

	snapshotSupported   bool
	snapshotClassClient snapshotclient.VolumeSnapshotClassInterface
	snapshotClassLister snapshotlisters.VolumeSnapshotClassLister
	snapshotClassSynced cache.InformerSynced

	csiDriverLister storagelistersv1beta1.CSIDriverLister
	csiDriverSynced cache.InformerSynced

	csiAddressGetter csiAddressGetter

	workQueue workqueue.RateLimitingInterface
}

// This controller is responsible to watch StorageClass, SnapshotClass.
// And then update StorageClassCapability CRD resource object to the newest status.
func NewController(
	capabilityClient capabilityclient.StorageClassCapabilityInterface,
	capabilityInformer capabilityinformers.Interface,
	storageClassClient storageclient.StorageClassInterface,
	storageClassInformer storageinformersv1.StorageClassInformer,
	snapshotSupported bool,
	snapshotClassClient snapshotclient.VolumeSnapshotClassInterface,
	snapshotClassInformer snapinformers.VolumeSnapshotClassInformer,
	csiDriverInformer storageinformersv1beta1.CSIDriverInformer,
) *StorageCapabilityController {

	utilruntime.Must(crdscheme.AddToScheme(scheme.Scheme))

	controller := &StorageCapabilityController{
		storageClassCapabilityClient: capabilityClient,
		storageCapabilityLister:      capabilityInformer.StorageClassCapabilities().Lister(),
		storageClassCapabilitySynced: capabilityInformer.StorageClassCapabilities().Informer().HasSynced,
		provisionerCapabilityLister:  capabilityInformer.ProvisionerCapabilities().Lister(),
		provisionerCapabilitySynced:  capabilityInformer.ProvisionerCapabilities().Informer().HasSynced,
		storageClassClient:           storageClassClient,
		storageClassLister:           storageClassInformer.Lister(),
		storageClassSynced:           storageClassInformer.Informer().HasSynced,
		snapshotSupported:            snapshotSupported,
		csiDriverLister:              csiDriverInformer.Lister(),
		csiDriverSynced:              csiDriverInformer.Informer().HasSynced,
		csiAddressGetter:             csiAddress,
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

	csiDriverInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handlerCSIDriver,
		UpdateFunc: func(oldObj, newObj interface{}) {
			return
		},
		DeleteFunc: controller.handlerCSIDriver,
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
		c.csiDriverSynced,
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

func (c *StorageCapabilityController) handlerCSIDriver(obj interface{}) {
	csiDriver := obj.(*storagev1beta1.CSIDriver)
	storageClasses, err := c.storageClassLister.List(labels.Everything())
	if err != nil {
		klog.Error("list StorageClass error when handler csiDriver", err)
		return
	}
	for _, storageClass := range storageClasses {
		if storageClass.Provisioner == csiDriver.Name {
			klog.V(4).Infof("enqueue StorageClass %s when handling csiDriver", storageClass.Name)
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
	// No capability because csi-plugin not installed
	if capabilitySpec == nil {
		klog.Infof("StorageClass %s has no capability", name)
		err = c.updateStorageClassSnapshotSupported(storageClass, false)
		if err != nil {
			return err
		}
		return c.deleteStorageCapability(name)
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
				_, err = c.snapshotClassClient.Create(volumeSnapshotClassCreate)
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
	storageClassCapabilityExist, err := c.storageCapabilityLister.Get(storageClass.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// If StorageClassCapability doesn't exist, create it
			storageClassCapabilityCreate := &capability.StorageClassCapability{ObjectMeta: metav1.ObjectMeta{Name: storageClass.Name}}
			storageClassCapabilityCreate.Spec = *capabilitySpec
			klog.Info("Create StorageClassCapability: ", storageClassCapabilityCreate)
			_, err = c.storageClassCapabilityClient.Create(storageClassCapabilityCreate)
			return err
		}
		return err
	}
	// If StorageClassCapability exist, update it.
	storageClassCapabilityUpdate := storageClassCapabilityExist.DeepCopy()
	storageClassCapabilityUpdate.Spec = *capabilitySpec
	if !reflect.DeepEqual(storageClassCapabilityExist, storageClassCapabilityUpdate) {
		klog.Info("Update StorageClassCapability: ", storageClassCapabilityUpdate)
		_, err = c.storageClassCapabilityClient.Update(storageClassCapabilityUpdate)
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
		_, err = c.storageClassClient.Update(storageClass)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *StorageCapabilityController) deleteStorageCapability(name string) error {
	_, err := c.storageCapabilityLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	klog.Infof("Delete StorageClassCapability %s", name)
	return c.storageClassCapabilityClient.Delete(name, &metav1.DeleteOptions{})
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
	return c.snapshotClassClient.Delete(name, &metav1.DeleteOptions{})
}

func (c *StorageCapabilityController) capabilityFromProvisioner(provisioner string) (*capability.StorageClassCapabilitySpec, error) {
	provisionerCapability, err := c.provisionerCapabilityLister.Get(getProvisionerCapabilityName(provisioner))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	klog.V(4).Infof("get provisioner capability:%s %s", provisioner, provisionerCapability.Name)
	capabilitySpec := &capability.StorageClassCapabilitySpec{
		Features: provisionerCapability.Spec.Features,
	}
	return capabilitySpec, nil
}

func (c *StorageCapabilityController) getCapabilitySpec(storageClass *storagev1.StorageClass) (*capability.StorageClassCapabilitySpec, error) {
	// get from provisioner capability first
	klog.V(4).Info("get cap ", storageClass.Provisioner)
	capabilitySpec, err := c.capabilityFromProvisioner(storageClass.Provisioner)
	if err != nil {
		return nil, err
	}

	// csi of storage capability
	if capabilitySpec == nil {
		isCsi, err := c.isCSIStorage(storageClass.Provisioner)
		if err != nil {
			return nil, err
		}
		if isCsi {
			capabilitySpec, err = csiCapability(c.csiAddressGetter(storageClass.Provisioner))
			if err != nil {
				return nil, err
			}
		}
	}

	if capabilitySpec != nil {
		capabilitySpec.Provisioner = storageClass.Provisioner
		if storageClass.AllowVolumeExpansion == nil || !*storageClass.AllowVolumeExpansion {
			capabilitySpec.Features.Volume.Expand = capability.ExpandModeUnknown
		}
		if !c.snapshotSupported {
			capabilitySpec.Features.Snapshot.Create = false
			capabilitySpec.Features.Snapshot.List = false
		}
	}
	return capabilitySpec, nil
}

func (c *StorageCapabilityController) isCSIStorage(provisioner string) (bool, error) {
	_, err := c.csiDriverLister.Get(provisioner)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// this is used for test of CSIDriver on windows
func (c *StorageCapabilityController) setCSIAddressGetter(getter csiAddressGetter) {
	c.csiAddressGetter = getter
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

func csiAddress(provisioner string) string {
	return fmt.Sprintf(csiAddressFormat, provisioner)
}

func getProvisionerCapabilityName(provisioner string) string {
	return strings.NewReplacer(".", "-", "/", "-").Replace(provisioner)
}
