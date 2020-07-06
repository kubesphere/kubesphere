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
	"github.com/google/go-cmp/cmp"
	"math/rand"

	//"github.com/google/go-cmp/cmp"
	snapbeta1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	snapfake "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned/fake"
	snapinformers "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	storagev1 "k8s.io/api/storage/v1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	k8sinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	ksv1alpha1 "kubesphere.io/kubesphere/pkg/apis/storage/v1alpha1"
	ksfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"reflect"
	"testing"
	"time"
)

var (
	noReSyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t                 *testing.T
	snapshotSupported bool
	// Clients
	k8sClient           *k8sfake.Clientset
	snapshotClassClient *snapfake.Clientset
	ksClient            *ksfake.Clientset
	// Objects from here preload into NewSimpleFake.
	storageObjects       []runtime.Object // include StorageClass and CSIDriver
	snapshotClassObjects []runtime.Object
	capabilityObjects    []runtime.Object //  include StorageClassCapability and ProvisionerCapability
	// Objects to put in the store.
	storageClassLister           []*storagev1.StorageClass
	snapshotClassLister          []*snapbeta1.VolumeSnapshotClass
	storageClassCapabilityLister []*ksv1alpha1.StorageClassCapability
	provisionerCapabilityLister  []*ksv1alpha1.ProvisionerCapability
	csiDriverLister              []*storagev1beta1.CSIDriver
	// Actions expected to happen on the client.
	actions []core.Action
	// CSI server
	runCSIServer  bool
	fakeCSIServer *fakeCSIServer
}

func newFixture(t *testing.T, snapshotSupported bool, runCSIServer bool) *fixture {
	return &fixture{
		t:                 t,
		snapshotSupported: snapshotSupported,
		runCSIServer:      runCSIServer,
	}
}

func (f *fixture) newController() (*StorageCapabilityController,
	k8sinformers.SharedInformerFactory,
	ksinformers.SharedInformerFactory,
	snapinformers.SharedInformerFactory) {

	f.k8sClient = k8sfake.NewSimpleClientset(f.storageObjects...)
	f.ksClient = ksfake.NewSimpleClientset(f.capabilityObjects...)
	f.snapshotClassClient = snapfake.NewSimpleClientset(f.snapshotClassObjects...)

	k8sInformers := k8sinformers.NewSharedInformerFactory(f.k8sClient, noReSyncPeriodFunc())
	ksInformers := ksinformers.NewSharedInformerFactory(f.ksClient, noReSyncPeriodFunc())
	snapshotInformers := snapinformers.NewSharedInformerFactory(f.snapshotClassClient, noReSyncPeriodFunc())

	c := NewController(
		f.ksClient.StorageV1alpha1().StorageClassCapabilities(),
		ksInformers.Storage().V1alpha1(),
		f.k8sClient.StorageV1().StorageClasses(),
		k8sInformers.Storage().V1().StorageClasses(),
		f.snapshotSupported,
		f.snapshotClassClient.SnapshotV1beta1().VolumeSnapshotClasses(),
		snapshotInformers.Snapshot().V1beta1().VolumeSnapshotClasses(),
		k8sInformers.Storage().V1beta1().CSIDrivers(),
	)
	if f.runCSIServer {
		port := 30000 + rand.Intn(100)
		fakeCSIServer, address := newTestCSIServer(port)
		f.fakeCSIServer = fakeCSIServer
		c.setCSIAddressGetter(func(storageClassProvisioner string) string { return address })
	}

	for _, storageClass := range f.storageClassLister {
		_ = k8sInformers.Storage().V1().StorageClasses().Informer().GetIndexer().Add(storageClass)
	}
	for _, csiDriver := range f.csiDriverLister {
		_ = k8sInformers.Storage().V1beta1().CSIDrivers().Informer().GetIndexer().Add(csiDriver)
	}
	for _, snapshotClass := range f.snapshotClassLister {
		_ = snapshotInformers.Snapshot().V1beta1().VolumeSnapshotClasses().Informer().GetIndexer().Add(snapshotClass)
	}
	for _, storageClassCapability := range f.storageClassCapabilityLister {
		_ = ksInformers.Storage().V1alpha1().StorageClassCapabilities().Informer().GetIndexer().Add(storageClassCapability)
	}
	for _, provisionerCapability := range f.provisionerCapabilityLister {
		_ = ksInformers.Storage().V1alpha1().ProvisionerCapabilities().Informer().GetIndexer().Add(provisionerCapability)
	}

	return c, k8sInformers, ksInformers, snapshotInformers
}

