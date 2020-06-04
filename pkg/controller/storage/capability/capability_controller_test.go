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
	snapbeta1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	snapfake "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned/fake"
	snapinformers "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	crdv1alpha1 "kubesphere.io/kubesphere/pkg/apis/storage/v1alpha1"
	crdfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	crdinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"reflect"
	"testing"
	"time"
)

var (
	alwaysReady        = func() bool { return true }
	noReSyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T
	// Clients
	k8sClient                      *k8sfake.Clientset
	snapshotClassClient            *snapfake.Clientset
	storageClassCapabilitiesClient *crdfake.Clientset
	// Objects from here preload into NewSimpleFake.
	storageClassObjects           []runtime.Object
	snapshotClassObjects          []runtime.Object
	storageClassCapabilityObjects []runtime.Object
	// Objects to put in the store.
	storageClassLister           []*storagev1.StorageClass
	snapshotClassLister          []*snapbeta1.VolumeSnapshotClass
	storageClassCapabilityLister []*crdv1alpha1.StorageClassCapability
	// Actions expected to happen on the client.
	storageClassCapabilitiesActions []core.Action
	// CSI server
	fakeCSIServer *fakeCSIServer
}

func newFixture(t *testing.T) *fixture {
	return &fixture{
		t: t,
	}
}

func (f *fixture) newController() (*StorageCapabilityController, kubeinformers.SharedInformerFactory,
	crdinformers.SharedInformerFactory, snapinformers.SharedInformerFactory) {

	fakeCSIServer, address := newTestCSIServer()
	f.fakeCSIServer = fakeCSIServer

	f.k8sClient = k8sfake.NewSimpleClientset(f.storageClassObjects...)
	f.storageClassCapabilitiesClient = crdfake.NewSimpleClientset(f.storageClassCapabilityObjects...)
	f.snapshotClassClient = snapfake.NewSimpleClientset(f.snapshotClassObjects...)

	k8sI := kubeinformers.NewSharedInformerFactory(f.k8sClient, noReSyncPeriodFunc())
	crdI := crdinformers.NewSharedInformerFactory(f.storageClassCapabilitiesClient, noReSyncPeriodFunc())
	snapI := snapinformers.NewSharedInformerFactory(f.snapshotClassClient, noReSyncPeriodFunc())

	c := NewController(
		f.k8sClient,
		f.storageClassCapabilitiesClient,
		k8sI.Storage().V1().StorageClasses(),
		snapI.Snapshot().V1beta1().VolumeSnapshotClasses(),
		crdI.Storage().V1alpha1().StorageClassCapabilities(),
		func(storageClassProvisioner string) string { return address },
	)

	for _, storageClass := range f.storageClassLister {
		_ = k8sI.Storage().V1().StorageClasses().Informer().GetIndexer().Add(storageClass)
	}
	for _, snapshotClass := range f.snapshotClassLister {
		_ = snapI.Snapshot().V1beta1().VolumeSnapshotClasses().Informer().GetIndexer().Add(snapshotClass)
	}
	for _, storageClassCapability := range f.storageClassCapabilityLister {
		_ = crdI.Storage().V1alpha1().StorageClassCapabilities().Informer().GetIndexer().Add(storageClassCapability)
	}

	return c, k8sI, crdI, snapI
}

func (f *fixture) runController(scName string, startInformers bool, expectError bool) {
	c, k8sI, crdI, snapI := f.newController()

	f.fakeCSIServer.run()
	defer f.fakeCSIServer.stop()

	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		k8sI.Start(stopCh)
		crdI.Start(stopCh)
		snapI.Start(stopCh)
	}
	c.storageClassSynced = alwaysReady
	c.snapshotClassSynced = alwaysReady
	c.storageClassCapabilitySynced = alwaysReady

	err := c.syncHandler(scName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing, got nil")
	}

	actions := filterInformerActions(f.storageClassCapabilitiesClient.Actions())
	for i, action := range actions {
		if len(f.storageClassCapabilitiesActions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.storageClassCapabilitiesActions), actions[i:])
			break
		}
		expectedAction := f.storageClassCapabilitiesActions[i]
		checkAction(expectedAction, action, f.t)
	}
}

func (f *fixture) run(scName string) {
	f.runController(scName, true, false)
}

func (f *fixture) expectCreateStorageClassCapabilitiesAction(storageClassCapability *crdv1alpha1.StorageClassCapability) {
	f.storageClassCapabilitiesActions = append(f.storageClassCapabilitiesActions, core.NewCreateAction(
		schema.GroupVersionResource{Resource: "storageclasscapabilities"}, storageClassCapability.Namespace, storageClassCapability))
}

