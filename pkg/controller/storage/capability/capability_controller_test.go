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
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	storagev1 "k8s.io/api/storage/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	k8sinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	ksfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
)

var (
	noReSyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t                 *testing.T
	snapshotSupported bool
	// Clients
	k8sClient *k8sfake.Clientset
	//nolint:unused
	ksClient *ksfake.Clientset
	// Objects from here preload into NewSimpleFake.
	storageObjects []runtime.Object // include StorageClass
	// Objects to put in the store.
	storageClassLister []*storagev1.StorageClass
	csiDriverLister    []*storagev1.CSIDriver
	// Actions expected to happen on the client.
	actions []core.Action
}

func newFixture(t *testing.T, snapshotSupported bool) *fixture {
	return &fixture{
		t:                 t,
		snapshotSupported: snapshotSupported,
	}
}

func (f *fixture) newController() (*StorageCapabilityController,
	k8sinformers.SharedInformerFactory) {

	f.k8sClient = k8sfake.NewSimpleClientset(f.storageObjects...)

	k8sInformers := k8sinformers.NewSharedInformerFactory(f.k8sClient, noReSyncPeriodFunc())
	c := NewController(
		f.k8sClient.StorageV1().StorageClasses(),
		k8sInformers.Storage().V1().StorageClasses(),
		k8sInformers.Storage().V1().CSIDrivers(),
	)

	for _, storageClass := range f.storageClassLister {
		_ = k8sInformers.Storage().V1().StorageClasses().Informer().GetIndexer().Add(storageClass)
	}
	for _, csiDriver := range f.csiDriverLister {
		_ = k8sInformers.Storage().V1().CSIDrivers().Informer().GetIndexer().Add(csiDriver)
	}

	return c, k8sInformers
}

func (f *fixture) runController(scName string, startInformers bool, expectError bool) {
	c, k8sI := f.newController()

	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		k8sI.Start(stopCh)
	}

	err := c.syncHandler(scName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing, got nil")
	}

	var actions []core.Action
	actions = append(actions, f.k8sClient.Actions()...)
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

func (f *fixture) expectUpdateStorageClassAction(storageClass *storagev1.StorageClass) {
	f.actions = append(f.actions, core.NewUpdateAction(
		schema.GroupVersionResource{Resource: "storageclasses"}, storageClass.Namespace, storageClass))
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

func newCSIDriver(name string) *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
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
	fixture := newFixture(t, true)
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClassUpdate := storageClass.DeepCopy()
	storageClassUpdate.Annotations = map[string]string{annotationAllowSnapshot: "true", annotationAllowClone: "true"}
	csiDriver := newCSIDriver("csi.example.com")

	// Objects exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)
	fixture.csiDriverLister = append(fixture.csiDriverLister, csiDriver)

	// Action expected
	fixture.expectUpdateStorageClassAction(storageClassUpdate)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestStorageClassHadAnnotation(t *testing.T) {
	fixture := newFixture(t, true)
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClass.Annotations = make(map[string]string)
	storageClassUpdate := storageClass.DeepCopy()
	storageClass.Annotations = map[string]string{annotationAllowSnapshot: "false", annotationAllowClone: "false"}

	// Object exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)

	// Action expected
	fixture.expectUpdateStorageClassAction(storageClassUpdate)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestStorageClassHadOneAnnotation(t *testing.T) {
	fixture := newFixture(t, true)
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClass.Annotations = map[string]string{annotationAllowSnapshot: "false"}
	storageClassUpdate := storageClass.DeepCopy()
	storageClassUpdate.Annotations[annotationAllowClone] = "true"
	csiDriver := newCSIDriver("csi.example.com")

	// Object exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)
	fixture.csiDriverLister = append(fixture.csiDriverLister, csiDriver)

	// Action expected
	fixture.expectUpdateStorageClassAction(storageClassUpdate)

	// Run test
	fixture.run(getKey(storageClass, t))
}

func TestStorageClassHadNoCSIDriver(t *testing.T) {
	fixture := newFixture(t, true)
	storageClass := newStorageClass("csi-example", "csi.example.com")
	storageClass.Annotations = map[string]string{}
	storageClassUpdate := storageClass.DeepCopy()
	storageClass.Annotations = map[string]string{annotationAllowSnapshot: "false"}
	storageClass.Annotations = map[string]string{annotationAllowClone: "false"}

	// Object exist
	fixture.storageObjects = append(fixture.storageObjects, storageClass)
	fixture.storageClassLister = append(fixture.storageClassLister, storageClass)

	// Action expected
	fixture.expectUpdateStorageClassAction(storageClassUpdate)

	// Run test
	fixture.run(getKey(storageClass, t))
}
