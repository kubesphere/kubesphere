/*
Copyright 2017 The Kubernetes Authors.

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

package storageobjectinuseprotection

import (
	"fmt"
	"io"

	"github.com/golang/glog"

	admission "k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/util/feature"
	api "k8s.io/kubernetes/pkg/apis/core"
	informers "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion"
	corelisters "k8s.io/kubernetes/pkg/client/listers/core/internalversion"
	"k8s.io/kubernetes/pkg/features"
	kubeapiserveradmission "k8s.io/kubernetes/pkg/kubeapiserver/admission"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
)

const (
	// PluginName is the name of this admission controller plugin
	PluginName = "StorageObjectInUseProtection"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		plugin := newPlugin()
		return plugin, nil
	})
}

// storageProtectionPlugin holds state for and implements the admission plugin.
type storageProtectionPlugin struct {
	*admission.Handler

	pvcLister corelisters.PersistentVolumeClaimLister
	pvLister  corelisters.PersistentVolumeLister
}

var _ admission.Interface = &storageProtectionPlugin{}
var _ = kubeapiserveradmission.WantsInternalKubeInformerFactory(&storageProtectionPlugin{})

// newPlugin creates a new admission plugin.
func newPlugin() *storageProtectionPlugin {
	return &storageProtectionPlugin{
		Handler: admission.NewHandler(admission.Create),
	}
}

func (c *storageProtectionPlugin) SetInternalKubeInformerFactory(f informers.SharedInformerFactory) {
	pvcInformer := f.Core().InternalVersion().PersistentVolumeClaims()
	c.pvcLister = pvcInformer.Lister()
	pvInformer := f.Core().InternalVersion().PersistentVolumes()
	c.pvLister = pvInformer.Lister()
	c.SetReadyFunc(func() bool {
		return pvcInformer.Informer().HasSynced() && pvInformer.Informer().HasSynced()
	})
}

// ValidateInitialization ensures lister is set.
func (c *storageProtectionPlugin) ValidateInitialization() error {
	if c.pvcLister == nil {
		return fmt.Errorf("missing PVC lister")
	}
	if c.pvLister == nil {
		return fmt.Errorf("missing PV lister")
	}
	return nil
}

var (
	pvResource  = api.Resource("persistentvolumes")
	pvcResource = api.Resource("persistentvolumeclaims")
)

// Admit sets finalizer on all PVCs(PVs). The finalizer is removed by
// PVCProtectionController(PVProtectionController) when it's not referenced.
//
// This prevents users from deleting a PVC that's used by a running pod.
// This also prevents admin from deleting a PV that's bound by a PVC
func (c *storageProtectionPlugin) Admit(a admission.Attributes) error {
	if !feature.DefaultFeatureGate.Enabled(features.StorageObjectInUseProtection) {
		return nil
	}

	switch a.GetResource().GroupResource() {
	case pvResource:
		return c.admitPV(a)
	case pvcResource:
		return c.admitPVC(a)

	default:
		return nil
	}
}

func (c *storageProtectionPlugin) admitPV(a admission.Attributes) error {
	if len(a.GetSubresource()) != 0 {
		return nil
	}

	pv, ok := a.GetObject().(*api.PersistentVolume)
	// if we can't convert the obj to PV, just return
	if !ok {
		return nil
	}
	for _, f := range pv.Finalizers {
		if f == volumeutil.PVProtectionFinalizer {
			// Finalizer is already present, nothing to do
			return nil
		}
	}
	glog.V(4).Infof("adding PV protection finalizer to %s", pv.Name)
	pv.Finalizers = append(pv.Finalizers, volumeutil.PVProtectionFinalizer)

	return nil
}

func (c *storageProtectionPlugin) admitPVC(a admission.Attributes) error {
	if len(a.GetSubresource()) != 0 {
		return nil
	}

	pvc, ok := a.GetObject().(*api.PersistentVolumeClaim)
	// if we can't convert the obj to PVC, just return
	if !ok {
		return nil
	}

	for _, f := range pvc.Finalizers {
		if f == volumeutil.PVCProtectionFinalizer {
			// Finalizer is already present, nothing to do
			return nil
		}
	}

	glog.V(4).Infof("adding PVC protection finalizer to %s/%s", pvc.Namespace, pvc.Name)
	pvc.Finalizers = append(pvc.Finalizers, volumeutil.PVCProtectionFinalizer)
	return nil
}