func (f *fixture) expectUpdateStorageClassCapabilitiesAction(storageClassCapability *crdv1alpha1.StorageClassCapability) {
	f.storageClassCapabilitiesActions = append(f.storageClassCapabilitiesActions, core.NewUpdateAction(
		schema.GroupVersionResource{Resource: "storageclasscapabilities"}, storageClassCapability.Namespace, storageClassCapability))
}

func (f *fixture) expectDeleteStorageClassCapabilitiesAction(storageClassCapability *crdv1alpha1.StorageClassCapability) {
	f.storageClassCapabilitiesActions = append(f.storageClassCapabilitiesActions, core.NewDeleteAction(
		schema.GroupVersionResource{Resource: "storageclasscapabilities"}, storageClassCapability.Namespace, storageClassCapability.Name))
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
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
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

func newStorageClassCapability(storageClass *storagev1.StorageClass) *crdv1alpha1.StorageClassCapability {
	storageClassCapability := &crdv1alpha1.StorageClassCapability{}
	storageClassCapability.Name = storageClass.Name
	storageClassCapability.Spec = *newStorageClassCapabilitySpec()
	storageClassCapability.Spec.Provisioner = storageClass.Provisioner
	return storageClassCapability
}

func newSnapshotClass(storageClass *storagev1.StorageClass) *snapbeta1.VolumeSnapshotClass {
	return &snapbeta1.VolumeSnapshotClass{
		ObjectMeta: v1.ObjectMeta{
			Name: storageClass.Name,
		},
		Driver: storageClass.Provisioner,
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
	fixture := newFixture(t)
	storageClass := newStorageClass("csi-example", "csi.example.com")
	snapshotClass := newSnapshotClass(storageClass)
	storageClassCapability := newStorageClassCapability(storageClass)

	// Objects exist
	fixture.storageClassObjects = append(fixture.storageClassObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)
	fixture.snapshotClassObjects = append(fixture.snapshotClassObjects, snapshotClass)
	fixture.snapshotClassLister = append(fixture.snapshotClassLister, snapshotClass)

	// Action expected
	fixture.expectCreateStorageClassCapabilitiesAction(storageClassCapability)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestUpdateStorageClass(t *testing.T) {
	storageClass := newStorageClass("csi-example", "csi.example.com")
	snapshotClass := newSnapshotClass(storageClass)
	storageClassCapability := newStorageClassCapability(storageClass)

	fixture := newFixture(t)
	// Object exist
	fixture.storageClassObjects = append(fixture.storageClassObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)
	fixture.snapshotClassObjects = append(fixture.snapshotClassObjects, snapshotClass)
	fixture.snapshotClassLister = append(fixture.snapshotClassLister, snapshotClass)
	fixture.storageClassCapabilityObjects = append(fixture.storageClassCapabilityObjects, storageClassCapability)
	fixture.storageClassCapabilityLister = append(fixture.storageClassCapabilityLister, storageClassCapability)

	// Action expected
	fixture.expectUpdateStorageClassCapabilitiesAction(storageClassCapability)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestDeleteStorageClass(t *testing.T) {
	storageClass := newStorageClass("csi-example", "csi.example.com")
	snapshotClass := newSnapshotClass(storageClass)
	storageClassCapability := newStorageClassCapability(storageClass)

	fixture := newFixture(t)
	// Object exist
	fixture.snapshotClassObjects = append(fixture.snapshotClassObjects, snapshotClass)
	fixture.snapshotClassLister = append(fixture.snapshotClassLister, snapshotClass)
	fixture.storageClassCapabilityObjects = append(fixture.storageClassCapabilityObjects, storageClassCapability)
	fixture.storageClassCapabilityLister = append(fixture.storageClassCapabilityLister, storageClassCapability)

	// Action expected
	fixture.expectDeleteStorageClassCapabilitiesAction(storageClassCapability)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestDeleteSnapshotClass(t *testing.T) {
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClassCapability := newStorageClassCapability(storageClass)

	fixture := newFixture(t)
	// Object exist
	fixture.storageClassCapabilityObjects = append(fixture.storageClassCapabilityObjects, storageClassCapability)
	fixture.storageClassCapabilityLister = append(fixture.storageClassCapabilityLister, storageClassCapability)
	fixture.storageClassObjects = append(fixture.storageClassObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)

	// Action expected
	storageClassCapabilityUpdate := storageClassCapability.DeepCopy()
	storageClassCapabilityUpdate.Spec.Features.Snapshot.Create = false
	storageClassCapabilityUpdate.Spec.Features.Snapshot.List = false
	fixture.expectUpdateStorageClassCapabilitiesAction(storageClassCapabilityUpdate)

	// Run test
	fixture.run(getKey(storageClass, t))
}