func (f *fixture) runController(scName string, startInformers bool, expectError bool) {
	c, k8sI, crdI, snapI := f.newController()

	if f.runCSIServer {
		f.fakeCSIServer.run()
		defer f.fakeCSIServer.stop()
	}

	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		k8sI.Start(stopCh)
		crdI.Start(stopCh)
		snapI.Start(stopCh)
	}

	err := c.syncHandler(scName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing, got nil")
	}

	var actions []core.Action
	actions = append(actions, f.snapshotClassClient.Actions()...)
	actions = append(actions, f.k8sClient.Actions()...)
	actions = append(actions, f.ksClient.Actions()...)
	filerActions := filterInformerActions(actions)
	if len(filerActions) != len(f.actions) {
		f.t.Errorf("count of actions: differ (-got, +want): %s", cmp.Diff(filerActions, f.actions))
		return
	}
	for i, action := range filerActions {
		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}
}

func (f *fixture) run(scName string) {
	f.runController(scName, true, false)
}

func (f *fixture) expectCreateStorageClassCapabilitiesAction(storageClassCapability *ksv1alpha1.StorageClassCapability) {
	f.actions = append(f.actions, core.NewCreateAction(
		schema.GroupVersionResource{Resource: "storageclasscapabilities"}, storageClassCapability.Namespace, storageClassCapability))
}

func (f *fixture) expectUpdateStorageClassCapabilitiesAction(storageClassCapability *ksv1alpha1.StorageClassCapability) {
	f.actions = append(f.actions, core.NewUpdateAction(
		schema.GroupVersionResource{Resource: "storageclasscapabilities"}, storageClassCapability.Namespace, storageClassCapability))
}

func (f *fixture) expectDeleteStorageClassCapabilitiesAction(storageClassCapability *ksv1alpha1.StorageClassCapability) {
	f.actions = append(f.actions, core.NewDeleteAction(
		schema.GroupVersionResource{Resource: "storageclasscapabilities"}, storageClassCapability.Namespace, storageClassCapability.Name))
}

func (f *fixture) expectUpdateStorageClassAction(storageClass *storagev1.StorageClass) {
	f.actions = append(f.actions, core.NewUpdateAction(
		schema.GroupVersionResource{Resource: "storageclasses"}, storageClass.Namespace, storageClass))
}

func (f *fixture) expectCreateSnapshotClassAction(snapshotClass *snapbeta1.VolumeSnapshotClass) {
	f.actions = append(f.actions, core.NewCreateAction(
		schema.GroupVersionResource{Resource: "volumesnapshotclasses"}, snapshotClass.Namespace, snapshotClass))
}

func (f *fixture) expectDeleteSnapshotClassAction(snapshotClass *snapbeta1.VolumeSnapshotClass) {
	f.actions = append(f.actions, core.NewDeleteAction(
		schema.GroupVersionResource{Resource: "volumesnapshotclasses"}, snapshotClass.Namespace, snapshotClass.Name))
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	var ret []core.Action
	for _, action := range actions {
		if action.GetVerb() == "list" || action.GetVerb() == "watch" {
			continue
		}
		ret = append(ret, action)
	}
	return ret
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("\nExpected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()
		if difference := cmp.Diff(object, expObject); len(difference) > 0 {
			t.Errorf("[CreateAction] %T differ (-got, +want): %s", expObject, difference)
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()
		if difference := cmp.Diff(object, expObject); len(difference) > 0 {
			t.Errorf("[UpdateAction] %T differ (-got, +want): %s", expObject, difference)
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()
		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
		}
	case core.DeleteActionImpl:
		e, _ := expected.(core.DeleteActionImpl)
		if difference := cmp.Diff(e.Name, a.Name); len(difference) > 0 {
			t.Errorf("[UpdateAction] %T differ (-got, +want): %s", e.Name, difference)
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

func newStorageClass(name string, provisioner string) *storagev1.StorageClass {
	isExpansion := true
	return &storagev1.StorageClass{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Provisioner:          provisioner,
		AllowVolumeExpansion: &isExpansion,
	}
}

func newStorageClassCapability(storageClass *storagev1.StorageClass) *ksv1alpha1.StorageClassCapability {
	storageClassCapability := &ksv1alpha1.StorageClassCapability{}
	storageClassCapability.Name = storageClass.Name
	storageClassCapability.Spec = *newStorageClassCapabilitySpec()
	storageClassCapability.Spec.Provisioner = storageClass.Provisioner
	return storageClassCapability
}

func newProvisionerCapability(storageClass *storagev1.StorageClass) *ksv1alpha1.ProvisionerCapability {
	provisionerCapability := &ksv1alpha1.ProvisionerCapability{}
	provisionerCapability.Name = getProvisionerCapabilityName(storageClass.Provisioner)
	provisionerCapability.Spec.PluginInfo.Name = storageClass.Provisioner
	provisionerCapability.Spec.Features = newStorageClassCapabilitySpec().Features
	// ProvisionerCapability snapshot is always false
	provisionerCapability.Spec.Features.Snapshot.Create = false
	return provisionerCapability
}

func newCSIDriver(storageClass *storagev1.StorageClass) *storagev1beta1.CSIDriver {
	csiDriver := &storagev1beta1.CSIDriver{}
	csiDriver.Name = storageClass.Provisioner
	return csiDriver
}

func newSnapshotClass(storageClass *storagev1.StorageClass) *snapbeta1.VolumeSnapshotClass {
	return &snapbeta1.VolumeSnapshotClass{
		ObjectMeta: v1.ObjectMeta{
			Name: storageClass.Name,
		},
		Driver:         storageClass.Provisioner,
		DeletionPolicy: snapbeta1.VolumeSnapshotContentDelete,
	}
}

func getKey(sc *storagev1.StorageClass, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sc)
	if err != nil {
		t.Errorf("Unexpected error getting key for %v: %v", sc.Name, err)
		return ""
	}
	return key
}

func TestCreateStorageClass(t *testing.T) {
	fixture := newFixture(t, true, true)
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClassUpdate := storageClass.DeepCopy()
	storageClassUpdate.Annotations = map[string]string{annotationSupportSnapshot: "true"}
	snapshotClass := newSnapshotClass(storageClass)
	storageClassCapability := newStorageClassCapability(storageClass)
	csiDriver := newCSIDriver(storageClass)

	// Objects exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass, csiDriver)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)
	fixture.csiDriverLister = append(fixture.csiDriverLister, csiDriver)

	// Action expected
	fixture.expectCreateSnapshotClassAction(snapshotClass)
	fixture.expectUpdateStorageClassAction(storageClassUpdate)
	fixture.expectCreateStorageClassCapabilitiesAction(storageClassCapability)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestUpdateStorageClass(t *testing.T) {
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClass.Annotations = map[string]string{annotationSupportSnapshot: "true"}
	snapshotClass := newSnapshotClass(storageClass)
	storageClassCapabilityUpdate := newStorageClassCapability(storageClass)
	storageClassCapability := newStorageClassCapability(storageClass)
	//old and new should have deference
	storageClassCapability.Spec.Features.Volume.Create = !storageClassCapability.Spec.Features.Volume.Create
	csiDriver := newCSIDriver(storageClass)

	fixture := newFixture(t, true, true)
	// Object exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass, csiDriver)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)
	fixture.csiDriverLister = append(fixture.csiDriverLister, csiDriver)
	fixture.snapshotClassObjects = append(fixture.snapshotClassObjects, snapshotClass)
	fixture.snapshotClassLister = append(fixture.snapshotClassLister, snapshotClass)
	fixture.capabilityObjects = append(fixture.capabilityObjects, storageClassCapability)
	fixture.storageClassCapabilityLister = append(fixture.storageClassCapabilityLister, storageClassCapability)

	// Action expected
	fixture.expectUpdateStorageClassCapabilitiesAction(storageClassCapabilityUpdate)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestDeleteStorageClass(t *testing.T) {
	storageClass := newStorageClass("csi-example", "csi.example.com")
	snapshotClass := newSnapshotClass(storageClass)
	storageClassCapability := newStorageClassCapability(storageClass)

	csiDriver := newCSIDriver(storageClass)

	fixture := newFixture(t, true, true)
	// Object exist
	fixture.storageObjects = append(fixture.storageObjects, csiDriver)
	fixture.csiDriverLister = append(fixture.csiDriverLister, csiDriver)
	fixture.snapshotClassObjects = append(fixture.snapshotClassObjects, snapshotClass)
	fixture.snapshotClassLister = append(fixture.snapshotClassLister, snapshotClass)
	fixture.capabilityObjects = append(fixture.capabilityObjects, storageClassCapability)
	fixture.storageClassCapabilityLister = append(fixture.storageClassCapabilityLister, storageClassCapability)

	// Action expected
	fixture.expectDeleteSnapshotClassAction(snapshotClass)
	fixture.expectDeleteStorageClassCapabilitiesAction(storageClassCapability)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestCreateStorageClassNotSupportSnapshot(t *testing.T) {
	fixture := newFixture(t, false, true)
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClassUpdate := storageClass.DeepCopy()
	storageClassUpdate.Annotations = map[string]string{annotationSupportSnapshot: "false"}
	storageClassCapability := newStorageClassCapability(storageClass)
	storageClassCapability.Spec.Features.Snapshot.Create = false
	storageClassCapability.Spec.Features.Snapshot.List = false
	provisionerCapability := newProvisionerCapability(storageClass)
	csiDriver := newCSIDriver(storageClass)

	// Objects exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass, csiDriver)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)
	fixture.csiDriverLister = append(fixture.csiDriverLister, csiDriver)
	fixture.capabilityObjects = append(fixture.capabilityObjects, provisionerCapability)
	fixture.provisionerCapabilityLister = append(fixture.provisionerCapabilityLister, provisionerCapability)

	// Action expected
	fixture.expectUpdateStorageClassAction(storageClassUpdate)
	fixture.expectCreateStorageClassCapabilitiesAction(storageClassCapability)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestCreateStorageClassInTree(t *testing.T) {
	// InTree Storage has no snapshot capability
	fixture := newFixture(t, true, true)
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClassUpdate := storageClass.DeepCopy()
	storageClassUpdate.Annotations = map[string]string{annotationSupportSnapshot: "false"}
	storageClassCapability := newStorageClassCapability(storageClass)
	storageClassCapability.Spec.Features.Snapshot.Create = false
	provisionerCapability := newProvisionerCapability(storageClass)

	// Objects exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)
	fixture.capabilityObjects = append(fixture.capabilityObjects, provisionerCapability)
	fixture.provisionerCapabilityLister = append(fixture.provisionerCapabilityLister, provisionerCapability)

	// Action expected
	fixture.expectUpdateStorageClassAction(storageClassUpdate)
	fixture.expectCreateStorageClassCapabilitiesAction(storageClassCapability)

	// Run test
	fixture.run(getKey(storageClass, t))
}
